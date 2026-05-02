package vm

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/dec64"
	"slug/internal/object"
	"strconv"
	"strings"
)

type Executor struct {
	env          *object.Environment
	stack        []object.Object
	externalCall func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object
}

func NewExecutor(
	env *object.Environment,
	externalCall func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object,
) *Executor {
	return &Executor{
		env:          env,
		externalCall: externalCall,
	}
}

func (e *Executor) EvalProgram(program *ast.Program) object.Object {
	chunk, err := Compile(program)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	return e.run(chunk)
}

func (e *Executor) run(chunk *Chunk) object.Object {
	e.stack = e.stack[:0]
	var last object.Object = object.NIL

	for ip := 0; ip < len(chunk.Instructions); ip++ {
		ins := chunk.Instructions[ip]
		switch ins.Op {
		case OpConstant:
			e.push(e.bindClosureIfNeeded(chunk.Constants[ins.IntArg]))
		case OpNil:
			e.push(object.NIL)
		case OpTrue:
			e.push(object.TRUE)
		case OpFalse:
			e.push(object.FALSE)
		case OpGetGlobal:
			val, ok := e.env.Get(ins.StrArg)
			if !ok {
				return e.errorAt(ins.Position, "identifier not found: %s", ins.StrArg)
			}
			val, errObj := e.resolveValue(ins.Position, val)
			if errObj != nil {
				return errObj
			}
			e.push(val)
		case OpSetGlobalConst:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for const bind")
			}
			if _, err := e.env.DefineConstant(ins.StrArg, val, false, false); err != nil {
				return e.errorAt(ins.Position, "%s", err.Error())
			}
			e.push(val)
		case OpSetGlobalVar:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for var bind")
			}
			if _, err := e.env.Define(ins.StrArg, val, false, false); err != nil {
				return e.errorAt(ins.Position, "%s", err.Error())
			}
			e.push(val)
		case OpBindMapAllConst, OpBindMapAllVar:
			mObj, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for map select-all binding")
			}
			m, ok := mObj.(*object.Map)
			if !ok {
				return e.errorAt(ins.Position, "map select-all expects map value, got %s", mObj.Type())
			}
			for _, pair := range m.Pairs {
				name, ok := mapBindingName(pair.Key)
				if !ok || name == "" {
					continue
				}
				var err error
				if ins.Op == OpBindMapAllConst {
					_, err = e.env.DefineConstant(name, pair.Value, false, false)
				} else {
					_, err = e.env.Define(name, pair.Value, false, false)
				}
				if err != nil {
					return e.errorAt(ins.Position, "%s", err.Error())
				}
			}
			e.push(mObj)
		case OpAssignGlobal:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for assignment")
			}
			if _, err := e.env.Assign(ins.StrArg, val); err != nil {
				return e.errorAt(ins.Position, "%s", err.Error())
			}
			e.push(val)
		case OpArray:
			if ins.IntArg < 0 {
				return e.errorAt(ins.Position, "invalid array arity")
			}
			elements := make([]object.Object, ins.IntArg)
			for i := ins.IntArg - 1; i >= 0; i-- {
				val, ok := e.pop()
				if !ok {
					return e.errorAt(ins.Position, "stack underflow for array literal")
				}
				elements[i] = val
			}
			e.push(&object.List{Elements: elements})
		case OpHash:
			if ins.IntArg < 0 {
				return e.errorAt(ins.Position, "invalid map arity")
			}
			pairs := make(map[object.MapKey]object.MapPair, ins.IntArg)
			for i := 0; i < ins.IntArg; i++ {
				value, ok := e.pop()
				if !ok {
					return e.errorAt(ins.Position, "stack underflow for map value")
				}
				keyObj, ok := e.pop()
				if !ok {
					return e.errorAt(ins.Position, "stack underflow for map key")
				}
				hashable, ok := keyObj.(object.Hashable)
				if !ok {
					return e.errorAt(ins.Position, "unusable as map key: %s", keyObj.Type())
				}
				mapKey := hashable.MapKey()
				pairs[mapKey] = object.MapPair{Key: keyObj, Value: value}
			}
			e.push(&object.Map{Pairs: pairs})
		case OpSlice:
			step, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for slice step")
			}
			end, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for slice end")
			}
			start, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for slice start")
			}
			e.push(&object.Slice{
				Start: nilIfNilObject(start),
				End:   nilIfNilObject(end),
				Step:  nilIfNilObject(step),
			})
		case OpIndex, OpIndexDot:
			index, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for index operand")
			}
			left, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for indexed value")
			}
			value, errObj := e.evalIndex(left, index, ins.Position, ins.Op == OpIndexDot)
			if errObj != nil {
				return errObj
			}
			value, errObj = e.resolveValue(ins.Position, value)
			if errObj != nil {
				return errObj
			}
			e.push(value)
		case OpDup:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for dup")
			}
			e.push(val)
			e.push(val)
		case OpSpawn:
			callee, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for spawn")
			}
			fn, ok := callee.(*VMFunction)
			if !ok {
				return e.errorAt(ins.Position, "spawn expects a VM function body, got %s", callee.Type())
			}
			handle := NewVMTaskHandle()
			closure := fn.Closure
			if closure == nil {
				closure = e.env
			}
			taskEnv := closure.ShallowCopy()

			go func() {
				child := NewExecutor(taskEnv, e.externalCall)
				result := child.run(fn.Chunk)
				if result == nil {
					result = object.NIL
				}
				if spawnedFn, ok := result.(*VMFunction); ok && len(spawnedFn.Params) == 0 {
					result = child.invokeVMFunction(spawnedFn, nil, nil, ins.Position)
				}
				handle.Complete(result)
			}()

			e.push(handle)
		case OpAwait:
			taskObj, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for await")
			}
			handle, ok := taskObj.(*VMTaskHandle)
			if !ok {
				return e.errorAt(ins.Position, "await expects a task handle, got %s", taskObj.Type())
			}
			<-handle.done
			result := handle.result
			if result == nil {
				result = object.NIL
			}
			if result.Type() == object.ERROR_OBJ {
				return result
			}
			e.push(result)
		case OpListPrepend:
			right, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for list prepend")
			}
			left, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for list prepend")
			}
			rList, ok := right.(*object.List)
			if !ok {
				return e.errorAt(ins.Position, "type mismatch: %s +: %s", left.Type(), right.Type())
			}
			newElements := make([]object.Object, 0, len(rList.Elements)+1)
			newElements = append(newElements, left)
			newElements = append(newElements, rList.Elements...)
			e.push(&object.List{Elements: newElements})
		case OpListAppend:
			right, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for list append")
			}
			left, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for list append")
			}
			lList, ok := left.(*object.List)
			if !ok {
				return e.errorAt(ins.Position, "type mismatch: %s :+ %s", left.Type(), right.Type())
			}
			newElements := make([]object.Object, 0, len(lList.Elements)+1)
			newElements = append(newElements, lList.Elements...)
			newElements = append(newElements, right)
			e.push(&object.List{Elements: newElements})
		case OpMatchListEmpty:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for list match")
			}
			lst, ok := val.(*object.List)
			e.push(e.nativeBool(ok && len(lst.Elements) == 0))
		case OpMatchListHeadTail:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for list match")
			}
			lst, ok := val.(*object.List)
			if !ok || len(lst.Elements) == 0 {
				e.push(object.FALSE)
				continue
			}
			head := lst.Elements[0]
			tail := &object.List{Elements: append([]object.Object(nil), lst.Elements[1:]...)}
			if _, err := e.env.DefineConstant(ins.StrArg, head, false, false); err != nil {
				return e.errorAt(ins.Position, "%s", err.Error())
			}
			if _, err := e.env.DefineConstant(ins.StrArg2, tail, false, false); err != nil {
				return e.errorAt(ins.Position, "%s", err.Error())
			}
			e.push(object.TRUE)
		case OpAdd, OpSub, OpMul, OpDiv, OpEqual, OpNotEqual, OpGreaterThan, OpLessThan:
			if errObj := e.evalBinary(ins.Op, ins.Position); errObj != nil {
				return errObj
			}
		case OpBang:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for ! operator")
			}
			switch val {
			case object.TRUE:
				e.push(object.FALSE)
			case object.FALSE, object.NIL:
				e.push(object.TRUE)
			default:
				e.push(object.FALSE)
			}
		case OpNegate:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for unary - operator")
			}
			num, ok := val.(*object.Number)
			if !ok {
				return e.errorAt(ins.Position, "unknown operator: -%s", val.Type())
			}
			e.push(&object.Number{Value: num.Value.Neg()})
		case OpPop:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for pop")
			}
			last = val
		case OpJump:
			ip = ins.IntArg - 1
		case OpJumpIfFalse:
			val, ok := e.pop()
			if !ok {
				return e.errorAt(ins.Position, "stack underflow for conditional jump")
			}
			if !isTruthy(val) {
				ip = ins.IntArg - 1
			}
		case OpCall:
			if errObj := e.evalCall(ins.IntArg, ins.CallPlan, ins.Position); errObj != nil {
				return errObj
			}
		case OpReturn:
			if val, ok := e.pop(); ok {
				return val
			}
			return last
		default:
			return e.errorAt(ins.Position, "unsupported opcode: %d", ins.Op)
		}
	}
	if val, ok := e.pop(); ok {
		return val
	}
	return last
}

func mapBindingName(key object.Object) (string, bool) {
	switch k := key.(type) {
	case *object.Symbol:
		return k.Name, true
	case *object.String:
		return k.Value, true
	default:
		return "", false
	}
}

func (e *Executor) evalBinary(op Opcode, pos int) object.Object {
	right, ok := e.pop()
	if !ok {
		return e.errorAt(pos, "stack underflow for binary operator")
	}
	left, ok := e.pop()
	if !ok {
		return e.errorAt(pos, "stack underflow for binary operator")
	}

	switch op {
	case OpAdd:
		if l, ok := left.(*object.Number); ok {
			if r, ok := right.(*object.Number); ok {
				e.push(&object.Number{Value: l.Value.Add(r.Value)})
				return nil
			}
		}
		if l, ok := left.(*object.String); ok {
			if r, ok := right.(*object.String); ok {
				e.push(&object.String{Value: l.Value + r.Value})
				return nil
			}
		}
		if l, ok := left.(*object.String); ok {
			e.push(&object.String{Value: l.Value + right.Inspect()})
			return nil
		}
		if r, ok := right.(*object.String); ok {
			e.push(&object.String{Value: left.Inspect() + r.Value})
			return nil
		}
		if l, ok := left.(*object.List); ok {
			if r, ok := right.(*object.List); ok {
				combined := make([]object.Object, 0, len(l.Elements)+len(r.Elements))
				combined = append(combined, l.Elements...)
				combined = append(combined, r.Elements...)
				e.push(&object.List{Elements: combined})
				return nil
			}
		}
		return e.errorAt(pos, "type mismatch: %s + %s", left.Type(), right.Type())
	case OpSub:
		l, lok := left.(*object.Number)
		r, rok := right.(*object.Number)
		if !lok || !rok {
			return e.errorAt(pos, "type mismatch: %s - %s", left.Type(), right.Type())
		}
		e.push(&object.Number{Value: l.Value.Sub(r.Value)})
	case OpMul:
		l, lok := left.(*object.Number)
		r, rok := right.(*object.Number)
		if lok && rok {
			e.push(&object.Number{Value: l.Value.Mul(r.Value)})
			return nil
		}
		if s, ok := left.(*object.String); ok {
			if n, ok := right.(*object.Number); ok {
				times := n.Value.ToInt64()
				if times < 0 {
					return e.errorAt(pos, "repetition count must be a non-negative NUMBER, got %s", strconv.FormatInt(times, 10))
				}
				e.push(&object.String{Value: strings.Repeat(s.Value, int(times))})
				return nil
			}
		}
		return e.errorAt(pos, "type mismatch: %s * %s", left.Type(), right.Type())
	case OpDiv:
		l, lok := left.(*object.Number)
		r, rok := right.(*object.Number)
		if !lok || !rok {
			return e.errorAt(pos, "type mismatch: %s / %s", left.Type(), right.Type())
		}
		e.push(&object.Number{Value: l.Value.Div(r.Value, 14, dec64.RoundHalfEven)})
	case OpEqual:
		e.push(e.nativeBool(objectsEqual(left, right)))
	case OpNotEqual:
		e.push(e.nativeBool(!objectsEqual(left, right)))
	case OpGreaterThan, OpLessThan:
		l, lok := left.(*object.Number)
		r, rok := right.(*object.Number)
		if !lok || !rok {
			opName := ">"
			if op == OpLessThan {
				opName = "<"
			}
			return e.errorAt(pos, "type mismatch: %s %s %s", left.Type(), opName, right.Type())
		}
		if op == OpGreaterThan {
			e.push(e.nativeBool(l.Value.Gt(r.Value)))
		} else {
			e.push(e.nativeBool(l.Value.Lt(r.Value)))
		}
	}
	return nil
}

func (e *Executor) push(obj object.Object) {
	e.stack = append(e.stack, obj)
}

func (e *Executor) pop() (object.Object, bool) {
	if len(e.stack) == 0 {
		return nil, false
	}
	obj := e.stack[len(e.stack)-1]
	e.stack = e.stack[:len(e.stack)-1]
	return obj, true
}

func (e *Executor) nativeBool(v bool) *object.Boolean {
	if v {
		return object.TRUE
	}
	return object.FALSE
}

func (e *Executor) bindClosureIfNeeded(obj object.Object) object.Object {
	fn, ok := obj.(*VMFunction)
	if !ok {
		return obj
	}
	// Capture current lexical environment when function literal is evaluated.
	return &VMFunction{
		Name:    fn.Name,
		Params:  append([]VMParam(nil), fn.Params...),
		Chunk:   fn.Chunk,
		Closure: e.env,
	}
}

func (e *Executor) evalCall(argCount int, plan []CallArgSpec, pos int) object.Object {
	if argCount < 0 {
		return e.errorAt(pos, "invalid call arity")
	}
	if len(plan) != argCount {
		return e.errorAt(pos, "invalid call metadata: expected %d args in plan, got %d", argCount, len(plan))
	}
	args := make([]object.Object, argCount)
	for i := argCount - 1; i >= 0; i-- {
		arg, ok := e.pop()
		if !ok {
			return e.errorAt(pos, "stack underflow for call arguments")
		}
		args[i] = arg
	}
	callee, ok := e.pop()
	if !ok {
		return e.errorAt(pos, "stack underflow for callee")
	}

	positional := make([]object.Object, 0, argCount)
	var named map[string]object.Object
	for i, spec := range plan {
		val := args[i]
		switch spec.Kind {
		case CallArgPositional:
			positional = append(positional, val)
		case CallArgSpread:
			list, ok := val.(*object.List)
			if !ok {
				return e.errorAt(pos, "spread operator can only be used on lists, got %s", val.Type())
			}
			positional = append(positional, list.Elements...)
		case CallArgNamed:
			if named == nil {
				named = make(map[string]object.Object)
			}
			if _, exists := named[spec.Name]; exists {
				return e.errorAt(pos, "duplicate named argument: %s", spec.Name)
			}
			named[spec.Name] = val
		default:
			return e.errorAt(pos, "invalid call argument kind: %d", spec.Kind)
		}
	}

	fn, ok := callee.(*VMFunction)
	if !ok {
		if e.externalCall == nil {
			return e.errorAt(pos, "attempted to call non-vm function value (%s)", callee.Type())
		}
		result := e.externalCall(pos, callee, positional, named)
		if result == nil {
			result = object.NIL
		}
		if result.Type() == object.ERROR_OBJ {
			return result
		}
		e.push(result)
		return nil
	}

	result := e.invokeVMFunction(fn, positional, named, pos)
	if result.Type() == object.ERROR_OBJ {
		return result
	}
	e.push(result)
	return nil
}

func (e *Executor) invokeVMFunction(fn *VMFunction, positional []object.Object, named map[string]object.Object, pos int) object.Object {
	bound, errObj := bindVMArguments(fn, positional, named, pos, e)
	if errObj != nil {
		return errObj
	}

	closure := fn.Closure
	if closure == nil {
		closure = e.env
	}
	callEnv := object.NewEnclosedEnvironment(closure, nil)
	for i, p := range fn.Params {
		if _, err := callEnv.DefineConstant(p.Name, bound[i], false, false); err != nil {
			return e.errorAt(pos, "%s", err.Error())
		}
	}

	child := NewExecutor(callEnv, e.externalCall)
	result := child.run(fn.Chunk)
	if result == nil {
		return object.NIL
	}
	return result
}

func bindVMArguments(
	fn *VMFunction,
	positional []object.Object,
	named map[string]object.Object,
	pos int,
	e *Executor,
) ([]object.Object, *object.Error) {
	paramCount := len(fn.Params)
	values := make([]object.Object, paramCount)
	provided := make([]bool, paramCount)
	hasVariadic := paramCount > 0 && fn.Params[paramCount-1].IsVariadic
	variadicIndex := paramCount - 1

	if len(named) > 0 {
		index := make(map[string]int, paramCount)
		for i, p := range fn.Params {
			index[p.Name] = i
		}
		for name, val := range named {
			idx, ok := index[name]
			if !ok {
				return nil, e.errorAt(pos, "unknown named parameter: %s", name)
			}
			if provided[idx] {
				return nil, e.errorAt(pos, "duplicate assignment to parameter: %s", name)
			}
			if fn.Params[idx].IsVariadic {
				if _, ok := val.(*object.List); !ok {
					return nil, e.errorAt(pos, "variadic parameter '%s' must be a list when passed by name", name)
				}
			}
			values[idx] = val
			provided[idx] = true
		}
	}

	posIndex := 0
	if hasVariadic {
		for i := 0; i < variadicIndex; i++ {
			if posIndex >= len(positional) {
				break
			}
			if provided[i] {
				continue
			}
			values[i] = positional[posIndex]
			provided[i] = true
			posIndex++
		}
		remaining := positional[posIndex:]
		if provided[variadicIndex] {
			if len(remaining) > 0 {
				return nil, e.errorAt(pos, "too many positional arguments")
			}
		} else {
			values[variadicIndex] = &object.List{Elements: remaining}
			provided[variadicIndex] = true
		}
	} else {
		for i := 0; i < paramCount; i++ {
			if posIndex >= len(positional) {
				break
			}
			if provided[i] {
				continue
			}
			values[i] = positional[posIndex]
			provided[i] = true
			posIndex++
		}
		if posIndex < len(positional) {
			return nil, e.errorAt(pos, "too many positional arguments")
		}
	}

	defEnv := fn.Closure
	if defEnv == nil {
		defEnv = e.env
	}
	if defEnv != nil {
		for defEnv.Outer != nil {
			defEnv = defEnv.Outer
		}
	}

	for i := 0; i < paramCount; i++ {
		if provided[i] {
			continue
		}
		param := fn.Params[i]
		if param.IsVariadic {
			values[i] = &object.List{Elements: []object.Object{}}
			provided[i] = true
			continue
		}
		if param.Default != nil {
			defaultExec := NewExecutor(defEnv, e.externalCall)
			defaultVal := defaultExec.run(param.Default)
			if defaultVal == nil {
				defaultVal = object.NIL
			}
			if defaultVal.Type() == object.ERROR_OBJ {
				if errObj, ok := defaultVal.(*object.Error); ok {
					return nil, errObj
				}
				return nil, e.errorAt(pos, "%s", defaultVal.Inspect())
			}
			values[i] = defaultVal
			provided[i] = true
			continue
		}
		return nil, e.errorAt(pos, "missing required parameter: %s", param.Name)
	}
	return values, nil
}

func (e *Executor) errorAt(pos int, format string, args ...interface{}) *object.Error {
	msg := fmt.Sprintf(format, args...)
	if pos > 0 {
		return &object.Error{Message: fmt.Sprintf("vm runtime error at pos %d: %s", pos, msg)}
	}
	return &object.Error{Message: "vm runtime error: " + msg}
}

func isTruthy(obj object.Object) bool {
	return obj != object.FALSE && obj != object.NIL
}

func nilIfNilObject(v object.Object) object.Object {
	if v == object.NIL {
		return nil
	}
	return v
}

func computeSliceIndices(length int, slice *object.Slice) (int, int, int) {
	start := 0
	end := length
	step := 1

	if slice.Start != nil {
		start = int(slice.Start.(*object.Number).Value.ToInt64())
	}
	if slice.End != nil {
		end = int(slice.End.(*object.Number).Value.ToInt64())
	}
	if slice.Step != nil {
		step = int(slice.Step.(*object.Number).Value.ToInt64())
	}

	if start < 0 {
		start += length
	}
	if end < 0 {
		end += length
	}
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if step <= 0 {
		step = 1
	}

	return start, end, step
}

func (e *Executor) resolveValue(pos int, obj object.Object) (object.Object, *object.Error) {
	for {
		ref, ok := obj.(*object.BindingRef)
		if !ok {
			return obj, nil
		}
		if ref.Env == nil {
			return nil, e.errorAt(pos, "invalid binding reference: %s", ref.Inspect())
		}
		val, _, ok := ref.Env.GetLocalBindingValue(ref.Name)
		if !ok {
			return nil, e.errorAt(pos, "binding reference not found: %s", ref.Name)
		}
		if val == object.BINDING_UNINITIALIZED {
			return nil, e.errorAt(pos, "%s used before initialization (likely circular import)", ref.Name)
		}
		obj = val
	}
}

func objectsEqual(a, b object.Object) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type() != b.Type() {
		return false
	}

	switch av := a.(type) {
	case *object.Number:
		return av.Value.Eq(b.(*object.Number).Value)
	case *object.Boolean:
		return av.Value == b.(*object.Boolean).Value
	case *object.String:
		return av.Value == b.(*object.String).Value
	case *object.Symbol:
		return av == b.(*object.Symbol)
	case *object.Nil:
		return true
	case *object.List:
		bv := b.(*object.List)
		if len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !objectsEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	case *object.Bytes:
		bv := b.(*object.Bytes)
		if len(av.Value) != len(bv.Value) {
			return false
		}
		for i := range av.Value {
			if av.Value[i] != bv.Value[i] {
				return false
			}
		}
		return true
	case *object.Map:
		bv := b.(*object.Map)
		if len(av.Pairs) != len(bv.Pairs) {
			return false
		}
		for k, pairA := range av.Pairs {
			pairB, ok := bv.Pairs[k]
			if !ok || !objectsEqual(pairA.Value, pairB.Value) {
				return false
			}
		}
		return true
	default:
		// Fallback for object kinds not yet fully modeled by VM.
		return a == b
	}
}

func (e *Executor) evalIndex(left, index object.Object, pos int, isDotLookup bool) (object.Object, *object.Error) {
	switch l := left.(type) {
	case *object.List:
		if slice, ok := index.(*object.Slice); ok {
			start, end, step := computeSliceIndices(len(l.Elements), slice)
			result := make([]object.Object, 0)
			for i := start; i < end; i += step {
				result = append(result, l.Elements[i])
			}
			return &object.List{Elements: result}, nil
		}
		num, ok := index.(*object.Number)
		if !ok {
			return nil, e.errorAt(pos, "index operator not supported: %s", index.Type())
		}
		idx := num.Value.ToInt64()
		max := int64(len(l.Elements) - 1)
		if idx < 0 {
			idx = max + idx + 1
		}
		if idx < 0 || idx > max {
			return object.NIL, nil
		}
		return l.Elements[idx], nil
	case *object.String:
		if slice, ok := index.(*object.Slice); ok {
			runes := []rune(l.Value)
			start, end, step := computeSliceIndices(len(runes), slice)
			var b strings.Builder
			for i := start; i < end; i += step {
				b.WriteRune(runes[i])
			}
			return &object.String{Value: b.String()}, nil
		}
		num, ok := index.(*object.Number)
		if !ok {
			return nil, e.errorAt(pos, "index operator not supported: %s", index.Type())
		}
		idx := num.Value.ToInt64()
		runes := []rune(l.Value)
		max := int64(len(runes) - 1)
		if idx < 0 {
			idx = max + idx + 1
		}
		if idx < 0 || idx > max {
			return object.NIL, nil
		}
		return &object.String{Value: string(runes[idx])}, nil
	case *object.Bytes:
		if slice, ok := index.(*object.Slice); ok {
			start, end, step := computeSliceIndices(len(l.Value), slice)
			result := make([]byte, 0)
			for i := start; i < end; i += step {
				result = append(result, l.Value[i])
			}
			return &object.Bytes{Value: result}, nil
		}
		num, ok := index.(*object.Number)
		if !ok {
			return nil, e.errorAt(pos, "index operator not supported: %s", index.Type())
		}
		idx := num.Value.ToInt64()
		max := int64(len(l.Value) - 1)
		if idx < 0 {
			idx = max + idx + 1
		}
		if idx < 0 || idx > max {
			return object.NIL, nil
		}
		return &object.Number{Value: dec64.FromInt(int(l.Value[idx]))}, nil
	case *object.Map:
		key, ok := index.(object.Hashable)
		if !ok {
			return nil, e.errorAt(pos, "unusable as map key: %s", index.Type())
		}
		pair, ok := l.Pairs[key.MapKey()]
		if !ok {
			if isDotLookup {
				if symbol, ok := index.(*object.Symbol); ok {
					strKey := (&object.String{Value: symbol.Name}).MapKey()
					if strPair, ok := l.Pairs[strKey]; ok {
						return strPair.Value, nil
					}
				}
			}
			return object.NIL, nil
		}
		return pair.Value, nil
	default:
		return nil, e.errorAt(pos, "index operator not supported: %s", left.Type())
	}
}

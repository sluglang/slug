package vm

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/dec64"
	"slug/internal/object"
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
			e.push(value)
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
		if !lok || !rok {
			return e.errorAt(pos, "type mismatch: %s * %s", left.Type(), right.Type())
		}
		e.push(&object.Number{Value: l.Value.Mul(r.Value)})
	case OpDiv:
		l, lok := left.(*object.Number)
		r, rok := right.(*object.Number)
		if !lok || !rok {
			return e.errorAt(pos, "type mismatch: %s / %s", left.Type(), right.Type())
		}
		e.push(&object.Number{Value: l.Value.Div(r.Value, 14, dec64.RoundHalfEven)})
	case OpEqual:
		e.push(e.nativeBool(left == right))
	case OpNotEqual:
		e.push(e.nativeBool(left != right))
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
		Params:  append([]string(nil), fn.Params...),
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
		if _, err := callEnv.DefineConstant(p, bound[i], false, false); err != nil {
			return e.errorAt(pos, "%s", err.Error())
		}
	}

	child := NewExecutor(callEnv, e.externalCall)
	result := child.run(fn.Chunk)
	if result == nil {
		result = object.NIL
	}
	if result.Type() == object.ERROR_OBJ {
		return result
	}
	e.push(result)
	return nil
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

	if len(named) > 0 {
		index := make(map[string]int, paramCount)
		for i, name := range fn.Params {
			index[name] = i
		}
		for name, val := range named {
			idx, ok := index[name]
			if !ok {
				return nil, e.errorAt(pos, "unknown named parameter: %s", name)
			}
			if provided[idx] {
				return nil, e.errorAt(pos, "duplicate assignment to parameter: %s", name)
			}
			values[idx] = val
			provided[idx] = true
		}
	}

	posIndex := 0
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

	for i, ok := range provided {
		if !ok {
			return nil, e.errorAt(pos, "missing required parameter: %s", fn.Params[i])
		}
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

func (e *Executor) evalIndex(left, index object.Object, pos int, isDotLookup bool) (object.Object, *object.Error) {
	switch l := left.(type) {
	case *object.List:
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

package runtime

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/foreign"
	"slug/internal/object"
	"slug/internal/util"
	"slug/internal/vm"
	"sync"
)

type Task struct {
	ID           int64
	Runtime      *Runtime
	OwnerNursery *NurseryScope
	Result       object.Object
	Err          *object.RuntimeError
	Done         chan struct{}
	Observed     bool
	IsFinished   bool
	mu           sync.Mutex

	envStack     []*object.Environment
	nurseryStack []*NurseryScope
}

func (e *Task) NextHandleID() int64                  { return e.Runtime.NextHandleID() }
func (e *Task) GetConfiguration() util.Configuration { return e.Runtime.Config }
func (e *Task) Nil() *object.Nil                     { return object.NIL }
func (e *Task) CurrentEnvStackSize() int             { return len(e.envStack) }

func (e *Task) PushEnv(env *object.Environment) {
	if env.IsThreadNurseryScope {
		e.PushNurseryScope(&NurseryScope{Limit: make(chan struct{}, env.Limit)})
	}
	e.envStack = append(e.envStack, env)
}

func (e *Task) CurrentEnv() *object.Environment {
	if len(e.envStack) == 0 {
		panic("Environment stack is empty in the current frame")
	}
	return e.envStack[len(e.envStack)-1]
}

func (e *Task) PopEnv(result object.Object) object.Object {
	if len(e.envStack) == 0 {
		panic("Attempted to pop from an empty environment stack")
	}
	currentEnv := e.CurrentEnv()
	if currentEnv.IsThreadNurseryScope {
		result, _ = e.popNurseryScope(result)
	}
	// VM runtime handles defer execution internally.
	e.envStack = e.envStack[:len(e.envStack)-1]
	return result
}

func (e *Task) PushNurseryScope(scope *NurseryScope) {
	e.nurseryStack = append(e.nurseryStack, scope)
}

func (e *Task) currentNurseryScope() *NurseryScope {
	if len(e.nurseryStack) == 0 {
		panic("Nursery stack is empty in the current frame")
	}
	return e.nurseryStack[len(e.nurseryStack)-1]
}

func (e *Task) popNurseryScope(result object.Object) (object.Object, bool) {
	currentScope := e.currentNurseryScope()
	nurseryInjected := false
	switch result.(type) {
	case *object.ReturnValue:
		currentScope.CancelChildren(nil, nil, "parent scope exited early")
	case *object.RuntimeError:
		currentScope.CancelChildren(nil, result.(*object.RuntimeError), "parent scope failed")
	case *object.Error:
		currentScope.CancelChildren(nil, nil, "parent scope failed")
	}
	currentScope.WaitChildren()
	if currentScope.NurseryErr != nil {
		if result == nil || (result.Type() != object.ERROR_OBJ && result.Type() != object.RETURN_VALUE_OBJ) {
			result = currentScope.NurseryErr
			currentScope.NurseryErr = nil
			nurseryInjected = true
		}
	}
	e.nurseryStack = e.nurseryStack[:len(e.nurseryStack)-1]
	return result, nurseryInjected
}

func (e *Task) LoadModule(modName string) (*object.Module, error) {
	return e.Runtime.LoadModule(modName)
}

func (e *Task) NewError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func (e *Task) NativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return object.TRUE
	}
	return object.FALSE
}

func buildParamIndex(params []*ast.FunctionParameter) map[string]int {
	index := make(map[string]int, len(params))
	for i, param := range params {
		index[param.Name.Value] = i
	}
	return index
}

type boundArguments struct {
	Values   []object.Object
	Provided []bool
}

func (e *Task) bindArguments(
	pos int,
	fnObj object.Object,
	params []*ast.FunctionParameter,
	positional []object.Object,
	named map[string]object.Object,
) (*boundArguments, object.Object) {
	if params == nil {
		if len(named) > 0 {
			return nil, e.NewError("named arguments are not supported for this function")
		}
		provided := make([]bool, len(positional))
		for i := range provided {
			provided[i] = true
		}
		return &boundArguments{Values: positional, Provided: provided}, nil
	}

	paramCount := len(params)
	values := make([]object.Object, paramCount)
	provided := make([]bool, paramCount)
	hasVariadic := paramCount > 0 && params[paramCount-1].IsVariadic
	variadicIndex := paramCount - 1

	if len(named) > 0 {
		var paramIndex map[string]int
		switch f := fnObj.(type) {
		case *object.Foreign:
			if f.ParamIndex == nil {
				f.ParamIndex = buildParamIndex(params)
			}
			paramIndex = f.ParamIndex
		default:
			paramIndex = buildParamIndex(params)
		}
		for name, val := range named {
			idx, ok := paramIndex[name]
			if !ok {
				return nil, e.NewError("unknown named parameter: %s", name)
			}
			if provided[idx] {
				return nil, e.NewError("duplicate assignment to parameter: %s", name)
			}
			if params[idx].IsVariadic {
				if _, ok := val.(*object.List); !ok {
					return nil, e.NewError("variadic parameter '%s' must be a list when passed by name", name)
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
				return nil, e.NewError("too many positional arguments")
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
			return nil, e.NewError("too many positional arguments")
		}
	}

	for i, param := range params {
		if provided[i] {
			continue
		}
		if param.IsVariadic {
			values[i] = &object.List{Elements: []object.Object{}}
			continue
		}
		if param.Default != nil {
			defaultValue := e.evalDefaultParam(fnObj, param.Default)
			if isError(defaultValue) {
				return nil, defaultValue
			}
			values[i] = defaultValue
			continue
		}
		return nil, e.NewError("missing required parameter: %s", param.Name.Value)
	}

	return &boundArguments{Values: values, Provided: provided}, nil
}

func (e *Task) evalDefaultParam(fnObj object.Object, expr ast.Expression) object.Object {
	if expr == nil {
		return object.NIL
	}
	defEnv := e.CurrentEnv()
	if f, ok := fnObj.(*object.Foreign); ok {
		_ = f
	}
	for defEnv != nil && defEnv.Outer != nil {
		defEnv = defEnv.Outer
	}
	val, err := evalExprWithVM(e.Runtime, defEnv, expr)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	return val
}

func (e *Task) ApplyFunction(pos int, fnName string, fnObj object.Object, positional []object.Object, named map[string]object.Object) object.Object {
	callEnv := e.CurrentEnv()
	fnObj = resolveValue(fnObj)
	if isError(fnObj) {
		return fnObj
	}
	switch fn := fnObj.(type) {
	case *object.FunctionGroup:
		f, err := fn.DispatchToFunction(fnName, positional, named)
		if err != nil {
			return e.NewError("error calling function '%s': %s", fnName, err.Error())
		}
		return e.ApplyFunction(pos, fnName, f, positional, named)
	case *object.Function:
		return e.NewError("treewalking evaluator removed: cannot call AST function value %q", fnName)
	case *object.Foreign:
		var result object.Object
		bound, errObj := e.bindArguments(pos, fn, fn.Parameters, positional, named)
		if errObj != nil {
			return errObj
		}
		callArgs := bound.Values
		if fn.Parameters != nil && len(fn.Parameters) > 0 && fn.Parameters[len(fn.Parameters)-1].IsVariadic {
			variadicIndex := len(fn.Parameters) - 1
			callArgs = append([]object.Object{}, bound.Values[:variadicIndex]...)
			if variadicVal, ok := bound.Values[variadicIndex].(*object.List); ok {
				callArgs = append(callArgs, variadicVal.Elements...)
			} else if bound.Values[variadicIndex] != nil {
				callArgs = append(callArgs, bound.Values[variadicIndex])
			}
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					result = e.NewError("error calling foreign function '%s'", fn.Name)
				}
			}()
			result = fn.Fn(e, callArgs...)
		}()
		if errObj, ok := result.(*object.Error); ok {
			payload := &object.Map{Pairs: map[object.MapKey]object.MapPair{}}
			foreign.PutString(payload, "type", "error")
			foreign.PutString(payload, "foreign", fn.Name)
			foreign.PutString(payload, "msg", errObj.Message)
			return e.runtimeError(pos, "error", payload)
		}
		return result
	case *vm.VMFunction:
		execEnv := callEnv
		if fn.Closure != nil {
			execEnv = fn.Closure
		}
		vmExec := vm.NewExecutorWithBridgeFactory(execEnv, func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
			return makeVMCallBridge(e.Runtime, callEnv)
		})
		return vmExec.EvalFunction(fn, positional, named, pos)
	default:
		if fn == nil {
			return e.NewError("no function found")
		}
		return e.NewError("not a function: %s", fn.Type())
	}
}

func resolveValue(obj object.Object) object.Object {
	for {
		ref, ok := obj.(*object.BindingRef)
		if !ok {
			return obj
		}
		if ref == nil || ref.Env == nil {
			return &object.Error{Message: "invalid binding reference"}
		}
		v, ok := ref.Env.Get(ref.Name)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("identifier not found: %s", ref.Name)}
		}
		obj = v
	}
}

func isError(obj object.Object) bool {
	if obj == nil {
		return false
	}
	return obj.Type() == object.ERROR_OBJ
}

func (e *Task) runtimeError(pos int, typ string, payload object.Object) *object.RuntimeError {
	return &object.RuntimeError{
		Payload:    payload,
		StackTrace: e.GatherStackTrace(nil),
	}
}

func (e *Task) GatherStackTrace(frame *object.StackFrame) []*object.StackFrame {
	if frame != nil {
		return []*object.StackFrame{frame}
	}
	if env := e.CurrentEnv(); env != nil && env.StackInfo != nil {
		return []*object.StackFrame{env.StackInfo}
	}
	return nil
}

func (th *Task) Type() object.ObjectType   { return object.TASK_HANDLE_OBJ }
func (th *Task) Inspect() string           { return fmt.Sprintf("<task %d>", th.ID) }
func (th *Task) DoneChan() <-chan struct{} { return th.Done }
func (th *Task) AwaitResult() object.Object {
	if th.Err != nil {
		return th.Err
	}
	if th.Result == nil {
		return object.NIL
	}
	return th.Result
}

func (th *Task) Complete(res object.Object) {
	th.mu.Lock()
	defer th.mu.Unlock()
	if th.IsFinished {
		return
	}
	if rtErr, ok := res.(*object.RuntimeError); ok {
		th.Err = rtErr
		th.Result = rtErr
	} else {
		th.Result = res
	}
	th.IsFinished = true
	if th.Done != nil {
		close(th.Done)
	}
}

func (th *Task) Cancel(cause *object.RuntimeError, reason string) {
	payload := &object.Map{Pairs: map[object.MapKey]object.MapPair{}}
	foreign.PutString(payload, "type", "cancelled")
	foreign.PutString(payload, "reason", reason)
	th.Complete(&object.RuntimeError{Payload: payload, Cause: cause})
}

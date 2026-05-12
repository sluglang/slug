package vm

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/foreign"
	"slug/internal/object"
	"slug/internal/util"
	"sync"
)

type BridgeFactory func(*object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object

type VMCallContextDeps struct {
	Config        util.Configuration
	LoadModule    func(string) (*object.Module, error)
	NextHandleID  func() int64
	BridgeFactory BridgeFactory
}

type VMCallContext struct {
	ID           int64
	OwnerNursery *NurseryScope
	Result       object.Object
	Err          *object.RuntimeError
	Done         chan struct{}
	Observed     bool
	IsFinished   bool
	mu           sync.Mutex

	deps VMCallContextDeps

	envStack     []*object.Environment
	nurseryStack []*NurseryScope
}

func NewVMCallContext(deps VMCallContextDeps) *VMCallContext {
	return &VMCallContext{deps: deps}
}

func (c *VMCallContext) NextHandleID() int64 {
	if c.deps.NextHandleID == nil {
		return 0
	}
	return c.deps.NextHandleID()
}
func (c *VMCallContext) GetConfiguration() util.Configuration { return c.deps.Config }
func (c *VMCallContext) Nil() *object.Nil                     { return object.NIL }
func (c *VMCallContext) CurrentEnvStackSize() int             { return len(c.envStack) }

func (c *VMCallContext) PushEnv(env *object.Environment) {
	if env.IsThreadNurseryScope {
		c.PushNurseryScope(&NurseryScope{Limit: make(chan struct{}, env.Limit)})
	}
	c.envStack = append(c.envStack, env)
}

func (c *VMCallContext) CurrentEnv() *object.Environment {
	if len(c.envStack) == 0 {
		panic("Environment stack is empty in the current frame")
	}
	return c.envStack[len(c.envStack)-1]
}

func (c *VMCallContext) PopEnv(result object.Object) object.Object {
	if len(c.envStack) == 0 {
		panic("Attempted to pop from an empty environment stack")
	}
	currentEnv := c.CurrentEnv()
	if currentEnv.IsThreadNurseryScope {
		result, _ = c.popNurseryScope(result)
	}
	// VM runtime handles defer execution internally.
	c.envStack = c.envStack[:len(c.envStack)-1]
	return result
}

func (c *VMCallContext) PushNurseryScope(scope *NurseryScope) {
	c.nurseryStack = append(c.nurseryStack, scope)
}

func (c *VMCallContext) currentNurseryScope() *NurseryScope {
	if len(c.nurseryStack) == 0 {
		panic("Nursery stack is empty in the current frame")
	}
	return c.nurseryStack[len(c.nurseryStack)-1]
}

func (c *VMCallContext) popNurseryScope(result object.Object) (object.Object, bool) {
	currentScope := c.currentNurseryScope()
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
	c.nurseryStack = c.nurseryStack[:len(c.nurseryStack)-1]
	return result, nurseryInjected
}

func (c *VMCallContext) LoadModule(modName string) (*object.Module, error) {
	if c.deps.LoadModule == nil {
		return nil, fmt.Errorf("module loader unavailable")
	}
	return c.deps.LoadModule(modName)
}

func (c *VMCallContext) NewError(format string, a ...interface{}) *object.Error {
	err := &object.Error{
		Message: fmt.Sprintf(format, a...),
		Kind:    "RuntimeError",
	}
	if env := c.CurrentEnv(); env != nil {
		err.Path = env.Path
		err.Src = env.Src
		if env.StackInfo != nil {
			err.Position = env.StackInfo.Position
			if err.Path == "" {
				err.Path = env.StackInfo.File
			}
			if err.Src == "" {
				err.Src = env.StackInfo.Src
			}
		}
	}
	return err
}

func (c *VMCallContext) NativeBoolToBooleanObject(input bool) *object.Boolean {
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

func (c *VMCallContext) bindArguments(
	pos int,
	fnObj object.Object,
	params []*ast.FunctionParameter,
	positional []object.Object,
	named map[string]object.Object,
) (*boundArguments, object.Object) {
	if params == nil {
		if len(named) > 0 {
			return nil, c.NewError("named arguments are not supported for this function")
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
				return nil, c.NewError("unknown named parameter: %s", name)
			}
			if provided[idx] {
				return nil, c.NewError("duplicate assignment to parameter: %s", name)
			}
			if params[idx].IsVariadic {
				if _, ok := val.(*object.List); !ok {
					return nil, c.NewError("variadic parameter '%s' must be a list when passed by name", name)
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
				return nil, c.NewError("too many positional arguments")
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
			return nil, c.NewError("too many positional arguments")
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
			defaultValue := c.evalDefaultParam(fnObj, param.Default)
			if isError(defaultValue) {
				return nil, defaultValue
			}
			values[i] = defaultValue
			continue
		}
		return nil, c.NewError("missing required parameter: %s", param.Name.Value)
	}

	return &boundArguments{Values: values, Provided: provided}, nil
}

func (c *VMCallContext) evalDefaultParam(fnObj object.Object, expr ast.Expression) object.Object {
	if expr == nil {
		return object.NIL
	}
	defEnv := c.CurrentEnv()
	if f, ok := fnObj.(*object.Foreign); ok {
		_ = f
	}
	for defEnv != nil && defEnv.Outer != nil {
		defEnv = defEnv.Outer
	}
	val, err := evalExpr(defEnv, expr, c.deps.BridgeFactory)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	return val
}

func (c *VMCallContext) ApplyFunction(pos int, fnName string, fnObj object.Object, positional []object.Object, named map[string]object.Object) object.Object {
	callEnv := c.CurrentEnv()
	fnObj = resolveValue(fnObj)
	if isError(fnObj) {
		return fnObj
	}
	switch fn := fnObj.(type) {
	case *object.FunctionGroup:
		f, err := fn.DispatchToFunction(fnName, positional, named)
		if err != nil {
			return c.NewError("error calling function '%s': %s", fnName, err.Error())
		}
		return c.ApplyFunction(pos, fnName, f, positional, named)
	case *object.Function:
		return c.NewError("runtime does not support direct AST function invocation: %q", fnName)
	case *object.Foreign:
		var result object.Object
		bound, errObj := c.bindArguments(pos, fn, fn.Parameters, positional, named)
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
					result = c.NewError("error calling foreign function '%s'", fn.Name)
				}
			}()
			result = fn.Fn(c, callArgs...)
		}()
		if errObj, ok := result.(*object.Error); ok {
			payload := &object.Map{Pairs: map[object.MapKey]object.MapPair{}}
			foreign.PutString(payload, "type", "error")
			foreign.PutString(payload, "foreign", fn.Name)
			foreign.PutString(payload, "msg", errObj.Message)
			return c.runtimeError(pos, "error", payload)
		}
		return result
	case *VMFunction:
		execEnv := callEnv
		if fn.Closure != nil {
			execEnv = fn.Closure
		}
		vmExec := NewExecutorWithBridgeFactory(execEnv, c.deps.BridgeFactory)
		return vmExec.EvalFunction(fn, positional, named, pos)
	default:
		if fn == nil {
			return c.NewError("no function found")
		}
		return c.NewError("not a function: %s", fn.Type())
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

func (c *VMCallContext) runtimeError(pos int, typ string, payload object.Object) *object.RuntimeError {
	frame := &object.StackFrame{Position: pos}
	if env := c.CurrentEnv(); env != nil {
		frame.File = env.Path
		frame.Src = env.Src
		if env.StackInfo != nil {
			frame.Function = env.StackInfo.Function
			if frame.File == "" {
				frame.File = env.StackInfo.File
			}
			if frame.Src == "" {
				frame.Src = env.StackInfo.Src
			}
		}
	}
	return &object.RuntimeError{
		Payload:    payload,
		StackTrace: c.GatherStackTrace(frame),
	}
}

func (c *VMCallContext) GatherStackTrace(frame *object.StackFrame) []*object.StackFrame {
	if frame != nil {
		return []*object.StackFrame{frame}
	}
	if env := c.CurrentEnv(); env != nil && env.StackInfo != nil {
		return []*object.StackFrame{env.StackInfo}
	}
	return nil
}

func (c *VMCallContext) Type() object.ObjectType   { return object.TASK_HANDLE_OBJ }
func (c *VMCallContext) Inspect() string           { return fmt.Sprintf("<task %d>", c.ID) }
func (c *VMCallContext) DoneChan() <-chan struct{} { return c.Done }
func (c *VMCallContext) AwaitResult() object.Object {
	if c.Err != nil {
		return c.Err
	}
	if c.Result == nil {
		return object.NIL
	}
	return c.Result
}

func (c *VMCallContext) Complete(res object.Object) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.IsFinished {
		return
	}
	if rtErr, ok := res.(*object.RuntimeError); ok {
		c.Err = rtErr
		c.Result = rtErr
	} else {
		c.Result = res
	}
	c.IsFinished = true
	if c.Done != nil {
		close(c.Done)
	}
}

func (c *VMCallContext) Cancel(cause *object.RuntimeError, reason string) {
	payload := &object.Map{Pairs: map[object.MapKey]object.MapPair{}}
	foreign.PutString(payload, "type", "cancelled")
	foreign.PutString(payload, "reason", reason)
	c.Complete(&object.RuntimeError{Payload: payload, Cause: cause})
}

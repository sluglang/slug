package runtime

import (
	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/vm"
)

// ExecuteProgram runs a parsed Slug program on the configured runtime backend.
func ExecuteProgram(mode string, rt *Runtime, env *object.Environment, program *ast.Program) object.Object {
	switch mode {
	case RuntimeVM:
		installBuiltinsIntoEnv(rt, env)
		exec := vm.NewExecutor(env, makeVMCallBridge(rt, env))
		result := exec.EvalProgram(program)
		if result == nil || result.Type() != object.ERROR_OBJ {
			entrypoint, err := FindMainEntrypoint(env)
			if err != nil {
				task := &Task{Runtime: rt}
				return task.NewError("%s", err.Error())
			}
			if entrypoint != nil {
				result = invokeEntrypoint(rt, env, entrypoint)
			}
		}
		return result
	default:
		task := &Task{
			Runtime: rt,
		}
		task.PushNurseryScope(&NurseryScope{
			Limit: make(chan struct{}, rt.Config.DefaultLimit),
		})
		callEnv := object.NewEnclosedEnvironment(env, nil)
		task.PushEnv(callEnv)

		result := task.Eval(program)
		if result == nil || result.Type() != object.ERROR_OBJ {
			entrypoint, err := FindMainEntrypoint(callEnv)
			if err != nil {
				result = task.NewError("%s", err.Error())
			} else if entrypoint != nil {
				result = task.ApplyFunction(0, "@main", entrypoint, nil, nil)
			}
		}
		result = task.PopEnv(result)
		if task.CurrentEnvStackSize() != 0 {
			panic("environment stack not empty after evaluation")
		}
		return result
	}
}

func invokeEntrypoint(rt *Runtime, moduleEnv *object.Environment, entrypoint object.Object) object.Object {
	task := &Task{Runtime: rt}
	task.PushNurseryScope(&NurseryScope{
		Limit: make(chan struct{}, rt.Config.DefaultLimit),
	})
	task.PushEnv(object.NewEnclosedEnvironment(moduleEnv, nil))
	out := task.ApplyFunction(0, "@main", entrypoint, nil, nil)
	out = task.PopEnv(out)
	return out
}

func installBuiltinsIntoEnv(rt *Runtime, env *object.Environment) {
	for name, fn := range rt.Builtins {
		if _, ok := env.Get(name); ok {
			continue
		}
		env.Bindings[name] = &object.Binding{
			Value:     fn,
			IsMutable: false,
			Meta: object.Meta{
				IsImport: false,
				IsExport: false,
			},
		}
	}
}

func makeVMCallBridge(rt *Runtime, env *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
	return func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
		task := &Task{
			Runtime: rt,
		}
		task.PushNurseryScope(&NurseryScope{
			Limit: make(chan struct{}, rt.Config.DefaultLimit),
		})
		// Use an isolated call environment for bridge calls so we inherit bindings
		// via the outer chain without reusing the caller's deferred stack.
		task.PushEnv(object.NewEnclosedEnvironment(env, nil))
		adaptedPositional := positional
		adaptedNamed := named
		if shouldAdaptArgsForIntrospection(callee) {
			adaptedPositional = make([]object.Object, len(positional))
			for i, v := range positional {
				adaptedPositional[i] = adaptVMObjectForTreewalk(v, env)
			}
			if len(named) > 0 {
				adaptedNamed = make(map[string]object.Object, len(named))
				for k, v := range named {
					adaptedNamed[k] = adaptVMObjectForTreewalk(v, env)
				}
			}
		}
		out := task.ApplyFunction(pos, "<vm>", callee, adaptedPositional, adaptedNamed)
		out = task.PopEnv(out)
		return out
	}
}

func shouldAdaptArgsForIntrospection(callee object.Object) bool {
	switch c := callee.(type) {
	case *object.Foreign:
		return c.Name == "describe"
	case *object.FunctionGroup:
		for _, fn := range c.Functions {
			if f, ok := fn.(*object.Foreign); ok && f.Name == "describe" {
				return true
			}
		}
		for _, g := range c.Delegates {
			for _, fn := range g.Functions {
				if f, ok := fn.(*object.Foreign); ok && f.Name == "describe" {
					return true
				}
			}
		}
	}
	return false
}

func adaptVMObjectForTreewalk(obj object.Object, env *object.Environment) object.Object {
	switch v := obj.(type) {
	case *vm.VMFunction:
		return &object.Function{
			Signature:  v.Signature,
			Tags:       v.Tags,
			Parameters: v.Parameters,
			ParamIndex: buildParamIndex(v.Parameters),
			Env:        env,
			Body:       &ast.BlockStatement{},
		}
	case *object.FunctionGroup:
		out := &object.FunctionGroup{
			Functions: make(map[ast.FSig]object.Object, len(v.Functions)),
		}
		for sig, fn := range v.Functions {
			out.Functions[sig] = adaptVMObjectForTreewalk(fn, env)
		}
		return out
	case *object.List:
		els := make([]object.Object, len(v.Elements))
		for i, e := range v.Elements {
			els[i] = adaptVMObjectForTreewalk(e, env)
		}
		return &object.List{Elements: els}
	case *object.Map:
		pairs := make(map[object.MapKey]object.MapPair, len(v.Pairs))
		for k, p := range v.Pairs {
			pairs[k] = object.MapPair{Key: p.Key, Value: adaptVMObjectForTreewalk(p.Value, env)}
		}
		return &object.Map{Pairs: pairs}
	default:
		return obj
	}
}

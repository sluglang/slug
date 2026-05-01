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
		return exec.EvalProgram(program)
	default:
		task := &Task{
			Runtime: rt,
		}
		task.PushNurseryScope(&NurseryScope{
			Limit: make(chan struct{}, rt.Config.DefaultLimit),
		})
		task.PushEnv(env)

		result := task.Eval(program)
		if result == nil || result.Type() != object.ERROR_OBJ {
			entrypoint, err := FindMainEntrypoint(env)
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
		task.PushEnv(env)
		out := task.ApplyFunction(pos, "<vm>", callee, positional, named)
		out = task.PopEnv(out)
		return out
	}
}

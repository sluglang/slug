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
		exec := vm.NewExecutor(env)
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

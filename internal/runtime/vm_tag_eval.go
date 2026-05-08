package runtime

import (
	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/vm"
)

func evalTagArgsWithVM(rt *Runtime, env *object.Environment, tags []*ast.Tag) (map[string]object.List, error) {
	return vm.EvalTagArgs(env, tags, func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
		return makeVMCallBridge(rt, callEnv)
	})
}

func evalExprWithVM(rt *Runtime, env *object.Environment, expr ast.Expression) (object.Object, error) {
	return vm.EvalExpr(env, expr, func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
		return makeVMCallBridge(rt, callEnv)
	})
}

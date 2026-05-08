package runtime

import (
	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/vm"
)

// prepareProgramForVM resolves `foreign` declarations against the runtime registry
// and returns a copy of the program without foreign declaration statements.
func prepareProgramForVM(rt *Runtime, env *object.Environment, program *ast.Program) (*ast.Program, error) {
	return vm.PrepareProgram(env, program, rt.LookupForeign, hasExportTag, buildParamIndex)
}

// applyForeignTagsForVM evaluates and applies tag payloads for foreign declarations
// after module code has executed, so tag expressions can reference module bindings.
func applyForeignTagsForVM(rt *Runtime, env *object.Environment, program *ast.Program) error {
	return vm.ApplyForeignTags(env, program, rt.LookupForeign, func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
		return makeVMCallBridge(rt, callEnv)
	})
}

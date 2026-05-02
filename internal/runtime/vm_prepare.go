package runtime

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/object"
)

// prepareProgramForVM resolves `foreign` declarations against the runtime registry
// and returns a copy of the program without foreign declaration statements.
func prepareProgramForVM(rt *Runtime, env *object.Environment, program *ast.Program) (*ast.Program, error) {
	if program == nil {
		return nil, nil
	}

	out := &ast.Program{
		Statements:   make([]ast.Statement, 0, len(program.Statements)),
		ModuleDoc:    program.ModuleDoc,
		HasModuleDoc: program.HasModuleDoc,
	}

	for _, stmt := range program.Statements {
		ff, ok := stmt.(*ast.ForeignFunctionDeclaration)
		if !ok {
			out.Statements = append(out.Statements, stmt)
			continue
		}

		functionName := ff.Name.Value
		fqn := env.ModuleFqn + "." + functionName
		foreignFn, exists := rt.LookupForeign(fqn)
		if !exists {
			return nil, fmt.Errorf("unknown foreign function %s", fqn)
		}

		foreignFn.Tags = make(map[string]object.List)
		foreignFn.Parameters = ff.Parameters
		foreignFn.ParamIndex = buildParamIndex(ff.Parameters)
		foreignFn.Name = functionName
		foreignFn.Signature = ff.Signature
		isExported := hasExportTag(ff.Tags)
		if _, err := env.DefineConstant(functionName, foreignFn, isExported, false); err != nil {
			return nil, err
		}
		if ff.HasDoc {
			env.SetLocalDoc(functionName, ff.Doc)
		}
	}

	return out, nil
}

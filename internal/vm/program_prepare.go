package vm

import (
	"errors"
	"fmt"
	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/token"
)

// PrepareProgram resolves `foreign` declarations against the provided registry
// and returns a copy of the program without foreign declaration statements.
func PrepareProgram(
	env *object.Environment,
	program *ast.Program,
	lookupForeign func(string) (*object.Foreign, bool),
	hasExportTag func([]*ast.Tag) bool,
	buildParamIndex func([]*ast.FunctionParameter) map[string]int,
) (*ast.Program, error) {
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
		foreignFn, exists := lookupForeign(fqn)
		if !exists {
			return nil, fmt.Errorf("unknown foreign function %s", fqn)
		}

		foreignFn.Tags = make(map[string]object.List)
		foreignFn.Parameters = ff.Parameters
		foreignFn.ParamIndex = buildParamIndex(ff.Parameters)
		foreignFn.ReturnType = ff.ReturnType
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

// ApplyForeignTags evaluates and applies tag payloads for foreign declarations
// after module code has executed, so tag expressions can reference module bindings.
func ApplyForeignTags(
	env *object.Environment,
	program *ast.Program,
	lookupForeign func(string) (*object.Foreign, bool),
	bridgeFactory func(*object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object,
) error {
	if program == nil {
		return nil
	}
	for _, stmt := range program.Statements {
		ff, ok := stmt.(*ast.ForeignFunctionDeclaration)
		if !ok {
			continue
		}
		functionName := ff.Name.Value
		fqn := env.ModuleFqn + "." + functionName
		foreignFn, exists := lookupForeign(fqn)
		if !exists {
			return fmt.Errorf("unknown foreign function %s", fqn)
		}
		tags, err := EvalTagArgs(env, ff.Tags, bridgeFactory)
		if err != nil {
			return err
		}
		foreignFn.Tags = tags
	}
	return nil
}

func EvalTagArgs(
	env *object.Environment,
	tags []*ast.Tag,
	bridgeFactory func(*object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object,
) (map[string]object.List, error) {
	result := make(map[string]object.List, len(tags))
	for _, tag := range tags {
		args := make([]object.Object, 0, len(tag.Args))
		for _, arg := range tag.Args {
			val, err := evalExpr(env, arg, bridgeFactory)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate tag %s argument: %w", tag.Name, err)
			}
			args = append(args, val)
		}
		result[tag.Name] = object.List{Elements: args}
	}
	return result, nil
}

func evalExpr(
	env *object.Environment,
	expr ast.Expression,
	bridgeFactory func(*object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object,
) (object.Object, error) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				Token:       token.Token{Type: token.RETURN, Literal: "return"},
				ReturnValue: expr,
			},
		},
	}
	exec := NewExecutorWithBridgeFactory(env, bridgeFactory)
	out := exec.EvalProgram(prog)
	switch v := out.(type) {
	case nil:
		return object.NIL, nil
	case *object.Error:
		return nil, errors.New(v.Message)
	case *object.RuntimeError:
		return nil, errors.New(v.Inspect())
	default:
		return out, nil
	}
}

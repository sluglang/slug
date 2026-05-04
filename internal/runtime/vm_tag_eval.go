package runtime

import (
	"errors"
	"fmt"
	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/token"
	"slug/internal/vm"
)

func evalTagArgsWithVM(rt *Runtime, env *object.Environment, tags []*ast.Tag) (map[string]object.List, error) {
	result := make(map[string]object.List, len(tags))
	for _, tag := range tags {
		args := make([]object.Object, 0, len(tag.Args))
		for _, arg := range tag.Args {
			val, err := evalExprWithVM(rt, env, arg)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate tag %s argument: %w", tag.Name, err)
			}
			args = append(args, val)
		}
		result[tag.Name] = object.List{Elements: args}
	}
	return result, nil
}

func evalExprWithVM(rt *Runtime, env *object.Environment, expr ast.Expression) (object.Object, error) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				Token:       token.Token{Type: token.RETURN, Literal: "return"},
				ReturnValue: expr,
			},
		},
	}
	exec := vm.NewExecutorWithBridgeFactory(env, func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
		return makeVMCallBridge(rt, callEnv)
	})
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

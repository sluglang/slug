package runtime

import (
	"strings"
	"testing"

	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/token"
)

func TestEvalCallExpressionUsesCalleeNameInDispatchErrors(t *testing.T) {
	env := object.NewEnvironment()
	env.Path = "test.slug"
	env.Src = "foo()"

	fn := &object.Function{
		Signature: ast.FSig{Min: 1, Max: 1},
		Parameters: []*ast.FunctionParameter{
			{
				Name: &ast.Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "arg"},
					Value: "arg",
				},
			},
		},
	}
	env.Bindings["foo"] = &object.Binding{
		Value: &object.FunctionGroup{
			Functions: map[ast.FSig]object.Object{
				fn.Signature: fn,
			},
		},
	}

	task := &Task{
		Runtime: &Runtime{
			Builtins: map[string]*object.Foreign{},
		},
		envStack: []*object.Environment{env},
	}

	call := &ast.CallExpression{
		Token: token.Token{Type: token.LPAREN, Literal: "(", Position: 3},
		Function: &ast.Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "foo", Position: 0},
			Value: "foo",
		},
	}

	result := task.Eval(call)
	errObj, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("expected error object, got %T", result)
	}

	if !strings.Contains(errObj.Message, "error calling function 'foo'") {
		t.Fatalf("expected callee name in error, got %q", errObj.Message)
	}
	if strings.Contains(errObj.Message, "error calling function '('") {
		t.Fatalf("unexpected token literal used as function name: %q", errObj.Message)
	}
	if !strings.Contains(errObj.Message, "Stacktrace:") {
		t.Fatalf("expected stacktrace in error output, got %q", errObj.Message)
	}
	if !strings.Contains(errObj.Message, "test.slug") {
		t.Fatalf("expected file path in stacktrace, got %q", errObj.Message)
	}
}

func TestCallDisplayNameForDotLookup(t *testing.T) {
	task := &Task{}

	name := task.callDisplayName(&ast.IndexExpression{
		Left: &ast.Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "std"},
			Value: "std",
		},
		Index: &ast.SymbolLiteral{
			Token: token.Token{Type: token.IDENT, Literal: "put"},
			Value: "put",
		},
		IsDotLookup: true,
	})

	if name != "std.put" {
		t.Fatalf("expected std.put, got %q", name)
	}
}

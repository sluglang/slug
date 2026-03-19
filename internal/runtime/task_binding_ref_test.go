package runtime

import (
	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/token"
	"testing"
)

func TestEvalCallArgumentsResolvesBindingRefFromDotLookup(t *testing.T) {
	task := &Task{
		Runtime: &Runtime{
			Builtins: map[string]*object.Foreign{},
		},
	}

	moduleEnv := object.NewEnvironment()
	if _, err := moduleEnv.DefineConstant("SQLITE_DRIVER", &object.String{Value: "sqlite3"}, true, false); err != nil {
		t.Fatalf("define SQLITE_DRIVER: %v", err)
	}

	imported := &object.Map{Pairs: map[object.MapKey]object.MapPair{}}
	imported.Put(object.InternSymbol("SQLITE_DRIVER"), &object.BindingRef{Env: moduleEnv, Name: "SQLITE_DRIVER"})

	root := object.NewEnvironment()
	if _, err := root.DefineConstant("db", imported, false, false); err != nil {
		t.Fatalf("define db: %v", err)
	}
	task.PushEnv(root)

	positional, named, errObj := task.evalCallArguments(1, []ast.Expression{
		&ast.IndexExpression{
			Token:       token.Token{Type: token.PERIOD, Literal: ".", Position: 1},
			Left:        &ast.Identifier{Token: token.Token{Type: token.IDENT, Literal: "db", Position: 1}, Value: "db"},
			Index:       &ast.SymbolLiteral{Token: token.Token{Type: token.SYMBOL, Literal: "SQLITE_DRIVER", Position: 1}, Value: "SQLITE_DRIVER"},
			IsDotLookup: true,
		},
	})
	if errObj != nil {
		t.Fatalf("unexpected evalCallArguments error: %s", errObj.Inspect())
	}
	if named != nil {
		t.Fatalf("expected no named arguments, got %v", named)
	}
	if len(positional) != 1 {
		t.Fatalf("expected one positional argument, got %d", len(positional))
	}
	driver, ok := positional[0].(*object.String)
	if !ok {
		t.Fatalf("expected resolved driver to be string, got %T (%s)", positional[0], positional[0].Type())
	}
	if driver.Value != "sqlite3" {
		t.Fatalf("expected resolved driver sqlite3, got %q", driver.Value)
	}
}

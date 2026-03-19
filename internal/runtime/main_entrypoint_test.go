package runtime

import (
	"slug/internal/ast"
	"slug/internal/object"
	"testing"
)

func TestFindMainEntrypointReturnsNilWhenMissing(t *testing.T) {
	env := object.NewEnvironment()
	env.Bindings["x"] = &object.Binding{Value: &object.Number{}}

	mainFn, err := FindMainEntrypoint(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mainFn != nil {
		t.Fatalf("expected nil entrypoint, got %T", mainFn)
	}
}

func TestFindMainEntrypointFindsTaggedFunction(t *testing.T) {
	env := object.NewEnvironment()
	fn := &object.Function{
		Signature: ast.FSig{Min: 0, Max: 0},
	}
	fn.SetTag(object.MAIN_TAG, object.List{})

	env.Bindings["start"] = &object.Binding{
		Value: &object.FunctionGroup{
			Functions: map[ast.FSig]object.Object{
				fn.Signature: fn,
			},
		},
	}

	mainFn, err := FindMainEntrypoint(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mainFn != fn {
		t.Fatalf("expected tagged function, got %T", mainFn)
	}
}

func TestFindMainEntrypointRejectsDuplicates(t *testing.T) {
	env := object.NewEnvironment()

	fnA := &object.Function{Signature: ast.FSig{Min: 0, Max: 0}}
	fnA.SetTag(object.MAIN_TAG, object.List{})
	fnB := &object.Function{Signature: ast.FSig{Min: 0, Max: 0}}
	fnB.SetTag(object.MAIN_TAG, object.List{})

	env.Bindings["a"] = &object.Binding{
		Value: &object.FunctionGroup{
			Functions: map[ast.FSig]object.Object{
				fnA.Signature: fnA,
			},
		},
	}
	env.Bindings["b"] = &object.Binding{
		Value: &object.FunctionGroup{
			Functions: map[ast.FSig]object.Object{
				fnB.Signature: fnB,
			},
		},
	}

	_, err := FindMainEntrypoint(env)
	if err == nil {
		t.Fatal("expected duplicate @main error, got nil")
	}
}

func TestFindMainEntrypointIgnoresImportedBindings(t *testing.T) {
	env := object.NewEnvironment()

	localMain := &object.Function{Signature: ast.FSig{Min: 0, Max: 0}}
	localMain.SetTag(object.MAIN_TAG, object.List{})
	importedMain := &object.Function{Signature: ast.FSig{Min: 0, Max: 0}}
	importedMain.SetTag(object.MAIN_TAG, object.List{})

	env.Bindings["start"] = &object.Binding{
		Value: &object.FunctionGroup{
			Functions: map[ast.FSig]object.Object{
				localMain.Signature: localMain,
			},
		},
	}
	env.Bindings["importedStart"] = &object.Binding{
		Value: &object.FunctionGroup{
			Functions: map[ast.FSig]object.Object{
				importedMain.Signature: importedMain,
			},
		},
		Meta: object.Meta{IsImport: true},
	}

	mainFn, err := FindMainEntrypoint(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mainFn != localMain {
		t.Fatalf("expected local @main to win, got %T", mainFn)
	}
}

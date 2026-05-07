package runtime

import (
	"fmt"
	"io"
	"os"
	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/util"
	"strings"
	"testing"
)

type importTestContext struct {
	env     *object.Environment
	modules map[string]*object.Module
}

func (c *importTestContext) CurrentEnv() *object.Environment {
	return c.env
}

func (c *importTestContext) ApplyFunction(int, string, object.Object, []object.Object, map[string]object.Object) object.Object {
	return nil
}

func (c *importTestContext) NewError(message string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(message, a...)}
}

func (c *importTestContext) Nil() *object.Nil {
	return object.NIL
}

func (c *importTestContext) NativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return object.TRUE
	}
	return object.FALSE
}

func (c *importTestContext) LoadModule(pathParts string) (*object.Module, error) {
	module, ok := c.modules[pathParts]
	if !ok {
		return nil, fmt.Errorf("module not found: %s", pathParts)
	}
	return module, nil
}

func (c *importTestContext) GetConfiguration() util.Configuration {
	return util.Configuration{}
}

func (c *importTestContext) NextHandleID() int64 {
	return 1
}

func exportedFunctionModule(moduleName string, fnName string, sig ast.FSig) *object.Module {
	fn := &object.Foreign{
		Name:      fnName,
		Signature: sig,
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			return object.NIL
		},
	}
	group := &object.FunctionGroup{
		Functions: map[ast.FSig]object.Object{
			sig: fn,
		},
	}
	env := object.NewEnvironment()
	env.Bindings[fnName] = &object.Binding{
		Value: group,
		Meta: object.Meta{
			IsExport: true,
		},
	}
	return &object.Module{
		Name: moduleName,
		Env:  env,
	}
}

func withCapturedStderr(t *testing.T, fn func()) string {
	t.Helper()

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe creation failed: %v", err)
	}
	defer func() {
		os.Stderr = origStderr
	}()

	os.Stderr = w
	fn()
	_ = w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}
	_ = r.Close()
	return string(out)
}

func TestImportIgnoresDuplicateModuleArgsWithoutWarning(t *testing.T) {
	importFn := fnBuiltinImport()
	ctx := &importTestContext{
		env: &object.Environment{ModuleFqn: "doc"},
		modules: map[string]*object.Module{
			"slug.list": exportedFunctionModule("slug.list", "indexOf", ast.FSig{
				Tags:       "@list||@num|",
				Min:        2,
				Max:        3,
				IsVariadic: false,
			}),
			"slug.string": exportedFunctionModule("slug.string", "indexOf", ast.FSig{
				Tags:       "@str|@str|@num|",
				Min:        2,
				Max:        3,
				IsVariadic: false,
			}),
		},
	}

	var result object.Object
	out := withCapturedStderr(t, func() {
		result = importFn.Fn(
			ctx,
			&object.String{Value: "slug.list"},
			&object.String{Value: "slug.string"},
			&object.String{Value: "slug.list"},
			&object.String{Value: "slug.string"},
		)
	})

	if strings.Contains(out, "duplicate signature") {
		t.Fatalf("unexpected duplicate signature warning: %q", out)
	}

	m, ok := result.(*object.Map)
	if !ok {
		t.Fatalf("expected import to return map, got %T", result)
	}
	indexOf, ok := m.Get(object.InternSymbol("indexOf"))
	if !ok {
		t.Fatal("expected merged import to contain indexOf")
	}
	fg, ok := indexOf.(*object.FunctionGroup)
	if !ok {
		t.Fatalf("expected indexOf to be a function group, got %T", indexOf)
	}
	if len(fg.Delegates) != 1 {
		t.Fatalf("expected exactly one delegate after deduped imports, got %d", len(fg.Delegates))
	}
}

func TestImportWarnsOnRealDuplicateSignature(t *testing.T) {
	importFn := fnBuiltinImport()
	sharedSig := ast.FSig{
		Tags:       "@str|@str|@num|",
		Min:        2,
		Max:        3,
		IsVariadic: false,
	}
	ctx := &importTestContext{
		env: &object.Environment{ModuleFqn: "doc"},
		modules: map[string]*object.Module{
			"slug.string": exportedFunctionModule("slug.string", "indexOf", sharedSig),
			"slug.regex":  exportedFunctionModule("slug.regex", "indexOf", sharedSig),
		},
	}

	out := withCapturedStderr(t, func() {
		_ = importFn.Fn(
			ctx,
			&object.String{Value: "slug.string"},
			&object.String{Value: "slug.regex"},
		)
	})

	if !strings.Contains(out, "duplicate signature") {
		t.Fatalf("expected duplicate signature warning for real collision, got %q", out)
	}
}

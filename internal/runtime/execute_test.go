package runtime

import (
	"slug/internal/lexer"
	"slug/internal/object"
	"slug/internal/parser"
	"slug/internal/util"
	"testing"
)

func TestExecuteProgramVMCallsBuiltinThroughBridge(t *testing.T) {
	source := `len("slug")`

	l := lexer.New(source)
	p := parser.New(l, "vm-test.slug", source)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	cfg := util.Configuration{
		RootPath:     ".",
		ProjectRoot:  ".",
		Cwd:          ".",
		DefaultLimit: 4,
		MainModule:   "vm-test",
		RuntimeMode:  RuntimeVM,
	}
	rt := NewRuntime(cfg)
	env := object.NewRootEnvironment(cfg.DefaultLimit)
	env.Path = "vm-test.slug"
	env.Src = source
	env.ModuleFqn = "vm-test"

	result := ExecuteProgram(rt, env, program)
	num, ok := result.(*object.Number)
	if !ok {
		t.Fatalf("expected number result, got %T (%s)", result, result.Inspect())
	}
	if num.Value.String() != "4" {
		t.Fatalf("expected 4, got %s", num.Value.String())
	}
}

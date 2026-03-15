package object

import (
	"slug/internal/ast"
	"strings"
	"testing"
)

func TestDefineConstant_AllowsMixedCallableOverloads(t *testing.T) {
	env := NewEnvironment()

	foreignSig := ast.FSig{Tags: "@str", Min: 1, Max: 1}
	localSig := ast.FSig{Tags: "@str,@str", Min: 2, Max: 2}

	if _, err := env.DefineConstant("trim", &Foreign{Name: "trim", Signature: foreignSig}, false, false); err != nil {
		t.Fatalf("unexpected foreign define error: %v", err)
	}
	if _, err := env.DefineConstant("trim", &Function{Signature: localSig}, false, false); err != nil {
		t.Fatalf("unexpected local overload define error: %v", err)
	}

	binding, ok := env.GetLocalBinding("trim")
	if !ok {
		t.Fatalf("expected trim binding")
	}
	fg, ok := binding.Value.(*FunctionGroup)
	if !ok {
		t.Fatalf("expected FunctionGroup, got %T", binding.Value)
	}
	if len(fg.Functions) != 2 {
		t.Fatalf("expected 2 overloads, got %d", len(fg.Functions))
	}
}

func TestDefineConstant_RejectsDuplicateCallableSignature(t *testing.T) {
	env := NewEnvironment()
	sig := ast.FSig{Tags: "@str", Min: 1, Max: 1}

	if _, err := env.DefineConstant("trim", &Foreign{Name: "trim", Signature: sig}, false, false); err != nil {
		t.Fatalf("unexpected foreign define error: %v", err)
	}

	_, err := env.DefineConstant("trim", &Function{Signature: sig}, false, false)
	if err == nil {
		t.Fatalf("expected duplicate signature error")
	}
	if !strings.Contains(err.Error(), "already has an overload with signature") {
		t.Fatalf("unexpected duplicate signature error: %v", err)
	}
}

func TestDefineConstant_NonCallableValRemainsImmutable(t *testing.T) {
	env := NewEnvironment()

	if _, err := env.DefineConstant("x", &String{Value: "a"}, false, false); err != nil {
		t.Fatalf("unexpected define error: %v", err)
	}
	_, err := env.DefineConstant("x", &String{Value: "b"}, false, false)
	if err == nil {
		t.Fatalf("expected immutability error")
	}
	if !strings.Contains(err.Error(), "already defined as a 'val'") {
		t.Fatalf("unexpected immutability error: %v", err)
	}
}

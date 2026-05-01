package vm

import (
	"slug/internal/lexer"
	"slug/internal/object"
	"slug/internal/parser"
	"testing"
)

func runVM(t *testing.T, input string) object.Object {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l, "test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	exec := NewExecutor(object.NewRootEnvironment(4))
	return exec.EvalProgram(program)
}

func TestExecutorArithmetic(t *testing.T) {
	got := runVM(t, "1 + 2 * 3")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "7" {
		t.Fatalf("expected 7, got %s", num.Value.String())
	}
}

func TestExecutorValAndIdentifier(t *testing.T) {
	got := runVM(t, "val x = 41\nx + 1")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "42" {
		t.Fatalf("expected 42, got %s", num.Value.String())
	}
}

func TestExecutorIfExpression(t *testing.T) {
	got := runVM(t, "if (true) { 10 } else { 20 }")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "10" {
		t.Fatalf("expected 10, got %s", num.Value.String())
	}
}

func TestExecutorUnsupportedExpression(t *testing.T) {
	got := runVM(t, "[1, 2, 3]")
	if got.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error object, got %T (%s)", got, got.Inspect())
	}
}

func TestExecutorVarAssignment(t *testing.T) {
	got := runVM(t, "var x = 2\nx = x + 5\nx")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "7" {
		t.Fatalf("expected 7, got %s", num.Value.String())
	}
}

func TestExecutorFunctionCall(t *testing.T) {
	got := runVM(t, "val add = fn(a, b) { a + b }\nadd(40, 2)")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "42" {
		t.Fatalf("expected 42, got %s", num.Value.String())
	}
}

func TestExecutorClosureCapture(t *testing.T) {
	got := runVM(t, "val base = 10\nval addBase = fn(x) { x + base }\naddBase(5)")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "15" {
		t.Fatalf("expected 15, got %s", num.Value.String())
	}
}

func TestExecutorShortCircuitAndOr(t *testing.T) {
	got := runVM(t, "var x = 0\nfalse && (x = 1)\ntrue || (x = 2)\nx")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "0" {
		t.Fatalf("expected 0, got %s", num.Value.String())
	}
}

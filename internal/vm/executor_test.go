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

	exec := NewExecutor(object.NewRootEnvironment(4), nil)
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
	got := runVM(t, "match 1 { 1 => 2 }")
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

func TestExecutorNamedArgumentsVMFunction(t *testing.T) {
	got := runVM(t, "val sub = fn(a, b) { a - b }\nsub(b = 2, a = 9)")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "7" {
		t.Fatalf("expected 7, got %s", num.Value.String())
	}
}

func TestExecutorPositionalAfterNamedRejected(t *testing.T) {
	got := runVM(t, "val sub = fn(a, b) { a - b }\nsub(a = 9, 2)")
	if got.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error object, got %T (%s)", got, got.Inspect())
	}
}

func TestExecutorSpreadArgumentsVMFunction(t *testing.T) {
	got := runVM(t, "val pair = fn(a, b) { a + b }\nval xs = [20, 22]\npair(...xs)")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "42" {
		t.Fatalf("expected 42, got %s", num.Value.String())
	}
}

func TestExecutorListStringBytesIndexing(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{code: "val xs = [10, 20, 30]\nxs[1]", want: "20"},
		{code: "\"slug\"[2]", want: "u"},
		{code: "0x\"0102ff\"[2]", want: "255"},
		{code: "[1][-99]", want: "nil"},
	}

	for _, tt := range tests {
		got := runVM(t, tt.code)
		if got.Inspect() != tt.want {
			t.Fatalf("for %q expected %q got %q", tt.code, tt.want, got.Inspect())
		}
	}
}

func TestExecutorListStringBytesSlicing(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{code: "[10, 20, 30, 40][1:3]", want: "[20, 30]"},
		{code: "\"slug\"[1:]", want: "lug"},
		{code: "\"slug\"[0:-1]", want: "slu"},
		{code: "0x\"01020304\"[1:4:2]", want: "0x\"0204\""},
		{code: "[1,2,3,4,5][0:5:2]", want: "[1, 3, 5]"},
	}

	for _, tt := range tests {
		got := runVM(t, tt.code)
		if got.Inspect() != tt.want {
			t.Fatalf("for %q expected %q got %q", tt.code, tt.want, got.Inspect())
		}
	}
}

func TestExecutorMapLiteralAndIndex(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{code: "{:a: 1, \"a\": 2}[:a]", want: "1"},
		{code: "{:a: 1, \"a\": 2}[\"a\"]", want: "2"},
		{code: "{:a: 1}[:missing]", want: "nil"},
	}

	for _, tt := range tests {
		got := runVM(t, tt.code)
		if got.Inspect() != tt.want {
			t.Fatalf("for %q expected %q got %q", tt.code, tt.want, got.Inspect())
		}
	}
}

func TestExecutorMapKeyMustBeHashable(t *testing.T) {
	got := runVM(t, "{:a: 1}[[]]")
	if got.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error object, got %T (%s)", got, got.Inspect())
	}
}

func TestExecutorMapDotLookupTolerance(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{code: "{\"name\": \"slug\"}.name", want: "slug"},
		{code: "{:name: \"sym\", \"name\": \"str\"}.name", want: "sym"},
		{code: "{}.missing", want: "nil"},
	}

	for _, tt := range tests {
		got := runVM(t, tt.code)
		if got.Inspect() != tt.want {
			t.Fatalf("for %q expected %q got %q", tt.code, tt.want, got.Inspect())
		}
	}
}

func TestExecutorFunctionDefaultParameter(t *testing.T) {
	got := runVM(t, "val add1 = fn(a = 41) { a + 1 }\nadd1()")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "42" {
		t.Fatalf("expected 42, got %s", num.Value.String())
	}
}

func TestExecutorFunctionVariadicParameter(t *testing.T) {
	got := runVM(t, "val pick = fn(...xs) { xs[2] }\npick(1, 2, 3)")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "3" {
		t.Fatalf("expected 3, got %s", num.Value.String())
	}
}

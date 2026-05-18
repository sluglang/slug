package vm

import (
	"slug/internal/lexer"
	"slug/internal/object"
	"slug/internal/parser"
	"slug/internal/semantic"
	"strings"
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
	if errs := semantic.Analyze("test.slug", input, program); len(errs) > 0 {
		t.Fatalf("semantic errors: %v", errs)
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
	got := runVM(t, "throw 1")
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

func TestExecutorRuntimeChecksAnnotatedReturnType(t *testing.T) {
	got := runVM(t, "val id = fn(x):num { x }\nid(\"oops\")")
	if got.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error object, got %T (%s)", got, got.Inspect())
	}
	if got.Inspect() == "" || !strings.Contains(got.Inspect(), "function return type mismatch") {
		t.Fatalf("expected function return type mismatch error, got %s", got.Inspect())
	}
}

func TestExecutorRuntimeChecksAnnotatedReturnTypeUsesStandardSlugFormat(t *testing.T) {
	input := "val id = fn(x):num { x }\nid(nil)"
	l := lexer.New(input)
	p := parser.New(l, "test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if errs := semantic.Analyze("test.slug", input, program); len(errs) > 0 {
		t.Fatalf("semantic errors: %v", errs)
	}

	exec := NewExecutor(object.NewRootEnvironment(4), nil)
	exec.env.Path = "test.slug"
	exec.env.Src = input

	got := exec.EvalProgram(program)
	if got.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error object, got %T (%s)", got, got.Inspect())
	}
	rendered := got.Inspect()
	if !strings.Contains(rendered, "RuntimeError: function return type mismatch") {
		t.Fatalf("expected runtime error prefix, got %s", rendered)
	}
	if !strings.Contains(rendered, "--> test.slug:2:3") {
		t.Fatalf("expected source context, got %s", rendered)
	}
}

func TestExecutorRuntimeChecksAnnotatedListReturnTypeUsesFormalShape(t *testing.T) {
	got := runVM(t, `
val zipWithIndex = fn():list<[any, str]> {
  [[1, 0], [2, 1]]
}
zipWithIndex()
`)
	if got.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error object, got %T (%s)", got, got.Inspect())
	}
	if got.Inspect() == "" || !strings.Contains(got.Inspect(), "got list<[num, num]>") {
		t.Fatalf("expected formal list type description in error, got %s", got.Inspect())
	}
}

func TestExecutorRuntimeChecksAnnotatedFunctionReturnType(t *testing.T) {
	got := runVM(t, `
val factory = fn():fn<num, num> {
  fn(n:num):num { n + 1 }
}
val inc = factory()
inc(1)
`)
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "2" {
		t.Fatalf("expected 2, got %s", num.Value.String())
	}
}

func TestExecutorRuntimeErrorFormattingIncludesSourceContext(t *testing.T) {
	exec := NewExecutor(object.NewRootEnvironment(4), nil)
	exec.env.Path = "test.slug"
	exec.env.Src = "val x = 1\n1 + true\n"
	err := exec.errorAt(len("val x = 1\n"), "type mismatch: num vs bool")
	if err == nil {
		t.Fatal("expected runtime error")
	}
	rendered := err.Inspect()
	if !strings.Contains(rendered, "RuntimeError: type mismatch: num vs bool") {
		t.Fatalf("expected runtime error prefix, got %s", rendered)
	}
	if !strings.Contains(rendered, "--> test.slug:2:1") {
		t.Fatalf("expected source location, got %s", rendered)
	}
	if !strings.Contains(rendered, "^ unexpected here") {
		t.Fatalf("expected source context, got %s", rendered)
	}
}

func TestExecutorRuntimeErrorFormattingForAssertionFailure(t *testing.T) {
	rtErr := &object.RuntimeError{
		Payload: &object.String{Value: "AssertionError: expected 1 got 2"},
		StackTrace: []*object.StackFrame{
			{
				File:     "test.slug",
				Src:      "assertEqual(1, 2)\n",
				Position: 0,
			},
		},
	}
	rendered := rtErr.Inspect()
	if !strings.Contains(rendered, "RuntimeError: AssertionError: expected 1 got 2") {
		t.Fatalf("expected assertion error payload, got %s", rendered)
	}
	if !strings.Contains(rendered, "--> test.slug:1:1") || !strings.Contains(rendered, "^ unexpected here") {
		t.Fatalf("expected source context in runtime error, got %s", rendered)
	}
}

func TestExecutorRuntimeChecksGenericAnnotatedReturnType(t *testing.T) {
	got := runVM(t, `
val zipWithIndex = fn<T>(lst:list<T>):list<[T, num]> {
  [[lst[0], 0]]
}
["a", "b", "c"] /> zipWithIndex
`)
	if got.Type() != object.LIST_OBJ {
		t.Fatalf("expected list result, got %T (%s)", got, got.Inspect())
	}
	if got.Inspect() != "[[a, 0]]" {
		t.Fatalf("expected concrete generic return, got %s", got.Inspect())
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

func TestExecutorMapCopyShallowMerge(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{
			code: "val m = {:a: 1, \"b\": 2}\nval n = m copy {a: 3, \"c\": 4}\n[n[:a], n[\"b\"], n[\"c\"], m[:a], m[\"c\"]]",
			want: "[3, 2, 4, 1, nil]",
		},
		{
			code: "val m = {:name: \"sym\"}\nval n = m copy {\"name\": \"str\"}\n[n.name, n[:name], n[\"name\"]]",
			want: "[sym, sym, str]",
		},
	}

	for _, tt := range tests {
		got := runVM(t, tt.code)
		if got.Inspect() != tt.want {
			t.Fatalf("for %q expected %q got %q", tt.code, tt.want, got.Inspect())
		}
	}
}

func TestExecutorCopyRejectsUnsupportedSource(t *testing.T) {
	got := runVM(t, "42 copy {value: 1}")
	if got.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error object, got %T (%s)", got, got.Inspect())
	}
}

func TestExecutorCopyRejectsNonMapPayload(t *testing.T) {
	got := runVM(t, "val m = {:a: 1}\nm copy 2")
	if got.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error object, got %T (%s)", got, got.Inspect())
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

func TestExecutorListPrependAppendOperators(t *testing.T) {
	got := runVM(t, "val xs = [1, 2]\nval ys = 0 +: xs\nval zs = ys :+ 3\nzs")
	if got.Inspect() != "[0, 1, 2, 3]" {
		t.Fatalf("expected [0, 1, 2, 3], got %s", got.Inspect())
	}
}

func TestExecutorMatchLiteralAndWildcard(t *testing.T) {
	got := runVM(t, "match 2 {\n2 => { 20 }\n_ => { 30 }\n}")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "20" {
		t.Fatalf("expected 20, got %s", num.Value.String())
	}
}

func TestExecutorMatchWithGuard(t *testing.T) {
	got := runVM(t, "match 2 {\n2 if false => { 20 }\n2 if true => { 25 }\n_ => { 30 }\n}")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "25" {
		t.Fatalf("expected 25, got %s", num.Value.String())
	}
}

func TestExecutorMatchListPatternHeadTail(t *testing.T) {
	got := runVM(t, "val sum = fn(ns, acc = 0) { match ns { [h, ...t] => { sum(t, acc + h) } [] => { acc } } }\nsum([1,2,3])")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "6" {
		t.Fatalf("expected 6, got %s", num.Value.String())
	}
}

func TestExecutorSpawnAndAwait(t *testing.T) {
	got := runVM(t, "val t = spawn { 40 + 2 }\nt")
	if got.Type() != object.TASK_HANDLE_OBJ {
		t.Fatalf("expected task handle, got %T (%s)", got, got.Inspect())
	}
}

func TestExecutorSpawnWithoutAwait(t *testing.T) {
	got := runVM(t, "spawn { 1 }\n42")
	num, ok := got.(*object.Number)
	if !ok {
		t.Fatalf("expected *object.Number, got %T (%s)", got, got.Inspect())
	}
	if num.Value.String() != "42" {
		t.Fatalf("expected 42, got %s", num.Value.String())
	}
}

func TestExecutorDispatchPrefersSpecificStructTagDeterministically(t *testing.T) {
	input := `
val User = struct {
	name,
	age:num,
	active = true,
}

val u = User {
	name: "Slug",
	age: 42,
}

var f = fn(v:struct) { 'struct' }
var f = fn(v:User) { 'user' }

f(u)
`

	for i := 0; i < 100; i++ {
		got := runVM(t, input)
		if got.Inspect() != "user" {
			t.Fatalf("run %d: expected user dispatch, got %T (%s)", i, got, got.Inspect())
		}
	}
}

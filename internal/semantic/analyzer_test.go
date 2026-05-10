package semantic_test

import (
	"bytes"
	"slug/internal/ast"
	"slug/internal/lexer"
	"slug/internal/parser"
	"slug/internal/semantic"
	"strings"
	"testing"
)

func parseProgram(t *testing.T, input string) (*ast.Program, []string) {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	errs := p.Errors()
	errs = append(errs, semantic.Analyze("semantic-test.slug", input, program)...)
	return program, errs
}

func TestSemanticMarksTailCallFlags(t *testing.T) {
	input := `
val f = fn(x) {
  g(x)
}
`
	program, errs := parseProgram(t, input)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected expression statement, got %T", program.Statements[0])
	}
	valExpr, ok := exprStmt.Expression.(*ast.ValExpression)
	if !ok {
		t.Fatalf("expected val expression, got %T", exprStmt.Expression)
	}
	fn, ok := valExpr.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected function literal, got %T", valExpr.Value)
	}
	if !fn.HasTailCall {
		t.Fatal("expected function literal to be tagged with HasTailCall")
	}
	bodyExpr, ok := fn.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected function body expression statement, got %T", fn.Body.Statements[0])
	}
	call, ok := bodyExpr.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected tail call expression, got %T", bodyExpr.Expression)
	}
	if !call.IsTailCall {
		t.Fatal("expected final call expression to be tagged IsTailCall")
	}
}

func TestSemanticRejectsNonTailRecur(t *testing.T) {
	input := `
val f = fn(x) {
  recur(x) + 1
}
`
	_, errs := parseProgram(t, input)
	if len(errs) == 0 {
		t.Fatal("expected non-tail recur error, got none")
	}
	if !containsError(errs, "'recur' is only allowed in tail position") {
		t.Fatalf("expected recur tail-position error, got: %v", errs)
	}
}

func TestSemanticRejectsMainOnNonFunction(t *testing.T) {
	input := `
@main
val start = 1
`
	_, errs := parseProgram(t, input)
	if len(errs) == 0 {
		t.Fatal("expected @main validation error, got none")
	}
	if !containsError(errs, "@main may only annotate functions") {
		t.Fatalf("expected @main non-function error, got: %v", errs)
	}
}

func TestSemanticAllowsMainWithDefaultedParameters(t *testing.T) {
	input := `
@main
val start = fn(args = argv(), limit = cfg("limit", 10)) {
  nil
}
`
	_, errs := parseProgram(t, input)
	if len(errs) > 0 {
		t.Fatalf("expected no @main errors for defaulted params, got: %v", errs)
	}
}

func TestSemanticRejectsMainNonZeroArity(t *testing.T) {
	input := `
@main
val start = fn(required) {
  nil
}
`
	_, errs := parseProgram(t, input)
	if len(errs) == 0 {
		t.Fatal("expected @main arity error, got none")
	}
	if !containsError(errs, "@main function must be callable with zero arguments") {
		t.Fatalf("expected @main arity error, got: %v", errs)
	}
}

func TestSemanticRejectsMultipleMainDeclarations(t *testing.T) {
	input := `
@main
val one = fn() { nil }

@main
val two = fn() { nil }
`
	_, errs := parseProgram(t, input)
	if len(errs) == 0 {
		t.Fatal("expected duplicate @main error, got none")
	}
	if !containsError(errs, "at most one @main function") {
		t.Fatalf("expected @main uniqueness error, got: %v", errs)
	}
}

func TestSemanticRejectsStructSchemaOutsideBindingRHS(t *testing.T) {
	input := `
val make = fn() {
  return struct {
    name,
  }
}
`
	_, errs := parseProgram(t, input)
	if len(errs) == 0 {
		t.Fatal("expected struct schema placement error, got none")
	}
	if !containsError(errs, "struct schemas are only allowed on the right-hand side of val/var bindings") {
		t.Fatalf("expected struct schema placement error, got: %v", errs)
	}
}

func containsError(errs []string, want string) bool {
	for _, err := range errs {
		if strings.Contains(err, want) {
			return true
		}
	}
	return false
}

func TestSemanticTypeCheckCanBeDisabled(t *testing.T) {
	input := `
val x = "hello" - 1
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: false,
	})
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors when type check disabled, got: %v", errs)
	}
	if len(warns) > 0 {
		t.Fatalf("expected no type warnings when type check disabled, got: %v", warns)
	}
}

func TestSemanticTypeCheckReportsMismatchAsErrorByDefault(t *testing.T) {
	input := `
val x = "hello" - 1
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected semantic errors when type check enabled, got none")
	}
	if !containsError(errs, "numeric operator type mismatch") {
		t.Fatalf("expected numeric operator type mismatch error, got: %v", errs)
	}
}

func TestSemanticTypeCheckChecksStructFieldTagsInInit(t *testing.T) {
	input := `
val User = struct {
  @num age,
}

val u = User { age: "bad" }
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected struct field type mismatch error, got none")
	}
	if !containsError(errs, "struct field User.age") {
		t.Fatalf("expected struct field mismatch error, got: %v", errs)
	}
}

func TestSemanticTypeCheckTraceEmitsEvents(t *testing.T) {
	input := `
val f = fn(flag, x) {
  var acc = nil
  if (flag && isList(x)) {
    acc = x
  } else {
    acc = []
  }
  acc
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	var buf bytes.Buffer
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
		TypeCheckTrace:  true,
		TraceWriter:     &buf,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
	if buf.Len() == 0 {
		t.Fatal("expected trace output, got none")
	}
	if !strings.Contains(buf.String(), "TypeTrace:") {
		t.Fatalf("expected TypeTrace output, got: %s", buf.String())
	}
}

func TestSemanticTypeCheckChecksMatchPatternNarrowing(t *testing.T) {
	input := `
val xs = [1, 2]
val out = match xs {
  [h, ...t] => h - "bad"
  _ => 0
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected narrowed-pattern type mismatch error, got none")
	}
	if !containsError(errs, "numeric operator type mismatch") {
		t.Fatalf("expected numeric operator mismatch error from pattern narrowing, got: %v", errs)
	}
}

func TestSemanticTypeCheckHasCallArgumentMismatchMessage(t *testing.T) {
	input := `
val f = fn(@num n) { n + 1 }
val out = f("bad")
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected call argument type mismatch error, got none")
	}
	if !containsError(errs, "call argument type mismatch") {
		t.Fatalf("expected call argument type mismatch message, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsTaggedDefaultNil(t *testing.T) {
	input := `
val assertThrows = fn(@fn f, expected, @str msg = nil) {
  true
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for tagged default nil, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsIfBranchWithThrowElse(t *testing.T) {
	input := `
val Error = struct {
  @str type = "Error",
  @str msg,
}

val assert = fn(a, msg = nil) {
  if (a) {
    true
  } else {
    throw Error { type: "AssertionError", msg: "nope" }
  }
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for if/throw branch, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsCallsUsingDefaultArity(t *testing.T) {
	input := `
val g = fn(a, b = 2, c = 3) { a + b + c }
val ok1 = g(1)
val ok2 = g(1, 4)
val ok3 = g(1, 4, 5)
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for default-arity calls, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsMatchBodyFunctionWithHeterogeneousParams(t *testing.T) {
	input := `
val mapLike = fn(@list vs, @fn f, acc = []) match {
  [[], ...] => acc
  [[h, ...t], ...] => recur(t, f, acc :+ h /> f())
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for heterogeneous match-body params, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsGenericFnTagAcrossArities(t *testing.T) {
	input := `
val apply = fn(@fn f) { f() }
val v = apply(fn() { 1 })

val useMapLike = fn(@list xs, @fn f) {
  xs /> map(fn(v) { [v, f()] })
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for generic @fn constraints, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsMultiPatternLiteralAlternatives(t *testing.T) {
	input := `
val parseBoolLike = fn(v) {
  match v {
    "true", "t", "yes", "y", "1" => true
    _ => false
  }
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for multi-pattern literal alternatives, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsNestedMatchAfterTypeDispatch(t *testing.T) {
	input := `
val BOOLEAN_TYPE = :bool
val NUMBER_TYPE = :num
val NIL_TYPE = :nil
val STRING_TYPE = :str

val Error = struct {
  @str type = "Error",
  @str msg,
}

val toBoolean = fn(v) {
  match v /> type() {
    ^BOOLEAN_TYPE => v
    ^NUMBER_TYPE  => v == 1
    ^NIL_TYPE     => false
    ^STRING_TYPE  => {
      match v {
        "true", "t", "yes", "y", "TRUE", "T", "YES", "Y", "1" => true
        "false", "f", "no", "n", "FALSE", "F", "NO", "N", "0" => false
        _ => throw Error { type: "TypeError", msg: "Cannot convert '{{v}}' to boolean" }
      }
    }
    t => throw Error { type: "TypeError", msg: "Cannot convert type {{t}} to boolean" }
  }
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for nested match after type dispatch, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsListPatternMatchOnBytes(t *testing.T) {
	input := `
val list = 0x"0102"

val out = match list {
  [] => false
  [...] => true
  _ => false
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for bytes list-pattern match, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsHeterogeneousMapPatternArms(t *testing.T) {
	input := `
val f = fn(x) {
  match x {
    {} => "empty map"
    {"k":1} => "map with k == 1"
    {"k":k} if k == "a" => "map with " + k
    {"k":k} => "map with " + k
    {"k1", "k2":k, ...a} => "map with " + k + " '" + a + "'"
    {...} => "map with data"
    _ => {
      return "default"
    }
  }
}

val out = {"k":"v"} /> f
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for heterogeneous map pattern arms, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsMultipleMapValueShapesAcrossCalls(t *testing.T) {
	input := `
val f = fn(x) {
  match x {
    {} => "empty map"
    {"k":1} => "map with k == 1"
    {"k":k} if k == "a" => "map with " + k
    {"k":k} => "map with " + k
    {"k1", "k2":k, ...a} => "map with " + k + " '" + a + "'"
    {...} => "map with data"
    _ => {
      return "default"
    }
  }
}

val a = {"k":"v"} /> f
val b = {"k":1} /> f
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for multiple map value shapes across calls, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsFunctionTagOverloadSetCalls(t *testing.T) {
	input := `
var add = fn(@num b) { b + 255 }
var add = fn(@bool b) { !b }
var add = fn(@str b) { b + " slug" }
var add = fn(@list b) { b :+ 255 }
var add = fn(@map b) { b }
var add = fn(@bytes b) { b :+ 255 }
var add = fn(@fn b) { b() + 255 }

val a = "hello" /> add
val b = false /> add
val c = 1 /> add
val d = [] /> add
val e = {} /> add
val f = 0x"63ff00" /> add
val g = fn() { 12 } /> add
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for function-tag overload set calls, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsHeterogeneousMapLiteralValues(t *testing.T) {
	input := `
val a = {"name": "Alice", "age": 30}
val b = {"mix": {"a": [1, {"b": "c"}]}}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for heterogeneous map literal values, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsDynamicUntaggedStructField(t *testing.T) {
	input := `
val ParseResult = struct {
  value,
  @num nextIdx,
}

val a = ParseResult { value: "x", nextIdx: 1 }
val b = ParseResult { value: true, nextIdx: 2 }
val c = ParseResult { value: 123, nextIdx: 3 }
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for dynamic untagged struct field, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsIfUsedForSideEffectsWithMixedBranchValues(t *testing.T) {
	input := `
val f = fn(c) {
  var line = 0
  var lineStart = true
  if (c == "\n") {
    line = line + 1
  } else {
    lineStart = false
  }
  line
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for side-effect if branches, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsMatchWithHeterogeneousStructCaseResults(t *testing.T) {
	input := `
val SectionNode = struct {
	name,
}
val ParentNode = struct {
	name,
}
val frame = { "kind": "section", "name": "x" }
val node = match frame["kind"] {
	"section" => SectionNode { name: frame["name"] }
	"parent" => ParentNode { name: frame["name"] }
	_ => nil
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsSpreadToSatisfyFixedArity(t *testing.T) {
	input := `
val rgbStyle = fn(r, g, b) { [r, g, b] }
val vals = [1, 2, 3]
val out = rgbStyle(...vals)
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsDeferredBitwiseBytesModeInConcatFlow(t *testing.T) {
	input := `
val paddedKey = 0x"0102"
val ipad = 0x"0304"
val message = 0x"05"
val inner = (paddedKey ^ ipad) + message
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckTracksIfBranchUnionForCalls(t *testing.T) {
	input := `
val x = if (true) { 1 } else { "nope" }
val useNum = fn(@num n) { n + 1 }
val out = useNum(x)
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected strict type-check error from union<num|str> passed to @num")
	}
	if !containsError(errs, "call argument type mismatch") {
		t.Fatalf("expected call argument type mismatch, got: %v", errs)
	}
}

func TestSemanticTypeCheckTracksMatchCaseUnionForCalls(t *testing.T) {
	input := `
val x = match "s" {
	"s" => 1
	_ => true
}
val useNum = fn(@num n) { n + 1 }
val out = useNum(x)
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected strict type-check error from union<num|bool> passed to @num")
	}
	if !containsError(errs, "call argument type mismatch") {
		t.Fatalf("expected call argument type mismatch, got: %v", errs)
	}
}

func TestSemanticTypeCheckNarrowsIfTypeGuardTrueBranch(t *testing.T) {
	input := `
val f = fn(x) {
	if (type(x) == STRING_TYPE) {
		x + "!"
	} else {
		"ok"
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckNarrowsIfNilGuardElseBranch(t *testing.T) {
	input := `
val x = if (true) { nil } else { 1 }
val f = fn() {
	if (x == nil) {
		0
	} else {
		x + 1
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckNarrowsMatchGuardTypePredicate(t *testing.T) {
	input := `
val f = fn(x) {
	match x {
		v if type(v) == NUMBER_TYPE => v + 1
		_ => 0
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckNarrowsIfPredicateTrueBranch(t *testing.T) {
	input := `
val x = if (true) { [1,2] } else { {"a":1} }
val f = fn() {
	if (isList(x)) {
		x + [3]
	} else {
		x
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckNarrowsIfPredicateElseBranch(t *testing.T) {
	input := `
val x = if (true) { [1,2] } else { {"a":1} }
val f = fn() {
	if (isList(x)) {
		0
	} else {
		x["a"]
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckNarrowsMatchGuardPredicate(t *testing.T) {
	input := `
val f = fn(x) {
	match x {
		v if isBytes(v) => v & 255
		_ => 0
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckNarrowsLenGuardShape(t *testing.T) {
	input := `
val x = if (true) { [1,2] } else { {"a":1} }
val f = fn() {
	if (len(x) > 0) {
		len(x)
	} else {
		0
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckReportsUnreachableIfBranchOnContradictoryGuard(t *testing.T) {
	input := `
val x = if (true) { [1,2] } else { {"a":1} }
val y = if (isList(x) && isMap(x)) { 1 } else { 2 }
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected unreachable branch diagnostic, got none")
	}
	if !containsError(errs, "unreachable if-branch") {
		t.Fatalf("expected unreachable if-branch diagnostic, got: %v", errs)
	}
}

func TestSemanticTypeCheckReportsUnreachableMatchGuardOnContradiction(t *testing.T) {
	input := `
val f = fn(x) {
	match x {
		v if isBytes(v) && isMap(v) => 1
		_ => 0
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected unreachable match case diagnostic, got none")
	}
	if !containsError(errs, "unreachable match case") {
		t.Fatalf("expected unreachable match case diagnostic, got: %v", errs)
	}
}

func TestSemanticTypeCheckTracksVarReassignmentWidening(t *testing.T) {
	input := `
val f = fn() {
	var acc = nil
	acc = 1
	acc = acc + 2
	acc
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckTracksVarReassignmentAcrossIfBranches(t *testing.T) {
	input := `
val f = fn(flag) {
	var x = nil
	if (flag) {
		x = 1
	} else {
		x = 2
	}
	x + 1
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckStabilizesRecurParameterTypes(t *testing.T) {
	input := `
val f = fn(x, i = 0) {
	if (i >= 3) {
		x + 1
	} else {
		recur(x + 1, i + 1)
	}
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckSimplifiesUnionListFamilies(t *testing.T) {
	input := `
val x = if (true) { [1] } else { ["a"] }
val y = x :+ 2
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no strict type-check errors, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsStringInterpolationWithNumbers(t *testing.T) {
	input := `
val count = 10
val failed = 2
println("Test executed: {{count}} / failed {{failed}}")
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for string interpolation with numbers, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsChainedStringTimesNumberTimesNumber(t *testing.T) {
	input := `
val makeIndent = fn(@num spaces, @num depth) {
  " " * spaces * depth
}
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for chained string repetition, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsBytesAppendAndPrependOperators(t *testing.T) {
	input := `
val a = 0x"0102" :+ 3
val b = 0 +: 0x"0102"
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for bytes append/prepend operators, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsBytesBitwiseOperators(t *testing.T) {
	input := `
val a = 0x"ff00" & 0x"0ff0"
val b = 0x"ff00" | 0x"0ff0"
val c = 0x"ff00" ^ 0x"0ff0"
val d = 255 & 0x"0ff0"
val e = 0x"0ff0" ^ 255
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for bytes bitwise operators, got: %v", errs)
	}
}

func TestSemanticTypeCheckAllowsBytesBitNotPrefixOperator(t *testing.T) {
	input := `
val a = ~0x"00ff"
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in type-check enabled, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for bytes bit-not operator, got: %v", errs)
	}
}

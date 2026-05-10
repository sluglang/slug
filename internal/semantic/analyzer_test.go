package semantic_test

import (
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

func TestSemanticTypeCheckWarnsByDefaultForTypeMismatch(t *testing.T) {
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
		StrictTypeCheck: false,
	})
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors in warn mode, got: %v", errs)
	}
	if len(warns) == 0 {
		t.Fatal("expected type warnings, got none")
	}
	if !containsError(warns, "numeric operator type mismatch") {
		t.Fatalf("expected numeric operator type mismatch warning, got: %v", warns)
	}
}

func TestSemanticTypeCheckStrictPromotesMismatchToError(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected semantic errors in strict mode, got none")
	}
	if !containsError(errs, "numeric operator type mismatch") {
		t.Fatalf("expected numeric operator type mismatch error, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictChecksStructFieldTagsInInit(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected struct field type mismatch error, got none")
	}
	if !containsError(errs, "struct field User.age") {
		t.Fatalf("expected struct field mismatch error, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictChecksMatchPatternNarrowing(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected narrowed-pattern type mismatch error, got none")
	}
	if !containsError(errs, "numeric operator type mismatch") {
		t.Fatalf("expected numeric operator mismatch error from pattern narrowing, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictHasCallArgumentMismatchMessage(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) == 0 {
		t.Fatal("expected call argument type mismatch error, got none")
	}
	if !containsError(errs, "call argument type mismatch") {
		t.Fatalf("expected call argument type mismatch message, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsTaggedDefaultNil(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for tagged default nil, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsIfBranchWithThrowElse(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for if/throw branch, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsCallsUsingDefaultArity(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for default-arity calls, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsMatchBodyFunctionWithHeterogeneousParams(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for heterogeneous match-body params, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsGenericFnTagAcrossArities(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for generic @fn constraints, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsMultiPatternLiteralAlternatives(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for multi-pattern literal alternatives, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsNestedMatchAfterTypeDispatch(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for nested match after type dispatch, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsListPatternMatchOnBytes(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for bytes list-pattern match, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsBytesAppendAndPrependOperators(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for bytes append/prepend operators, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsBytesBitwiseOperators(t *testing.T) {
	input := `
val a = 0x"ff00" & 0x"0ff0"
val b = 0x"ff00" | 0x"0ff0"
val c = 0x"ff00" ^ 0x"0ff0"
`
	l := lexer.New(input)
	p := parser.New(l, "semantic-test.slug", input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	errs, warns := semantic.AnalyzeWithOptions("semantic-test.slug", input, program, semantic.AnalyzeOptions{
		EnableTypeCheck: true,
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for bytes bitwise operators, got: %v", errs)
	}
}

func TestSemanticTypeCheckStrictAllowsBytesBitNotPrefixOperator(t *testing.T) {
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
		StrictTypeCheck: true,
	})
	if len(warns) != 0 {
		t.Fatalf("expected no warnings in strict mode, got: %v", warns)
	}
	if len(errs) > 0 {
		t.Fatalf("expected no semantic errors for bytes bit-not operator, got: %v", errs)
	}
}

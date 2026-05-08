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

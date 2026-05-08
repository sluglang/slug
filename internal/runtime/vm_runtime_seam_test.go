package runtime

import (
	"slug/internal/ast"
	"slug/internal/dec64"
	"slug/internal/object"
	"slug/internal/token"
	"slug/internal/util"
	"testing"
)

func TestPrepareProgramForVMBindsForeignAndStripsDeclaration(t *testing.T) {
	rt := NewRuntime(util.Configuration{DefaultLimit: 4})
	foreignFn := &object.Foreign{Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object { return object.NIL }}
	rt.ForeignFunctions["test.mod.echo"] = foreignFn

	env := object.NewRootEnvironment(4)
	env.ModuleFqn = "test.mod"

	param := &ast.FunctionParameter{Name: &ast.Identifier{Token: token.Token{Literal: "value"}, Value: "value"}}
	foreignDecl := &ast.ForeignFunctionDeclaration{
		Name:       &ast.Identifier{Token: token.Token{Literal: "echo"}, Value: "echo"},
		Parameters: []*ast.FunctionParameter{param},
		Signature:  ast.FSig{Min: 1, Max: 1},
	}
	prog := &ast.Program{Statements: []ast.Statement{
		foreignDecl,
		&ast.ExpressionStatement{Expression: &ast.NumberLiteral{Token: token.Token{Literal: "1"}, Value: dec64.FromInt(1)}},
	}}

	prepared, err := prepareProgramForVM(rt, env, prog)
	if err != nil {
		t.Fatalf("prepareProgramForVM returned error: %v", err)
	}
	if len(prepared.Statements) != 1 {
		t.Fatalf("expected 1 non-foreign statement, got %d", len(prepared.Statements))
	}
	if _, ok := prepared.Statements[0].(*ast.ForeignFunctionDeclaration); ok {
		t.Fatal("foreign declaration should be stripped from prepared program")
	}

	bound, ok := env.Get("echo")
	if !ok {
		t.Fatal("expected foreign function binding in module env")
	}
	group, ok := bound.(*object.FunctionGroup)
	if !ok {
		t.Fatalf("expected echo binding to be function group, got %T", bound)
	}
	if len(group.Functions) != 1 {
		t.Fatalf("expected exactly one function in group, got %d", len(group.Functions))
	}
	found := false
	for _, fn := range group.Functions {
		if fn == foreignFn {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected function group to contain the registered foreign function")
	}
	if foreignFn.Name != "echo" {
		t.Fatalf("expected foreign name echo, got %q", foreignFn.Name)
	}
	if len(foreignFn.Parameters) != 1 || foreignFn.ParamIndex["value"] != 0 {
		t.Fatalf("expected foreign parameters/param index to be populated, got %#v %#v", foreignFn.Parameters, foreignFn.ParamIndex)
	}
}

func TestApplyForeignTagsForVMEvaluatesTagArgs(t *testing.T) {
	rt := NewRuntime(util.Configuration{DefaultLimit: 4})
	foreignFn := &object.Foreign{Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object { return object.NIL }}
	rt.ForeignFunctions["test.mod.fn"] = foreignFn

	env := object.NewRootEnvironment(4)
	env.ModuleFqn = "test.mod"
	if _, err := env.DefineConstant("x", &object.Number{Value: dec64.FromInt(41)}, false, false); err != nil {
		t.Fatalf("define constant x: %v", err)
	}

	foreignDecl := &ast.ForeignFunctionDeclaration{
		Name: &ast.Identifier{Token: token.Token{Literal: "fn"}, Value: "fn"},
		Tags: []*ast.Tag{{
			Name: "doc",
			Args: []ast.Expression{&ast.InfixExpression{
				Token:    token.Token{Literal: "+"},
				Left:     &ast.Identifier{Token: token.Token{Literal: "x"}, Value: "x"},
				Operator: "+",
				Right:    &ast.NumberLiteral{Token: token.Token{Literal: "1"}, Value: dec64.FromInt(1)},
			}},
		}},
	}
	prog := &ast.Program{Statements: []ast.Statement{foreignDecl}}

	if err := applyForeignTagsForVM(rt, env, prog); err != nil {
		t.Fatalf("applyForeignTagsForVM returned error: %v", err)
	}
	tag, ok := foreignFn.Tags["doc"]
	if !ok {
		t.Fatal("expected evaluated doc tag")
	}
	if len(tag.Elements) != 1 {
		t.Fatalf("expected one evaluated tag arg, got %d", len(tag.Elements))
	}
	num, ok := tag.Elements[0].(*object.Number)
	if !ok {
		t.Fatalf("expected number tag arg, got %T", tag.Elements[0])
	}
	if num.Value.String() != "42" {
		t.Fatalf("expected evaluated tag arg 42, got %s", num.Value.String())
	}
}

func TestMakeVMCallBridgeBindsNamedAndDefaultArgsForForeign(t *testing.T) {
	rt := NewRuntime(util.Configuration{DefaultLimit: 4})
	env := object.NewRootEnvironment(4)
	bridge := makeVMCallBridge(rt, env)

	foreignFn := &object.Foreign{
		Name: "sum",
		Parameters: []*ast.FunctionParameter{
			{Name: &ast.Identifier{Token: token.Token{Literal: "a"}, Value: "a"}},
			{Name: &ast.Identifier{Token: token.Token{Literal: "b"}, Value: "b"}, Default: &ast.NumberLiteral{Token: token.Token{Literal: "40"}, Value: dec64.FromInt(40)}},
		},
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			a := args[0].(*object.Number)
			b := args[1].(*object.Number)
			return &object.Number{Value: a.Value.Add(b.Value)}
		},
	}

	out := bridge(0, foreignFn, nil, map[string]object.Object{"a": &object.Number{Value: dec64.FromInt(2)}})
	num, ok := out.(*object.Number)
	if !ok {
		t.Fatalf("expected number result, got %T (%s)", out, out.Inspect())
	}
	if num.Value.String() != "42" {
		t.Fatalf("expected 42, got %s", num.Value.String())
	}
}

func TestMakeVMCallBridgeFlattensNamedVariadicListForForeign(t *testing.T) {
	rt := NewRuntime(util.Configuration{DefaultLimit: 4})
	env := object.NewRootEnvironment(4)
	bridge := makeVMCallBridge(rt, env)

	foreignFn := &object.Foreign{
		Name: "join",
		Parameters: []*ast.FunctionParameter{
			{Name: &ast.Identifier{Token: token.Token{Literal: "head"}, Value: "head"}},
			{Name: &ast.Identifier{Token: token.Token{Literal: "rest"}, Value: "rest"}, IsVariadic: true},
		},
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			if len(args) != 3 {
				return ctx.NewError("wrong arg count after variadic flatten: %d", len(args))
			}
			return &object.Number{Value: dec64.FromInt(len(args))}
		},
	}

	out := bridge(0, foreignFn, nil, map[string]object.Object{
		"head": &object.Number{Value: dec64.FromInt(1)},
		"rest": &object.List{Elements: []object.Object{
			&object.Number{Value: dec64.FromInt(2)},
			&object.Number{Value: dec64.FromInt(3)},
		}},
	})
	num, ok := out.(*object.Number)
	if !ok {
		t.Fatalf("expected number result, got %T (%s)", out, out.Inspect())
	}
	if num.Value.String() != "3" {
		t.Fatalf("expected flattened variadic arg count 3, got %s", num.Value.String())
	}
}

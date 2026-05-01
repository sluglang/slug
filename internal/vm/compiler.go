package vm

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/object"
)

func Compile(program *ast.Program) (*Chunk, error) {
	c := &compiler{chunk: &Chunk{}}
	if err := c.compileProgram(program); err != nil {
		return nil, err
	}
	c.emit(Instruction{Op: OpReturn})
	return c.chunk, nil
}

type compiler struct {
	chunk *Chunk
}

func (c *compiler) compileProgram(program *ast.Program) error {
	for _, stmt := range program.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *compiler) compileStatement(stmt ast.Statement) error {
	switch node := stmt.(type) {
	case *ast.ExpressionStatement:
		if err := c.compileExpression(node.Expression); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpPop, Position: node.Token.Position})
		return nil
	case *ast.ReturnStatement:
		if err := c.compileExpression(node.ReturnValue); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpReturn, Position: node.Token.Position})
		return nil
	default:
		return unsupportedNodeErr("statement", stmt)
	}
}

func (c *compiler) compileExpression(expr ast.Expression) error {
	switch node := expr.(type) {
	case *ast.NumberLiteral:
		idx := c.addConstant(&object.Number{Value: node.Value})
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
		return nil
	case *ast.StringLiteral:
		idx := c.addConstant(&object.String{Value: node.Value})
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
		return nil
	case *ast.BytesLiteral:
		idx := c.addConstant(&object.Bytes{Value: node.Value})
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
		return nil
	case *ast.SymbolLiteral:
		idx := c.addConstant(object.InternSymbol(node.Value))
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
		return nil
	case *ast.Boolean:
		op := OpFalse
		if node.Value {
			op = OpTrue
		}
		c.emit(Instruction{Op: op, Position: node.Token.Position})
		return nil
	case *ast.Nil:
		c.emit(Instruction{Op: OpNil, Position: node.Token.Position})
		return nil
	case *ast.Identifier:
		c.emit(Instruction{Op: OpGetGlobal, StrArg: node.Value, Position: node.Token.Position})
		return nil
	case *ast.ValExpression:
		name, err := patternName(node.Pattern)
		if err != nil {
			return fmt.Errorf("vm compile error at %d: %w", node.Token.Position, err)
		}
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpSetGlobalConst, StrArg: name, Position: node.Token.Position})
		return nil
	case *ast.VarExpression:
		name, err := patternName(node.Pattern)
		if err != nil {
			return fmt.Errorf("vm compile error at %d: %w", node.Token.Position, err)
		}
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpSetGlobalVar, StrArg: name, Position: node.Token.Position})
		return nil
	case *ast.PrefixExpression:
		if err := c.compileExpression(node.Right); err != nil {
			return err
		}
		switch node.Operator {
		case "!":
			c.emit(Instruction{Op: OpBang, Position: node.Token.Position})
		case "-":
			c.emit(Instruction{Op: OpNegate, Position: node.Token.Position})
		default:
			return fmt.Errorf("vm compile error at %d: unsupported prefix operator %q", node.Token.Position, node.Operator)
		}
		return nil
	case *ast.InfixExpression:
		if node.Operator == "=" {
			ident, ok := node.Left.(*ast.Identifier)
			if !ok {
				return fmt.Errorf("vm compile error at %d: left side of assignment must be an identifier", node.Token.Position)
			}
			if err := c.compileExpression(node.Right); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpAssignGlobal, StrArg: ident.Value, Position: node.Token.Position})
			return nil
		}
		if node.Operator == "&&" || node.Operator == "||" {
			return c.compileShortCircuit(node)
		}
		if err := c.compileExpression(node.Left); err != nil {
			return err
		}
		if err := c.compileExpression(node.Right); err != nil {
			return err
		}
		switch node.Operator {
		case "+":
			c.emit(Instruction{Op: OpAdd, Position: node.Token.Position})
		case "-":
			c.emit(Instruction{Op: OpSub, Position: node.Token.Position})
		case "*":
			c.emit(Instruction{Op: OpMul, Position: node.Token.Position})
		case "/":
			c.emit(Instruction{Op: OpDiv, Position: node.Token.Position})
		case "==":
			c.emit(Instruction{Op: OpEqual, Position: node.Token.Position})
		case "!=":
			c.emit(Instruction{Op: OpNotEqual, Position: node.Token.Position})
		case ">":
			c.emit(Instruction{Op: OpGreaterThan, Position: node.Token.Position})
		case "<":
			c.emit(Instruction{Op: OpLessThan, Position: node.Token.Position})
		default:
			return fmt.Errorf("vm compile error at %d: unsupported infix operator %q", node.Token.Position, node.Operator)
		}
		return nil
	case *ast.BlockStatement:
		return c.compileBlock(node)
	case *ast.IfExpression:
		if err := c.compileExpression(node.Condition); err != nil {
			return err
		}
		jumpIfFalseIdx := c.emit(Instruction{Op: OpJumpIfFalse, Position: node.Token.Position})
		if err := c.compileBlock(node.ThenBranch); err != nil {
			return err
		}
		jumpToEndIdx := c.emit(Instruction{Op: OpJump, Position: node.Token.Position})
		c.patchJump(jumpIfFalseIdx, len(c.chunk.Instructions))
		if node.ElseBranch != nil {
			if err := c.compileBlock(node.ElseBranch); err != nil {
				return err
			}
		} else {
			c.emit(Instruction{Op: OpNil, Position: node.Token.Position})
		}
		c.patchJump(jumpToEndIdx, len(c.chunk.Instructions))
		return nil
	case *ast.ListLiteral:
		for _, el := range node.Elements {
			if err := c.compileExpression(el); err != nil {
				return err
			}
		}
		c.emit(Instruction{Op: OpArray, IntArg: len(node.Elements), Position: node.Token.Position})
		return nil
	case *ast.MapLiteral:
		pairCount := 0
		for key, value := range node.Pairs {
			if err := c.compileExpression(key); err != nil {
				return err
			}
			if err := c.compileExpression(value); err != nil {
				return err
			}
			pairCount++
		}
		c.emit(Instruction{Op: OpHash, IntArg: pairCount, Position: node.Token.Position})
		return nil
	case *ast.IndexExpression:
		if err := c.compileExpression(node.Left); err != nil {
			return err
		}
		if err := c.compileExpression(node.Index); err != nil {
			return err
		}
		op := OpIndex
		if node.IsDotLookup {
			op = OpIndexDot
		}
		c.emit(Instruction{Op: op, Position: node.Token.Position})
		return nil
	case *ast.SliceExpression:
		if err := c.compileMaybeExpression(node.Start); err != nil {
			return err
		}
		if err := c.compileMaybeExpression(node.End); err != nil {
			return err
		}
		if err := c.compileMaybeExpression(node.Step); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpSlice, Position: node.Token.Position})
		return nil
	case *ast.FunctionLiteral:
		fnObj, err := c.compileFunctionLiteral(node)
		if err != nil {
			return err
		}
		idx := c.addConstant(fnObj)
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
		return nil
	case *ast.CallExpression:
		if err := c.compileExpression(node.Function); err != nil {
			return err
		}
		plan := make([]CallArgSpec, 0, len(node.Arguments))
		sawNamed := false
		for _, arg := range node.Arguments {
			switch a := arg.(type) {
			case *ast.NamedArgument:
				sawNamed = true
				if err := c.compileExpression(a.Value); err != nil {
					return err
				}
				plan = append(plan, CallArgSpec{Kind: CallArgNamed, Name: a.Name.Value})
			case *ast.SpreadExpression:
				if sawNamed {
					return fmt.Errorf("vm compile error at %d: positional arguments must appear before named arguments", a.Token.Position)
				}
				if err := c.compileExpression(a.Value); err != nil {
					return err
				}
				plan = append(plan, CallArgSpec{Kind: CallArgSpread})
			default:
				if sawNamed {
					return fmt.Errorf("vm compile error at %d: positional arguments must appear before named arguments", node.Token.Position)
				}
				if err := c.compileExpression(arg); err != nil {
					return err
				}
				plan = append(plan, CallArgSpec{Kind: CallArgPositional})
			}
		}
		c.emit(Instruction{
			Op:       OpCall,
			IntArg:   len(node.Arguments),
			CallPlan: plan,
			Position: node.Token.Position,
		})
		return nil
	default:
		return unsupportedNodeErr("expression", expr)
	}
}

func (c *compiler) compileMaybeExpression(expr ast.Expression) error {
	if expr == nil {
		c.emit(Instruction{Op: OpNil})
		return nil
	}
	return c.compileExpression(expr)
}

func (c *compiler) compileBlock(block *ast.BlockStatement) error {
	if block == nil {
		c.emit(Instruction{Op: OpNil})
		return nil
	}
	if len(block.Statements) == 0 {
		c.emit(Instruction{Op: OpNil, Position: block.Token.Position})
		return nil
	}
	for i, stmt := range block.Statements {
		isLast := i == len(block.Statements)-1
		exprStmt, ok := stmt.(*ast.ExpressionStatement)
		if ok {
			if err := c.compileExpression(exprStmt.Expression); err != nil {
				return err
			}
			if !isLast {
				c.emit(Instruction{Op: OpPop, Position: exprStmt.Token.Position})
			}
			continue
		}
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
		if !isLast {
			c.emit(Instruction{Op: OpPop})
		}
	}
	return nil
}

func (c *compiler) addConstant(obj object.Object) int {
	c.chunk.Constants = append(c.chunk.Constants, obj)
	return len(c.chunk.Constants) - 1
}

func (c *compiler) emit(ins Instruction) int {
	c.chunk.Instructions = append(c.chunk.Instructions, ins)
	return len(c.chunk.Instructions) - 1
}

func (c *compiler) patchJump(at int, target int) {
	c.chunk.Instructions[at].IntArg = target
}

func patternName(pattern ast.MatchPattern) (string, error) {
	switch p := pattern.(type) {
	case *ast.IdentifierPattern:
		return p.Value.Value, nil
	case *ast.BindingPattern:
		if p.Name == nil {
			return "", fmt.Errorf("binding pattern missing binding name")
		}
		return p.Name.Value, nil
	default:
		return "", fmt.Errorf("unsupported binding pattern %T", pattern)
	}
}

func unsupportedNodeErr(kind string, node ast.Node) error {
	return fmt.Errorf("vm compile error: unsupported %s node %T", kind, node)
}

func (c *compiler) compileShortCircuit(node *ast.InfixExpression) error {
	if err := c.compileExpression(node.Left); err != nil {
		return err
	}
	if node.Operator == "&&" {
		jumpFalse := c.emit(Instruction{Op: OpJumpIfFalse, Position: node.Token.Position})
		c.emit(Instruction{Op: OpPop, Position: node.Token.Position})
		if err := c.compileExpression(node.Right); err != nil {
			return err
		}
		jumpEnd := c.emit(Instruction{Op: OpJump, Position: node.Token.Position})
		c.patchJump(jumpFalse, len(c.chunk.Instructions))
		c.emit(Instruction{Op: OpFalse, Position: node.Token.Position})
		c.patchJump(jumpEnd, len(c.chunk.Instructions))
		return nil
	}

	// ||
	jumpFalse := c.emit(Instruction{Op: OpJumpIfFalse, Position: node.Token.Position})
	c.emit(Instruction{Op: OpTrue, Position: node.Token.Position})
	jumpEnd := c.emit(Instruction{Op: OpJump, Position: node.Token.Position})
	c.patchJump(jumpFalse, len(c.chunk.Instructions))
	if err := c.compileExpression(node.Right); err != nil {
		return err
	}
	c.patchJump(jumpEnd, len(c.chunk.Instructions))
	return nil
}

func (c *compiler) compileFunctionLiteral(node *ast.FunctionLiteral) (*VMFunction, error) {
	params := make([]string, 0, len(node.Parameters))
	for _, p := range node.Parameters {
		if len(p.Tags) > 0 || p.Default != nil || p.IsVariadic {
			return nil, fmt.Errorf("vm compile error at %d: tagged/default/variadic params are not yet supported", node.Token.Position)
		}
		params = append(params, p.Name.Value)
	}

	child := &compiler{chunk: &Chunk{}}
	if err := child.compileBlock(node.Body); err != nil {
		return nil, err
	}
	child.emit(Instruction{Op: OpReturn, Position: node.Token.Position})

	return &VMFunction{
		Params: params,
		Chunk:  child.chunk,
	}, nil
}

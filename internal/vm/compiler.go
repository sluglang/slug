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
			return fmt.Errorf("vm compile error at %d: assignment is not yet supported by vm backend", node.Token.Position)
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
	default:
		return unsupportedNodeErr("expression", expr)
	}
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

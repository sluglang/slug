package vm

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/dec64"
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
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		if err := c.compileBindPattern(node.Pattern, true, node.Token.Position); err != nil {
			return err
		}
		return nil
	case *ast.VarExpression:
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		if err := c.compileBindPattern(node.Pattern, false, node.Token.Position); err != nil {
			return err
		}
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
		case "~":
			c.emit(Instruction{Op: OpBitNot, Position: node.Token.Position})
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
		case "+:":
			c.emit(Instruction{Op: OpListPrepend, Position: node.Token.Position})
		case ":+":
			c.emit(Instruction{Op: OpListAppend, Position: node.Token.Position})
		case "-":
			c.emit(Instruction{Op: OpSub, Position: node.Token.Position})
		case "*":
			c.emit(Instruction{Op: OpMul, Position: node.Token.Position})
		case "/":
			c.emit(Instruction{Op: OpDiv, Position: node.Token.Position})
		case "%":
			c.emit(Instruction{Op: OpMod, Position: node.Token.Position})
		case "==":
			c.emit(Instruction{Op: OpEqual, Position: node.Token.Position})
		case "!=":
			c.emit(Instruction{Op: OpNotEqual, Position: node.Token.Position})
		case ">":
			c.emit(Instruction{Op: OpGreaterThan, Position: node.Token.Position})
		case ">=":
			c.emit(Instruction{Op: OpGreaterThanEqual, Position: node.Token.Position})
		case "<":
			c.emit(Instruction{Op: OpLessThan, Position: node.Token.Position})
		case "<=":
			c.emit(Instruction{Op: OpLessThanEqual, Position: node.Token.Position})
		case "&":
			c.emit(Instruction{Op: OpBitAnd, Position: node.Token.Position})
		case "|":
			c.emit(Instruction{Op: OpBitOr, Position: node.Token.Position})
		case "^":
			c.emit(Instruction{Op: OpBitXor, Position: node.Token.Position})
		case "<<":
			c.emit(Instruction{Op: OpShiftLeft, Position: node.Token.Position})
		case ">>":
			c.emit(Instruction{Op: OpShiftRight, Position: node.Token.Position})
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
	case *ast.MatchExpression:
		return c.compileMatchExpression(node)
	case *ast.SpawnExpression:
		fnObj, err := c.compileSpawnBody(node)
		if err != nil {
			return err
		}
		idx := c.addConstant(fnObj)
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
		c.emit(Instruction{Op: OpSpawn, Position: node.Token.Position})
		return nil
	case *ast.AwaitExpression:
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpAwait, Position: node.Token.Position})
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
	case *ast.RecurExpression:
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
			Op:       OpRecur,
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
	c.emit(Instruction{Op: OpPushScope, Position: block.Token.Position})
	if len(block.Statements) == 0 {
		c.emit(Instruction{Op: OpNil, Position: block.Token.Position})
		c.emit(Instruction{Op: OpPopScope, Position: block.Token.Position})
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
	c.emit(Instruction{Op: OpPopScope, Position: block.Token.Position})
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

func (c *compiler) compileBindPattern(pattern ast.MatchPattern, isConst bool, pos int) error {
	switch p := pattern.(type) {
	case *ast.IdentifierPattern:
		op := OpSetGlobalVar
		if isConst {
			op = OpSetGlobalConst
		}
		c.emit(Instruction{Op: op, StrArg: p.Value.Value, Position: pos})
		return nil
	case *ast.BindingPattern:
		if p.Name == nil {
			return fmt.Errorf("vm compile error at %d: binding pattern missing name", pos)
		}
		op := OpSetGlobalVar
		if isConst {
			op = OpSetGlobalConst
		}
		c.emit(Instruction{Op: OpDup, Position: pos})
		c.emit(Instruction{Op: op, StrArg: p.Name.Value, Position: pos})
		c.emit(Instruction{Op: OpPop, Position: pos})
		return c.compileBindPattern(p.Pattern, isConst, pos)
	case *ast.MapPattern:
		if p.Exact {
			return fmt.Errorf("vm compile error at %d: exact map patterns are not yet supported", pos)
		}
		if p.Spread != nil {
			sp, ok := p.Spread.(*ast.SpreadPattern)
			if !ok || sp.Value != nil {
				return fmt.Errorf("vm compile error at %d: spread map binding capture is not yet supported", pos)
			}
		}
		if p.SelectAll {
			op := OpBindMapAllVar
			if isConst {
				op = OpBindMapAllConst
			}
			c.emit(Instruction{Op: op, Position: pos})
		}
		for _, entry := range p.Pairs {
			if entry.Pattern == nil {
				continue
			}
			c.emit(Instruction{Op: OpDup, Position: pos})
			if err := c.compileMapPatternKey(entry.Key, pos); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpIndex, Position: pos})
			if err := c.compileBindPattern(entry.Pattern, isConst, pos); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpPop, Position: pos})
		}
		return nil
	case *ast.ListPattern:
		spreadIndex := -1
		for i, el := range p.Elements {
			if sp, ok := el.(*ast.SpreadPattern); ok {
				if i != len(p.Elements)-1 {
					return fmt.Errorf("vm compile error at %d: spread (...) must be final element in list binding pattern", pos)
				}
				spreadIndex = i
				if sp.Value == nil {
					break
				}
				c.emit(Instruction{Op: OpDup, Position: pos})
				c.emitNumberConstant(pos, i)
				c.emit(Instruction{Op: OpNil, Position: pos})
				c.emit(Instruction{Op: OpNil, Position: pos})
				c.emit(Instruction{Op: OpSlice, Position: pos})
				if err := c.compileBindPattern(&ast.IdentifierPattern{Value: sp.Value}, isConst, pos); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpPop, Position: pos})
				break
			}

			switch pt := el.(type) {
			case *ast.WildcardPattern:
				continue
			default:
				c.emit(Instruction{Op: OpDup, Position: pos})
				c.emitNumberConstant(pos, i)
				c.emit(Instruction{Op: OpIndex, Position: pos})
				if err := c.compileBindPattern(pt, isConst, pos); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpPop, Position: pos})
			}
		}
		if spreadIndex == -1 {
			// no spread: nothing extra to bind
		}
		return nil
	default:
		return fmt.Errorf("vm compile error at %d: unsupported binding pattern %T", pos, pattern)
	}
}

func (c *compiler) compileMapPatternKey(key ast.Expression, pos int) error {
	switch k := key.(type) {
	case *ast.Identifier:
		idx := c.addConstant(object.InternSymbol(k.Value))
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: pos})
		return nil
	case *ast.SymbolLiteral:
		idx := c.addConstant(object.InternSymbol(k.Value))
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: pos})
		return nil
	case *ast.StringLiteral:
		idx := c.addConstant(&object.String{Value: k.Value})
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: pos})
		return nil
	default:
		return c.compileExpression(key)
	}
}

func (c *compiler) emitNumberConstant(pos int, n int) {
	idx := c.addConstant(&object.Number{Value: dec64.FromInt(n)})
	c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: pos})
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
	params := make([]VMParam, 0, len(node.Parameters))
	for _, p := range node.Parameters {
		if len(p.Tags) > 0 {
			return nil, fmt.Errorf("vm compile error at %d: tagged params are not yet supported", node.Token.Position)
		}
		var defaultChunk *Chunk
		if p.Default != nil {
			dc := &compiler{chunk: &Chunk{}}
			if err := dc.compileExpression(p.Default); err != nil {
				return nil, err
			}
			dc.emit(Instruction{Op: OpReturn, Position: node.Token.Position})
			defaultChunk = dc.chunk
		}
		params = append(params, VMParam{
			Name:       p.Name.Value,
			IsVariadic: p.IsVariadic,
			Default:    defaultChunk,
		})
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

func (c *compiler) compileSpawnBody(node *ast.SpawnExpression) (*VMFunction, error) {
	child := &compiler{chunk: &Chunk{}}
	if err := child.compileExpression(node.Body); err != nil {
		return nil, err
	}
	child.emit(Instruction{Op: OpReturn, Position: node.Token.Position})

	return &VMFunction{
		Params: []VMParam{},
		Chunk:  child.chunk,
	}, nil
}

func (c *compiler) compileMatchExpression(node *ast.MatchExpression) error {
	if node.Value == nil {
		return fmt.Errorf("vm compile error at %d: pipeline-style match without explicit value is not yet supported", node.Token.Position)
	}
	if len(node.Cases) == 0 {
		return fmt.Errorf("vm compile error at %d: match requires at least one case", node.Token.Position)
	}
	if err := c.compileExpression(node.Value); err != nil {
		return err
	}

	endJumps := make([]int, 0, len(node.Cases))

	for _, cs := range node.Cases {
		pat := cs.Pattern

		// wildcard case: consume scrutinee and execute body
		if _, ok := pat.(*ast.WildcardPattern); ok {
			if cs.Guard != nil {
				if err := c.compileExpression(cs.Guard); err != nil {
					return err
				}
				guardJump := c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position})
				c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
				if err := c.compileExpression(cs.Body); err != nil {
					return err
				}
				j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
				endJumps = append(endJumps, j)
				c.patchJump(guardJump, len(c.chunk.Instructions))
				continue
			}
			c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			if err := c.compileExpression(cs.Body); err != nil {
				return err
			}
			j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
			endJumps = append(endJumps, j)
			continue
		}

		if idp, ok := pat.(*ast.IdentifierPattern); ok {
			c.emit(Instruction{Op: OpPushScope, Position: cs.Token.Position})
			c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
			if err := c.compileBindPattern(idp, true, cs.Token.Position); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			if cs.Guard != nil {
				if err := c.compileExpression(cs.Guard); err != nil {
					return err
				}
				guardJump := c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position})
				c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
				if err := c.compileExpression(cs.Body); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpPopScope, Position: cs.Token.Position})
				j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
				endJumps = append(endJumps, j)
				c.patchJump(guardJump, len(c.chunk.Instructions))
				c.emit(Instruction{Op: OpPopScope, Position: cs.Token.Position})
				continue
			}

			c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			if err := c.compileExpression(cs.Body); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpPopScope, Position: cs.Token.Position})
			j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
			endJumps = append(endJumps, j)
			continue
		}

		if lp, ok := pat.(*ast.ListPattern); ok {
			spreadIndex := -1
			var spreadPat *ast.SpreadPattern
			for i, el := range lp.Elements {
				if sp, ok := el.(*ast.SpreadPattern); ok {
					if spreadIndex != -1 {
						return fmt.Errorf("vm compile error at %d: list pattern supports only one spread", cs.Token.Position)
					}
					if i != len(lp.Elements)-1 {
						return fmt.Errorf("vm compile error at %d: spread (...) must be final element in list pattern", cs.Token.Position)
					}
					spreadIndex = i
					spreadPat = sp
				}
			}
			c.emit(Instruction{Op: OpPushScope, Position: cs.Token.Position})
			c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
			if spreadIndex >= 0 {
				c.emit(Instruction{Op: OpMatchSeqLenGte, IntArg: spreadIndex, Position: cs.Token.Position})
			} else {
				c.emit(Instruction{Op: OpMatchSeqLenEq, IntArg: len(lp.Elements), Position: cs.Token.Position})
			}
			failJumps := []int{c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position})}

			for i, el := range lp.Elements {
				if spreadIndex >= 0 && i == spreadIndex {
					if spreadPat != nil && spreadPat.Value != nil {
						c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
						c.emit(Instruction{Op: OpMatchSeqTail, IntArg: spreadIndex, Position: cs.Token.Position})
						if err := c.compileBindPattern(&ast.IdentifierPattern{Value: spreadPat.Value}, true, cs.Token.Position); err != nil {
							return err
						}
						c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
					}
					break
				}

				if _, ok := el.(*ast.WildcardPattern); ok {
					continue
				}

				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				c.emitNumberConstant(cs.Token.Position, i)
				c.emit(Instruction{Op: OpIndex, Position: cs.Token.Position})
				switch ep := el.(type) {
				case *ast.IdentifierPattern:
					if err := c.compileBindPattern(ep, true, cs.Token.Position); err != nil {
						return err
					}
					c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
				case *ast.LiteralPattern:
					if err := c.compileExpression(ep.Value); err != nil {
						return err
					}
					c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
					failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))
				default:
					return fmt.Errorf("vm compile error at %d: unsupported list pattern shape", cs.Token.Position)
				}
			}

			if cs.Guard != nil {
				if err := c.compileExpression(cs.Guard); err != nil {
					return err
				}
				failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))
			}

			c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			if err := c.compileExpression(cs.Body); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpPopScope, Position: cs.Token.Position})
			j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
			endJumps = append(endJumps, j)

			failTarget := len(c.chunk.Instructions)
			for _, fj := range failJumps {
				c.patchJump(fj, failTarget)
			}
			c.emit(Instruction{Op: OpPopScope, Position: cs.Token.Position})
			continue
		}

		// Compare scrutinee with literal-like pattern
		c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
		if err := c.compilePatternValue(pat); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
		nextCaseJump := c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position})

		if cs.Guard != nil {
			if err := c.compileExpression(cs.Guard); err != nil {
				return err
			}
			guardJump := c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position})
			c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			if err := c.compileExpression(cs.Body); err != nil {
				return err
			}
			j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
			endJumps = append(endJumps, j)
			c.patchJump(guardJump, len(c.chunk.Instructions))
			c.patchJump(nextCaseJump, len(c.chunk.Instructions))
			continue
		}

		// Matched: consume scrutinee and run body
		c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
		if err := c.compileExpression(cs.Body); err != nil {
			return err
		}
		j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
		endJumps = append(endJumps, j)
		c.patchJump(nextCaseJump, len(c.chunk.Instructions))
	}

	// No case matched: consume scrutinee and produce nil.
	c.emit(Instruction{Op: OpPop, Position: node.Token.Position})
	c.emit(Instruction{Op: OpNil, Position: node.Token.Position})

	end := len(c.chunk.Instructions)
	for _, j := range endJumps {
		c.patchJump(j, end)
	}
	return nil
}

func (c *compiler) compilePatternValue(pattern ast.MatchPattern) error {
	switch p := pattern.(type) {
	case *ast.LiteralPattern:
		return c.compileExpression(p.Value)
	default:
		return fmt.Errorf("vm compile error: unsupported match pattern %T", pattern)
	}
}

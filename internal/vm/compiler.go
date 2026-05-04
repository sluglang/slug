package vm

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/dec64"
	"slug/internal/object"
	"strings"
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
	case *ast.DeferStatement:
		return c.compileDeferStatement(node)
	case *ast.ThrowStatement:
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpThrow, Position: node.Token.Position})
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
		if len(node.Tags) > 0 {
			c.emit(Instruction{Op: OpApplyTags, StrArg: encodeTagNames(node.Tags), Position: node.Token.Position})
		}
		if err := c.compileBindPattern(node.Pattern, true, node.Token.Position); err != nil {
			return err
		}
		if node.HasDoc {
			if idp, ok := node.Pattern.(*ast.IdentifierPattern); ok {
				c.emit(Instruction{Op: OpSetDoc, StrArg: idp.Value.Value, StrArg2: node.Doc, Position: node.Token.Position})
			}
		}
		return nil
	case *ast.VarExpression:
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		if len(node.Tags) > 0 {
			c.emit(Instruction{Op: OpApplyTags, StrArg: encodeTagNames(node.Tags), Position: node.Token.Position})
		}
		if err := c.compileBindPattern(node.Pattern, false, node.Token.Position); err != nil {
			return err
		}
		if node.HasDoc {
			if idp, ok := node.Pattern.(*ast.IdentifierPattern); ok {
				c.emit(Instruction{Op: OpSetDoc, StrArg: idp.Value.Value, StrArg2: node.Doc, Position: node.Token.Position})
			}
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
	case *ast.SelectExpression:
		return c.compileSelectExpression(node)
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
	case *ast.StructSchemaExpression:
		schema := &object.StructSchema{
			Fields:     make([]object.StructSchemaField, 0, len(node.Fields)),
			FieldIndex: make(map[string]int, len(node.Fields)),
		}
		for _, field := range node.Fields {
			if _, exists := schema.FieldIndex[field.Name]; exists {
				return fmt.Errorf("vm compile error at %d: duplicate struct field: %s", field.Token.Position, field.Name)
			}
			schema.FieldIndex[field.Name] = len(schema.Fields)
			schema.Fields = append(schema.Fields, object.StructSchemaField{
				Name:    field.Name,
				Default: field.Default,
				Tags:    field.Tags,
			})
		}
		idx := c.addConstant(schema)
		c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
		c.emit(Instruction{Op: OpStructSchema, Position: node.Token.Position})
		return nil
	case *ast.StructInitExpression:
		if err := c.compileExpression(node.Schema); err != nil {
			return err
		}
		for _, field := range node.Fields {
			idx := c.addConstant(object.InternSymbol(field.Name))
			c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: field.Token.Position})
			if err := c.compileExpression(field.Value); err != nil {
				return err
			}
		}
		c.emit(Instruction{Op: OpHash, IntArg: len(node.Fields), Position: node.Token.Position})
		c.emit(Instruction{Op: OpStructInit, Position: node.Token.Position})
		return nil
	case *ast.StructCopyExpression:
		if err := c.compileExpression(node.Source); err != nil {
			return err
		}
		if err := c.compileExpression(node.Fields); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpStructCopy, Position: node.Token.Position})
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
	scopeLimit := 0
	if block.IsNursery {
		// -1 means "use inherited/default limit at runtime".
		scopeLimit = -1
		if block.Limit != nil {
			n, ok := block.Limit.(*ast.NumberLiteral)
			if !ok {
				return fmt.Errorf("vm compile error at %d: nursery limit must be a numeric literal", block.Token.Position)
			}
			limit := int(n.Value.ToInt64())
			if limit < 1 {
				return fmt.Errorf("vm compile error at %d: nursery limit must be >= 1", block.Token.Position)
			}
			scopeLimit = limit
		}
	}
	c.emit(Instruction{Op: OpPushScope, IntArg: scopeLimit, Position: block.Token.Position})
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

func (c *compiler) compileDeferStatement(node *ast.DeferStatement) error {
	fnObj, err := c.compileDeferBody(node)
	if err != nil {
		return err
	}
	idx := c.addConstant(fnObj)
	c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
	errName := ""
	if node.ErrorName != nil {
		errName = node.ErrorName.Value
	}
	c.emit(Instruction{
		Op:       OpDefer,
		IntArg:   int(node.Mode),
		StrArg:   errName,
		Position: node.Token.Position,
	})
	c.emit(Instruction{Op: OpNil, Position: node.Token.Position})
	return nil
}

func (c *compiler) compileDeferBody(node *ast.DeferStatement) (*VMFunction, error) {
	child := &compiler{chunk: &Chunk{}}
	if node.Call == nil {
		child.emit(Instruction{Op: OpNil, Position: node.Token.Position})
	} else {
		switch s := node.Call.(type) {
		case *ast.ExpressionStatement:
			if err := child.compileExpression(s.Expression); err != nil {
				return nil, err
			}
		case *ast.BlockStatement:
			if err := child.compileBlock(s); err != nil {
				return nil, err
			}
		default:
			if err := child.compileStatement(s); err != nil {
				return nil, err
			}
		}
	}
	child.emit(Instruction{Op: OpReturn, Position: node.Token.Position})
	return &VMFunction{
		Params: []VMParam{},
		Chunk:  child.chunk,
	}, nil
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
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil || field.Pattern == nil {
				continue
			}
			c.emit(Instruction{Op: OpDup, Position: pos})
			symIdx := c.addConstant(object.InternSymbol(field.Name))
			c.emit(Instruction{Op: OpConstant, IntArg: symIdx, Position: pos})
			c.emit(Instruction{Op: OpIndexDot, Position: pos})
			if err := c.compileBindPattern(field.Pattern, isConst, pos); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpPop, Position: pos})
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

func mapPatternKeyName(key ast.Expression) string {
	switch k := key.(type) {
	case *ast.Identifier:
		return k.Value
	case *ast.SymbolLiteral:
		return k.Value
	case *ast.StringLiteral:
		return k.Value
	default:
		return ""
	}
}

func encodeTagNames(tags []*ast.Tag) string {
	parts := make([]string, 0, len(tags))
	for _, t := range tags {
		if t == nil || t.Name == "" {
			continue
		}
		parts = append(parts, t.Name)
	}
	return strings.Join(parts, ",")
}

func (c *compiler) compileListPatternFromValue(p *ast.ListPattern, pos int, failJumps *[]int) error {
	localFailJumps := make([]int, 0, 8)

	spreadIndex := -1
	var spreadPat *ast.SpreadPattern
	for i, el := range p.Elements {
		if sp, ok := el.(*ast.SpreadPattern); ok {
			if spreadIndex != -1 {
				return fmt.Errorf("vm compile error at %d: list pattern supports only one spread", pos)
			}
			if i != len(p.Elements)-1 {
				return fmt.Errorf("vm compile error at %d: spread (...) must be final element in list pattern", pos)
			}
			spreadIndex = i
			spreadPat = sp
		}
	}

	c.emit(Instruction{Op: OpDup, Position: pos})
	if spreadIndex >= 0 {
		c.emit(Instruction{Op: OpMatchSeqLenGte, IntArg: spreadIndex, Position: pos})
	} else {
		c.emit(Instruction{Op: OpMatchSeqLenEq, IntArg: len(p.Elements), Position: pos})
	}
	localFailJumps = append(localFailJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: pos}))

	for i, el := range p.Elements {
		if spreadIndex >= 0 && i == spreadIndex {
			if spreadPat != nil && spreadPat.Value != nil {
				c.emit(Instruction{Op: OpDup, Position: pos})
				c.emit(Instruction{Op: OpMatchSeqTail, IntArg: spreadIndex, Position: pos})
				if err := c.compileBindPattern(&ast.IdentifierPattern{Value: spreadPat.Value}, true, pos); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpPop, Position: pos})
			}
			break
		}
		if _, ok := el.(*ast.WildcardPattern); ok {
			continue
		}

		c.emit(Instruction{Op: OpDup, Position: pos})
		c.emitNumberConstant(pos, i)
		c.emit(Instruction{Op: OpIndex, Position: pos})
		switch ep := el.(type) {
		case *ast.IdentifierPattern:
			if err := c.compileBindPattern(ep, true, pos); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpPop, Position: pos})
		case *ast.MapPattern:
			if err := c.compileMapPatternFromValue(ep, pos, &localFailJumps); err != nil {
				return err
			}
		case *ast.PinnedIdentifierPattern:
			c.emit(Instruction{Op: OpGetGlobal, StrArg: ep.Value.Value, Position: pos})
			c.emit(Instruction{Op: OpEqual, Position: pos})
			localFailJumps = append(localFailJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: pos}))
		case *ast.LiteralPattern:
			if err := c.compileExpression(ep.Value); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpEqual, Position: pos})
			localFailJumps = append(localFailJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: pos}))
		case *ast.ListPattern:
			nestedFailJumps := make([]int, 0, 4)
			if err := c.compileListPatternFromValue(ep, pos, &nestedFailJumps); err != nil {
				return err
			}
			localFailJumps = append(localFailJumps, nestedFailJumps...)
		default:
			return fmt.Errorf("vm compile error at %d: unsupported list pattern shape", pos)
		}
	}

	c.emit(Instruction{Op: OpPop, Position: pos})
	doneJump := c.emit(Instruction{Op: OpJump, Position: pos})
	localFailTarget := len(c.chunk.Instructions)
	for _, fj := range localFailJumps {
		c.patchJump(fj, localFailTarget)
	}
	c.emit(Instruction{Op: OpPop, Position: pos})
	*failJumps = append(*failJumps, c.emit(Instruction{Op: OpJump, Position: pos}))
	c.patchJump(doneJump, len(c.chunk.Instructions))
	return nil
}

func (c *compiler) compileMapPatternFromValue(mp *ast.MapPattern, pos int, failJumps *[]int) error {
	c.emit(Instruction{Op: OpDup, Position: pos})
	if mp.Exact {
		c.emit(Instruction{Op: OpMatchMapLenEq, IntArg: len(mp.Pairs), Position: pos})
	} else if len(mp.Pairs) > 0 {
		c.emit(Instruction{Op: OpMatchMapLenGte, IntArg: len(mp.Pairs), Position: pos})
	} else if mp.Spread != nil || mp.SelectAll {
		c.emit(Instruction{Op: OpMatchMapLenGte, IntArg: 0, Position: pos})
	} else {
		c.emit(Instruction{Op: OpMatchMapLenEq, IntArg: 0, Position: pos})
	}
	*failJumps = append(*failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: pos}))

	if mp.SelectAll {
		c.emit(Instruction{Op: OpBindMapAllConst, Position: pos})
	}

	for _, entry := range mp.Pairs {
		c.emit(Instruction{Op: OpDup, Position: pos})
		if err := c.compileMapPatternKey(entry.Key, pos); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpMapHasKey, Position: pos})
		*failJumps = append(*failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: pos}))

		c.emit(Instruction{Op: OpDup, Position: pos})
		if err := c.compileMapPatternKey(entry.Key, pos); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpIndex, Position: pos})
		switch ep := entry.Pattern.(type) {
		case *ast.IdentifierPattern:
			if err := c.compileBindPattern(ep, true, pos); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpPop, Position: pos})
		case *ast.PinnedIdentifierPattern:
			c.emit(Instruction{Op: OpGetGlobal, StrArg: ep.Value.Value, Position: pos})
			c.emit(Instruction{Op: OpEqual, Position: pos})
			*failJumps = append(*failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: pos}))
		case *ast.LiteralPattern:
			if err := c.compileExpression(ep.Value); err != nil {
				return err
			}
			c.emit(Instruction{Op: OpEqual, Position: pos})
			*failJumps = append(*failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: pos}))
		case *ast.MapPattern:
			if err := c.compileMapPatternFromValue(ep, pos, failJumps); err != nil {
				return err
			}
		case *ast.ListPattern:
			if err := c.compileListPatternFromValue(ep, pos, failJumps); err != nil {
				return err
			}
		default:
			return fmt.Errorf("vm compile error: unsupported map entry pattern %T", ep)
		}
	}

	c.emit(Instruction{Op: OpPop, Position: pos})
	return nil
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
			Tags:       p.Tags,
		})
	}

	child := &compiler{chunk: &Chunk{}}
	if err := child.compileBlock(node.Body); err != nil {
		return nil, err
	}
	// Mark the function body's outer lexical scope so VM recur cleanup can
	// distinguish it from inner block scopes.
	if len(child.chunk.Instructions) >= 2 {
		child.chunk.Instructions[0].StrArg = "fnroot"
		child.chunk.Instructions[len(child.chunk.Instructions)-1].StrArg = "fnroot"
	}
	child.emit(Instruction{Op: OpReturn, Position: node.Token.Position})

	return &VMFunction{
		Params:     params,
		Chunk:      child.chunk,
		Signature:  node.Signature,
		Parameters: node.Parameters,
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
		var wholeBinding *ast.Identifier
		if bp, ok := pat.(*ast.BindingPattern); ok {
			wholeBinding = bp.Name
			pat = bp.Pattern
		}

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
			if wholeBinding != nil {
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				if err := c.compileBindPattern(&ast.IdentifierPattern{Value: wholeBinding}, true, cs.Token.Position); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			}
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
			if wholeBinding != nil {
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				if err := c.compileBindPattern(&ast.IdentifierPattern{Value: wholeBinding}, true, cs.Token.Position); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			}
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
				case *ast.MapPattern:
					if err := c.compileMapPatternFromValue(ep, cs.Token.Position, &failJumps); err != nil {
						return err
					}
				case *ast.PinnedIdentifierPattern:
					c.emit(Instruction{Op: OpGetGlobal, StrArg: ep.Value.Value, Position: cs.Token.Position})
					c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
					failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))
				case *ast.LiteralPattern:
					if err := c.compileExpression(ep.Value); err != nil {
						return err
					}
					c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
					failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))
				case *ast.ListPattern:
					if err := c.compileListPatternFromValue(ep, cs.Token.Position, &failJumps); err != nil {
						return err
					}
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

		if mp, ok := pat.(*ast.MapPattern); ok {
			c.emit(Instruction{Op: OpPushScope, Position: cs.Token.Position})
			if wholeBinding != nil {
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				if err := c.compileBindPattern(&ast.IdentifierPattern{Value: wholeBinding}, true, cs.Token.Position); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			}
			c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
			if mp.Exact {
				c.emit(Instruction{Op: OpMatchMapLenEq, IntArg: len(mp.Pairs), Position: cs.Token.Position})
			} else if len(mp.Pairs) > 0 {
				c.emit(Instruction{Op: OpMatchMapLenGte, IntArg: len(mp.Pairs), Position: cs.Token.Position})
			} else if mp.Spread != nil || mp.SelectAll {
				c.emit(Instruction{Op: OpMatchMapLenGte, IntArg: 0, Position: cs.Token.Position})
			} else {
				// {} must be exact-empty in Slug patterns
				c.emit(Instruction{Op: OpMatchMapLenEq, IntArg: 0, Position: cs.Token.Position})
			}
			failJumps := []int{c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position})}

			matchedKeys := make([]string, 0, len(mp.Pairs))
			if mp.SelectAll {
				c.emit(Instruction{Op: OpBindMapAllConst, Position: cs.Token.Position})
			}

			for _, entry := range mp.Pairs {
				keyName := mapPatternKeyName(entry.Key)
				if keyName != "" {
					matchedKeys = append(matchedKeys, keyName)
				}
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				if err := c.compileMapPatternKey(entry.Key, cs.Token.Position); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpMapHasKey, Position: cs.Token.Position})
				failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))

				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				if err := c.compileMapPatternKey(entry.Key, cs.Token.Position); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpIndex, Position: cs.Token.Position})
				switch ep := entry.Pattern.(type) {
				case *ast.IdentifierPattern:
					if err := c.compileBindPattern(ep, true, cs.Token.Position); err != nil {
						return err
					}
					c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
				case *ast.PinnedIdentifierPattern:
					c.emit(Instruction{Op: OpGetGlobal, StrArg: ep.Value.Value, Position: cs.Token.Position})
					c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
					failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))
				case *ast.LiteralPattern:
					if err := c.compileExpression(ep.Value); err != nil {
						return err
					}
					c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
					failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))
				default:
					return fmt.Errorf("vm compile error: unsupported map entry pattern %T", ep)
				}
			}

			if sp, ok := mp.Spread.(*ast.SpreadPattern); ok && sp.Value != nil {
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				c.emit(Instruction{
					Op:       OpMatchMapBindRemainder,
					StrArg:   sp.Value.Value,
					StrArg2:  strings.Join(matchedKeys, ","),
					Position: cs.Token.Position,
				})
				c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			} else if idp, ok := mp.Spread.(*ast.IdentifierPattern); ok && idp.Value != nil {
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				c.emit(Instruction{
					Op:       OpMatchMapBindRemainder,
					StrArg:   idp.Value.Value,
					StrArg2:  strings.Join(matchedKeys, ","),
					Position: cs.Token.Position,
				})
				c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
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

		if sp, ok := pat.(*ast.StructPattern); ok {
			c.emit(Instruction{Op: OpPushScope, Position: cs.Token.Position})
			if wholeBinding != nil {
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				if err := c.compileBindPattern(&ast.IdentifierPattern{Value: wholeBinding}, true, cs.Token.Position); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			}
			c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
			c.emit(Instruction{Op: OpMatchStructSchema, StrArg: sp.Schema.Value, Position: cs.Token.Position})
			failJumps := []int{c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position})}
			for _, field := range sp.Fields {
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				idx := c.addConstant(object.InternSymbol(field.Name))
				c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: cs.Token.Position})
				c.emit(Instruction{Op: OpIndex, Position: cs.Token.Position})
				switch ep := field.Pattern.(type) {
				case *ast.IdentifierPattern:
					if err := c.compileBindPattern(ep, true, cs.Token.Position); err != nil {
						return err
					}
					c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
				case *ast.PinnedIdentifierPattern:
					c.emit(Instruction{Op: OpGetGlobal, StrArg: ep.Value.Value, Position: cs.Token.Position})
					c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
					failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))
				case *ast.LiteralPattern:
					if err := c.compileExpression(ep.Value); err != nil {
						return err
					}
					c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
					failJumps = append(failJumps, c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position}))
				default:
					return fmt.Errorf("vm compile error: unsupported struct field pattern %T", ep)
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

		if multi, ok := pat.(*ast.MultiPattern); ok {
			if len(multi.Patterns) == 0 {
				return fmt.Errorf("vm compile error at %d: empty multi pattern", cs.Token.Position)
			}
			matchedJumps := make([]int, 0, len(multi.Patterns))
			for _, sub := range multi.Patterns {
				c.emit(Instruction{Op: OpDup, Position: cs.Token.Position})
				if err := c.compilePatternValue(sub); err != nil {
					return err
				}
				c.emit(Instruction{Op: OpEqual, Position: cs.Token.Position})
				jf := c.emit(Instruction{Op: OpJumpIfFalse, Position: cs.Token.Position})
				c.emit(Instruction{Op: OpTrue, Position: cs.Token.Position})
				j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
				matchedJumps = append(matchedJumps, j)
				c.patchJump(jf, len(c.chunk.Instructions))
			}
			c.emit(Instruction{Op: OpFalse, Position: cs.Token.Position})
			matchedLabel := len(c.chunk.Instructions)
			for _, j := range matchedJumps {
				c.patchJump(j, matchedLabel)
			}

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
			c.emit(Instruction{Op: OpPop, Position: cs.Token.Position})
			if err := c.compileExpression(cs.Body); err != nil {
				return err
			}
			j := c.emit(Instruction{Op: OpJump, Position: cs.Token.Position})
			endJumps = append(endJumps, j)
			c.patchJump(nextCaseJump, len(c.chunk.Instructions))
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

func (c *compiler) compileSelectExpression(node *ast.SelectExpression) error {
	// Temporary lowering: delegate select scheduling semantics to runtime thunk call.
	// This preserves concurrency behavior while native VM select opcodes are pending.
	thunk := &object.Function{
		Signature:  ast.FSig{Min: 0, Max: 0},
		Parameters: []*ast.FunctionParameter{},
		ParamIndex: map[string]int{},
		Body: &ast.BlockStatement{
			Token: node.Token,
			Statements: []ast.Statement{
				&ast.ExpressionStatement{
					Token:      node.Token,
					Expression: node,
				},
			},
		},
	}
	idx := c.addConstant(thunk)
	c.emit(Instruction{Op: OpConstant, IntArg: idx, Position: node.Token.Position})
	c.emit(Instruction{Op: OpCall, IntArg: 0, CallPlan: []CallArgSpec{}, Position: node.Token.Position})
	return nil
}

func (c *compiler) compilePatternValue(pattern ast.MatchPattern) error {
	switch p := pattern.(type) {
	case *ast.LiteralPattern:
		return c.compileExpression(p.Value)
	case *ast.PinnedIdentifierPattern:
		c.emit(Instruction{Op: OpGetGlobal, StrArg: p.Value.Value, Position: p.Token.Position})
		return nil
	case *ast.MultiPattern:
		return fmt.Errorf("vm compile error: multi pattern requires special case lowering")
	default:
		return fmt.Errorf("vm compile error: unsupported match pattern %T", pattern)
	}
}

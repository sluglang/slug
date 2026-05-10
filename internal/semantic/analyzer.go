package semantic

import (
	"bytes"
	"fmt"
	"io"
	"slug/internal/ast"
	"slug/internal/util"
)

type AnalyzeOptions struct {
	EnableTypeCheck bool
	TypeCheckTrace  bool
	TraceWriter     io.Writer
}

// Analyze performs semantic validation and annotations on a parsed AST.
// It returns formatted diagnostics compatible with parser error formatting.
func Analyze(path, src string, program *ast.Program) []string {
	errs, _ := AnalyzeWithOptions(path, src, program, AnalyzeOptions{})
	return errs
}

// AnalyzeWithOptions performs semantic checks and returns both errors and warnings.
func AnalyzeWithOptions(path, src string, program *ast.Program, opts AnalyzeOptions) ([]string, []string) {
	if program == nil {
		return nil, nil
	}
	a := &analyzer{path: path, src: src}
	a.annotateAndValidateFunctions(program)
	a.validateStructSchemaUsage(program)
	a.validateMainTagUsage(program)
	if opts.EnableTypeCheck {
		a.runInferredTypeChecks(program, opts.TypeCheckTrace, opts.TraceWriter)
	}
	return a.errors, a.warnings
}

type analyzer struct {
	path     string
	src      string
	errors   []string
	warnings []string
}

func (a *analyzer) addErrorAt(pos int, message string, args ...interface{}) {
	line, col := util.GetLineAndColumn(a.src, pos)
	m := fmt.Sprintf(message, args...)

	var errorMsg bytes.Buffer
	errorMsg.WriteString(fmt.Sprintf("\nParseError: %s\n", m))
	errorMsg.WriteString(fmt.Sprintf("    --> %s:%d:%d\n", a.path, line, col))
	errorMsg.WriteString(util.GetContextLines(a.src, line, col))

	a.errors = append(a.errors, errorMsg.String())
}

func (a *analyzer) addWarningAt(pos int, message string, args ...interface{}) {
	line, col := util.GetLineAndColumn(a.src, pos)
	m := fmt.Sprintf(message, args...)

	var warningMsg bytes.Buffer
	warningMsg.WriteString(fmt.Sprintf("\nTypeWarning: %s\n", m))
	warningMsg.WriteString(fmt.Sprintf("    --> %s:%d:%d\n", a.path, line, col))
	warningMsg.WriteString(util.GetContextLines(a.src, line, col))

	a.warnings = append(a.warnings, warningMsg.String())
}

func (a *analyzer) annotateAndValidateFunctions(program *ast.Program) {
	for _, stmt := range program.Statements {
		a.walkStatementForFunctions(stmt)
	}
}

func (a *analyzer) walkStatementForFunctions(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		a.walkExpressionForFunctions(s.Expression)
	case *ast.ReturnStatement:
		a.walkExpressionForFunctions(s.ReturnValue)
	case *ast.ThrowStatement:
		a.walkExpressionForFunctions(s.Value)
	case *ast.BlockStatement:
		for _, child := range s.Statements {
			a.walkStatementForFunctions(child)
		}
	case *ast.DeferStatement:
		if s.Call != nil {
			a.walkStatementForFunctions(s.Call)
		}
	}
}

func (a *analyzer) walkExpressionForFunctions(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.FunctionLiteral:
		a.setTailCallFlags(e)
		a.validateRecurUsage(e)
		if e.Body != nil {
			for _, stmt := range e.Body.Statements {
				a.walkStatementForFunctions(stmt)
			}
		}
	case *ast.StructSchemaExpression:
		for _, field := range e.Fields {
			if field.Default != nil {
				a.walkExpressionForFunctions(field.Default)
			}
		}
	case *ast.StructInitExpression:
		a.walkExpressionForFunctions(e.Schema)
		for _, field := range e.Fields {
			a.walkExpressionForFunctions(field.Value)
		}
	case *ast.StructCopyExpression:
		a.walkExpressionForFunctions(e.Source)
		a.walkExpressionForFunctions(e.Fields)
	case *ast.VarExpression:
		a.walkExpressionForFunctions(e.Value)
	case *ast.ValExpression:
		a.walkExpressionForFunctions(e.Value)
	case *ast.IfExpression:
		a.walkExpressionForFunctions(e.Condition)
		a.walkExpressionForFunctions(e.ThenBranch)
		a.walkExpressionForFunctions(e.ElseBranch)
	case *ast.MatchExpression:
		a.walkExpressionForFunctions(e.Value)
		for _, c := range e.Cases {
			if c == nil {
				continue
			}
			a.walkExpressionForFunctions(c.Guard)
			if c.Body != nil {
				for _, stmt := range c.Body.Statements {
					a.walkStatementForFunctions(stmt)
				}
			}
		}
	case *ast.SelectExpression:
		for _, c := range e.Cases {
			if c == nil {
				continue
			}
			a.walkExpressionForFunctions(c.Channel)
			a.walkExpressionForFunctions(c.Value)
			a.walkExpressionForFunctions(c.After)
			a.walkExpressionForFunctions(c.Await)
			a.walkExpressionForFunctions(c.Handler)
		}
	case *ast.CallExpression:
		a.walkExpressionForFunctions(e.Function)
		for _, arg := range e.Arguments {
			a.walkExpressionForFunctions(arg)
		}
	case *ast.InfixExpression:
		a.walkExpressionForFunctions(e.Left)
		a.walkExpressionForFunctions(e.Right)
	case *ast.PrefixExpression:
		a.walkExpressionForFunctions(e.Right)
	case *ast.ListLiteral:
		for _, el := range e.Elements {
			a.walkExpressionForFunctions(el)
		}
	case *ast.MapLiteral:
		for k, v := range e.Pairs {
			a.walkExpressionForFunctions(k)
			a.walkExpressionForFunctions(v)
		}
	case *ast.IndexExpression:
		a.walkExpressionForFunctions(e.Left)
		a.walkExpressionForFunctions(e.Index)
	case *ast.SliceExpression:
		a.walkExpressionForFunctions(e.Start)
		a.walkExpressionForFunctions(e.End)
		a.walkExpressionForFunctions(e.Step)
	case *ast.SpreadExpression:
		a.walkExpressionForFunctions(e.Value)
	case *ast.BlockStatement:
		if e == nil {
			return
		}
		for _, stmt := range e.Statements {
			a.walkStatementForFunctions(stmt)
		}
	}
}

func (a *analyzer) setTailCallFlags(fn *ast.FunctionLiteral) {
	if fn.Body == nil || len(fn.Body.Statements) == 0 {
		return
	}
	fn.HasTailCall = a.checkTailCallsInBlock(fn.Body)
}

func (a *analyzer) checkTailCallsInBlock(block *ast.BlockStatement) bool {
	if block == nil || len(block.Statements) == 0 {
		return false
	}
	for _, stmt := range block.Statements {
		if _, ok := stmt.(*ast.DeferStatement); ok {
			return false
		}
	}
	lastStmt := block.Statements[len(block.Statements)-1]
	return a.checkTailCallsInStatement(lastStmt)
}

func (a *analyzer) checkTailCallsInStatement(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		return a.markTailCall(s.ReturnValue)
	case *ast.ExpressionStatement:
		return a.markTailCall(s.Expression)
	default:
		return false
	}
}

func (a *analyzer) markTailCall(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.CallExpression:
		e.IsTailCall = true
		return true
	case *ast.RecurExpression:
		return true
	case *ast.IfExpression:
		thenHasTail := a.checkTailCallsInBlock(e.ThenBranch)
		elseHasTail := false
		if e.ElseBranch != nil {
			elseHasTail = a.checkTailCallsInBlock(e.ElseBranch)
		}
		return thenHasTail || elseHasTail
	case *ast.MatchExpression:
		hasAnyTailCall := false
		for _, matchCase := range e.Cases {
			if matchCase.Body != nil && a.checkTailCallsInBlock(matchCase.Body) {
				hasAnyTailCall = true
			}
		}
		return hasAnyTailCall
	default:
		return false
	}
}

func (a *analyzer) validateRecurUsage(fn *ast.FunctionLiteral) {
	if fn.Body == nil {
		return
	}
	a.validateRecurInBlock(fn.Body, true)
}

func (a *analyzer) validateRecurInBlock(block *ast.BlockStatement, inTail bool) {
	if block == nil || len(block.Statements) == 0 {
		return
	}
	for i, stmt := range block.Statements {
		stmtInTail := inTail && (i == len(block.Statements)-1)
		a.validateRecurInStatement(stmt, stmtInTail)
	}
}

func (a *analyzer) validateRecurInStatement(stmt ast.Statement, inTail bool) {
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		a.validateRecurInExpr(s.ReturnValue, true)
	case *ast.ExpressionStatement:
		a.validateRecurInExpr(s.Expression, inTail)
	}
}

func (a *analyzer) validateRecurInExpr(expr ast.Expression, inTail bool) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.RecurExpression:
		if !inTail {
			a.addErrorAt(e.Token.Position, "'recur' is only allowed in tail position")
		}
	case *ast.IfExpression:
		a.validateRecurInExpr(e.Condition, false)
		a.validateRecurInBlock(e.ThenBranch, inTail)
		if e.ElseBranch != nil {
			a.validateRecurInBlock(e.ElseBranch, inTail)
		}
	case *ast.MatchExpression:
		if e.Value != nil {
			a.validateRecurInExpr(e.Value, false)
		}
		for _, c := range e.Cases {
			if c == nil {
				continue
			}
			if c.Guard != nil {
				a.validateRecurInExpr(c.Guard, false)
			}
			if c.Body != nil {
				a.validateRecurInBlock(c.Body, inTail)
			}
		}
	case *ast.SelectExpression:
		for _, c := range e.Cases {
			if c == nil {
				continue
			}
			switch c.Kind {
			case ast.SelectRecv:
				a.validateRecurInExpr(c.Channel, false)
			case ast.SelectSend:
				a.validateRecurInExpr(c.Channel, false)
				a.validateRecurInExpr(c.Value, false)
			case ast.SelectAfter:
				a.validateRecurInExpr(c.After, false)
			case ast.SelectAwait:
				a.validateRecurInExpr(c.Await, false)
			}
			if c.Handler != nil {
				a.validateRecurInExpr(c.Handler, inTail)
			}
		}
	case *ast.CallExpression:
		a.validateRecurInExpr(e.Function, false)
		for _, arg := range e.Arguments {
			a.validateRecurInExpr(arg, false)
		}
	case *ast.PrefixExpression:
		a.validateRecurInExpr(e.Right, false)
	case *ast.InfixExpression:
		a.validateRecurInExpr(e.Left, false)
		a.validateRecurInExpr(e.Right, false)
	case *ast.ListLiteral:
		for _, el := range e.Elements {
			a.validateRecurInExpr(el, false)
		}
	case *ast.MapLiteral:
		for k, v := range e.Pairs {
			a.validateRecurInExpr(k, false)
			a.validateRecurInExpr(v, false)
		}
	case *ast.IndexExpression:
		a.validateRecurInExpr(e.Left, false)
		a.validateRecurInExpr(e.Index, false)
	case *ast.SliceExpression:
		a.validateRecurInExpr(e.Start, false)
		a.validateRecurInExpr(e.End, false)
		a.validateRecurInExpr(e.Step, false)
	case *ast.SpreadExpression:
		a.validateRecurInExpr(e.Value, false)
	case *ast.FunctionLiteral:
		a.validateRecurUsage(e)
	}
}

func (a *analyzer) validateStructSchemaUsage(program *ast.Program) {
	for _, stmt := range program.Statements {
		a.validateStructSchemaInStatement(stmt)
	}
}

func (a *analyzer) validateStructSchemaInStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		switch expr := s.Expression.(type) {
		case *ast.VarExpression:
			if _, ok := expr.Value.(*ast.StructSchemaExpression); ok {
				return
			}
			if a.containsStructSchema(expr.Value) {
				a.addErrorAt(expr.Token.Position, "struct schemas are only allowed on the right-hand side of val/var bindings")
			}
		case *ast.ValExpression:
			if _, ok := expr.Value.(*ast.StructSchemaExpression); ok {
				return
			}
			if a.containsStructSchema(expr.Value) {
				a.addErrorAt(expr.Token.Position, "struct schemas are only allowed on the right-hand side of val/var bindings")
			}
		default:
			if a.containsStructSchema(s.Expression) {
				a.addErrorAt(s.Token.Position, "struct schemas are only allowed on the right-hand side of val/var bindings")
			}
		}
	case *ast.ReturnStatement:
		if a.containsStructSchema(s.ReturnValue) {
			a.addErrorAt(s.Token.Position, "struct schemas are only allowed on the right-hand side of val/var bindings")
		}
	case *ast.ThrowStatement:
		if a.containsStructSchema(s.Value) {
			a.addErrorAt(s.Token.Position, "struct schemas are only allowed on the right-hand side of val/var bindings")
		}
	case *ast.BlockStatement:
		for _, child := range s.Statements {
			a.validateStructSchemaInStatement(child)
		}
	case *ast.DeferStatement:
		if a.containsStructSchemaInStatement(s.Call) {
			a.addErrorAt(s.Token.Position, "struct schemas are only allowed on the right-hand side of val/var bindings")
		}
	}
}

func (a *analyzer) containsStructSchema(expr ast.Expression) bool {
	if expr == nil {
		return false
	}

	switch e := expr.(type) {
	case *ast.StructSchemaExpression:
		return true
	case *ast.StructInitExpression:
		if a.containsStructSchema(e.Schema) {
			return true
		}
		for _, field := range e.Fields {
			if a.containsStructSchema(field.Value) {
				return true
			}
		}
		return false
	case *ast.StructCopyExpression:
		if a.containsStructSchema(e.Source) {
			return true
		}
		if a.containsStructSchema(e.Fields) {
			return true
		}
		return false
	case *ast.FunctionLiteral:
		if e.Body != nil {
			for _, stmt := range e.Body.Statements {
				if a.containsStructSchemaInStatement(stmt) {
					return true
				}
			}
		}
		return false
	case *ast.BlockStatement:
		for _, stmt := range e.Statements {
			if a.containsStructSchemaInStatement(stmt) {
				return true
			}
		}
		return false
	case *ast.CallExpression:
		if a.containsStructSchema(e.Function) {
			return true
		}
		for _, arg := range e.Arguments {
			if a.containsStructSchema(arg) {
				return true
			}
		}
		return false
	case *ast.InfixExpression:
		return a.containsStructSchema(e.Left) || a.containsStructSchema(e.Right)
	case *ast.PrefixExpression:
		return a.containsStructSchema(e.Right)
	case *ast.IfExpression:
		if a.containsStructSchema(e.Condition) {
			return true
		}
		if e.ThenBranch != nil && a.containsStructSchema(e.ThenBranch) {
			return true
		}
		if e.ElseBranch != nil && a.containsStructSchema(e.ElseBranch) {
			return true
		}
		return false
	case *ast.MatchExpression:
		if a.containsStructSchema(e.Value) {
			return true
		}
		for _, c := range e.Cases {
			if c == nil {
				continue
			}
			if a.containsStructSchema(c.Guard) {
				return true
			}
			if c.Body != nil && a.containsStructSchema(c.Body) {
				return true
			}
		}
		return false
	case *ast.ListLiteral:
		for _, el := range e.Elements {
			if a.containsStructSchema(el) {
				return true
			}
		}
		return false
	case *ast.MapLiteral:
		for k, v := range e.Pairs {
			if a.containsStructSchema(k) || a.containsStructSchema(v) {
				return true
			}
		}
		return false
	case *ast.IndexExpression:
		return a.containsStructSchema(e.Left) || a.containsStructSchema(e.Index)
	case *ast.SliceExpression:
		return a.containsStructSchema(e.Start) || a.containsStructSchema(e.End) || a.containsStructSchema(e.Step)
	case *ast.SpreadExpression:
		return a.containsStructSchema(e.Value)
	default:
		return false
	}
}

func (a *analyzer) containsStructSchemaInStatement(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return a.containsStructSchema(s.Expression)
	case *ast.ReturnStatement:
		return a.containsStructSchema(s.ReturnValue)
	case *ast.ThrowStatement:
		return a.containsStructSchema(s.Value)
	case *ast.BlockStatement:
		for _, child := range s.Statements {
			if a.containsStructSchemaInStatement(child) {
				return true
			}
		}
		return false
	case *ast.DeferStatement:
		return a.containsStructSchemaInStatement(s.Call)
	default:
		return false
	}
}

func (a *analyzer) validateMainTagUsage(program *ast.Program) {
	mainCount := 0
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.ExpressionStatement:
			switch expr := s.Expression.(type) {
			case *ast.ValExpression:
				a.validateMainOnBinding(expr.Tags, expr.Value, &mainCount)
			case *ast.VarExpression:
				a.validateMainOnBinding(expr.Tags, expr.Value, &mainCount)
			}
		case *ast.ForeignFunctionDeclaration:
			a.validateMainOnForeignFunction(s, &mainCount)
		}
	}
}

func (a *analyzer) validateMainOnBinding(tags []*ast.Tag, value ast.Expression, mainCount *int) {
	mainTag := findTagByName(tags, "@main")
	if mainTag == nil {
		return
	}

	*mainCount++
	if *mainCount > 1 {
		a.addErrorAt(mainTag.Token.Position, "a module may define at most one @main function")
		return
	}

	fn, ok := value.(*ast.FunctionLiteral)
	if !ok {
		a.addErrorAt(mainTag.Token.Position, "@main may only annotate functions")
		return
	}
	if fn.Signature.Min != 0 {
		a.addErrorAt(mainTag.Token.Position, "@main function must be callable with zero arguments")
	}
}

func (a *analyzer) validateMainOnForeignFunction(fn *ast.ForeignFunctionDeclaration, mainCount *int) {
	mainTag := findTagByName(fn.Tags, "@main")
	if mainTag == nil {
		return
	}

	*mainCount++
	if *mainCount > 1 {
		a.addErrorAt(mainTag.Token.Position, "a module may define at most one @main function")
		return
	}
	if fn.Signature.Min != 0 {
		a.addErrorAt(mainTag.Token.Position, "@main function must be callable with zero arguments")
	}
}

func findTagByName(tags []*ast.Tag, name string) *ast.Tag {
	for _, tag := range tags {
		if tag == nil {
			continue
		}
		if tag.Name == name {
			return tag
		}
	}
	return nil
}

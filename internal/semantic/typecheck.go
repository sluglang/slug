package semantic

import (
	"fmt"
	"slug/internal/ast"
	"slug/internal/object"
)

type typeKind string

const (
	typeUnknown typeKind = "unknown"
	typeAny     typeKind = "any"
	typeNil     typeKind = "nil"
	typeBool    typeKind = "bool"
	typeNum     typeKind = "num"
	typeStr     typeKind = "str"
	typeBytes   typeKind = "bytes"
	typeSym     typeKind = "sym"
	typeList    typeKind = "list"
	typeMap     typeKind = "map"
	typeFn      typeKind = "fn"
	typeTask    typeKind = "task"
	typeChan    typeKind = "chan"
	typeStruct  typeKind = "struct"
)

type tnode struct {
	kind     typeKind
	id       int
	parent   *tnode
	elem     *tnode
	key      *tnode
	val      *tnode
	params   []*tnode
	ret      *tnode
	variadic bool
	name     string
}

type tconstraint struct {
	lhs    *tnode
	rhs    *tnode
	pos    int
	reason string
}

type tdiag struct {
	pos int
	msg string
}

type typeChecker struct {
	a           *analyzer
	nextID      int
	constraints []tconstraint
	diags       []tdiag
	scopes      []map[string]*tnode
}

func (a *analyzer) runInferredTypeChecks(program *ast.Program, strict bool) {
	c := newTypeChecker(a)
	c.checkProgram(program)
	for _, d := range c.diags {
		if strict {
			a.addErrorAt(d.pos, "%s", d.msg)
		} else {
			a.addWarningAt(d.pos, "%s", d.msg)
		}
	}
}

func newTypeChecker(a *analyzer) *typeChecker {
	c := &typeChecker{a: a}
	c.pushScope()
	return c
}

func (c *typeChecker) checkProgram(program *ast.Program) {
	for _, stmt := range program.Statements {
		c.inferStatement(stmt)
	}
	c.solveConstraints()
}

func (c *typeChecker) pushScope() {
	c.scopes = append(c.scopes, map[string]*tnode{})
}

func (c *typeChecker) popScope() {
	if len(c.scopes) > 0 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

func (c *typeChecker) bind(name string, t *tnode) {
	if len(c.scopes) == 0 {
		c.pushScope()
	}
	c.scopes[len(c.scopes)-1][name] = t
}

func (c *typeChecker) lookup(name string) *tnode {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i][name]; ok {
			return t
		}
	}
	return nil
}

func (c *typeChecker) freshUnknown() *tnode {
	c.nextID++
	return &tnode{kind: typeUnknown, id: c.nextID}
}

func (c *typeChecker) scalar(kind typeKind) *tnode {
	return &tnode{kind: kind}
}

func (c *typeChecker) listType(elem *tnode) *tnode {
	if elem == nil {
		elem = c.freshUnknown()
	}
	return &tnode{kind: typeList, elem: elem}
}

func (c *typeChecker) mapType(key, val *tnode) *tnode {
	if key == nil {
		key = c.freshUnknown()
	}
	if val == nil {
		val = c.freshUnknown()
	}
	return &tnode{kind: typeMap, key: key, val: val}
}

func (c *typeChecker) fnType(params []*tnode, ret *tnode, variadic bool) *tnode {
	if ret == nil {
		ret = c.freshUnknown()
	}
	return &tnode{kind: typeFn, params: params, ret: ret, variadic: variadic}
}

func (c *typeChecker) cloneScalarTagType(tag string) *tnode {
	switch tag {
	case "@num":
		return c.scalar(typeNum)
	case "@str":
		return c.scalar(typeStr)
	case "@bool":
		return c.scalar(typeBool)
	case "@bytes":
		return c.scalar(typeBytes)
	case "@sym", "@symbol":
		return c.scalar(typeSym)
	case "@list":
		return c.listType(c.freshUnknown())
	case "@map":
		return c.mapType(c.freshUnknown(), c.freshUnknown())
	case "@fn":
		return c.fnType(nil, c.freshUnknown(), false)
	case "@task":
		return c.scalar(typeTask)
	case "@chan":
		return c.scalar(typeChan)
	case "@struct":
		return &tnode{kind: typeStruct}
	default:
		if _, ok := object.TypeTags[tag]; ok {
			return c.scalar(typeAny)
		}
		return nil
	}
}

func (c *typeChecker) addConstraint(lhs, rhs *tnode, pos int, reason string) {
	if lhs == nil || rhs == nil {
		return
	}
	c.constraints = append(c.constraints, tconstraint{lhs: lhs, rhs: rhs, pos: pos, reason: reason})
}

func (c *typeChecker) addDiag(pos int, format string, args ...interface{}) {
	c.diags = append(c.diags, tdiag{pos: pos, msg: fmt.Sprintf(format, args...)})
}

func (c *typeChecker) inferStatement(stmt ast.Statement) *tnode {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return c.inferExpr(s.Expression)
	case *ast.ReturnStatement:
		return c.inferExpr(s.ReturnValue)
	case *ast.ThrowStatement:
		return c.inferExpr(s.Value)
	case *ast.BlockStatement:
		return c.inferBlock(s)
	case *ast.DeferStatement:
		if s.Call != nil {
			c.inferStatement(s.Call)
		}
		return c.scalar(typeNil)
	default:
		return c.freshUnknown()
	}
}

func (c *typeChecker) inferBlock(block *ast.BlockStatement) *tnode {
	if block == nil {
		return c.scalar(typeNil)
	}
	c.pushScope()
	defer c.popScope()
	last := c.scalar(typeNil)
	for _, stmt := range block.Statements {
		last = c.inferStatement(stmt)
	}
	return last
}

func (c *typeChecker) inferExpr(expr ast.Expression) *tnode {
	if expr == nil {
		return c.scalar(typeNil)
	}
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return c.scalar(typeNum)
	case *ast.StringLiteral:
		return c.scalar(typeStr)
	case *ast.BytesLiteral:
		return c.scalar(typeBytes)
	case *ast.Boolean:
		return c.scalar(typeBool)
	case *ast.Nil:
		return c.scalar(typeNil)
	case *ast.SymbolLiteral:
		return c.scalar(typeSym)
	case *ast.Identifier:
		if t := c.lookup(e.Value); t != nil {
			return t
		}
		unknown := c.freshUnknown()
		c.bind(e.Value, unknown)
		return unknown
	case *ast.ListLiteral:
		elemType := c.freshUnknown()
		for _, item := range e.Elements {
			itemType := c.inferExpr(item)
			c.addConstraint(elemType, itemType, e.Token.Position, "list element type")
		}
		return c.listType(elemType)
	case *ast.MapLiteral:
		kt := c.freshUnknown()
		vt := c.freshUnknown()
		for k, v := range e.Pairs {
			kType := c.inferExpr(k)
			vType := c.inferExpr(v)
			c.addConstraint(kt, kType, e.Token.Position, "map key type")
			c.addConstraint(vt, vType, e.Token.Position, "map value type")
		}
		return c.mapType(kt, vt)
	case *ast.PrefixExpression:
		r := c.inferExpr(e.Right)
		switch e.Operator {
		case "-", "~":
			c.addConstraint(r, c.scalar(typeNum), e.Token.Position, "prefix numeric operator")
			return c.scalar(typeNum)
		case "!":
			return c.scalar(typeBool)
		default:
			return c.freshUnknown()
		}
	case *ast.InfixExpression:
		return c.inferInfix(e)
	case *ast.IfExpression:
		cond := c.inferExpr(e.Condition)
		c.addConstraint(cond, c.scalar(typeBool), e.Token.Position, "if condition must be boolean")
		thenType := c.inferBlock(e.ThenBranch)
		elseType := c.scalar(typeNil)
		if e.ElseBranch != nil {
			elseType = c.inferBlock(e.ElseBranch)
		}
		result := c.freshUnknown()
		c.addConstraint(result, thenType, e.Token.Position, "if then branch")
		c.addConstraint(result, elseType, e.Token.Position, "if else branch")
		return result
	case *ast.FunctionLiteral:
		return c.inferFunctionLiteral(e)
	case *ast.CallExpression:
		return c.inferCall(e)
	case *ast.NamedArgument:
		return c.inferExpr(e.Value)
	case *ast.SpreadExpression:
		return c.inferExpr(e.Value)
	case *ast.IndexExpression:
		left := c.inferExpr(e.Left)
		idx := c.inferExpr(e.Index)
		_ = idx
		result := c.freshUnknown()
		// List/map/bytes/string index shapes are handled at runtime; keep broad but constrained.
		container := c.freshUnknown()
		c.addConstraint(container, left, e.Token.Position, "index container")
		c.addConstraint(result, c.freshUnknown(), e.Token.Position, "index result")
		return result
	case *ast.StructSchemaExpression:
		return &tnode{kind: typeStruct, name: "schema"}
	case *ast.StructInitExpression:
		_ = c.inferExpr(e.Schema)
		for _, f := range e.Fields {
			c.inferExpr(f.Value)
		}
		return &tnode{kind: typeStruct}
	case *ast.StructCopyExpression:
		s := c.inferExpr(e.Source)
		_ = c.inferExpr(e.Fields)
		return s
	case *ast.VarExpression:
		rhs := c.inferExpr(e.Value)
		c.enforceTags(rhs, e.Tags, e.Token.Position)
		c.bindPattern(e.Pattern, rhs)
		return rhs
	case *ast.ValExpression:
		rhs := c.inferExpr(e.Value)
		c.enforceTags(rhs, e.Tags, e.Token.Position)
		c.bindPattern(e.Pattern, rhs)
		return rhs
	case *ast.MatchExpression:
		if e.Value != nil {
			c.inferExpr(e.Value)
		}
		out := c.freshUnknown()
		for _, cs := range e.Cases {
			if cs == nil {
				continue
			}
			if cs.Guard != nil {
				guardType := c.inferExpr(cs.Guard)
				c.addConstraint(guardType, c.scalar(typeBool), cs.Token.Position, "match guard must be boolean")
			}
			bodyType := c.inferBlock(cs.Body)
			c.addConstraint(out, bodyType, cs.Token.Position, "match case result")
		}
		return out
	case *ast.SelectExpression:
		out := c.freshUnknown()
		for _, cs := range e.Cases {
			if cs == nil {
				continue
			}
			if cs.Channel != nil {
				c.inferExpr(cs.Channel)
			}
			if cs.Value != nil {
				c.inferExpr(cs.Value)
			}
			if cs.After != nil {
				c.inferExpr(cs.After)
			}
			if cs.Await != nil {
				h := c.inferExpr(cs.Await)
				c.addConstraint(h, c.scalar(typeTask), cs.Token.Position, "select await expects task")
			}
			if cs.Handler != nil {
				hType := c.inferExpr(cs.Handler)
				c.addConstraint(out, hType, cs.Token.Position, "select handler result")
			}
		}
		return out
	case *ast.RecurExpression:
		for _, arg := range e.Arguments {
			c.inferExpr(arg)
		}
		return c.freshUnknown()
	case *ast.SpawnExpression:
		c.inferExpr(e.Body)
		return c.scalar(typeTask)
	case *ast.BlockStatement:
		return c.inferBlock(e)
	default:
		return c.freshUnknown()
	}
}

func (c *typeChecker) inferFunctionLiteral(fn *ast.FunctionLiteral) *tnode {
	params := make([]*tnode, 0, len(fn.Parameters))
	c.pushScope()
	for _, p := range fn.Parameters {
		pt := c.freshUnknown()
		for _, tag := range p.Tags {
			if tag == nil {
				continue
			}
			if tt := c.cloneScalarTagType(tag.Name); tt != nil {
				c.addConstraint(pt, tt, p.Name.Token.Position, "parameter tag constraint")
			}
		}
		if p.Default != nil {
			dt := c.inferExpr(p.Default)
			c.addConstraint(pt, dt, p.Name.Token.Position, "default argument type")
		}
		params = append(params, pt)
		c.bind(p.Name.Value, pt)
	}
	ret := c.scalar(typeNil)
	if fn.Body != nil {
		ret = c.inferBlock(fn.Body)
	}
	c.popScope()
	return c.fnType(params, ret, len(fn.Parameters) > 0 && fn.Parameters[len(fn.Parameters)-1].IsVariadic)
}

func (c *typeChecker) inferCall(call *ast.CallExpression) *tnode {
	callee := c.inferExpr(call.Function)
	result := c.freshUnknown()
	args := make([]*tnode, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		args = append(args, c.inferExpr(arg))
	}
	ct := c.find(callee)
	if ct.kind == typeFn {
		if !ct.variadic && len(args) != len(ct.params) {
			c.addDiag(call.Token.Position, "call arity mismatch: expected %d arguments, got %d", len(ct.params), len(args))
		}
		limit := len(args)
		if len(ct.params) < limit {
			limit = len(ct.params)
		}
		for i := 0; i < limit; i++ {
			c.addConstraint(args[i], ct.params[i], call.Token.Position, "call argument type")
		}
		c.addConstraint(result, ct.ret, call.Token.Position, "call return type")
	}
	return result
}

func (c *typeChecker) inferInfix(e *ast.InfixExpression) *tnode {
	left := c.inferExpr(e.Left)
	right := c.inferExpr(e.Right)
	out := c.freshUnknown()
	switch e.Operator {
	case "=":
		c.addConstraint(left, right, e.Token.Position, "assignment")
		return right
	case "+":
		lf := c.find(left)
		rf := c.find(right)
		if lf.kind == typeStr || rf.kind == typeStr {
			c.addConstraint(out, c.scalar(typeStr), e.Token.Position, "string concatenation")
			return out
		}
		c.addConstraint(left, right, e.Token.Position, "+ operand compatibility")
		// supports num+num, list+list, bytes+bytes; unknown resolved by constraints.
		return out
	case "-", "/", "%", "<<", ">>", "&", "|", "^":
		c.addConstraint(left, c.scalar(typeNum), e.Token.Position, "numeric operator")
		c.addConstraint(right, c.scalar(typeNum), e.Token.Position, "numeric operator")
		c.addConstraint(out, c.scalar(typeNum), e.Token.Position, "numeric result")
		return out
	case "*":
		// runtime supports num*num and str*num
		lf := c.find(left)
		rf := c.find(right)
		if lf.kind == typeStr {
			c.addConstraint(right, c.scalar(typeNum), e.Token.Position, "string repetition count")
			c.addConstraint(out, c.scalar(typeStr), e.Token.Position, "string repetition result")
			return out
		}
		if rf.kind == typeStr {
			c.addConstraint(left, c.scalar(typeNum), e.Token.Position, "string repetition count")
			c.addConstraint(out, c.scalar(typeStr), e.Token.Position, "string repetition result")
			return out
		}
		c.addConstraint(left, c.scalar(typeNum), e.Token.Position, "numeric operator")
		c.addConstraint(right, c.scalar(typeNum), e.Token.Position, "numeric operator")
		c.addConstraint(out, c.scalar(typeNum), e.Token.Position, "numeric result")
		return out
	case "==", "!=":
		c.addConstraint(out, c.scalar(typeBool), e.Token.Position, "equality result")
		return out
	case "<", "<=", ">", ">=":
		c.addConstraint(left, right, e.Token.Position, "comparison operand compatibility")
		c.addConstraint(out, c.scalar(typeBool), e.Token.Position, "comparison result")
		return out
	case "&&", "||":
		c.addConstraint(left, c.scalar(typeBool), e.Token.Position, "logical operator")
		c.addConstraint(right, c.scalar(typeBool), e.Token.Position, "logical operator")
		c.addConstraint(out, c.scalar(typeBool), e.Token.Position, "logical result")
		return out
	case ":+", "+:":
		c.addConstraint(left, c.listType(c.freshUnknown()), e.Token.Position, "list concat operator")
		c.addConstraint(out, c.listType(c.freshUnknown()), e.Token.Position, "list concat result")
		return out
	default:
		return out
	}
}

func (c *typeChecker) bindPattern(pattern ast.MatchPattern, valueType *tnode) {
	switch p := pattern.(type) {
	case *ast.IdentifierPattern:
		if p != nil && p.Value != nil {
			c.bind(p.Value.Value, valueType)
		}
	case *ast.BindingPattern:
		if p.Name != nil {
			c.bind(p.Name.Value, valueType)
		}
		c.bindPattern(p.Pattern, valueType)
	case *ast.ListPattern:
		elem := c.freshUnknown()
		c.addConstraint(valueType, c.listType(elem), p.Token.Position, "list pattern type")
		for _, item := range p.Elements {
			c.bindPattern(item, elem)
		}
	case *ast.MapPattern:
		mv := c.freshUnknown()
		c.addConstraint(valueType, c.mapType(c.freshUnknown(), mv), p.Token.Position, "map pattern type")
		for _, entry := range p.Pairs {
			c.bindPattern(entry.Pattern, mv)
		}
		if p.Spread != nil {
			c.bindPattern(p.Spread, c.mapType(c.freshUnknown(), mv))
		}
	case *ast.StructPattern:
		for _, f := range p.Fields {
			c.bindPattern(f.Pattern, c.freshUnknown())
		}
	case *ast.SpreadPattern:
		if p.Value != nil {
			c.bind(p.Value.Value, c.freshUnknown())
		}
	}
}

func (c *typeChecker) enforceTags(t *tnode, tags []*ast.Tag, pos int) {
	for _, tag := range tags {
		if tag == nil {
			continue
		}
		tt := c.cloneScalarTagType(tag.Name)
		if tt == nil {
			continue
		}
		c.addConstraint(t, tt, pos, fmt.Sprintf("tag constraint %s", tag.Name))
	}
}

func (c *typeChecker) solveConstraints() {
	for _, cs := range c.constraints {
		if !c.unify(cs.lhs, cs.rhs) {
			c.addDiag(cs.pos, "inferred type mismatch (%s): %s vs %s", cs.reason, c.describe(cs.lhs), c.describe(cs.rhs))
		}
	}
}

func (c *typeChecker) find(t *tnode) *tnode {
	if t == nil {
		return nil
	}
	if t.parent == nil {
		return t
	}
	t.parent = c.find(t.parent)
	return t.parent
}

func (c *typeChecker) bindUnknown(u, target *tnode) bool {
	u = c.find(u)
	target = c.find(target)
	if u == target {
		return true
	}
	u.parent = target
	return true
}

func (c *typeChecker) unify(a, b *tnode) bool {
	a = c.find(a)
	b = c.find(b)
	if a == nil || b == nil {
		return true
	}
	if a == b {
		return true
	}
	if a.kind == typeAny || b.kind == typeAny {
		return true
	}
	if a.kind == typeUnknown {
		return c.bindUnknown(a, b)
	}
	if b.kind == typeUnknown {
		return c.bindUnknown(b, a)
	}
	if a.kind != b.kind {
		return false
	}
	switch a.kind {
	case typeList:
		return c.unify(a.elem, b.elem)
	case typeMap:
		return c.unify(a.key, b.key) && c.unify(a.val, b.val)
	case typeFn:
		if a.variadic != b.variadic {
			return false
		}
		if len(a.params) != len(b.params) {
			return false
		}
		for i := range a.params {
			if !c.unify(a.params[i], b.params[i]) {
				return false
			}
		}
		return c.unify(a.ret, b.ret)
	default:
		if a.kind == typeStruct && a.name != "" && b.name != "" && a.name != b.name {
			return false
		}
		return true
	}
}

func (c *typeChecker) describe(t *tnode) string {
	t = c.find(t)
	if t == nil {
		return "unknown"
	}
	switch t.kind {
	case typeUnknown:
		return fmt.Sprintf("t%d", t.id)
	case typeList:
		return "list<" + c.describe(t.elem) + ">"
	case typeMap:
		return "map<" + c.describe(t.key) + ", " + c.describe(t.val) + ">"
	case typeFn:
		return "fn"
	case typeStruct:
		if t.name != "" {
			return "struct(" + t.name + ")"
		}
		return "struct"
	default:
		return string(t.kind)
	}
}

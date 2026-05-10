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
	minArgs  int
	maxArgs  int
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

type callCheck struct {
	pos      int
	got      *tnode
	expected *tnode
}

type plusCheck struct {
	pos   int
	left  *tnode
	right *tnode
}

type mulCheck struct {
	pos   int
	left  *tnode
	right *tnode
}

type typeChecker struct {
	a           *analyzer
	nextID      int
	constraints []tconstraint
	diags       []tdiag
	callChecks  []callCheck
	plusChecks  []plusCheck
	mulChecks   []mulCheck
	scopes      []map[string]*tnode
	schemas     map[string]map[string]*tnode
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
	c := &typeChecker{a: a, schemas: map[string]map[string]*tnode{}}
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

func (c *typeChecker) pushIsolatedScopeFromVisible() {
	visible := c.visibleBindings()
	cloned := map[string]*tnode{}
	for name, t := range visible {
		cloned[name] = c.cloneForCaseScope(t, map[*tnode]*tnode{})
	}
	c.scopes = append(c.scopes, cloned)
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
	scope := c.scopes[len(c.scopes)-1]
	if existing, ok := scope[name]; ok {
		ef := c.find(existing)
		tf := c.find(t)
		// Mirror runtime overload-merging semantics for repeated function bindings:
		// keep a generic function-group-like view instead of letting the latest
		// overload erase prior ones for call-site checks.
		if ef != nil && tf != nil && ef.kind == typeFn && tf.kind == typeFn {
			scope[name] = c.fnType(nil, c.freshUnknown(), false, 0, -1)
			return
		}
	}
	scope[name] = t
}

func (c *typeChecker) lookup(name string) *tnode {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i][name]; ok {
			return t
		}
	}
	return nil
}

func (c *typeChecker) visibleBindings() map[string]*tnode {
	out := map[string]*tnode{}
	for i := 0; i < len(c.scopes); i++ {
		for k, v := range c.scopes[i] {
			out[k] = v
		}
	}
	return out
}

func (c *typeChecker) cloneForCaseScope(t *tnode, memo map[*tnode]*tnode) *tnode {
	t = c.find(t)
	if t == nil {
		return c.freshUnknown()
	}
	if existing, ok := memo[t]; ok {
		return existing
	}
	cp := &tnode{
		kind:     t.kind,
		id:       t.id,
		variadic: t.variadic,
		minArgs:  t.minArgs,
		maxArgs:  t.maxArgs,
		name:     t.name,
	}
	memo[t] = cp
	switch t.kind {
	case typeUnknown:
		cp.id = 0
		cp.kind = typeUnknown
	case typeList:
		cp.elem = c.cloneForCaseScope(t.elem, memo)
	case typeMap:
		cp.key = c.cloneForCaseScope(t.key, memo)
		cp.val = c.cloneForCaseScope(t.val, memo)
	case typeFn:
		cp.params = make([]*tnode, len(t.params))
		for i := range t.params {
			cp.params[i] = c.cloneForCaseScope(t.params[i], memo)
		}
		cp.ret = c.cloneForCaseScope(t.ret, memo)
	}
	return cp
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

func (c *typeChecker) fnType(params []*tnode, ret *tnode, variadic bool, minArgs, maxArgs int) *tnode {
	if ret == nil {
		ret = c.freshUnknown()
	}
	return &tnode{kind: typeFn, params: params, ret: ret, variadic: variadic, minArgs: minArgs, maxArgs: maxArgs}
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
		return c.fnType(nil, c.freshUnknown(), false, 0, -1)
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
		// Throw does not produce a normal value in-flow; model as any-compatible
		// so branch-join logic doesn't force incompatible value unification.
		_ = c.inferExpr(s.Value)
		return c.scalar(typeAny)
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
	return c.inferBlockInCurrentScope(block)
}

func (c *typeChecker) inferBlockInCurrentScope(block *ast.BlockStatement) *tnode {
	if block == nil {
		return c.scalar(typeNil)
	}
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
		// Slug lists are heterogeneous at runtime. Infer each element for local
		// constraints/side effects, but do not force all elements to unify.
		elemType := c.scalar(typeAny)
		for _, item := range e.Elements {
			_ = c.inferExpr(item)
		}
		return c.listType(elemType)
	case *ast.MapLiteral:
		// Slug maps are heterogeneous at runtime for both keys and values.
		// Infer nested expressions for local checks without forcing global
		// key/value homogeneity across the literal.
		kt := c.scalar(typeAny)
		vt := c.scalar(typeAny)
		for k, v := range e.Pairs {
			_ = c.inferExpr(k)
			_ = c.inferExpr(v)
		}
		return c.mapType(kt, vt)
	case *ast.PrefixExpression:
		r := c.inferExpr(e.Right)
		switch e.Operator {
		case "-":
			c.addConstraint(r, c.scalar(typeNum), e.Token.Position, "prefix numeric operator")
			return c.scalar(typeNum)
		case "~":
			rt := c.find(r)
			if rt.kind == typeBytes {
				return c.scalar(typeBytes)
			}
			if rt.kind == typeUnknown || rt.kind == typeAny {
				return c.freshUnknown()
			}
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
		schemaType := c.inferExpr(e.Schema)
		c.addConstraint(schemaType, &tnode{kind: typeStruct}, e.Token.Position, "struct init schema type")
		var schemaName string
		if id, ok := e.Schema.(*ast.Identifier); ok {
			schemaName = id.Value
		}
		for _, f := range e.Fields {
			fv := c.inferExpr(f.Value)
			if schemaName != "" {
				if fields, ok := c.schemas[schemaName]; ok {
					if expected, ok := fields[f.Name]; ok {
						c.addConstraint(fv, expected, f.Token.Position, fmt.Sprintf("struct field %s.%s", schemaName, f.Name))
					}
				}
			}
		}
		return &tnode{kind: typeStruct, name: schemaName}
	case *ast.StructCopyExpression:
		s := c.inferExpr(e.Source)
		_ = c.inferExpr(e.Fields)
		return s
	case *ast.VarExpression:
		rhs := c.inferExpr(e.Value)
		c.enforceTags(rhs, e.Tags, e.Token.Position)
		c.bindPattern(e.Pattern, rhs)
		c.registerStructSchema(e.Pattern, e.Value)
		return rhs
	case *ast.ValExpression:
		rhs := c.inferExpr(e.Value)
		c.enforceTags(rhs, e.Tags, e.Token.Position)
		c.bindPattern(e.Pattern, rhs)
		c.registerStructSchema(e.Pattern, e.Value)
		return rhs
	case *ast.MatchExpression:
		var scrutinee *tnode
		if e.Value != nil {
			scrutinee = c.inferExpr(e.Value)
		}
		out := c.freshUnknown()
		for _, cs := range e.Cases {
			if cs == nil {
				continue
			}
			c.pushIsolatedScopeFromVisible()
			caseMemo := map[*tnode]*tnode{}
			caseScrutinee := scrutinee
			if caseScrutinee != nil {
				caseScrutinee = c.cloneForCaseScope(caseScrutinee, caseMemo)
			}
			if caseScrutinee != nil {
				c.narrowPattern(cs.Pattern, caseScrutinee, cs.Token.Position)
			}
			if cs.Guard != nil {
				guardType := c.inferExpr(cs.Guard)
				c.addConstraint(guardType, c.scalar(typeBool), cs.Token.Position, "match guard must be boolean")
			}
			bodyType := c.inferBlockInCurrentScope(cs.Body)
			c.popScope()
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
	minArgs := fn.Signature.Min
	maxArgs := fn.Signature.Max
	if maxArgs > 1_000_000 {
		maxArgs = -1
	}
	return c.fnType(params, ret, len(fn.Parameters) > 0 && fn.Parameters[len(fn.Parameters)-1].IsVariadic, minArgs, maxArgs)
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
		argc := len(args)
		minArgs := ct.minArgs
		maxArgs := ct.maxArgs
		if maxArgs == 0 && len(ct.params) > 0 && !ct.variadic {
			maxArgs = len(ct.params)
		}
		if maxArgs == 0 && len(ct.params) == 0 {
			maxArgs = -1
		}
		if argc < minArgs || (maxArgs >= 0 && argc > maxArgs) {
			if maxArgs >= 0 {
				c.addDiag(call.Token.Position, "call arity mismatch: expected %d..%d arguments, got %d", minArgs, maxArgs, argc)
			} else {
				c.addDiag(call.Token.Position, "call arity mismatch: expected at least %d arguments, got %d", minArgs, argc)
			}
		}
		limit := len(args)
		if len(ct.params) < limit {
			limit = len(ct.params)
		}
		for i := 0; i < limit; i++ {
			c.callChecks = append(c.callChecks, callCheck{
				pos:      call.Token.Position,
				got:      args[i],
				expected: ct.params[i],
			})
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
		// Defer '+' compatibility check to runtime-aligned rule after solving.
		c.plusChecks = append(c.plusChecks, plusCheck{pos: e.Token.Position, left: left, right: right})
		return out
	case "-", "/", "%", "<<", ">>":
		c.addConstraint(left, c.scalar(typeNum), e.Token.Position, "numeric operator")
		c.addConstraint(right, c.scalar(typeNum), e.Token.Position, "numeric operator")
		c.addConstraint(out, c.scalar(typeNum), e.Token.Position, "numeric result")
		return out
	case "&", "|", "^":
		lf := c.find(left)
		rf := c.find(right)
		// runtime supports:
		// - num<op>num -> num
		// - bytes<op>bytes -> bytes
		// - bytes<op>num / num<op>bytes -> bytes (num converted to one byte at runtime)
		if lf.kind == typeBytes || rf.kind == typeBytes {
			if lf.kind != typeBytes {
				c.addConstraint(left, c.scalar(typeNum), e.Token.Position, "bytes bitwise mixed operator")
			}
			if rf.kind != typeBytes {
				c.addConstraint(right, c.scalar(typeNum), e.Token.Position, "bytes bitwise mixed operator")
			}
			c.addConstraint(out, c.scalar(typeBytes), e.Token.Position, "bytes bitwise result")
			return out
		}
		c.addConstraint(left, c.scalar(typeNum), e.Token.Position, "numeric operator")
		c.addConstraint(right, c.scalar(typeNum), e.Token.Position, "numeric operator")
		c.addConstraint(out, c.scalar(typeNum), e.Token.Position, "numeric result")
		return out
	case "*":
		// Defer '*' compatibility check (runtime allows num*num and str*num).
		c.mulChecks = append(c.mulChecks, mulCheck{pos: e.Token.Position, left: left, right: right})
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
		lf := c.find(left)
		rf := c.find(right)
		if e.Operator == ":+" {
			// list :+ any -> list
			if lf.kind == typeList {
				c.addConstraint(out, lf, e.Token.Position, "list concat result")
				return out
			}
			// bytes :+ num -> bytes
			if lf.kind == typeBytes {
				c.addConstraint(right, c.scalar(typeNum), e.Token.Position, "bytes append operator")
				c.addConstraint(out, c.scalar(typeBytes), e.Token.Position, "bytes append result")
				return out
			}
			// Unknown lhs: allow either list or bytes mode without eager mismatch.
			if lf.kind == typeUnknown || lf.kind == typeAny {
				c.addConstraint(out, c.freshUnknown(), e.Token.Position, "list concat result")
				return out
			}
			c.addDiag(e.Token.Position, "list/bytes append type mismatch: expected list or bytes on left, got %s", c.describe(left))
			return out
		}
		// +: operator
		// any +: list -> list
		if rf.kind == typeList {
			c.addConstraint(out, rf, e.Token.Position, "list prepend result")
			return out
		}
		// num +: bytes -> bytes
		if rf.kind == typeBytes {
			c.addConstraint(left, c.scalar(typeNum), e.Token.Position, "bytes prepend operator")
			c.addConstraint(out, c.scalar(typeBytes), e.Token.Position, "bytes prepend result")
			return out
		}
		if rf.kind == typeUnknown || rf.kind == typeAny {
			c.addConstraint(out, c.freshUnknown(), e.Token.Position, "list concat result")
			return out
		}
		c.addDiag(e.Token.Position, "list/bytes prepend type mismatch: expected list or bytes on right, got %s", c.describe(right))
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
		elem := c.sequenceElementType(valueType)
		c.constrainSequenceLike(valueType, elem, p.Token.Position, "list pattern type")
		for _, item := range p.Elements {
			c.bindPattern(item, elem)
		}
	case *ast.MapPattern:
		c.addConstraint(valueType, c.mapType(c.freshUnknown(), c.scalar(typeAny)), p.Token.Position, "map pattern type")
		for _, entry := range p.Pairs {
			c.bindPattern(entry.Pattern, c.freshUnknown())
		}
		if p.Spread != nil {
			c.bindPattern(p.Spread, c.mapType(c.freshUnknown(), c.scalar(typeAny)))
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

func (c *typeChecker) narrowPattern(pattern ast.MatchPattern, valueType *tnode, pos int) {
	if pattern == nil {
		return
	}
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return
	case *ast.IdentifierPattern, *ast.BindingPattern, *ast.SpreadPattern:
		c.bindPattern(pattern, valueType)
	case *ast.LiteralPattern:
		lit := c.inferExpr(p.Value)
		c.addConstraint(valueType, lit, pos, "match literal pattern")
	case *ast.ListPattern:
		elem := c.sequenceElementType(valueType)
		c.constrainSequenceLike(valueType, elem, pos, "match list pattern")
		for _, item := range p.Elements {
			c.narrowPattern(item, elem, pos)
		}
	case *ast.MapPattern:
		c.addConstraint(valueType, c.mapType(c.freshUnknown(), c.scalar(typeAny)), pos, "match map pattern")
		for _, entry := range p.Pairs {
			c.narrowPattern(entry.Pattern, c.freshUnknown(), pos)
		}
		if p.Spread != nil {
			c.narrowPattern(p.Spread, c.mapType(c.freshUnknown(), c.scalar(typeAny)), pos)
		}
	case *ast.StructPattern:
		c.addConstraint(valueType, &tnode{kind: typeStruct, name: p.Schema.Value}, pos, "match struct pattern")
		if fields, ok := c.schemas[p.Schema.Value]; ok {
			for _, f := range p.Fields {
				expected, found := fields[f.Name]
				if found {
					c.narrowPattern(f.Pattern, expected, pos)
				} else {
					c.narrowPattern(f.Pattern, c.freshUnknown(), pos)
				}
			}
			return
		}
		for _, f := range p.Fields {
			c.narrowPattern(f.Pattern, c.freshUnknown(), pos)
		}
	case *ast.MultiPattern:
		// Multi-pattern alternatives are OR semantics. Constraining against all
		// alternatives over-constrains (intersection) and creates false positives.
		// For v1, narrow by the first alternative only as a safe approximation.
		if len(p.Patterns) > 0 {
			c.narrowPattern(p.Patterns[0], valueType, pos)
		}
	}
}

func (c *typeChecker) sequenceElementType(seq *tnode) *tnode {
	s := c.find(seq)
	if s != nil {
		switch s.kind {
		case typeBytes:
			return c.scalar(typeNum)
		case typeList:
			if s.elem != nil {
				return s.elem
			}
		}
	}
	return c.freshUnknown()
}

func (c *typeChecker) constrainSequenceLike(seq, elem *tnode, pos int, reason string) {
	s := c.find(seq)
	if s == nil {
		return
	}
	switch s.kind {
	case typeBytes:
		// bytes are sequence-like at runtime and match list-pattern opcodes
		return
	case typeList, typeUnknown, typeAny:
		c.addConstraint(seq, c.listType(elem), pos, reason)
	default:
		// Keep diagnostic quality clear for obvious non-sequence matches.
		c.addDiag(pos, "list pattern type mismatch: expected list or bytes, got %s", c.describe(seq))
	}
}

func (c *typeChecker) registerStructSchema(pattern ast.MatchPattern, value ast.Expression) {
	if value == nil || pattern == nil {
		return
	}
	schemaExpr, ok := value.(*ast.StructSchemaExpression)
	if !ok {
		return
	}
	idPattern, ok := pattern.(*ast.IdentifierPattern)
	if !ok || idPattern.Value == nil {
		return
	}
	name := idPattern.Value.Value
	fieldTypes := map[string]*tnode{}
	for _, field := range schemaExpr.Fields {
		ft := c.freshUnknown()
		for _, tag := range field.Tags {
			if tag == nil {
				continue
			}
			if tt := c.cloneScalarTagType(tag.Name); tt != nil {
				c.addConstraint(ft, tt, field.Token.Position, fmt.Sprintf("struct schema tag %s.%s", name, field.Name))
			}
		}
		if field.Default != nil {
			dt := c.inferExpr(field.Default)
			c.addConstraint(ft, dt, field.Token.Position, fmt.Sprintf("struct schema default %s.%s", name, field.Name))
		}
		fieldTypes[field.Name] = ft
	}
	c.schemas[name] = fieldTypes
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
			c.addDiag(cs.pos, "%s", c.renderConstraintMismatch(cs))
		}
	}
	for _, cc := range c.callChecks {
		if !c.isCompatible(cc.got, cc.expected) {
			c.addDiag(cc.pos, "call argument type mismatch: expected %s, got %s", c.describe(cc.expected), c.describe(cc.got))
		}
	}
	for _, pc := range c.plusChecks {
		if !c.isPlusCompatible(pc.left, pc.right) {
			c.addDiag(pc.pos, "operator '+' type mismatch: %s vs %s", c.describe(pc.left), c.describe(pc.right))
		}
	}
	for _, mc := range c.mulChecks {
		if !c.isMulCompatible(mc.left, mc.right) {
			c.addDiag(mc.pos, "numeric operator type mismatch: expected num, got %s", c.describe(mc.left))
		}
	}
}

func (c *typeChecker) isCompatible(got, expected *tnode) bool {
	g := c.find(got)
	e := c.find(expected)
	if g == nil || e == nil {
		return true
	}
	if g == e {
		return true
	}
	if g.kind == typeAny || e.kind == typeAny || g.kind == typeUnknown || e.kind == typeUnknown {
		return true
	}
	if g.kind == typeNil || e.kind == typeNil {
		return true
	}
	if g.kind != e.kind {
		return false
	}
	switch g.kind {
	case typeList:
		return c.isCompatible(g.elem, e.elem)
	case typeMap:
		return c.isCompatible(g.key, e.key) && c.isCompatible(g.val, e.val)
	case typeFn:
		if (e.maxArgs == -1 && len(e.params) == 0) || (g.maxArgs == -1 && len(g.params) == 0) {
			return true
		}
		if g.variadic != e.variadic {
			return false
		}
		if len(g.params) != len(e.params) {
			return false
		}
		for i := range g.params {
			if !c.isCompatible(g.params[i], e.params[i]) {
				return false
			}
		}
		return c.isCompatible(g.ret, e.ret)
	case typeStruct:
		if g.name != "" && e.name != "" && g.name != e.name {
			return false
		}
		return true
	default:
		return true
	}
}

func (c *typeChecker) isPlusCompatible(left, right *tnode) bool {
	l := c.find(left)
	r := c.find(right)
	if l == nil || r == nil {
		return true
	}
	// Be permissive with unresolved types.
	if l.kind == typeAny || r.kind == typeAny || l.kind == typeUnknown || r.kind == typeUnknown {
		return true
	}
	// Runtime allows string concatenation with any value.
	if l.kind == typeStr || r.kind == typeStr {
		return true
	}
	// Non-string modes require same shape.
	switch l.kind {
	case typeNum:
		return r.kind == typeNum
	case typeList:
		return r.kind == typeList
	case typeBytes:
		return r.kind == typeBytes
	default:
		return false
	}
}

func (c *typeChecker) isMulCompatible(left, right *tnode) bool {
	l := c.find(left)
	r := c.find(right)
	if l == nil || r == nil {
		return true
	}
	if l.kind == typeAny || r.kind == typeAny || l.kind == typeUnknown || r.kind == typeUnknown {
		return true
	}
	// runtime: str * num
	if l.kind == typeStr {
		return r.kind == typeNum
	}
	// runtime: num * num
	if l.kind == typeNum {
		return r.kind == typeNum
	}
	return false
}

func (c *typeChecker) renderConstraintMismatch(cs tconstraint) string {
	left := c.describe(cs.lhs)
	right := c.describe(cs.rhs)
	switch cs.reason {
	case "assignment":
		return fmt.Sprintf("assignment type mismatch: cannot assign %s to %s", right, left)
	case "call argument type":
		return fmt.Sprintf("call argument type mismatch: expected %s, got %s", right, left)
	case "call return type":
		return fmt.Sprintf("call return type mismatch: expected %s, got %s", right, left)
	case "if condition must be boolean":
		return fmt.Sprintf("if condition type mismatch: expected bool, got %s", left)
	case "match guard must be boolean":
		return fmt.Sprintf("match guard type mismatch: expected bool, got %s", left)
	case "select await expects task":
		return fmt.Sprintf("select await type mismatch: expected task, got %s", left)
	case "prefix numeric operator":
		return fmt.Sprintf("prefix operator type mismatch: expected num, got %s", left)
	case "numeric operator":
		return fmt.Sprintf("numeric operator type mismatch: expected num, got %s", left)
	case "logical operator":
		return fmt.Sprintf("logical operator type mismatch: expected bool, got %s", left)
	case "+ operand compatibility":
		return fmt.Sprintf("operator '+' type mismatch: %s vs %s", left, right)
	case "comparison operand compatibility":
		return fmt.Sprintf("comparison operand type mismatch: %s vs %s", left, right)
	}
	return fmt.Sprintf("inferred type mismatch (%s): %s vs %s", cs.reason, left, right)
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
	// Slug runtime/tag dispatch treats nil as admissible across typed positions.
	// Keep semantic typing aligned by allowing nil to unify with any concrete type.
	if a.kind == typeNil || b.kind == typeNil {
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
		// Generic function constraints (e.g. from @fn tags) intentionally do not
		// pin arity/parameter shape; they only require function-typed values.
		if (a.maxArgs == -1 && len(a.params) == 0) || (b.maxArgs == -1 && len(b.params) == 0) {
			return true
		}
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

package semantic

import (
	"fmt"
	"io"
	"slug/internal/ast"
	"slug/internal/object"
	"slug/internal/util"
	"sort"
	"strings"
)

type typeKind string

const (
	typeUnknown typeKind = "unknown"
	typeNever   typeKind = "never"
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
	typeUnion   typeKind = "union"
)

const maxUnionOptions = 8

type tnode struct {
	kind     typeKind
	id       int
	parent   *tnode
	elem     *tnode
	key      *tnode
	val      *tnode
	params   []*tnode
	options  []*tnode
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

type bitwiseCheck struct {
	pos   int
	left  *tnode
	right *tnode
}

type typeChecker struct {
	a           *analyzer
	trace       bool
	traceWriter io.Writer
	nextID      int
	constraints []tconstraint
	diags       []tdiag
	diagSeen    map[string]bool
	callChecks  []callCheck
	plusChecks  []plusCheck
	mulChecks   []mulCheck
	bitChecks   []bitwiseCheck
	scopes      []map[string]*tnode
	declared    []map[string]*tnode
	overrides   []map[string]*tnode
	schemas     map[string]map[string]*tnode
}

func (a *analyzer) runInferredTypeChecks(program *ast.Program, trace bool, traceWriter io.Writer) {
	c := newTypeChecker(a, trace, traceWriter)
	c.checkProgram(program)
	for _, d := range c.diags {
		a.addErrorAt(d.pos, "%s", d.msg)
	}
}

func newTypeChecker(a *analyzer, trace bool, traceWriter io.Writer) *typeChecker {
	c := &typeChecker{
		a:           a,
		trace:       trace,
		traceWriter: traceWriter,
		schemas:     map[string]map[string]*tnode{},
		diagSeen:    map[string]bool{},
	}
	c.pushScope()
	c.pushOverride()
	return c
}

func (c *typeChecker) tracef(pos int, event string, details string) {
	if !c.trace || c.traceWriter == nil {
		return
	}
	if pos < 0 {
		_, _ = fmt.Fprintf(c.traceWriter, "TypeTrace: %s | %s\n", event, details)
		return
	}
	line, col := util.GetLineAndColumn(c.a.src, pos)
	_, _ = fmt.Fprintf(c.traceWriter, "TypeTrace: %s @ %s:%d:%d | %s\n", event, c.a.path, line, col, details)
}

func (c *typeChecker) checkProgram(program *ast.Program) {
	for _, stmt := range program.Statements {
		c.inferStatement(stmt)
	}
	c.solveConstraints()
}

func (c *typeChecker) pushScope() {
	c.scopes = append(c.scopes, map[string]*tnode{})
	c.declared = append(c.declared, map[string]*tnode{})
}

func (c *typeChecker) pushOverride() {
	c.overrides = append(c.overrides, map[string]*tnode{})
}

func (c *typeChecker) pushIsolatedScopeFromVisible() {
	visible := c.visibleBindings()
	cloned := map[string]*tnode{}
	for name, t := range visible {
		cloned[name] = c.cloneForCaseScope(t, map[*tnode]*tnode{})
	}
	c.scopes = append(c.scopes, cloned)
	c.declared = append(c.declared, map[string]*tnode{})
	ov := map[string]*tnode{}
	for name, t := range cloned {
		ov[name] = t
	}
	c.overrides = append(c.overrides, ov)
}

func (c *typeChecker) popScope() {
	if len(c.scopes) > 0 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
	if len(c.declared) > 0 {
		c.declared = c.declared[:len(c.declared)-1]
	}
}

func (c *typeChecker) popOverride() {
	if len(c.overrides) > 0 {
		c.overrides = c.overrides[:len(c.overrides)-1]
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
			c.setOverride(name, scope[name])
			return
		}
	}
	scope[name] = t
	c.setOverride(name, t)
}

func (c *typeChecker) bindDeclared(name string, t *tnode) {
	if name == "" || t == nil {
		return
	}
	if len(c.declared) == 0 {
		c.declared = append(c.declared, map[string]*tnode{})
	}
	c.declared[len(c.declared)-1][name] = t
}

func (c *typeChecker) lookupDeclared(name string) *tnode {
	for i := len(c.declared) - 1; i >= 0; i-- {
		if t, ok := c.declared[i][name]; ok {
			return t
		}
	}
	return nil
}

func (c *typeChecker) lookup(name string) *tnode {
	for i := len(c.overrides) - 1; i >= 0; i-- {
		if t, ok := c.overrides[i][name]; ok {
			return t
		}
	}
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i][name]; ok {
			return t
		}
	}
	return nil
}

func (c *typeChecker) setOverride(name string, t *tnode) {
	if name == "" || t == nil {
		return
	}
	if len(c.overrides) == 0 {
		c.pushOverride()
	}
	c.overrides[len(c.overrides)-1][name] = t
}

func (c *typeChecker) snapshotCurrentOverride() map[string]*tnode {
	out := map[string]*tnode{}
	if len(c.overrides) == 0 {
		return out
	}
	for k, v := range c.overrides[len(c.overrides)-1] {
		out[k] = v
	}
	return out
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
	case typeUnion:
		cp.options = make([]*tnode, 0, len(t.options))
		for _, opt := range t.options {
			cp.options = append(cp.options, c.cloneForCaseScope(opt, memo))
		}
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

func (c *typeChecker) unionType(types ...*tnode) *tnode {
	origCount := len(types)
	flat := make([]*tnode, 0, len(types))
	seen := map[string]bool{}
	var addOpt func(*tnode)
	addOpt = func(t *tnode) {
		if t == nil {
			return
		}
		tf := c.find(t)
		if tf == nil {
			return
		}
		if tf.kind == typeAny {
			flat = []*tnode{c.scalar(typeAny)}
			seen = map[string]bool{"any": true}
			return
		}
		if tf.kind == typeNever {
			return
		}
		if tf.kind == typeUnion {
			for _, opt := range tf.options {
				addOpt(opt)
				if len(flat) == 1 && flat[0].kind == typeAny {
					return
				}
			}
			return
		}
		sig := c.typeSig(tf)
		if seen[sig] {
			return
		}
		seen[sig] = true
		flat = append(flat, tf)
	}
	for _, t := range types {
		addOpt(t)
		if len(flat) == 1 && flat[0].kind == typeAny {
			return flat[0]
		}
	}
	flat = c.normalizeUnionOptions(flat)
	if len(flat) == 0 {
		return c.freshUnknown()
	}
	if len(flat) == 1 {
		return flat[0]
	}
	if len(flat) > maxUnionOptions {
		c.tracef(-1, "union-normalize", fmt.Sprintf("widen-to-any options=%d cap=%d", len(flat), maxUnionOptions))
		return c.scalar(typeAny)
	}
	u := &tnode{kind: typeUnion, options: flat}
	c.tracef(-1, "union-normalize", fmt.Sprintf("inputs=%d normalized=%d -> %s", origCount, len(flat), c.describe(u)))
	return u
}

func (c *typeChecker) normalizeUnionOptions(options []*tnode) []*tnode {
	if len(options) == 0 {
		return options
	}
	// any dominates all.
	for _, opt := range options {
		of := c.find(opt)
		if of != nil && of.kind == typeAny {
			return []*tnode{c.scalar(typeAny)}
		}
	}
	// Dedup + remove never.
	out := make([]*tnode, 0, len(options))
	seen := map[string]bool{}
	for _, opt := range options {
		of := c.find(opt)
		if of == nil || of.kind == typeNever {
			continue
		}
		sig := c.typeSig(of)
		if seen[sig] {
			continue
		}
		seen[sig] = true
		out = append(out, of)
	}
	// Merge list/map families where safe.
	out = c.mergeUnionFamilies(out)
	return out
}

func (c *typeChecker) mergeUnionFamilies(options []*tnode) []*tnode {
	if len(options) < 2 {
		return options
	}
	var listElem *tnode
	var mapKey *tnode
	var mapVal *tnode
	hasList := false
	hasMap := false
	rest := make([]*tnode, 0, len(options))
	for _, opt := range options {
		of := c.find(opt)
		if of == nil {
			continue
		}
		switch of.kind {
		case typeList:
			hasList = true
			if listElem == nil {
				listElem = of.elem
			} else {
				listElem = c.unionType(listElem, of.elem)
			}
		case typeMap:
			hasMap = true
			if mapKey == nil {
				mapKey = of.key
				mapVal = of.val
			} else {
				mapKey = c.unionType(mapKey, of.key)
				mapVal = c.unionType(mapVal, of.val)
			}
		default:
			rest = append(rest, of)
		}
	}
	if hasList {
		rest = append(rest, c.listType(listElem))
	}
	if hasMap {
		rest = append(rest, c.mapType(mapKey, mapVal))
	}
	return rest
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

func (c *typeChecker) parseDeclaredType(raw string) *tnode {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	if parts := splitTypeTopLevel(s, '|'); len(parts) > 1 {
		opts := make([]*tnode, 0, len(parts))
		for _, p := range parts {
			t := c.parseDeclaredType(p)
			if t != nil {
				opts = append(opts, t)
			}
		}
		if len(opts) == 0 {
			return nil
		}
		return c.unionType(opts...)
	}
	if strings.HasPrefix(s, "list<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("list<") : len(s)-1])
		elem := c.parseDeclaredType(inner)
		if elem == nil {
			elem = c.freshUnknown()
		}
		return c.listType(elem)
	}
	if strings.HasPrefix(s, "map<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("map<") : len(s)-1])
		parts := splitTypeTopLevel(inner, ',')
		if len(parts) == 2 {
			k := c.parseDeclaredType(parts[0])
			v := c.parseDeclaredType(parts[1])
			if k == nil {
				k = c.freshUnknown()
			}
			if v == nil {
				v = c.freshUnknown()
			}
			return c.mapType(k, v)
		}
		return c.mapType(c.freshUnknown(), c.freshUnknown())
	}
	if strings.HasPrefix(s, "chan<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("chan<") : len(s)-1])
		// channel payload typing is accepted in declarations for forward compatibility.
		// current checker treats channel as a scalar kind, consistent with existing behavior.
		_ = c.parseDeclaredType(inner)
		return c.scalar(typeChan)
	}
	if strings.HasPrefix(s, "task<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("task<") : len(s)-1])
		_ = c.parseDeclaredType(inner)
		return c.scalar(typeTask)
	}
	if strings.HasPrefix(s, "fn<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("fn<") : len(s)-1])
		parts := splitTypeTopLevel(inner, ',')
		if len(parts) == 0 {
			return c.fnType(nil, c.freshUnknown(), false, 0, -1)
		}
		if len(parts) == 1 {
			return c.fnType(nil, c.parseDeclaredType(parts[0]), false, 0, -1)
		}
		params := make([]*tnode, 0, len(parts)-1)
		for _, p := range parts[:len(parts)-1] {
			pt := c.parseDeclaredType(p)
			if pt == nil {
				pt = c.freshUnknown()
			}
			params = append(params, pt)
		}
		ret := c.parseDeclaredType(parts[len(parts)-1])
		if ret == nil {
			ret = c.freshUnknown()
		}
		return c.fnType(params, ret, false, len(params), len(params))
	}
	if strings.HasPrefix(s, "struct<") && strings.HasSuffix(s, ">") {
		name := strings.TrimSpace(s[len("struct<") : len(s)-1])
		return &tnode{kind: typeStruct, name: name}
	}
	switch s {
	case "num":
		return c.scalar(typeNum)
	case "str":
		return c.scalar(typeStr)
	case "bool":
		return c.scalar(typeBool)
	case "bytes":
		return c.scalar(typeBytes)
	case "sym", "symbol":
		return c.scalar(typeSym)
	case "list":
		return c.listType(c.freshUnknown())
	case "map":
		return c.mapType(c.freshUnknown(), c.freshUnknown())
	case "fn":
		return c.fnType(nil, c.freshUnknown(), false, 0, -1)
	case "task":
		return c.scalar(typeTask)
	case "chan":
		return c.scalar(typeChan)
	case "struct":
		return &tnode{kind: typeStruct}
	case "nil":
		return c.scalar(typeNil)
	case "any", "?":
		return c.scalar(typeAny)
	default:
		// Treat unknown bare names as struct references.
		if isSimpleTypeIdent(s) {
			return &tnode{kind: typeStruct, name: s}
		}
		return nil
	}
}

func (c *typeChecker) validateSpecialDeclaredType(raw string, pos int, context string) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return
	}
	var inner string
	switch {
	case strings.HasPrefix(s, "chan<") && strings.HasSuffix(s, ">"):
		inner = strings.TrimSpace(s[len("chan<") : len(s)-1])
	default:
		return
	}
	if inner == "" {
		c.addDiag(pos, "invalid %s type annotation: channel payload type is required", context)
		return
	}
	payload := c.parseDeclaredType(inner)
	if payload == nil {
		c.addDiag(pos, "invalid %s type annotation: unable to parse channel payload type '%s'", context, inner)
		return
	}
	if !c.typeAllowsNil(payload) {
		c.addDiag(pos, "invalid %s type annotation: channel payload must include nil (use chan<...|nil>)", context)
	}
}

func splitTypeTopLevel(s string, sep rune) []string {
	out := []string{}
	start := 0
	angles := 0
	parens := 0
	brackets := 0
	for i, r := range s {
		switch r {
		case '<':
			angles++
		case '>':
			if angles > 0 {
				angles--
			}
		case '(':
			parens++
		case ')':
			if parens > 0 {
				parens--
			}
		case '[':
			brackets++
		case ']':
			if brackets > 0 {
				brackets--
			}
		default:
			if r == sep && angles == 0 && parens == 0 && brackets == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	out = append(out, strings.TrimSpace(s[start:]))
	return out
}

func isSimpleTypeIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || (i > 0 && r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func (c *typeChecker) addConstraint(lhs, rhs *tnode, pos int, reason string) {
	if lhs == nil || rhs == nil {
		return
	}
	c.constraints = append(c.constraints, tconstraint{lhs: lhs, rhs: rhs, pos: pos, reason: reason})
}

func (c *typeChecker) addDiag(pos int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	key := fmt.Sprintf("%d:%s", pos, msg)
	if c.diagSeen[key] {
		return
	}
	c.diagSeen[key] = true
	c.diags = append(c.diags, tdiag{pos: pos, msg: msg})
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
	c.pushOverride()
	defer c.popScope()
	defer c.popOverride()
	return c.inferBlockInCurrentScope(block)
}

func (c *typeChecker) inferBlockWithRefinementsAndOverrides(block *ast.BlockStatement, refs map[string]*tnode) (*tnode, map[string]*tnode) {
	if block == nil {
		return c.scalar(typeNil), map[string]*tnode{}
	}
	if c.hasImpossibleRefinement(refs) {
		return c.scalar(typeNever), map[string]*tnode{}
	}
	c.pushScope()
	c.pushOverride()
	for name, t := range refs {
		c.bind(name, t)
	}
	outType := c.inferBlockInCurrentScope(block)
	ov := c.snapshotCurrentOverride()
	c.popOverride()
	c.popScope()
	return outType, ov
}

func (c *typeChecker) mergeIfOverrides(outer map[string]*tnode, thenOv, elseOv map[string]*tnode, hasElse bool) {
	for name, outerType := range outer {
		thenType, okThen := thenOv[name]
		if !okThen {
			thenType = outerType
		}
		elseType := outerType
		if hasElse {
			if t, ok := elseOv[name]; ok {
				elseType = t
			}
		}
		merged := c.unionType(thenType, elseType)
		c.tracef(-1, "merge", fmt.Sprintf("%s: then=%s else=%s -> %s", name, c.describe(thenType), c.describe(elseType), c.describe(merged)))
		if c.find(merged).kind == typeAny && c.find(thenType).kind != typeAny && c.find(elseType).kind != typeAny {
			c.tracef(-1, "union-normalize", fmt.Sprintf("context=if-merge var=%s widened-to-any", name))
		}
		c.setOverride(name, merged)
	}
}

func (c *typeChecker) hasImpossibleRefinement(refs map[string]*tnode) bool {
	for _, t := range refs {
		tf := c.find(t)
		if tf != nil && tf.kind == typeNever {
			c.tracef(-1, "contradiction", "refinement collapsed to never")
			return true
		}
	}
	return false
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
		outerVisible := c.visibleBindings()
		thenRefs := c.collectGuardRefinements(e.Condition, true)
		elseRefs := c.collectGuardRefinements(e.Condition, false)
		if c.hasImpossibleRefinement(thenRefs) {
			c.addDiag(e.Token.Position, "unreachable if-branch: guard refinements are contradictory")
		}
		if c.hasImpossibleRefinement(elseRefs) {
			c.addDiag(e.Token.Position, "unreachable else-branch: guard refinements are contradictory")
		}
		thenType, thenOv := c.inferBlockWithRefinementsAndOverrides(e.ThenBranch, thenRefs)
		elseType := c.scalar(typeNil)
		elseOv := map[string]*tnode{}
		if e.ElseBranch != nil {
			elseType, elseOv = c.inferBlockWithRefinementsAndOverrides(e.ElseBranch, elseRefs)
		}
		c.mergeIfOverrides(outerVisible, thenOv, elseOv, e.ElseBranch != nil)
		return c.unionType(thenType, elseType)
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
		c.validateSpecialDeclaredType(e.Type, e.Token.Position, "var")
		if tt := c.parseDeclaredType(e.Type); tt != nil {
			c.addConstraint(rhs, tt, e.Token.Position, "var annotation")
			if !c.isDeclaredCompatible(rhs, tt) {
				c.addDiag(e.Token.Position, "inferred type mismatch (var annotation): %s vs %s", c.describe(rhs), c.describe(tt))
			}
			c.enforceDeclaredLiteralCompatibility(e.Value, tt, e.Token.Position)
			c.bindPatternDeclared(e.Pattern, tt)
			if c.typeMayBeNilConcretely(rhs) && !c.typeAllowsNil(tt) {
				c.addDiag(e.Token.Position, "nilability mismatch (var annotation): expected %s, got %s", c.describe(tt), c.describe(rhs))
			}
		}
		c.enforceTags(rhs, e.Tags, e.Token.Position)
		c.bindPattern(e.Pattern, rhs)
		c.registerStructSchema(e.Pattern, e.Value)
		return rhs
	case *ast.ValExpression:
		rhs := c.inferExpr(e.Value)
		c.validateSpecialDeclaredType(e.Type, e.Token.Position, "val")
		if tt := c.parseDeclaredType(e.Type); tt != nil {
			c.addConstraint(rhs, tt, e.Token.Position, "val annotation")
			if !c.isDeclaredCompatible(rhs, tt) {
				c.addDiag(e.Token.Position, "inferred type mismatch (val annotation): %s vs %s", c.describe(rhs), c.describe(tt))
			}
			c.enforceDeclaredLiteralCompatibility(e.Value, tt, e.Token.Position)
			c.bindPatternDeclared(e.Pattern, tt)
			if c.typeMayBeNilConcretely(rhs) && !c.typeAllowsNil(tt) {
				c.addDiag(e.Token.Position, "nilability mismatch (val annotation): expected %s, got %s", c.describe(tt), c.describe(rhs))
			}
		}
		c.enforceTags(rhs, e.Tags, e.Token.Position)
		c.bindPattern(e.Pattern, rhs)
		c.registerStructSchema(e.Pattern, e.Value)
		return rhs
	case *ast.MatchExpression:
		var scrutinee *tnode
		if e.Value != nil {
			scrutinee = c.inferExpr(e.Value)
		}
		caseResults := make([]*tnode, 0, len(e.Cases))
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
				guardRefs := c.collectGuardRefinements(cs.Guard, true)
				if c.hasImpossibleRefinement(guardRefs) {
					c.addDiag(cs.Token.Position, "unreachable match case: guard refinements are contradictory")
					c.popOverride()
					c.popScope()
					caseResults = append(caseResults, c.scalar(typeNever))
					continue
				}
				for name, t := range guardRefs {
					c.bind(name, t)
				}
			}
			bodyType := c.inferBlockInCurrentScope(cs.Body)
			c.popOverride()
			c.popScope()
			caseResults = append(caseResults, bodyType)
		}
		if len(caseResults) == 0 {
			return c.scalar(typeNil)
		}
		return c.unionType(caseResults...)
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
	paramNames := make([]string, 0, len(fn.Parameters))
	c.pushScope()
	c.pushOverride()
	for _, p := range fn.Parameters {
		pt := c.freshUnknown()
		c.validateSpecialDeclaredType(p.Type, p.Name.Token.Position, "parameter")
		if tt := c.parseDeclaredType(p.Type); tt != nil {
			c.addConstraint(pt, tt, p.Name.Token.Position, "parameter annotation constraint")
		}
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
		paramNames = append(paramNames, p.Name.Value)
	}
	ret := c.scalar(typeNil)
	if fn.Body != nil {
		iterations := 1
		if containsRecurInCurrentFunction(fn.Body) {
			iterations = 4
		}
		for i := 0; i < iterations; i++ {
			iterRet, changed := c.inferFunctionIteration(fn.Body, paramNames, params)
			ret = c.unionType(ret, iterRet)
			if !changed {
				break
			}
		}
	}
	c.validateSpecialDeclaredType(fn.ReturnType, fn.Token.Position, "function return")
	if tt := c.parseDeclaredType(fn.ReturnType); tt != nil {
		c.addConstraint(ret, tt, fn.Token.Position, "function return annotation")
	}
	c.popOverride()
	c.popScope()
	minArgs := fn.Signature.Min
	maxArgs := fn.Signature.Max
	if maxArgs > 1_000_000 {
		maxArgs = -1
	}
	return c.fnType(params, ret, len(fn.Parameters) > 0 && fn.Parameters[len(fn.Parameters)-1].IsVariadic, minArgs, maxArgs)
}

func (c *typeChecker) inferFunctionIteration(body *ast.BlockStatement, paramNames []string, params []*tnode) (*tnode, bool) {
	c.pushScope()
	c.pushOverride()
	for i, name := range paramNames {
		pt := params[i]
		c.bind(name, pt)
	}
	iterRet := c.inferBlockInCurrentScope(body)
	iterOv := c.snapshotCurrentOverride()
	c.popOverride()
	c.popScope()
	changed := false
	for i, name := range paramNames {
		cur := params[i]
		next := cur
		if ov, ok := iterOv[name]; ok {
			next = c.unionType(cur, ov)
		}
		if c.typeSig(next) != c.typeSig(cur) {
			changed = true
			params[i] = next
		}
	}
	for i, name := range paramNames {
		if len(c.scopes) > 0 {
			c.scopes[len(c.scopes)-1][name] = params[i]
		}
		c.setOverride(name, params[i])
	}
	return iterRet, changed
}

func containsRecurInCurrentFunction(block *ast.BlockStatement) bool {
	if block == nil {
		return false
	}
	for _, s := range block.Statements {
		if containsRecurInStmt(s, false) {
			return true
		}
	}
	return false
}

func containsRecurInStmt(stmt ast.Statement, nestedFn bool) bool {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return containsRecurInExpr(s.Expression, nestedFn)
	case *ast.ReturnStatement:
		return containsRecurInExpr(s.ReturnValue, nestedFn)
	case *ast.ThrowStatement:
		return containsRecurInExpr(s.Value, nestedFn)
	case *ast.BlockStatement:
		return containsRecurInCurrentFunction(s)
	case *ast.DeferStatement:
		if s.Call == nil {
			return false
		}
		return containsRecurInStmt(s.Call, nestedFn)
	default:
		return false
	}
}

func containsRecurInExpr(expr ast.Expression, nestedFn bool) bool {
	switch e := expr.(type) {
	case *ast.RecurExpression:
		return !nestedFn
	case *ast.IfExpression:
		if containsRecurInExpr(e.Condition, nestedFn) {
			return true
		}
		return containsRecurInCurrentFunction(e.ThenBranch) || containsRecurInCurrentFunction(e.ElseBranch)
	case *ast.InfixExpression:
		return containsRecurInExpr(e.Left, nestedFn) || containsRecurInExpr(e.Right, nestedFn)
	case *ast.PrefixExpression:
		return containsRecurInExpr(e.Right, nestedFn)
	case *ast.CallExpression:
		if containsRecurInExpr(e.Function, nestedFn) {
			return true
		}
		for _, a := range e.Arguments {
			if containsRecurInExpr(a, nestedFn) {
				return true
			}
		}
		return false
	case *ast.FunctionLiteral:
		// Nested function literals should not trigger fixed-point iterations
		// for the outer function.
		return containsRecurInExpr(e.Body, true)
	case *ast.MatchExpression:
		if containsRecurInExpr(e.Value, nestedFn) {
			return true
		}
		for _, cs := range e.Cases {
			if cs == nil {
				continue
			}
			if containsRecurInExpr(cs.Guard, nestedFn) || containsRecurInCurrentFunction(cs.Body) {
				return true
			}
		}
		return false
	case *ast.ListLiteral:
		for _, it := range e.Elements {
			if containsRecurInExpr(it, nestedFn) {
				return true
			}
		}
		return false
	case *ast.MapLiteral:
		for k, v := range e.Pairs {
			if containsRecurInExpr(k, nestedFn) || containsRecurInExpr(v, nestedFn) {
				return true
			}
		}
		return false
	case *ast.IndexExpression:
		return containsRecurInExpr(e.Left, nestedFn) || containsRecurInExpr(e.Index, nestedFn)
	case *ast.StructInitExpression:
		if containsRecurInExpr(e.Schema, nestedFn) {
			return true
		}
		for _, f := range e.Fields {
			if containsRecurInExpr(f.Value, nestedFn) {
				return true
			}
		}
		return false
	case *ast.StructCopyExpression:
		return containsRecurInExpr(e.Source, nestedFn) || containsRecurInExpr(e.Fields, nestedFn)
	case *ast.VarExpression:
		return containsRecurInExpr(e.Value, nestedFn)
	case *ast.ValExpression:
		return containsRecurInExpr(e.Value, nestedFn)
	case *ast.SpawnExpression:
		return containsRecurInExpr(e.Body, nestedFn)
	case *ast.BlockStatement:
		return containsRecurInCurrentFunction(e)
	default:
		return false
	}
}

func (c *typeChecker) inferCall(call *ast.CallExpression) *tnode {
	callee := c.inferExpr(call.Function)
	result := c.freshUnknown()
	type callArg struct {
		t        *tnode
		isSpread bool
	}
	args := make([]callArg, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		if spread, ok := arg.(*ast.SpreadExpression); ok {
			args = append(args, callArg{t: c.inferExpr(spread.Value), isSpread: true})
			continue
		}
		args = append(args, callArg{t: c.inferExpr(arg)})
	}
	ct := c.find(callee)
	if ct.kind == typeFn {
		hasSpread := false
		fixedArgs := 0
		for _, a := range args {
			if a.isSpread {
				hasSpread = true
				continue
			}
			fixedArgs++
		}
		minArgs := ct.minArgs
		maxArgs := ct.maxArgs
		if maxArgs == 0 && len(ct.params) > 0 && !ct.variadic {
			maxArgs = len(ct.params)
		}
		if maxArgs == 0 && len(ct.params) == 0 {
			maxArgs = -1
		}
		if !hasSpread {
			argc := len(args)
			if argc < minArgs || (maxArgs >= 0 && argc > maxArgs) {
				if maxArgs >= 0 {
					c.addDiag(call.Token.Position, "call arity mismatch: expected %d..%d arguments, got %d", minArgs, maxArgs, argc)
				} else {
					c.addDiag(call.Token.Position, "call arity mismatch: expected at least %d arguments, got %d", minArgs, argc)
				}
			}
		} else if maxArgs >= 0 && fixedArgs > maxArgs {
			// Spread can contribute 0..N args, so only reject when fixed args
			// already exceed the maximum possible arity.
			c.addDiag(call.Token.Position, "call arity mismatch: expected %d..%d arguments, got at least %d", minArgs, maxArgs, fixedArgs)
		}
		paramIdx := 0
		for _, arg := range args {
			if paramIdx >= len(ct.params) {
				break
			}
			if !arg.isSpread {
				c.callChecks = append(c.callChecks, callCheck{
					pos:      call.Token.Position,
					got:      arg.t,
					expected: ct.params[paramIdx],
				})
				paramIdx++
				continue
			}
			elem := c.freshUnknown()
			at := c.find(arg.t)
			switch at.kind {
			case typeList:
				elem = at.elem
			case typeBytes:
				elem = c.scalar(typeNum)
			case typeStr:
				elem = c.scalar(typeStr)
			}
			for ; paramIdx < len(ct.params); paramIdx++ {
				c.callChecks = append(c.callChecks, callCheck{
					pos:      call.Token.Position,
					got:      elem,
					expected: ct.params[paramIdx],
				})
			}
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
		c.trackReassignmentType(e.Left, right, e.Token.Position)
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
		// Defer bitwise mode compatibility to post-unification to avoid eager
		// numeric pinning when operands are unresolved during inference.
		c.bitChecks = append(c.bitChecks, bitwiseCheck{pos: e.Token.Position, left: left, right: right})
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
			if lf.kind == typeUnion {
				hasList, hasBytes, hasOther := c.unionContainerModes(lf)
				// Ambiguous unions that may be list/bytes should not fail eagerly.
				if hasList || hasBytes {
					if hasBytes && !hasList {
						c.addConstraint(right, c.scalar(typeNum), e.Token.Position, "bytes append operator")
						c.addConstraint(out, c.scalar(typeBytes), e.Token.Position, "bytes append result")
						return out
					}
					c.addConstraint(out, c.freshUnknown(), e.Token.Position, "list concat result")
					return out
				}
				if !hasOther {
					c.addConstraint(out, c.freshUnknown(), e.Token.Position, "list concat result")
					return out
				}
			}
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
		if rf.kind == typeUnion {
			hasList, hasBytes, hasOther := c.unionContainerModes(rf)
			if hasList || hasBytes {
				if hasBytes && !hasList {
					c.addConstraint(left, c.scalar(typeNum), e.Token.Position, "bytes prepend operator")
					c.addConstraint(out, c.scalar(typeBytes), e.Token.Position, "bytes prepend result")
					return out
				}
				c.addConstraint(out, c.freshUnknown(), e.Token.Position, "list concat result")
				return out
			}
			if !hasOther {
				c.addConstraint(out, c.freshUnknown(), e.Token.Position, "list concat result")
				return out
			}
		}
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

func (c *typeChecker) unionContainerModes(t *tnode) (hasList bool, hasBytes bool, hasOther bool) {
	tf := c.find(t)
	if tf == nil || tf.kind != typeUnion {
		return false, false, false
	}
	for _, opt := range tf.options {
		of := c.find(opt)
		if of == nil {
			continue
		}
		switch of.kind {
		case typeList:
			hasList = true
		case typeBytes:
			hasBytes = true
		case typeUnknown, typeAny, typeNil, typeNever:
			// keep permissive for unresolved/flow states
		default:
			hasOther = true
		}
	}
	return hasList, hasBytes, hasOther
}

func (c *typeChecker) trackReassignmentType(lhs ast.Expression, rhs *tnode, pos int) {
	id, ok := lhs.(*ast.Identifier)
	if !ok || id == nil {
		return
	}
	name := id.Value
	if name == "" {
		return
	}
	if dt := c.lookupDeclared(name); dt != nil {
		if c.typeMayBeNilConcretely(rhs) && !c.typeAllowsNil(dt) {
			c.addDiag(pos, "nilability mismatch (assignment): expected %s, got %s", c.describe(dt), c.describe(rhs))
		}
	}
	for i := len(c.scopes) - 1; i >= 0; i-- {
		cur, exists := c.scopes[i][name]
		if !exists || cur == nil {
			continue
		}
		// Lightweight path-sensitive widening: preserve prior possibilities
		// and include assigned value shape to model mutable var evolution.
		merged := c.unionType(cur, rhs)
		c.tracef(pos, "widen", fmt.Sprintf("%s: %s + %s -> %s", name, c.describe(cur), c.describe(rhs), c.describe(merged)))
		if c.find(merged).kind == typeAny && c.find(cur).kind != typeAny && c.find(rhs).kind != typeAny {
			c.tracef(pos, "union-normalize", fmt.Sprintf("context=assignment var=%s widened-to-any", name))
		}
		c.scopes[i][name] = merged
		c.setOverride(name, rhs)
		return
	}
}

func (c *typeChecker) bindPatternDeclared(pattern ast.MatchPattern, declared *tnode) {
	switch p := pattern.(type) {
	case *ast.IdentifierPattern:
		if p != nil && p.Value != nil {
			c.bindDeclared(p.Value.Value, declared)
		}
	case *ast.BindingPattern:
		if p.Name != nil {
			c.bindDeclared(p.Name.Value, declared)
		}
		if p.Pattern != nil {
			c.bindPatternDeclared(p.Pattern, declared)
		}
	case *ast.ListPattern:
		for _, el := range p.Elements {
			c.bindPatternDeclared(el, declared)
		}
	case *ast.MapPattern:
		for _, pair := range p.Pairs {
			c.bindPatternDeclared(pair.Pattern, declared)
		}
		if p.Spread != nil {
			c.bindPatternDeclared(p.Spread, declared)
		}
	case *ast.StructPattern:
		for _, f := range p.Fields {
			c.bindPatternDeclared(f.Pattern, declared)
		}
	case *ast.SpreadPattern:
		if p.Value != nil {
			c.bindDeclared(p.Value.Value, declared)
		}
	}
}

func (c *typeChecker) typeAllowsNil(t *tnode) bool {
	tf := c.find(t)
	if tf == nil {
		return true
	}
	switch tf.kind {
	case typeAny, typeUnknown, typeNil:
		return true
	case typeUnion:
		for _, opt := range tf.options {
			if c.typeAllowsNil(opt) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (c *typeChecker) typeMayBeNil(t *tnode) bool {
	tf := c.find(t)
	if tf == nil {
		return false
	}
	switch tf.kind {
	case typeNil, typeUnknown, typeAny:
		return true
	case typeUnion:
		for _, opt := range tf.options {
			if c.typeMayBeNil(opt) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (c *typeChecker) typeMayBeNilConcretely(t *tnode) bool {
	tf := c.find(t)
	if tf == nil {
		return false
	}
	switch tf.kind {
	case typeNil:
		return true
	case typeUnknown, typeAny:
		return false
	case typeUnion:
		hasNil := false
		for _, opt := range tf.options {
			of := c.find(opt)
			if of == nil {
				continue
			}
			if of.kind == typeUnknown || of.kind == typeAny {
				return false
			}
			if c.typeMayBeNilConcretely(of) {
				hasNil = true
			}
		}
		return hasNil
	default:
		return false
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
		if strings.TrimSpace(field.Type) == "" && (field.Default == nil || isNilExpr(field.Default)) {
			// Untagged open field: keep dynamic by default.
			ft = c.scalar(typeAny)
		}
		if strings.TrimSpace(field.Type) != "" {
			c.validateSpecialDeclaredType(field.Type, field.Token.Position, "struct field")
			if tt := c.parseDeclaredType(field.Type); tt != nil {
				c.addConstraint(ft, tt, field.Token.Position, fmt.Sprintf("struct schema type %s.%s", name, field.Name))
			}
		}
		if field.Default != nil && !isNilExpr(field.Default) {
			dt := c.inferExpr(field.Default)
			c.addConstraint(ft, dt, field.Token.Position, fmt.Sprintf("struct schema default %s.%s", name, field.Name))
		}
		fieldTypes[field.Name] = ft
	}
	c.schemas[name] = fieldTypes
}

func isNilExpr(expr ast.Expression) bool {
	if expr == nil {
		return true
	}
	_, ok := expr.(*ast.Nil)
	return ok
}

func (c *typeChecker) collectGuardRefinements(expr ast.Expression, whenTrue bool) map[string]*tnode {
	out := map[string]*tnode{}
	c.collectGuardRefinementsInto(expr, whenTrue, out)
	return out
}

func (c *typeChecker) collectGuardRefinementsInto(expr ast.Expression, whenTrue bool, out map[string]*tnode) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.PrefixExpression:
		if e.Operator == "!" {
			c.collectGuardRefinementsInto(e.Right, !whenTrue, out)
		}
	case *ast.InfixExpression:
		switch e.Operator {
		case "&&":
			if whenTrue {
				c.collectGuardRefinementsInto(e.Left, true, out)
				c.collectGuardRefinementsInto(e.Right, true, out)
				return
			}
			leftFalse := c.collectGuardRefinements(e.Left, false)
			rightFalse := c.collectGuardRefinements(e.Right, false)
			for k, v := range c.intersectRefinements(leftFalse, rightFalse) {
				out[k] = v
			}
		case "||":
			if !whenTrue {
				c.collectGuardRefinementsInto(e.Left, false, out)
				c.collectGuardRefinementsInto(e.Right, false, out)
				return
			}
			leftTrue := c.collectGuardRefinements(e.Left, true)
			rightTrue := c.collectGuardRefinements(e.Right, true)
			for k, v := range c.intersectRefinements(leftTrue, rightTrue) {
				out[k] = v
			}
		case "==", "!=":
			c.collectComparisonRefinement(e.Left, e.Right, e.Operator, whenTrue, out)
			c.collectLenRefinement(e.Left, e.Right, e.Operator, whenTrue, out)
		case ">", ">=", "<", "<=":
			c.collectLenRefinement(e.Left, e.Right, e.Operator, whenTrue, out)
		}
	case *ast.CallExpression:
		c.collectPredicateCallRefinement(e, whenTrue, out)
	}
}

func (c *typeChecker) intersectRefinements(a, b map[string]*tnode) map[string]*tnode {
	out := map[string]*tnode{}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			continue
		}
		merged := c.unionType(av, bv)
		out[k] = merged
	}
	return out
}

func (c *typeChecker) collectComparisonRefinement(left, right ast.Expression, op string, whenTrue bool, out map[string]*tnode) {
	eq := (op == "==" && whenTrue) || (op == "!=" && !whenTrue)
	neq := (op == "!=" && whenTrue) || (op == "==" && !whenTrue)

	if name, ok := identifierName(left); ok && isNilExpr(right) {
		if eq {
			c.refineBinding(name, c.scalar(typeNil), true, out)
		}
		if neq {
			c.refineBindingExclude(name, c.scalar(typeNil), out)
		}
		return
	}
	if name, ok := identifierName(right); ok && isNilExpr(left) {
		if eq {
			c.refineBinding(name, c.scalar(typeNil), true, out)
		}
		if neq {
			c.refineBindingExclude(name, c.scalar(typeNil), out)
		}
		return
	}

	if name, tk, ok := c.parseTypePredicate(left, right); ok {
		target := c.scalar(tk)
		if eq {
			c.refineBinding(name, target, true, out)
			return
		}
		if neq {
			c.refineBindingExclude(name, target, out)
			return
		}
	}
	if name, tk, ok := c.parseTypePredicate(right, left); ok {
		target := c.scalar(tk)
		if eq {
			c.refineBinding(name, target, true, out)
			return
		}
		if neq {
			c.refineBindingExclude(name, target, out)
			return
		}
	}
}

func (c *typeChecker) collectPredicateCallRefinement(call *ast.CallExpression, whenTrue bool, out map[string]*tnode) {
	if call == nil {
		return
	}
	fn, ok := call.Function.(*ast.Identifier)
	if !ok || fn == nil || len(call.Arguments) != 1 {
		return
	}
	name, ok := identifierName(call.Arguments[0])
	if !ok {
		return
	}
	var target *tnode
	switch fn.Value {
	case "isList":
		target = c.listType(c.scalar(typeAny))
	case "isMap":
		target = c.mapType(c.scalar(typeAny), c.scalar(typeAny))
	case "isStruct":
		target = &tnode{kind: typeStruct}
	case "isFn":
		target = c.fnType(nil, c.freshUnknown(), false, 0, -1)
	case "isBytes":
		target = c.scalar(typeBytes)
	case "isStr", "isString":
		target = c.scalar(typeStr)
	case "isNum", "isNumber":
		target = c.scalar(typeNum)
	case "isBool", "isBoolean":
		target = c.scalar(typeBool)
	default:
		return
	}
	if whenTrue {
		c.refineBinding(name, target, true, out)
		return
	}
	c.refineBindingExclude(name, target, out)
}

func (c *typeChecker) collectLenRefinement(left, right ast.Expression, op string, whenTrue bool, out map[string]*tnode) {
	name, n, cmpOp, ok := parseLenComparison(left, right, op)
	if !ok {
		return
	}
	base := c.lookup(name)
	if base == nil {
		return
	}
	// Conservative shape refinement only:
	// len(x) > 0, >=1, !=0  => x is sequence/map-like in true branch
	// len(x) == 0           => x is sequence/map-like in true branch
	// len(x) == 0 false     => len(x) != 0 (sequence/map-like) but no element info.
	seqMap := c.unionType(
		c.listType(c.scalar(typeAny)),
		c.mapType(c.scalar(typeAny), c.scalar(typeAny)),
		c.scalar(typeBytes),
		c.scalar(typeStr),
	)
	holds := evalLenPredicate(n, cmpOp, whenTrue)
	if holds {
		c.refineBinding(name, seqMap, true, out)
	}
}

func parseLenComparison(left, right ast.Expression, op string) (string, int, string, bool) {
	if name, ok := parseLenCallName(left); ok {
		if n, ok := parseIntLiteral(right); ok {
			return name, n, op, ok
		}
	}
	if name, ok := parseLenCallName(right); ok {
		if n, ok := parseIntLiteral(left); ok {
			rop := reverseCmpOp(op)
			if rop == "" {
				return "", 0, "", false
			}
			return name, n, rop, true
		}
	}
	return "", 0, "", false
}

func reverseCmpOp(op string) string {
	switch op {
	case ">":
		return "<"
	case ">=":
		return "<="
	case "<":
		return ">"
	case "<=":
		return ">="
	case "==", "!=":
		return op
	default:
		return ""
	}
}

func parseLenCallName(expr ast.Expression) (string, bool) {
	call, ok := expr.(*ast.CallExpression)
	if !ok || call == nil {
		return "", false
	}
	fn, ok := call.Function.(*ast.Identifier)
	if !ok || fn == nil || fn.Value != "len" || len(call.Arguments) != 1 {
		return "", false
	}
	return identifierName(call.Arguments[0])
}

func parseIntLiteral(expr ast.Expression) (int, bool) {
	n, ok := expr.(*ast.NumberLiteral)
	if !ok || n == nil {
		return 0, false
	}
	if n.Value.IsFloat() {
		return 0, false
	}
	return n.Value.ToInt(), true
}

func evalLenPredicate(n int, op string, whenTrue bool) bool {
	switch op {
	case ">":
		return (n >= 0 && whenTrue) || (n < 0 && !whenTrue)
	case ">=":
		return (n > 0 && whenTrue) || (n <= 0 && !whenTrue)
	case "<":
		return (n <= 0 && whenTrue) || (n > 0 && !whenTrue)
	case "<=":
		return (n < 0 && whenTrue) || (n >= 0 && !whenTrue)
	case "==":
		return n == 0
	case "!=":
		return n == 0
	default:
		return false
	}
}

func identifierName(expr ast.Expression) (string, bool) {
	id, ok := expr.(*ast.Identifier)
	if !ok || id == nil {
		return "", false
	}
	return id.Value, true
}

func (c *typeChecker) parseTypePredicate(typeExpr, typeConst ast.Expression) (string, typeKind, bool) {
	call, ok := typeExpr.(*ast.CallExpression)
	if !ok || call == nil {
		return "", "", false
	}
	fn, ok := call.Function.(*ast.Identifier)
	if !ok || fn == nil || fn.Value != "type" || len(call.Arguments) != 1 {
		return "", "", false
	}
	argName, ok := identifierName(call.Arguments[0])
	if !ok {
		return "", "", false
	}
	if k, ok := typeConstToKind(typeConst); ok {
		return argName, k, true
	}
	return "", "", false
}

func typeConstToKind(expr ast.Expression) (typeKind, bool) {
	switch e := expr.(type) {
	case *ast.Identifier:
		switch strings.ToUpper(e.Value) {
		case "NIL_TYPE":
			return typeNil, true
		case "BOOLEAN_TYPE":
			return typeBool, true
		case "NUMBER_TYPE":
			return typeNum, true
		case "STRING_TYPE":
			return typeStr, true
		case "LIST_TYPE":
			return typeList, true
		case "MAP_TYPE":
			return typeMap, true
		case "FUNCTION_TYPE":
			return typeFn, true
		case "TASK_TYPE":
			return typeTask, true
		case "CHAN_TYPE", "CHANNEL_TYPE":
			return typeChan, true
		case "STRUCT_TYPE":
			return typeStruct, true
		case "BYTES_TYPE":
			return typeBytes, true
		case "SYMBOL_TYPE":
			return typeSym, true
		}
	case *ast.SymbolLiteral:
		switch strings.ToLower(e.Value) {
		case "nil":
			return typeNil, true
		case "bool", "boolean":
			return typeBool, true
		case "number", "num":
			return typeNum, true
		case "string", "str":
			return typeStr, true
		case "list":
			return typeList, true
		case "map":
			return typeMap, true
		case "function", "fn":
			return typeFn, true
		case "task":
			return typeTask, true
		case "chan", "channel":
			return typeChan, true
		case "struct":
			return typeStruct, true
		case "bytes":
			return typeBytes, true
		case "symbol", "sym":
			return typeSym, true
		}
	}
	return "", false
}

func (c *typeChecker) refineBinding(name string, target *tnode, _ bool, out map[string]*tnode) {
	base, fromRefinement := c.currentRefinementBase(name, out)
	if base == nil {
		return
	}
	narrowed := c.intersectType(base, target)
	if narrowed == nil {
		if fromRefinement {
			out[name] = c.scalar(typeNever)
			c.tracef(-1, "contradiction", fmt.Sprintf("refine %s: %s ∩ %s -> never", name, c.describe(base), c.describe(target)))
		}
		return
	}
	c.tracef(-1, "refine", fmt.Sprintf("%s: %s -> %s (target=%s)", name, c.describe(base), c.describe(narrowed), c.describe(target)))
	out[name] = narrowed
}

func (c *typeChecker) refineBindingExclude(name string, target *tnode, out map[string]*tnode) {
	base, fromRefinement := c.currentRefinementBase(name, out)
	if base == nil {
		return
	}
	narrowed := c.excludeType(base, target)
	if narrowed == nil {
		if fromRefinement {
			out[name] = c.scalar(typeNever)
			c.tracef(-1, "contradiction", fmt.Sprintf("exclude %s: %s minus %s -> never", name, c.describe(base), c.describe(target)))
		}
		return
	}
	c.tracef(-1, "refine", fmt.Sprintf("%s: %s exclude %s -> %s", name, c.describe(base), c.describe(target), c.describe(narrowed)))
	out[name] = narrowed
}

func (c *typeChecker) currentRefinementBase(name string, out map[string]*tnode) (*tnode, bool) {
	if t, ok := out[name]; ok && t != nil {
		return t, true
	}
	return c.lookup(name), false
}

func (c *typeChecker) intersectType(base, target *tnode) *tnode {
	b := c.find(base)
	t := c.find(target)
	if b == nil || t == nil {
		return base
	}
	if b.kind == typeAny || b.kind == typeUnknown {
		return t
	}
	if t.kind == typeAny || t.kind == typeUnknown {
		return b
	}
	if b.kind == typeUnion {
		opts := make([]*tnode, 0, len(b.options))
		for _, opt := range b.options {
			if c.matchesRefinementTarget(opt, t) {
				opts = append(opts, opt)
			}
		}
		if len(opts) == 0 {
			return nil
		}
		return c.unionType(opts...)
	}
	if c.matchesRefinementTarget(b, t) {
		return b
	}
	return nil
}

func (c *typeChecker) excludeType(base, target *tnode) *tnode {
	b := c.find(base)
	t := c.find(target)
	if b == nil || t == nil {
		return base
	}
	if b.kind == typeAny || b.kind == typeUnknown {
		return b
	}
	if b.kind == typeUnion {
		opts := make([]*tnode, 0, len(b.options))
		for _, opt := range b.options {
			if c.matchesRefinementTarget(opt, t) {
				continue
			}
			opts = append(opts, opt)
		}
		if len(opts) == 0 {
			return nil
		}
		return c.unionType(opts...)
	}
	if c.matchesRefinementTarget(b, t) {
		return nil
	}
	return b
}

func (c *typeChecker) matchesRefinementTarget(candidate, target *tnode) bool {
	cf := c.find(candidate)
	tf := c.find(target)
	if cf == nil || tf == nil {
		return false
	}
	if tf.kind == typeUnion {
		for _, opt := range tf.options {
			if c.matchesRefinementTarget(cf, opt) {
				return true
			}
		}
		return false
	}
	if tf.kind == typeAny || tf.kind == typeUnknown {
		return true
	}
	if cf.kind != tf.kind {
		return false
	}
	switch tf.kind {
	case typeStruct:
		if tf.name == "" {
			return true
		}
		if cf.name == "" {
			return false
		}
		return cf.name == tf.name
	case typeList:
		if tf.elem == nil {
			return true
		}
		return c.matchesRefinementTarget(cf.elem, tf.elem)
	case typeMap:
		keyOK := true
		valOK := true
		if tf.key != nil {
			keyOK = c.matchesRefinementTarget(cf.key, tf.key)
		}
		if tf.val != nil {
			valOK = c.matchesRefinementTarget(cf.val, tf.val)
		}
		return keyOK && valOK
	case typeFn:
		return true
	default:
		return true
	}
}

func (c *typeChecker) sameConcreteType(a, b *tnode) bool {
	af := c.find(a)
	bf := c.find(b)
	if af == nil || bf == nil {
		return false
	}
	if af.kind != bf.kind {
		return false
	}
	switch af.kind {
	case typeStruct:
		if af.name != "" && bf.name != "" {
			return af.name == bf.name
		}
		return true
	case typeList:
		return c.sameConcreteType(af.elem, bf.elem)
	case typeMap:
		return c.sameConcreteType(af.key, bf.key) && c.sameConcreteType(af.val, bf.val)
	case typeFn:
		return true
	case typeUnion:
		if len(af.options) != len(bf.options) {
			return false
		}
		for i := range af.options {
			if !c.sameConcreteType(af.options[i], bf.options[i]) {
				return false
			}
		}
		return true
	default:
		return true
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
	for _, bc := range c.bitChecks {
		if !c.isBitwiseCompatible(bc.left, bc.right) {
			c.addDiag(bc.pos, "bitwise operator type mismatch: %s vs %s", c.describe(bc.left), c.describe(bc.right))
		}
	}
}

func (c *typeChecker) typeSig(t *tnode) string {
	t = c.find(t)
	if t == nil {
		return "nilnode"
	}
	switch t.kind {
	case typeList:
		return "list<" + c.typeSig(t.elem) + ">"
	case typeMap:
		return "map<" + c.typeSig(t.key) + "," + c.typeSig(t.val) + ">"
	case typeFn:
		return "fn"
	case typeStruct:
		if t.name == "" {
			return "struct"
		}
		return "struct(" + t.name + ")"
	case typeUnion:
		out := "union("
		for i, opt := range t.options {
			if i > 0 {
				out += "|"
			}
			out += c.typeSig(opt)
		}
		return out + ")"
	default:
		return string(t.kind)
	}
}

func (c *typeChecker) isCompatible(got, expected *tnode) bool {
	g := c.find(got)
	e := c.find(expected)
	if g == nil || e == nil {
		return true
	}
	if g.kind == typeNever || e.kind == typeNever {
		return true
	}
	if g == e {
		return true
	}
	if g.kind == typeAny || e.kind == typeAny || g.kind == typeUnknown || e.kind == typeUnknown {
		return true
	}
	if g.kind == typeUnion {
		for _, opt := range g.options {
			if !c.isCompatible(opt, e) {
				return false
			}
		}
		return true
	}
	if e.kind == typeUnion {
		for _, opt := range e.options {
			if c.isCompatible(g, opt) {
				return true
			}
		}
		return false
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

// isDeclaredCompatible applies stricter matching for declared type annotations.
// In particular, nil only satisfies explicit nil branches (e.g. T|nil),
// which lets generic container declarations enforce element/value unions.
func (c *typeChecker) isDeclaredCompatible(got, expected *tnode) bool {
	g := c.find(got)
	e := c.find(expected)
	if g == nil || e == nil {
		return true
	}
	if g.kind == typeNever || e.kind == typeNever {
		return true
	}
	if g == e {
		return true
	}
	if g.kind == typeAny || e.kind == typeAny || g.kind == typeUnknown || e.kind == typeUnknown {
		return true
	}
	if g.kind == typeUnion {
		for _, opt := range g.options {
			if !c.isDeclaredCompatible(opt, e) {
				return false
			}
		}
		return true
	}
	if e.kind == typeUnion {
		for _, opt := range e.options {
			if c.isDeclaredCompatible(g, opt) {
				return true
			}
		}
		return false
	}
	if g.kind == typeNil || e.kind == typeNil {
		return g.kind == typeNil && e.kind == typeNil
	}
	if g.kind != e.kind {
		return false
	}
	switch g.kind {
	case typeList:
		return c.isDeclaredCompatible(g.elem, e.elem)
	case typeMap:
		return c.isDeclaredCompatible(g.key, e.key) && c.isDeclaredCompatible(g.val, e.val)
	case typeFn:
		if (e.maxArgs == -1 && len(e.params) == 0) || (g.maxArgs == -1 && len(g.params) == 0) {
			return true
		}
		if g.variadic != e.variadic || len(g.params) != len(e.params) {
			return false
		}
		for i := range g.params {
			if !c.isDeclaredCompatible(g.params[i], e.params[i]) {
				return false
			}
		}
		return c.isDeclaredCompatible(g.ret, e.ret)
	case typeStruct:
		if g.name != "" && e.name != "" && g.name != e.name {
			return false
		}
		return true
	default:
		return true
	}
}

func (c *typeChecker) enforceDeclaredLiteralCompatibility(expr ast.Expression, declared *tnode, pos int) {
	d := c.find(declared)
	if d == nil || expr == nil {
		return
	}
	switch lit := expr.(type) {
	case *ast.ListLiteral:
		if d.kind != typeList {
			return
		}
		for _, item := range lit.Elements {
			it := c.inferExpr(item)
			if !c.isDeclaredCompatible(it, d.elem) {
				c.addDiag(pos, "inferred type mismatch (list element type): %s vs %s", c.describe(it), c.describe(d.elem))
			}
		}
	case *ast.MapLiteral:
		if d.kind != typeMap {
			return
		}
		for k, v := range lit.Pairs {
			kt := c.inferExpr(k)
			vt := c.inferExpr(v)
			if !c.isDeclaredCompatible(kt, d.key) {
				c.addDiag(pos, "inferred type mismatch (map key type): %s vs %s", c.describe(kt), c.describe(d.key))
			}
			if !c.isDeclaredCompatible(vt, d.val) {
				c.addDiag(pos, "inferred type mismatch (map value type): %s vs %s", c.describe(vt), c.describe(d.val))
			}
		}
	}
}

func (c *typeChecker) isPlusCompatible(left, right *tnode) bool {
	l := c.find(left)
	r := c.find(right)
	if l == nil || r == nil {
		return true
	}
	if l.kind == typeNever || r.kind == typeNever {
		return true
	}
	// Be permissive with unresolved types.
	if l.kind == typeAny || r.kind == typeAny || l.kind == typeUnknown || r.kind == typeUnknown {
		return true
	}
	if l.kind == typeUnion {
		for _, opt := range l.options {
			if !c.isPlusCompatible(opt, r) {
				return false
			}
		}
		return true
	}
	if r.kind == typeUnion {
		for _, opt := range r.options {
			if !c.isPlusCompatible(l, opt) {
				return false
			}
		}
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
	if l.kind == typeNever || r.kind == typeNever {
		return true
	}
	if l.kind == typeAny || r.kind == typeAny || l.kind == typeUnknown || r.kind == typeUnknown {
		return true
	}
	if l.kind == typeUnion {
		for _, opt := range l.options {
			if !c.isMulCompatible(opt, r) {
				return false
			}
		}
		return true
	}
	if r.kind == typeUnion {
		for _, opt := range r.options {
			if !c.isMulCompatible(l, opt) {
				return false
			}
		}
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

func (c *typeChecker) isBitwiseCompatible(left, right *tnode) bool {
	l := c.find(left)
	r := c.find(right)
	if l == nil || r == nil {
		return true
	}
	if l.kind == typeNever || r.kind == typeNever {
		return true
	}
	if l.kind == typeAny || r.kind == typeAny || l.kind == typeUnknown || r.kind == typeUnknown {
		return true
	}
	if l.kind == typeUnion {
		for _, opt := range l.options {
			if !c.isBitwiseCompatible(opt, r) {
				return false
			}
		}
		return true
	}
	if r.kind == typeUnion {
		for _, opt := range r.options {
			if !c.isBitwiseCompatible(l, opt) {
				return false
			}
		}
		return true
	}
	// runtime:
	// - num <op> num
	// - bytes <op> bytes
	// - bytes <op> num / num <op> bytes
	if l.kind == typeNum && r.kind == typeNum {
		return true
	}
	if l.kind == typeBytes && r.kind == typeBytes {
		return true
	}
	if l.kind == typeBytes && r.kind == typeNum {
		return true
	}
	if l.kind == typeNum && r.kind == typeBytes {
		return true
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
	if a.kind == typeNever || b.kind == typeNever {
		return true
	}
	if a == b {
		return true
	}
	if a.kind == typeAny || b.kind == typeAny {
		return true
	}
	if a.kind == typeUnion {
		for _, opt := range a.options {
			if c.isCompatible(opt, b) {
				return true
			}
		}
		return false
	}
	if b.kind == typeUnion {
		for _, opt := range b.options {
			if c.isCompatible(a, opt) {
				return true
			}
		}
		return false
	}
	// Slug runtime/tag dispatch treats nil as admissible across typed positions.
	// Keep semantic typing aligned by allowing nil to unify with any concrete type.
	if a.kind == typeNil || b.kind == typeNil {
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
	case typeNever:
		return "never"
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
	case typeUnion:
		opts := make([]*tnode, len(t.options))
		copy(opts, t.options)
		sort.Slice(opts, func(i, j int) bool {
			return c.typeSig(opts[i]) < c.typeSig(opts[j])
		})
		out := "union<"
		for i, opt := range opts {
			if i > 0 {
				out += " | "
			}
			out += c.describe(opt)
		}
		return out + ">"
	default:
		return string(t.kind)
	}
}

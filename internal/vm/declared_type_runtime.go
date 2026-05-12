package vm

import (
	"slug/internal/ast"
	"slug/internal/object"
	"sort"
	"strings"
)

type rtType struct {
	kind     string
	elem     *rtType
	key      *rtType
	val      *rtType
	params   []*rtType
	ret      *rtType
	elems    []*rtType
	options  []*rtType
	name     string
	variadic bool
}

func parseRuntimeDeclaredType(raw string) *rtType {
	return parseRuntimeDeclaredTypeWithTypeParams(raw, nil)
}

func parseRuntimeDeclaredTypeWithTypeParams(raw string, typeParams map[string]struct{}) *rtType {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	if parts := splitRuntimeTypeTopLevel(s, '|'); len(parts) > 1 {
		opts := make([]*rtType, 0, len(parts))
		for _, p := range parts {
			t := parseRuntimeDeclaredTypeWithTypeParams(p, typeParams)
			if t != nil {
				opts = append(opts, t)
			}
		}
		if len(opts) == 0 {
			return nil
		}
		return &rtType{kind: "union", options: opts}
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := strings.TrimSpace(s[1 : len(s)-1])
		if inner == "" {
			return &rtType{kind: "tuple", elems: []*rtType{}}
		}
		parts := splitRuntimeTypeTopLevel(inner, ',')
		elems := make([]*rtType, 0, len(parts))
		for _, p := range parts {
			t := parseRuntimeDeclaredTypeWithTypeParams(p, typeParams)
			if t == nil {
				t = &rtType{kind: "any"}
			}
			elems = append(elems, t)
		}
		return &rtType{kind: "tuple", elems: elems}
	}
	if strings.HasPrefix(s, "list<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("list<") : len(s)-1])
		elem := parseRuntimeDeclaredTypeWithTypeParams(inner, typeParams)
		if elem == nil {
			elem = &rtType{kind: "any"}
		}
		return &rtType{kind: "list", elem: elem}
	}
	if strings.HasPrefix(s, "map<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("map<") : len(s)-1])
		parts := splitRuntimeTypeTopLevel(inner, ',')
		if len(parts) == 2 {
			k := parseRuntimeDeclaredTypeWithTypeParams(parts[0], typeParams)
			v := parseRuntimeDeclaredTypeWithTypeParams(parts[1], typeParams)
			if k == nil {
				k = &rtType{kind: "any"}
			}
			if v == nil {
				v = &rtType{kind: "any"}
			}
			return &rtType{kind: "map", key: k, val: v}
		}
		return &rtType{kind: "map", key: &rtType{kind: "any"}, val: &rtType{kind: "any"}}
	}
	if strings.HasPrefix(s, "chan<") && strings.HasSuffix(s, ">") {
		return &rtType{kind: "chan"}
	}
	if strings.HasPrefix(s, "task<") && strings.HasSuffix(s, ">") {
		return &rtType{kind: "task"}
	}
	if strings.HasPrefix(s, "fn<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("fn<") : len(s)-1])
		parts := splitRuntimeTypeTopLevel(inner, ',')
		if len(parts) == 0 {
			return nil
		}
		ret := parseRuntimeDeclaredTypeWithTypeParams(parts[0], typeParams)
		if ret == nil {
			ret = &rtType{kind: "any"}
		}
		if len(parts) == 1 {
			return &rtType{kind: "fn", ret: ret}
		}
		params := make([]*rtType, 0, len(parts)-1)
		for _, p := range parts[1:] {
			t := parseRuntimeDeclaredTypeWithTypeParams(p, typeParams)
			if t == nil {
				t = &rtType{kind: "any"}
			}
			params = append(params, t)
		}
		return &rtType{kind: "fn", ret: ret, params: params, variadic: false}
	}
	if strings.HasPrefix(s, "struct<") && strings.HasSuffix(s, ">") {
		name := strings.TrimSpace(s[len("struct<") : len(s)-1])
		return &rtType{kind: "struct", name: name}
	}
	switch s {
	case "any", "?":
		return &rtType{kind: "any"}
	case "nil":
		return &rtType{kind: "nil"}
	case "bool":
		return &rtType{kind: "bool"}
	case "num":
		return &rtType{kind: "num"}
	case "str":
		return &rtType{kind: "str"}
	case "bytes":
		return &rtType{kind: "bytes"}
	case "sym", "symbol":
		return &rtType{kind: "sym"}
	case "list":
		return &rtType{kind: "list", elem: &rtType{kind: "any"}}
	case "map":
		return &rtType{kind: "map", key: &rtType{kind: "any"}, val: &rtType{kind: "any"}}
	case "fn":
		return &rtType{kind: "fn"}
	case "task":
		return &rtType{kind: "task"}
	case "chan":
		return &rtType{kind: "chan"}
	case "struct":
		return &rtType{kind: "struct"}
	default:
		if typeParams != nil {
			if _, ok := typeParams[s]; ok {
				return &rtType{kind: "typeparam", name: s}
			}
		}
		if isRuntimeSimpleTypeIdent(s) {
			return &rtType{kind: "struct", name: s}
		}
		return nil
	}
}

func splitRuntimeTypeTopLevel(s string, sep rune) []string {
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

func isRuntimeSimpleTypeIdent(s string) bool {
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

func runtimeObjectMatchesDeclaredType(v object.Object, t *rtType) bool {
	if t == nil {
		return true
	}
	if t.kind == "any" {
		return true
	}
	if t.kind == "typeparam" {
		return true
	}
	if v == nil {
		return t.kind == "nil"
	}
	if t.kind == "union" {
		for _, opt := range t.options {
			if runtimeObjectMatchesDeclaredType(v, opt) {
				return true
			}
		}
		return false
	}
	switch t.kind {
	case "nil":
		return v == object.NIL
	case "bool":
		return v.Type() == object.BOOLEAN_OBJ
	case "num":
		return v.Type() == object.NUMBER_OBJ
	case "str":
		return v.Type() == object.STRING_OBJ
	case "bytes":
		return v.Type() == object.BYTE_OBJ
	case "sym":
		return v.Type() == object.SYMBOL_OBJ
	case "list":
		l, ok := v.(*object.List)
		if !ok {
			return false
		}
		for _, el := range l.Elements {
			if !runtimeObjectMatchesDeclaredType(el, t.elem) {
				return false
			}
		}
		return true
	case "tuple":
		l, ok := v.(*object.List)
		if !ok || len(l.Elements) != len(t.elems) {
			return false
		}
		for i, et := range t.elems {
			if !runtimeObjectMatchesDeclaredType(l.Elements[i], et) {
				return false
			}
		}
		return true
	case "map":
		m, ok := v.(*object.Map)
		if !ok {
			return false
		}
		okAll := true
		m.ForEach(func(_ object.MapKey, p object.MapPair) bool {
			if !runtimeObjectMatchesDeclaredType(p.Key, t.key) || !runtimeObjectMatchesDeclaredType(p.Value, t.val) {
				okAll = false
				return false
			}
			return true
		})
		return okAll
	case "fn":
		actual := runtimeDeclaredTypeFromObject(v)
		if actual == nil {
			switch v.Type() {
			case object.FUNCTION_OBJ, object.FUNCTION_GROUP_OBJ, object.FOREIGN_OBJ:
				return true
			default:
				return false
			}
		}
		return runtimeDeclaredTypeCompatible(actual, t)
	case "task":
		return v.Type() == object.TASK_HANDLE_OBJ
	case "chan":
		return v.Type() == object.CHANNEL_OBJ
	case "struct":
		if v.Type() != object.STRUCT_OBJ && v.Type() != object.STRUCT_SCHEMA_OBJ {
			return false
		}
		if t.name == "" {
			return true
		}
		switch sv := v.(type) {
		case *object.StructValue:
			return sv.Schema != nil && sv.Schema.Name == t.name
		case *object.StructSchema:
			return sv.Name == t.name
		default:
			return false
		}
	default:
		return true
	}
}

func runtimeDeclaredTypeFromObject(v object.Object) *rtType {
	switch fn := v.(type) {
	case *object.Function:
		return runtimeDeclaredTypeFromFunction(fn.Parameters, fn.ReturnType, fn.TypeParams)
	case *object.Foreign:
		return runtimeDeclaredTypeFromFunction(fn.Parameters, fn.ReturnType, fn.TypeParams)
	case *VMFunction:
		return runtimeDeclaredTypeFromVMFunction(fn)
	case *object.FunctionGroup:
		return &rtType{kind: "fn"}
	default:
		return nil
	}
}

func runtimeDeclaredTypeFromVMFunction(fn *VMFunction) *rtType {
	if fn == nil {
		return nil
	}
	typeParams := runtimeTypeParamSet(fn.TypeParams)
	params := make([]*rtType, 0, len(fn.Parameters))
	for _, p := range fn.Parameters {
		t := parseRuntimeDeclaredTypeWithTypeParams(p.Type, typeParams)
		if t == nil {
			t = &rtType{kind: "any"}
		}
		params = append(params, t)
	}
	ret := parseRuntimeDeclaredTypeWithTypeParams(fn.ReturnType, typeParams)
	if ret == nil {
		ret = &rtType{kind: "any"}
	}
	return &rtType{kind: "fn", params: params, ret: ret, variadic: len(fn.Parameters) > 0 && fn.Parameters[len(fn.Parameters)-1].IsVariadic}
}

func runtimeDeclaredTypeFromFunction(params []*ast.FunctionParameter, returnType string, typeParams []string) *rtType {
	typeParamSet := runtimeTypeParamSet(typeParams)
	fnParams := make([]*rtType, 0, len(params))
	for _, p := range params {
		t := parseRuntimeDeclaredTypeWithTypeParams(p.Type, typeParamSet)
		if t == nil {
			t = &rtType{kind: "any"}
		}
		fnParams = append(fnParams, t)
	}
	ret := parseRuntimeDeclaredTypeWithTypeParams(returnType, typeParamSet)
	if ret == nil {
		ret = &rtType{kind: "any"}
	}
	return &rtType{kind: "fn", params: fnParams, ret: ret}
}

func runtimeTypeParamSet(typeParams []string) map[string]struct{} {
	if len(typeParams) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(typeParams))
	for _, name := range typeParams {
		if name == "" {
			continue
		}
		out[name] = struct{}{}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func runtimeTypeFromObject(v object.Object) *rtType {
	if v == nil {
		return &rtType{kind: "nil"}
	}
	switch obj := v.(type) {
	case *object.Nil:
		return &rtType{kind: "nil"}
	case *object.Boolean:
		return &rtType{kind: "bool"}
	case *object.Number:
		return &rtType{kind: "num"}
	case *object.String:
		return &rtType{kind: "str"}
	case *object.Bytes:
		return &rtType{kind: "bytes"}
	case *object.Symbol:
		return &rtType{kind: "sym"}
	case *object.List:
		return runtimeTypeFromList(obj)
	case *object.Map:
		return runtimeTypeFromMap(obj)
	case *object.StructValue:
		if obj.Schema != nil && obj.Schema.Name != "" {
			return &rtType{kind: "struct", name: obj.Schema.Name}
		}
		return &rtType{kind: "struct"}
	case *object.StructSchema:
		if obj.Name != "" {
			return &rtType{kind: "struct", name: obj.Name}
		}
		return &rtType{kind: "struct"}
	case *object.Channel:
		return &rtType{kind: "chan"}
	case *VMTaskHandle:
		return &rtType{kind: "task"}
	case *object.Function:
		return runtimeDeclaredTypeFromFunction(obj.Parameters, obj.ReturnType, obj.TypeParams)
	case *object.Foreign:
		return runtimeDeclaredTypeFromFunction(obj.Parameters, obj.ReturnType, obj.TypeParams)
	case *VMFunction:
		return runtimeDeclaredTypeFromVMFunction(obj)
	case *object.FunctionGroup:
		return &rtType{kind: "fn"}
	default:
		switch obj.Type() {
		case object.NIL_OBJ:
			return &rtType{kind: "nil"}
		case object.BOOLEAN_OBJ:
			return &rtType{kind: "bool"}
		case object.NUMBER_OBJ:
			return &rtType{kind: "num"}
		case object.STRING_OBJ:
			return &rtType{kind: "str"}
		case object.BYTE_OBJ:
			return &rtType{kind: "bytes"}
		case object.SYMBOL_OBJ:
			return &rtType{kind: "sym"}
		case object.LIST_OBJ:
			return &rtType{kind: "list", elem: &rtType{kind: "any"}}
		case object.MAP_OBJ:
			return &rtType{kind: "map", key: &rtType{kind: "any"}, val: &rtType{kind: "any"}}
		case object.STRUCT_OBJ, object.STRUCT_SCHEMA_OBJ:
			return &rtType{kind: "struct"}
		case object.CHANNEL_OBJ:
			return &rtType{kind: "chan"}
		case object.TASK_HANDLE_OBJ:
			return &rtType{kind: "task"}
		case object.FUNCTION_OBJ, object.FUNCTION_GROUP_OBJ, object.FOREIGN_OBJ:
			return &rtType{kind: "fn"}
		default:
			return nil
		}
	}
}

func runtimeTypeFromList(list *object.List) *rtType {
	if list == nil || len(list.Elements) == 0 {
		return &rtType{kind: "list", elem: &rtType{kind: "any"}}
	}
	elems := make([]*rtType, 0, len(list.Elements))
	for _, el := range list.Elements {
		t := runtimeTypeFromObject(el)
		if t == nil {
			t = &rtType{kind: "any"}
		}
		elems = append(elems, t)
	}
	if allRuntimeTypesEqual(elems) {
		return &rtType{kind: "list", elem: elems[0]}
	}
	if len(elems) <= 4 {
		return &rtType{kind: "tuple", elems: elems}
	}
	return &rtType{kind: "list", elem: runtimeRuntimeTypeUnion(elems)}
}

func runtimeTypeFromMap(m *object.Map) *rtType {
	if m == nil || m.Len() == 0 {
		return &rtType{kind: "map", key: &rtType{kind: "any"}, val: &rtType{kind: "any"}}
	}
	keys := make([]*rtType, 0, m.Len())
	vals := make([]*rtType, 0, m.Len())
	m.ForEach(func(_ object.MapKey, p object.MapPair) bool {
		kt := runtimeTypeFromObject(p.Key)
		vt := runtimeTypeFromObject(p.Value)
		if kt == nil {
			kt = &rtType{kind: "any"}
		}
		if vt == nil {
			vt = &rtType{kind: "any"}
		}
		keys = append(keys, kt)
		vals = append(vals, vt)
		return true
	})
	return &rtType{kind: "map", key: runtimeRuntimeTypeUnion(keys), val: runtimeRuntimeTypeUnion(vals)}
}

func runtimeRuntimeTypeUnion(types []*rtType) *rtType {
	if len(types) == 0 {
		return &rtType{kind: "any"}
	}
	uniq := map[string]*rtType{}
	for _, t := range types {
		if t == nil {
			t = &rtType{kind: "any"}
		}
		uniq[describeRuntimeDeclaredType(t)] = t
	}
	if len(uniq) == 1 {
		for _, t := range uniq {
			return t
		}
	}
	parts := make([]string, 0, len(uniq))
	for k := range uniq {
		parts = append(parts, k)
	}
	sort.Strings(parts)
	opts := make([]*rtType, 0, len(parts))
	for _, k := range parts {
		opts = append(opts, uniq[k])
	}
	return &rtType{kind: "union", options: opts}
}

func allRuntimeTypesEqual(types []*rtType) bool {
	if len(types) <= 1 {
		return true
	}
	first := describeRuntimeDeclaredType(types[0])
	for _, t := range types[1:] {
		if describeRuntimeDeclaredType(t) != first {
			return false
		}
	}
	return true
}

func inferRuntimeTypeBindings(expected, actual *rtType, bindings map[string]*rtType) bool {
	if expected == nil || actual == nil {
		return true
	}
	if expected.kind == "typeparam" {
		if existing, ok := bindings[expected.name]; ok {
			return runtimeDeclaredTypeCompatible(actual, existing) && runtimeDeclaredTypeCompatible(existing, actual)
		}
		bindings[expected.name] = actual
		return true
	}
	if actual.kind == "typeparam" {
		return true
	}
	if expected.kind == "union" {
		for _, opt := range expected.options {
			trial := cloneRuntimeTypeBindings(bindings)
			if inferRuntimeTypeBindings(opt, actual, trial) {
				replaceRuntimeTypeBindings(bindings, trial)
				return true
			}
		}
		return false
	}
	if actual.kind == "union" {
		return false
	}
	if expected.kind != actual.kind {
		return false
	}
	switch expected.kind {
	case "list":
		return inferRuntimeTypeBindings(expected.elem, actual.elem, bindings)
	case "tuple":
		if len(expected.elems) != len(actual.elems) {
			return false
		}
		for i := range expected.elems {
			if !inferRuntimeTypeBindings(expected.elems[i], actual.elems[i], bindings) {
				return false
			}
		}
		return true
	case "map":
		return inferRuntimeTypeBindings(expected.key, actual.key, bindings) && inferRuntimeTypeBindings(expected.val, actual.val, bindings)
	case "fn":
		if actual.variadic != expected.variadic || len(actual.params) != len(expected.params) {
			return false
		}
		for i := range expected.params {
			if !inferRuntimeTypeBindings(expected.params[i], actual.params[i], bindings) {
				return false
			}
		}
		return inferRuntimeTypeBindings(expected.ret, actual.ret, bindings)
	case "struct":
		if expected.name == "" || actual.name == "" {
			return true
		}
		return expected.name == actual.name
	default:
		return true
	}
}

func cloneRuntimeTypeBindings(bindings map[string]*rtType) map[string]*rtType {
	out := make(map[string]*rtType, len(bindings))
	for k, v := range bindings {
		out[k] = v
	}
	return out
}

func replaceRuntimeTypeBindings(dst, src map[string]*rtType) {
	for k := range dst {
		delete(dst, k)
	}
	for k, v := range src {
		dst[k] = v
	}
}

func substituteRuntimeDeclaredType(t *rtType, bindings map[string]*rtType) *rtType {
	if t == nil {
		return nil
	}
	switch t.kind {
	case "typeparam":
		if bindings != nil {
			if v, ok := bindings[t.name]; ok && v != nil {
				return v
			}
		}
		return t
	case "union":
		opts := make([]*rtType, 0, len(t.options))
		for _, opt := range t.options {
			opts = append(opts, substituteRuntimeDeclaredType(opt, bindings))
		}
		return &rtType{kind: "union", options: opts}
	case "list":
		return &rtType{kind: "list", elem: substituteRuntimeDeclaredType(t.elem, bindings)}
	case "tuple":
		elems := make([]*rtType, 0, len(t.elems))
		for _, el := range t.elems {
			elems = append(elems, substituteRuntimeDeclaredType(el, bindings))
		}
		return &rtType{kind: "tuple", elems: elems}
	case "map":
		return &rtType{kind: "map", key: substituteRuntimeDeclaredType(t.key, bindings), val: substituteRuntimeDeclaredType(t.val, bindings)}
	case "fn":
		params := make([]*rtType, 0, len(t.params))
		for _, p := range t.params {
			params = append(params, substituteRuntimeDeclaredType(p, bindings))
		}
		return &rtType{kind: "fn", ret: substituteRuntimeDeclaredType(t.ret, bindings), params: params, variadic: t.variadic}
	default:
		return t
	}
}

func runtimeDeclaredTypeForCall(returnType string, typeParams []string, params []*ast.FunctionParameter, values []object.Object) *rtType {
	if strings.TrimSpace(returnType) == "" {
		return nil
	}
	typeParamSet := runtimeTypeParamSet(typeParams)
	expected := parseRuntimeDeclaredTypeWithTypeParams(returnType, typeParamSet)
	if expected == nil {
		return nil
	}
	if len(params) == 0 || len(values) == 0 {
		return expected
	}
	bindings := runtimeTypeBindingsFromCall(params, typeParamSet, values)
	return substituteRuntimeDeclaredType(expected, bindings)
}

func runtimeTypeBindingsFromCall(params []*ast.FunctionParameter, typeParams map[string]struct{}, values []object.Object) map[string]*rtType {
	bindings := map[string]*rtType{}
	for i, p := range params {
		if i >= len(values) {
			break
		}
		if p == nil || strings.TrimSpace(p.Type) == "" {
			continue
		}
		actual := runtimeTypeFromObject(values[i])
		if actual == nil {
			continue
		}
		expected := parseRuntimeDeclaredTypeWithTypeParams(p.Type, typeParams)
		if expected == nil {
			continue
		}
		trial := cloneRuntimeTypeBindings(bindings)
		if inferRuntimeTypeBindings(expected, actual, trial) {
			replaceRuntimeTypeBindings(bindings, trial)
		}
	}
	return bindings
}

func runtimeDeclaredTypeCompatible(actual, expected *rtType) bool {
	if expected == nil || expected.kind == "any" {
		return true
	}
	if actual == nil || actual.kind == "any" {
		return true
	}
	if expected.kind == "typeparam" || actual.kind == "typeparam" {
		return true
	}
	if actual.kind != expected.kind {
		return false
	}
	switch expected.kind {
	case "union":
		for _, opt := range expected.options {
			if runtimeDeclaredTypeCompatible(actual, opt) {
				return true
			}
		}
		return false
	case "list":
		return runtimeDeclaredTypeCompatible(actual.elem, expected.elem)
	case "tuple":
		if len(actual.elems) != len(expected.elems) {
			return false
		}
		for i := range expected.elems {
			if !runtimeDeclaredTypeCompatible(actual.elems[i], expected.elems[i]) {
				return false
			}
		}
		return true
	case "map":
		return runtimeDeclaredTypeCompatible(actual.key, expected.key) && runtimeDeclaredTypeCompatible(actual.val, expected.val)
	case "fn":
		if expected.ret == nil && len(expected.params) == 0 {
			return true
		}
		if actual.variadic != expected.variadic || len(actual.params) != len(expected.params) {
			return false
		}
		for i := range expected.params {
			if !runtimeDeclaredTypeCompatible(actual.params[i], expected.params[i]) {
				return false
			}
		}
		return runtimeDeclaredTypeCompatible(actual.ret, expected.ret)
	case "struct":
		if expected.name == "" || actual.name == "" {
			return true
		}
		return actual.name == expected.name
	default:
		return actual.kind == expected.kind
	}
}

func describeRuntimeDeclaredType(t *rtType) string {
	if t == nil {
		return "any"
	}
	switch t.kind {
	case "union":
		parts := make([]string, 0, len(t.options))
		for _, opt := range t.options {
			parts = append(parts, describeRuntimeDeclaredType(opt))
		}
		return strings.Join(parts, "|")
	case "list":
		return "list<" + describeRuntimeDeclaredType(t.elem) + ">"
	case "tuple":
		parts := make([]string, 0, len(t.elems))
		for _, el := range t.elems {
			parts = append(parts, describeRuntimeDeclaredType(el))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case "map":
		return "map<" + describeRuntimeDeclaredType(t.key) + ", " + describeRuntimeDeclaredType(t.val) + ">"
	case "fn":
		out := "fn<" + describeRuntimeDeclaredType(t.ret)
		for _, p := range t.params {
			out += ", " + describeRuntimeDeclaredType(p)
		}
		return out + ">"
	case "typeparam":
		if t.name != "" {
			return t.name
		}
		return "typeparam"
	case "struct":
		if t.name == "" {
			return "struct"
		}
		return "struct(" + t.name + ")"
	default:
		return t.kind
	}
}

func describeRuntimeObjectType(v object.Object) string {
	if v == nil {
		return "nil"
	}
	switch obj := v.(type) {
	case *object.Nil:
		return "nil"
	case *object.Boolean:
		return "bool"
	case *object.Number:
		return "num"
	case *object.String:
		return "str"
	case *object.Bytes:
		return "bytes"
	case *object.Symbol:
		return "sym"
	case *object.List:
		return describeRuntimeListType(obj)
	case *object.Map:
		return describeRuntimeMapType(obj)
	case *object.StructValue:
		if obj.Schema != nil && obj.Schema.Name != "" {
			return "struct(" + obj.Schema.Name + ")"
		}
		return "struct"
	case *object.StructSchema:
		if obj.Name != "" {
			return "struct(" + obj.Name + ")"
		}
		return "struct"
	case *object.Channel:
		return "chan"
	case *VMTaskHandle:
		return "task"
	case *object.Function:
		return describeRuntimeDeclaredType(runtimeDeclaredTypeFromFunction(obj.Parameters, obj.ReturnType, obj.TypeParams))
	case *object.Foreign:
		return describeRuntimeDeclaredType(runtimeDeclaredTypeFromFunction(obj.Parameters, obj.ReturnType, obj.TypeParams))
	case *VMFunction:
		return describeRuntimeDeclaredType(runtimeDeclaredTypeFromVMFunction(obj))
	case *object.FunctionGroup:
		return "fn"
	default:
		switch v.Type() {
		case object.NIL_OBJ:
			return "nil"
		case object.BOOLEAN_OBJ:
			return "bool"
		case object.NUMBER_OBJ:
			return "num"
		case object.STRING_OBJ:
			return "str"
		case object.BYTE_OBJ:
			return "bytes"
		case object.SYMBOL_OBJ:
			return "sym"
		case object.LIST_OBJ:
			return "list"
		case object.MAP_OBJ:
			return "map"
		case object.STRUCT_OBJ, object.STRUCT_SCHEMA_OBJ:
			return "struct"
		case object.CHANNEL_OBJ:
			return "chan"
		case object.TASK_HANDLE_OBJ:
			return "task"
		case object.FUNCTION_OBJ, object.FUNCTION_GROUP_OBJ, object.FOREIGN_OBJ:
			return "fn"
		default:
			return strings.ToLower(string(v.Type()))
		}
	}
}

func describeRuntimeListType(list *object.List) string {
	if list == nil || len(list.Elements) == 0 {
		return "list<any>"
	}
	parts := make([]string, 0, len(list.Elements))
	for _, el := range list.Elements {
		parts = append(parts, describeRuntimeObjectType(el))
	}
	if allEqualStrings(parts) {
		if strings.HasPrefix(parts[0], "[") {
			return "list<" + parts[0] + ">"
		}
		if len(parts) <= 4 {
			return "[" + strings.Join(parts, ", ") + "]"
		}
		return "list<" + parts[0] + ">"
	}
	if len(parts) <= 4 {
		return "[" + strings.Join(parts, ", ") + "]"
	}
	return "list<any>"
}

func describeRuntimeMapType(m *object.Map) string {
	if m == nil || m.Len() == 0 {
		return "map<any, any>"
	}
	keyTypes := map[string]struct{}{}
	valTypes := map[string]struct{}{}
	m.ForEach(func(_ object.MapKey, p object.MapPair) bool {
		keyTypes[describeRuntimeObjectType(p.Key)] = struct{}{}
		valTypes[describeRuntimeObjectType(p.Value)] = struct{}{}
		return true
	})
	return "map<" + joinTypeSet(keyTypes) + ", " + joinTypeSet(valTypes) + ">"
}

func joinTypeSet(set map[string]struct{}) string {
	if len(set) == 0 {
		return "any"
	}
	parts := make([]string, 0, len(set))
	for k := range set {
		parts = append(parts, k)
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

func allEqualStrings(parts []string) bool {
	if len(parts) == 0 {
		return true
	}
	first := parts[0]
	for _, part := range parts[1:] {
		if part != first {
			return false
		}
	}
	return true
}

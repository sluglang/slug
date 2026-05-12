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
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	if parts := splitRuntimeTypeTopLevel(s, '|'); len(parts) > 1 {
		opts := make([]*rtType, 0, len(parts))
		for _, p := range parts {
			t := parseRuntimeDeclaredType(p)
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
			t := parseRuntimeDeclaredType(p)
			if t == nil {
				t = &rtType{kind: "any"}
			}
			elems = append(elems, t)
		}
		return &rtType{kind: "tuple", elems: elems}
	}
	if strings.HasPrefix(s, "list<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("list<") : len(s)-1])
		elem := parseRuntimeDeclaredType(inner)
		if elem == nil {
			elem = &rtType{kind: "any"}
		}
		return &rtType{kind: "list", elem: elem}
	}
	if strings.HasPrefix(s, "map<") && strings.HasSuffix(s, ">") {
		inner := strings.TrimSpace(s[len("map<") : len(s)-1])
		parts := splitRuntimeTypeTopLevel(inner, ',')
		if len(parts) == 2 {
			k := parseRuntimeDeclaredType(parts[0])
			v := parseRuntimeDeclaredType(parts[1])
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
		ret := parseRuntimeDeclaredType(parts[0])
		if ret == nil {
			ret = &rtType{kind: "any"}
		}
		if len(parts) == 1 {
			return &rtType{kind: "fn", ret: ret}
		}
		params := make([]*rtType, 0, len(parts)-1)
		for _, p := range parts[1:] {
			t := parseRuntimeDeclaredType(p)
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
		return runtimeDeclaredTypeFromFunction(fn.Parameters, fn.ReturnType)
	case *object.Foreign:
		return runtimeDeclaredTypeFromFunction(fn.Parameters, fn.ReturnType)
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
	params := make([]*rtType, 0, len(fn.Parameters))
	for _, p := range fn.Parameters {
		t := parseRuntimeDeclaredType(p.Type)
		if t == nil {
			t = &rtType{kind: "any"}
		}
		params = append(params, t)
	}
	ret := parseRuntimeDeclaredType(fn.ReturnType)
	if ret == nil {
		ret = &rtType{kind: "any"}
	}
	return &rtType{kind: "fn", params: params, ret: ret, variadic: len(fn.Parameters) > 0 && fn.Parameters[len(fn.Parameters)-1].IsVariadic}
}

func runtimeDeclaredTypeFromFunction(params []*ast.FunctionParameter, returnType string) *rtType {
	fnParams := make([]*rtType, 0, len(params))
	for _, p := range params {
		t := parseRuntimeDeclaredType(p.Type)
		if t == nil {
			t = &rtType{kind: "any"}
		}
		fnParams = append(fnParams, t)
	}
	ret := parseRuntimeDeclaredType(returnType)
	if ret == nil {
		ret = &rtType{kind: "any"}
	}
	return &rtType{kind: "fn", params: fnParams, ret: ret}
}

func runtimeDeclaredTypeCompatible(actual, expected *rtType) bool {
	if expected == nil || expected.kind == "any" {
		return true
	}
	if actual == nil || actual.kind == "any" {
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
		return describeRuntimeDeclaredType(runtimeDeclaredTypeFromFunction(obj.Parameters, obj.ReturnType))
	case *object.Foreign:
		return describeRuntimeDeclaredType(runtimeDeclaredTypeFromFunction(obj.Parameters, obj.ReturnType))
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

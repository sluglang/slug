package foreign

import (
	"log/slog"
	"sort"
	"strings"

	"slug/internal/ast"
	"slug/internal/dec64"
	"slug/internal/object"
)

func fnMetaHasTag() *object.Foreign {
	return &object.Foreign{
		Name: "hasTag",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {

			if len(args) != 2 {
				return ctx.NewError("hasTag expects exactly 2 arguments: object and tagName name")
			}

			tagName, ok := args[1].(*object.String)
			if !ok {
				return ctx.NewError("second argument to hasTag must be a string")
			}

			switch o := args[0].(type) {
			case object.Taggable:
				return ctx.NativeBoolToBooleanObject(o.HasTag(tagName.Value))
			}
			return ctx.NativeBoolToBooleanObject(false)
		},
	}
}

func fnMetaGetTag() *object.Foreign {
	return &object.Foreign{
		Name: "getTag",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			if len(args) != 2 {
				return ctx.NewError("getTag expects exactly 2 arguments: object and tagName name")
			}

			tagName, ok := args[1].(*object.String)
			if !ok {
				return ctx.NewError("second argument to getTag must be a string")
			}

			switch o := args[0].(type) {
			case object.Taggable:
				if args, exists := o.GetTagParams(tagName.Value); exists {
					return &args
				}
			}
			return ctx.Nil()
		},
	}
}

func fnMetaSearchModuleTags() *object.Foreign {
	return &object.Foreign{
		Name: "searchModuleTags",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {

			if len(args) < 1 || len(args) > 3 {
				return ctx.NewError("searchModuleTags expects 1-3 arguments: module name, tag name, and optional includePrivate flag")
			}

			// Check module name
			moduleName, ok := args[0].(*object.String)
			if !ok {
				return ctx.NewError("first argument must be the module name as a string")
			}

			// Check tag name
			tagName, ok := args[1].(*object.String)
			if !ok {
				return ctx.NewError("second argument must be the tag name as a string")
			}

			// Check optional includePrivate flag
			includePrivate := false
			if len(args) == 3 {
				flag, ok := args[2].(*object.Boolean)
				if !ok {
					return ctx.NewError("third argument must be a boolean for includePrivate")
				}
				includePrivate = flag.Value
			}

			// Load the targeted module
			module, err := ctx.LoadModule(moduleName.Value)
			if err != nil {
				return ctx.NewError("failed to load module '%s': %s", moduleName.Value, err.Error())
			}

			slog.Debug("module loaded",
				slog.Any("module-name", module.Name),
				slog.Any("path", module.Path),
				slog.Any("binding-count", len(module.Env.Bindings)))

			m := &object.Map{}

			for name, binding := range module.Env.Bindings {

				slog.Debug("binding module value",
					slog.Any("module-name", module.Name),
					slog.Any("binding-name", name),
					slog.Any("bound-value", binding.Value.Type()),
				)

				if binding.Meta.IsImport {
					continue
				}

				if (includePrivate || binding.Meta.IsExport) &&
					hasTag(binding, tagName.Value) {

					m.Put(&object.String{Value: name}, binding.Value)
				}
			}

			return m
		},
	}
}

func fnMetaSearchScopeTags() *object.Foreign {
	return &object.Foreign{
		Name: "searchScopeTags",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			if len(args) != 1 {
				return ctx.NewError("searchScopeTags expects 1 argument: tag name")
			}

			tagName, ok := args[0].(*object.String)
			if !ok {
				return ctx.NewError("argument must be a string tag name")
			}

			var tuples []object.Object

			env := ctx.CurrentEnv()

			for env != nil {
				for name, binding := range env.Bindings {
					if hasTag(binding, tagName.Value) {
						taggable, ok := binding.Value.(object.Taggable)
						if ok {
							opts, _ := taggable.GetTagParams(tagName.Value)
							var tuple = make([]object.Object, 3)
							tuple[0] = &object.String{Value: name}
							tuple[1] = binding.Value
							tuple[2] = &opts
							tuples = append(tuples, &object.List{
								Elements: tuple,
							})
						} else {
							slog.Warn("this should not happen")
						}
					}
				}
				env = env.Outer
			}

			return &object.List{
				Elements: tuples,
			}
		},
	}
}

func fnMetaModuleDocs() *object.Foreign {
	return &object.Foreign{
		Name: "moduleDocs",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			if len(args) != 1 {
				return ctx.NewError("moduleDocs expects exactly 1 argument: module name")
			}

			moduleName, ok := args[0].(*object.String)
			if !ok {
				return ctx.NewError("moduleDocs argument must be a string module name")
			}

			module, err := ctx.LoadModule(moduleName.Value)
			if err != nil {
				return ctx.NewError("failed to load module '%s': %s", moduleName.Value, err.Error())
			}

			if !module.HasDoc {
				return ctx.Nil()
			}
			return &object.String{Value: module.Doc}
		},
	}
}

func fnMetaDescribe() *object.Foreign {
	return &object.Foreign{
		Name: "describe",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			if len(args) != 1 {
				return ctx.NewError("describe expects exactly 1 argument: value")
			}

			original := args[0]
			resolved := original
			if ref, ok := original.(*object.BindingRef); ok {
				if val, ok := resolveBindingValue(ref); ok {
					resolved = val
				}
			}

			result := &object.Map{}
			result.Put(object.InternSymbol("type"), object.InternSymbol(describeType(resolved)))
			result.Put(object.InternSymbol("docs"), &object.String{Value: describeDocs(ctx, original)})
			result.Put(object.InternSymbol("tags"), describeTags(resolved))
			result.Put(object.InternSymbol("details"), describeDetails(resolved))
			return result
		},
	}
}

func fnMetaDescribeSymbol() *object.Foreign {
	return &object.Foreign{
		Name: "describeSymbol",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			if len(args) != 2 {
				return ctx.NewError("describeSymbol expects exactly 2 arguments: module name and symbol name")
			}
			moduleName, ok := args[0].(*object.String)
			if !ok {
				return ctx.NewError("first argument to describeSymbol must be a string module name")
			}
			symbolName, ok := args[1].(*object.String)
			if !ok {
				return ctx.NewError("second argument to describeSymbol must be a string symbol name")
			}
			module, err := ctx.LoadModule(moduleName.Value)
			if err != nil {
				return ctx.NewError("failed to load module '%s': %s", moduleName.Value, err.Error())
			}
			binding, ok := module.Env.GetLocalBinding(symbolName.Value)
			if !ok || binding == nil {
				return ctx.Nil()
			}
			val := binding.Value
			if ref, ok := val.(*object.BindingRef); ok {
				if resolved, ok := resolveBindingValue(ref); ok {
					val = resolved
				}
			}
			describeVal := val
			if fg, ok := val.(*object.FunctionGroup); ok && len(fg.Functions) == 1 {
				for _, fn := range fg.Functions {
					describeVal = fn
					break
				}
			}
			result := &object.Map{}
			result.Put(object.InternSymbol("type"), object.InternSymbol(describeType(describeVal)))
			docs := ""
			if binding.Meta.HasDoc {
				docs = binding.Meta.Doc
			} else {
				docs = describeDocs(ctx, describeVal)
			}
			result.Put(object.InternSymbol("docs"), &object.String{Value: docs})
			result.Put(object.InternSymbol("tags"), describeTags(describeVal))
			result.Put(object.InternSymbol("details"), describeDetails(describeVal))
			return result
		},
	}
}

func describeType(value object.Object) string {
	if value == nil {
		return "nil"
	}

	switch value.Type() {
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
	case object.STRUCT_SCHEMA_OBJ, object.STRUCT_OBJ:
		return "struct"
	case object.MODULE_OBJ:
		return "module"
	case object.FUNCTION_OBJ, object.FUNCTION_GROUP_OBJ, object.FOREIGN_OBJ:
		if fg, ok := value.(*object.FunctionGroup); ok {
			if fg == nil || len(fg.Functions) <= 1 {
				return "fn"
			}
			return "grp"
		}
		return "fn"
	case object.CHANNEL_OBJ:
		return "chan"
	case object.TASK_HANDLE_OBJ:
		return "task"
	case object.ERROR_OBJ:
		return "error"
	default:
		return strings.ToLower(string(value.Type()))
	}
}

func describeDocs(ctx object.RuntimeContext, value object.Object) string {
	if value == nil {
		return ""
	}
	if doc, ok := findDocForValue(ctx, value); ok {
		return doc
	}
	if mod, ok := value.(*object.Module); ok {
		if mod.HasDoc {
			return mod.Doc
		}
	}
	return ""
}

func describeTags(value object.Object) *object.Map {
	result := &object.Map{}
	if value == nil {
		return result
	}

	taggable, ok := value.(object.Taggable)
	if !ok {
		return result
	}

	tags := taggable.GetTags()
	if len(tags) == 0 {
		return result
	}

	names := make([]string, 0, len(tags))
	for name := range tags {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		params := tags[name]
		copied := make([]object.Object, len(params.Elements))
		copy(copied, params.Elements)
		result.Put(&object.String{Value: name}, &object.List{Elements: copied})
	}
	return result
}

func describeDetails(value object.Object) *object.Map {
	details := &object.Map{}
	if value == nil {
		return details
	}

	switch v := value.(type) {
	case *object.StructValue:
		return describeStructDetails(v)
	case *object.StructSchema:
		return describeStructSchemaDetails(v)
	case *object.Function:
		return describeFunctionDetails(false, v.Signature, v.Parameters)
	case *object.Foreign:
		return describeFunctionDetails(true, v.Signature, v.Parameters)
	case interface {
		GetSignature() ast.FSig
		GetParameters() []*ast.FunctionParameter
	}:
		return describeFunctionDetails(false, v.GetSignature(), v.GetParameters())
	case *object.FunctionGroup:
		return describeFunctionGroupDetails(v)
	case *object.Module:
		return describeModuleDetails(v)
	default:
		return details
	}
}

func describeStructDetails(v *object.StructValue) *object.Map {
	details := &object.Map{}

	structType := "struct"
	fields := []object.Object{}
	fieldCount := 0

	if v.Schema != nil {
		if v.Schema.Name != "" {
			structType = v.Schema.Name
		}
		fieldCount = len(v.Schema.Fields)
		fields = make([]object.Object, 0, fieldCount)
		for _, field := range v.Schema.Fields {
			fieldMap := &object.Map{}
			fieldMap.Put(object.InternSymbol("name"), &object.String{Value: field.Name})
			fieldMap.Put(object.InternSymbol("tags"), tagsFromAstTags(field.Tags))
			hasDefault := field.Default != nil
			fieldMap.Put(object.InternSymbol("hasDefault"), boolObject(hasDefault))
			if hasDefault {
				fieldMap.Put(object.InternSymbol("default"), &object.String{Value: describeDefaultValue(field.Default)})
			}
			fields = append(fields, fieldMap)
		}
	}

	details.Put(object.InternSymbol("structType"), object.InternSymbol(structType))
	details.Put(object.InternSymbol("fields"), &object.List{Elements: fields})
	details.Put(object.InternSymbol("fieldCount"), &object.Number{Value: dec64.FromInt(fieldCount)})
	return details
}

func describeStructSchemaDetails(schema *object.StructSchema) *object.Map {
	details := &object.Map{}
	if schema == nil {
		return details
	}

	structType := "struct"
	if schema.Name != "" {
		structType = schema.Name
	}

	fieldCount := len(schema.Fields)
	fields := make([]object.Object, 0, fieldCount)
	for _, field := range schema.Fields {
		fieldMap := &object.Map{}
		fieldMap.Put(object.InternSymbol("name"), &object.String{Value: field.Name})
		fieldMap.Put(object.InternSymbol("tags"), tagsFromAstTags(field.Tags))
		hasDefault := field.Default != nil
		fieldMap.Put(object.InternSymbol("hasDefault"), boolObject(hasDefault))
		if hasDefault {
			fieldMap.Put(object.InternSymbol("default"), &object.String{Value: describeDefaultValue(field.Default)})
		}
		fields = append(fields, fieldMap)
	}

	details.Put(object.InternSymbol("structType"), object.InternSymbol(structType))
	details.Put(object.InternSymbol("fields"), &object.List{Elements: fields})
	details.Put(object.InternSymbol("fieldCount"), &object.Number{Value: dec64.FromInt(fieldCount)})
	return details
}

func describeFunctionGroupDetails(group *object.FunctionGroup) *object.Map {
	if group == nil || len(group.Functions) == 0 {
		return describeFunctionDetails(false, ast.FSig{}, nil)
	}

	entries := collectFunctionEntries(group.Functions)
	if len(entries) == 1 {
		entry := entries[0]
		return describeFunctionDetails(isForeignFunction(entry.fn), entry.sig, paramsFromFunctionObject(entry.fn))
	}

	groups := make([]object.Object, 0, len(entries))
	for _, entry := range entries {
		fnMap := &object.Map{}
		fnMap.Put(object.InternSymbol("type"), object.InternSymbol("fn"))
		if group.HasDoc {
			fnMap.Put(object.InternSymbol("docs"), &object.String{Value: group.Doc})
		} else {
			fnMap.Put(object.InternSymbol("docs"), &object.String{Value: ""})
		}
		fnMap.Put(object.InternSymbol("tags"), describeTags(entry.fn))
		fnMap.Put(object.InternSymbol("details"), describeFunctionDetails(isForeignFunction(entry.fn), entry.sig, paramsFromFunctionObject(entry.fn)))
		groups = append(groups, fnMap)
	}

	details := &object.Map{}
	details.Put(object.InternSymbol("groups"), &object.List{Elements: groups})
	return details
}

func collectFunctionEntries(functions map[ast.FSig]object.Object) []functionEntry {
	entries := make([]functionEntry, 0, len(functions))
	for sig, fn := range functions {
		entries = append(entries, functionEntry{sig: sig, fn: fn})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].sig.Min != entries[j].sig.Min {
			return entries[i].sig.Min < entries[j].sig.Min
		}
		if entries[i].sig.Max != entries[j].sig.Max {
			return entries[i].sig.Max < entries[j].sig.Max
		}
		if entries[i].sig.IsVariadic != entries[j].sig.IsVariadic {
			return !entries[i].sig.IsVariadic
		}
		if entries[i].sig.Tags != entries[j].sig.Tags {
			return entries[i].sig.Tags < entries[j].sig.Tags
		}
		return entries[i].fn.Type() < entries[j].fn.Type()
	})
	return entries
}

type functionEntry struct {
	sig ast.FSig
	fn  object.Object
}

func paramsFromFunctionObject(fn object.Object) []*ast.FunctionParameter {
	if f, ok := fn.(interface {
		GetParameters() []*ast.FunctionParameter
	}); ok {
		return f.GetParameters()
	}
	return nil
}

func isForeignFunction(fn object.Object) bool {
	_, ok := fn.(*object.Foreign)
	return ok
}

func describeFunctionDetails(isForeign bool, sig ast.FSig, params []*ast.FunctionParameter) *object.Map {
	details := &object.Map{}
	details.Put(object.InternSymbol("foreign"), boolObject(isForeign))
	details.Put(object.InternSymbol("arityMin"), &object.Number{Value: dec64.FromInt(sig.Min)})
	details.Put(object.InternSymbol("arityMax"), &object.Number{Value: dec64.FromInt(sig.Max)})

	paramList := &object.List{Elements: []object.Object{}}
	if len(params) > 0 {
		paramList.Elements = make([]object.Object, 0, len(params))
		for idx, param := range params {
			if param == nil || param.Name == nil {
				continue
			}
			paramMap := &object.Map{}
			paramMap.Put(object.InternSymbol("name"), &object.String{Value: param.Name.Value})
			paramMap.Put(object.InternSymbol("tags"), tagsFromAstTags(param.Tags))
			hasDefault := param.Default != nil
			paramMap.Put(object.InternSymbol("hasDefault"), boolObject(hasDefault))
			if hasDefault {
				paramMap.Put(object.InternSymbol("default"), &object.String{Value: describeDefaultValue(param.Default)})
			}
			if sig.IsVariadic && idx == len(params)-1 {
				paramMap.Put(object.InternSymbol("vargs"), object.TRUE)
			} else {
				paramMap.Put(object.InternSymbol("vargs"), object.FALSE)
			}
			paramList.Elements = append(paramList.Elements, paramMap)
		}
	}

	details.Put(object.InternSymbol("params"), paramList)
	return details
}

func describeModuleDetails(module *object.Module) *object.Map {
	details := &object.Map{}
	if module == nil || module.Env == nil {
		details.Put(object.InternSymbol("exports"), &object.List{Elements: []object.Object{}})
		return details
	}

	exports := []string{}
	for name, binding := range module.Env.Bindings {
		if binding == nil {
			continue
		}
		if binding.Meta.IsExport && !binding.Meta.IsImport {
			exports = append(exports, name)
		}
	}
	sort.Strings(exports)

	values := make([]object.Object, 0, len(exports))
	for _, name := range exports {
		values = append(values, object.InternSymbol(name))
	}

	details.Put(object.InternSymbol("exports"), &object.List{Elements: values})
	return details
}

func tagsFromAstTags(tags []*ast.Tag) *object.Map {
	result := &object.Map{}
	if len(tags) == 0 {
		return result
	}

	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag == nil {
			continue
		}
		names = append(names, tag.Name)
	}

	sort.Strings(names)
	for _, name := range names {
		var params []object.Object
		for _, tag := range tags {
			if tag == nil || tag.Name != name {
				continue
			}
			for _, arg := range tag.Args {
				params = append(params, &object.String{Value: arg.String()})
			}
			break
		}
		result.Put(&object.String{Value: name}, &object.List{Elements: params})
	}
	return result
}

func boolObject(value bool) *object.Boolean {
	if value {
		return object.TRUE
	}
	return object.FALSE
}

func describeDefaultValue(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
	if s, ok := expr.(*ast.StringLiteral); ok {
		return `"` + escapeStringLiteral(s.Value) + `"`
	}
	return expr.String()
}

func escapeStringLiteral(value string) string {
	var out strings.Builder
	for _, r := range value {
		switch r {
		case '\\':
			out.WriteString(`\\`)
		case '"':
			out.WriteString(`\"`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		case '\b':
			out.WriteString(`\b`)
		case '\f':
			out.WriteString(`\f`)
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}

func hasTag(binding *object.Binding, tagName string) bool {
	if binding == nil {
		return false
	}

	// Check if the binding contains a group of functions
	fg, ok := binding.Value.(object.Taggable)
	return ok && fg.HasTag(tagName)
}

func findDocForValue(ctx object.RuntimeContext, value object.Object) (string, bool) {
	if fg, ok := value.(*object.FunctionGroup); ok && fg.HasDoc {
		return fg.Doc, true
	}

	if ref, ok := value.(*object.BindingRef); ok {
		return docFromBinding(ref.Env, ref.Name)
	}

	env := ctx.CurrentEnv()
	if env == nil {
		return "", false
	}
	for cur := env; cur != nil; cur = cur.Outer {
		if doc, ok := docFromEnvValue(cur, value); ok {
			return doc, true
		}
	}
	return "", false
}

func docFromBinding(env *object.Environment, name string) (string, bool) {
	if env == nil {
		return "", false
	}
	binding, ok := env.GetLocalBinding(name)
	if !ok || binding == nil {
		return "", false
	}
	if !binding.Meta.HasDoc {
		return "", false
	}
	return binding.Meta.Doc, true
}

func docFromEnvValue(env *object.Environment, value object.Object) (string, bool) {
	if env == nil {
		return "", false
	}

	matches := 0
	var doc string
	hasDoc := false

	for _, binding := range env.Bindings {
		if binding == nil {
			continue
		}
		resolved, ok := resolveBindingValue(binding.Value)
		if !ok {
			continue
		}
		// Imported module maps may hold BindingRef members; map indexing resolves
		// those refs before describe() sees the value. Recover docs by matching the
		// resolved member value back to its source binding doc metadata.
		if m, ok := resolved.(*object.Map); ok {
			if doc, found := docFromMapMembers(m, value); found {
				return doc, true
			}
		}
		if resolved != value {
			continue
		}
		matches++
		if binding.Meta.HasDoc {
			doc = binding.Meta.Doc
			hasDoc = true
		}
	}

	if matches == 1 && hasDoc {
		return doc, true
	}
	return "", false
}

func docFromMapMembers(m *object.Map, value object.Object) (string, bool) {
	if m == nil {
		return "", false
	}
	found := ""
	hasDoc := false
	m.ForEach(func(_ object.MapKey, pair object.MapPair) bool {
		ref, ok := pair.Value.(*object.BindingRef)
		if !ok {
			return true
		}
		resolved, ok := resolveBindingValue(ref)
		if !ok || resolved != value {
			return true
		}
		doc, ok := docFromBinding(ref.Env, ref.Name)
		if !ok {
			return true
		}
		found = doc
		hasDoc = true
		return false
	})
	return found, hasDoc
}

func resolveBindingValue(value object.Object) (object.Object, bool) {
	for {
		ref, ok := value.(*object.BindingRef)
		if !ok {
			return value, true
		}
		if ref.Env == nil {
			return nil, false
		}
		val, _, ok := ref.Env.GetLocalBindingValue(ref.Name)
		if !ok {
			return nil, false
		}
		if val == object.BINDING_UNINITIALIZED {
			return nil, false
		}
		value = val
	}
}

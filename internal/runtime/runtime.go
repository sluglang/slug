package runtime

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"slug/internal/ast"
	"slug/internal/foreign"
	"slug/internal/lexer"
	"slug/internal/object"
	"slug/internal/parser"
	"slug/internal/semantic"
	"slug/internal/util"
	"slug/internal/vm"
	"strings"
	"sync/atomic"
)

type Runtime struct {
	Config           util.Configuration
	Modules          map[string]*object.Module
	Builtins         map[string]*object.Foreign
	ForeignFunctions map[string]*object.Foreign
	nextID           atomic.Int64
}

func NewRuntime(config util.Configuration) *Runtime {

	config.Store = util.NewConfigStore(config.RootPath, config.SlugHome, config.MainModule, config.Argv)

	builtinFunctions := map[string]*object.Foreign{
		"argv":       fnBuiltinArgv(),
		"argm":       fnBuiltinArgm(),
		"cfg":        fnBuiltinCfg(),
		"import":     fnBuiltinImport(),
		"len":        fnBuiltinLen(),
		"print":      fnBuiltinPrint(),
		"println":    fnBuiltinPrintLn(),
		"stacktrace": fnBuiltinStacktrace(),
	}

	functions := getForeignFunctions()
	functions["slug.channel.chan"] = fnChannelChan()
	functions["slug.channel.close"] = fnChannelClose()

	return &Runtime{
		Config:           config,
		Modules:          nil,
		Builtins:         builtinFunctions,
		ForeignFunctions: functions,
	}
}

func (r *Runtime) NextHandleID() int64 {
	return r.nextID.Add(1)<<16 | int64(rand.Intn(0xFFFF))
}

func (r *Runtime) LookupForeign(name string) (*object.Foreign, bool) {
	if fn, ok := r.ForeignFunctions[name]; ok {
		return fn, true
	}
	return nil, false
}

func (r *Runtime) LoadModule(modName string) (*object.Module, error) {

	if r.Modules == nil {
		r.Modules = make(map[string]*object.Module)
	}

	if mod, ok := r.Modules[modName]; ok {
		slog.Info("Module loaded from cache",
			slog.String("name", modName))
		return mod, nil
	}

	// 1. Resolve module name to relative file path (r.g., "slug.std" -> "slug/std.slug")
	pathParts := strings.Split(modName, ".")
	relPath := filepath.Join(pathParts...) + ".slug"

	// 2. Search Paths: Check local RootPath, then SLUG_HOME/lib
	var fullPath string
	var loadRoot string
	var source []byte
	var err error

	// Try local RootPath (directory of the entry script)
	loadRoot = r.Config.RootPath
	fullPath = filepath.Join(r.Config.RootPath, relPath)
	source, errFirst := os.ReadFile(fullPath)

	if errFirst != nil {
		//// Fallback to $SLUG_HOME/lib
		if r.Config.SlugHome != "" {
			loadRoot = filepath.Join(r.Config.SlugHome, "lib")
			fullPath = filepath.Join(r.Config.SlugHome, "lib", relPath)
			source, err = os.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("could not load module %s: local error: %v, lib error: %v", modName, errFirst, err)
			}
		} else {
			return nil, fmt.Errorf("could not load module %s: %v (SLUG_HOME not set)", modName, errFirst)
		}
	}

	// 3. Tokenize and Parse
	l := lexer.New(string(source))
	p := parser.New(l, fullPath, string(source))
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		slog.Warn("Error loading module",
			slog.String("name", modName),
			slog.String("fullPath", fullPath),
			slog.String("errors", strings.Join(p.Errors(), "\n")),
		)
		return nil, fmt.Errorf("parse errors in module %s:\n%s", modName, strings.Join(p.Errors(), "\n"))
	}
	semErrs, semWarns := semantic.AnalyzeWithOptions(fullPath, string(source), program, semantic.AnalyzeOptions{
		EnableTypeCheck: r.Config.EnableTypeCheck,
		StrictTypeCheck: r.Config.StrictTypeCheck,
	})
	if len(semWarns) > 0 {
		slog.Warn("Semantic type warnings loading module",
			slog.String("name", modName),
			slog.String("fullPath", fullPath),
			slog.String("warnings", strings.Join(semWarns, "\n")),
		)
	}
	if len(semErrs) > 0 {
		slog.Warn("Semantic error loading module",
			slog.String("name", modName),
			slog.String("fullPath", fullPath),
			slog.String("errors", strings.Join(semErrs, "\n")),
		)
		return nil, fmt.Errorf("semantic errors in module %s:\n%s", modName, strings.Join(semErrs, "\n"))
	}

	if r.Config.DebugJsonAST {
		json, err := parser.RenderASTAsJSON(program)
		if err != nil {
			slog.Error("Failed to render AST as JSON",
				slog.Any("error", err))
		} else {
			jsonPath := fullPath + ".ast.json"
			err = os.WriteFile(jsonPath, []byte(json), 0644)
			if err != nil {
				slog.Error("Failed to write AST as JSON")
			}
		}
	}
	if r.Config.DebugTxtAST {
		txtPath := fullPath + ".ast.txt"
		text := parser.RenderASTAsText(program, 0)
		err = os.WriteFile(txtPath, []byte(text), 0644)
		if err != nil {
			slog.Error("Failed to write AST as JSON")
		}
	}

	// 4. Setup Module Object and Environment
	moduleEnv := object.NewEnvironment()
	moduleEnv.Path = fullPath
	moduleEnv.LibRoot = moduleLibRoot(loadRoot, modName)
	moduleEnv.ModuleFqn = modName
	moduleEnv.Src = string(source)
	module := &object.Module{
		Name:    modName,
		Path:    fullPath,
		Src:     string(source),
		Program: program,
		Env:     moduleEnv,
		Doc:     program.ModuleDoc,
		HasDoc:  program.HasModuleDoc,
	}

	r.Modules[modName] = module

	// Declare pass: prebind top-level names to support circular imports.
	predeclareTopLevel(program, moduleEnv)
	slog.Info("Module loaded, added to cache",
		slog.String("name", modName),
		slog.String("fullPath", fullPath),
	)

	// 5. Evaluate the module in its own environment
	slog.Debug("loading module", slog.String("name", modName), slog.String("path", fullPath))
	return r.evalModuleWithVM(modName, module, moduleEnv, program)
}

func (r *Runtime) evalModuleWithVM(modName string, module *object.Module, moduleEnv *object.Environment, program *ast.Program) (*object.Module, error) {
	moduleEnv.Outer = moduleBuiltinsEnvironment(r, moduleEnv)
	vmProgram, prepErr := vm.PrepareProgram(moduleEnv, program, r.LookupForeign, hasExportTag, buildParamIndexForVMBridge)
	if prepErr != nil {
		return nil, fmt.Errorf("runtime error while loading module %s: %s", modName, prepErr.Error())
	}
	exec := vm.NewExecutorWithBridgeFactory(moduleEnv, func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
		return makeVMCallBridge(r, callEnv)
	})
	out := exec.EvalProgram(vmProgram)
	if out != nil && out.Type() == object.ERROR_OBJ {
		return nil, fmt.Errorf("runtime error while loading module %s: %s", modName, out.Inspect())
	}
	if err := vm.ApplyForeignTags(moduleEnv, program, r.LookupForeign, func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
		return makeVMCallBridge(r, callEnv)
	}); err != nil {
		return nil, fmt.Errorf("runtime error while loading module %s: %s", modName, err.Error())
	}
	if err := applyTopLevelTagsForVM(r, program, moduleEnv); err != nil {
		return nil, fmt.Errorf("runtime error while loading module %s: %s", modName, err.Error())
	}
	applyTopLevelExportMetadata(program, moduleEnv)
	return module, nil
}

func moduleLibRoot(loadRoot, _ string) string {
	root := filepath.Clean(loadRoot)
	if root == "." || root == "" {
		return root
	}
	return root
}

func predeclareTopLevel(program *ast.Program, env *object.Environment) error {
	for _, stmt := range program.Statements {
		// Statements may be wrapped in ExpressionStatement
		var expr ast.Expression
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			expr = exprStmt.Expression
		}

		switch s := expr.(type) {
		case *ast.ValExpression:
			if err := predeclarePattern(s.Pattern, true, hasExportTag(s.Tags), env); err != nil {
				return err
			}
		case *ast.VarExpression:
			if err := predeclarePattern(s.Pattern, false, hasExportTag(s.Tags), env); err != nil {
				return err
			}
		}
	}
	return nil
}

func predeclarePattern(pat ast.MatchPattern, isConst bool, isExport bool, env *object.Environment) error {
	switch p := pat.(type) {

	case *ast.BindingPattern:
		name := p.Name.Value
		if isConst {
			if _, err := env.DefineConstant(name, object.BINDING_UNINITIALIZED, isExport, false); err != nil {
				return err
			}
		} else {
			if _, err := env.Define(name, object.BINDING_UNINITIALIZED, isExport, false); err != nil {
				return err
			}
		}
		return predeclarePattern(p.Pattern, isConst, isExport, env)

	case *ast.IdentifierPattern:
		name := p.Value.Value
		if isConst {
			_, err := env.DefineConstant(name, object.BINDING_UNINITIALIZED, isExport, false)
			return err
		}
		_, err := env.Define(name, object.BINDING_UNINITIALIZED, isExport, false)
		return err

	case *ast.SpreadPattern:
		// spread can bind a name: var {..rest} = ...
		if p.Value == nil {
			return nil
		}
		name := p.Value.Value
		if isConst {
			_, err := env.DefineConstant(name, object.BINDING_UNINITIALIZED, isExport, false)
			return err
		}
		_, err := env.Define(name, object.BINDING_UNINITIALIZED, isExport, false)
		return err

	case *ast.ListPattern:
		for _, elem := range p.Elements {
			if err := predeclarePattern(elem, isConst, isExport, env); err != nil {
				return err
			}
		}
		return nil

	case *ast.MapPattern:
		// var {*} = ... : cannot predeclare (no names!)
		if p.SelectAll {
			return nil
		}
		// declare any identifiers inside explicit subpatterns
		for _, entry := range p.Pairs {
			if entry.Pattern == nil {
				continue
			}
			if err := predeclarePattern(entry.Pattern, isConst, isExport, env); err != nil {
				return err
			}
		}
		if p.Spread != nil {
			if err := predeclarePattern(p.Spread, isConst, isExport, env); err != nil {
				return err
			}
		}
		return nil

	case *ast.MultiPattern:
		// no new bindings are *guaranteed* across alternatives; safest is to predeclare none
		return nil

	default:
		// LiteralPattern, WildcardPattern, etc: no bindings
		return nil
	}
}

func applyTopLevelExportMetadata(program *ast.Program, env *object.Environment) {
	for _, stmt := range program.Statements {
		exprStmt, ok := stmt.(*ast.ExpressionStatement)
		if !ok {
			continue
		}
		var pat ast.MatchPattern
		var tags []*ast.Tag
		switch ex := exprStmt.Expression.(type) {
		case *ast.ValExpression:
			pat = ex.Pattern
			tags = ex.Tags
		case *ast.VarExpression:
			pat = ex.Pattern
			tags = ex.Tags
		default:
			continue
		}
		if !hasExportTag(tags) {
			continue
		}
		applyExportToPattern(pat, env)
	}
}

func applyTopLevelTagsForVM(rt *Runtime, program *ast.Program, env *object.Environment) error {
	if program == nil || env == nil {
		return nil
	}
	for _, stmt := range program.Statements {
		exprStmt, ok := stmt.(*ast.ExpressionStatement)
		if !ok {
			continue
		}
		var pat ast.MatchPattern
		var tags []*ast.Tag
		switch ex := exprStmt.Expression.(type) {
		case *ast.ValExpression:
			pat = ex.Pattern
			tags = ex.Tags
		case *ast.VarExpression:
			pat = ex.Pattern
			tags = ex.Tags
		default:
			continue
		}
		if len(tags) == 0 {
			continue
		}
		evaluated, err := vm.EvalTagArgs(env, tags, func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
			return makeVMCallBridge(rt, callEnv)
		})
		if err != nil {
			return err
		}
		applyTagsToPatternValues(pat, evaluated, env)
	}
	return nil
}

func applyTagsToPatternValues(pat ast.MatchPattern, tags map[string]object.List, env *object.Environment) {
	switch p := pat.(type) {
	case *ast.IdentifierPattern:
		applyTagMapToBindingValue(p.Value.Value, tags, env)
	case *ast.BindingPattern:
		applyTagMapToBindingValue(p.Name.Value, tags, env)
		applyTagsToPatternValues(p.Pattern, tags, env)
	case *ast.ListPattern:
		for _, el := range p.Elements {
			applyTagsToPatternValues(el, tags, env)
		}
	case *ast.MapPattern:
		for _, entry := range p.Pairs {
			if entry.Pattern != nil {
				applyTagsToPatternValues(entry.Pattern, tags, env)
			}
		}
		if p.Spread != nil {
			applyTagsToPatternValues(p.Spread, tags, env)
		}
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field.Pattern != nil {
				applyTagsToPatternValues(field.Pattern, tags, env)
			}
		}
	}
}

func applyTagMapToBindingValue(name string, tags map[string]object.List, env *object.Environment) {
	binding, ok := env.GetLocalBinding(name)
	if !ok || binding == nil || binding.Value == nil {
		return
	}
	taggable, ok := binding.Value.(object.Taggable)
	if !ok {
		return
	}
	for tag, params := range tags {
		taggable.SetTag(tag, params)
	}
}

func applyExportToPattern(pat ast.MatchPattern, env *object.Environment) {
	switch p := pat.(type) {
	case *ast.IdentifierPattern:
		if binding, ok := env.GetLocalBinding(p.Value.Value); ok {
			binding.Meta.IsExport = true
		}
	case *ast.BindingPattern:
		if binding, ok := env.GetLocalBinding(p.Name.Value); ok {
			binding.Meta.IsExport = true
		}
		applyExportToPattern(p.Pattern, env)
	case *ast.ListPattern:
		for _, el := range p.Elements {
			applyExportToPattern(el, env)
		}
	case *ast.MapPattern:
		for _, entry := range p.Pairs {
			if entry.Pattern != nil {
				applyExportToPattern(entry.Pattern, env)
			}
		}
		if p.Spread != nil {
			applyExportToPattern(p.Spread, env)
		}
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field.Pattern != nil {
				applyExportToPattern(field.Pattern, env)
			}
		}
	}
}

func moduleBuiltinsEnvironment(rt *Runtime, moduleEnv *object.Environment) *object.Environment {
	builtinsEnv := object.NewEnvironment()
	builtinsEnv.Path = moduleEnv.Path
	builtinsEnv.LibRoot = moduleEnv.LibRoot
	builtinsEnv.ModuleFqn = moduleEnv.ModuleFqn
	builtinsEnv.Src = moduleEnv.Src
	for name, fn := range rt.Builtins {
		builtinsEnv.Bindings[name] = &object.Binding{
			Value:     fn,
			IsMutable: false,
			Meta: object.Meta{
				IsImport: false,
				IsExport: false,
			},
		}
	}
	return builtinsEnv
}

func hasExportTag(tags []*ast.Tag) bool {
	for _, tag := range tags {
		if tag.Name == object.EXPORT_TAG {
			return true
		}
	}
	return false
}

func getForeignFunctions() map[string]*object.Foreign {
	foreignFunctions := map[string]*object.Foreign{}
	for k, v := range foreign.GetForeignFunctions() {
		foreignFunctions[k] = v
	}
	return foreignFunctions
}

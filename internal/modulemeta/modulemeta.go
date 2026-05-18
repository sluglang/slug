package modulemeta

import (
	"fmt"
	"os"
	"path/filepath"
	"slug/internal/ast"
	"slug/internal/lexer"
	"slug/internal/parser"
	"strings"
)

// Callable describes an exported callable shape recovered from source.
type Callable struct {
	Name       string
	TypeParams []string
	Parameters []*ast.FunctionParameter
	ReturnType string
	Signature  ast.FSig
}

// Symbol describes an exported top-level symbol recovered from source.
type Symbol struct {
	Name       string
	Kind       string
	Start      int
	End        int
	ScopeDepth int
	Callables  []Callable
}

// ResolveModuleSource resolves a module name to source text using the same
// lookup order as runtime module loading.
func ResolveModuleSource(originPath, module string) (string, string, error) {
	if module == "" {
		return "", "", fmt.Errorf("module name is empty")
	}

	relPath := filepath.Join(strings.Split(module, ".")...) + ".slug"
	candidates := []string{}
	if originPath != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(originPath), relPath))
	}
	if slugHome := strings.TrimSpace(os.Getenv("SLUG_HOME")); slugHome != "" {
		candidates = append(candidates, filepath.Join(slugHome, "lib", relPath))
	}

	for _, path := range candidates {
		src, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		absPath, absErr := filepath.Abs(path)
		if absErr != nil {
			absPath = filepath.Clean(path)
		}
		return absPath, string(src), nil
	}

	if originPath != "" {
		return "", "", fmt.Errorf("could not resolve module %s from %s", module, originPath)
	}
	return "", "", fmt.Errorf("could not resolve module %s", module)
}

// CollectExportedSymbols returns exported top-level symbols and callable shapes
// from a parsed module source.
func CollectExportedSymbols(src string) []Symbol {
	l := lexer.New(src)
	p := parser.New(l, "<modulemeta>", src)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil
	}

	out := []Symbol{}
	for _, st := range program.Statements {
		if ff, ok := st.(*ast.ForeignFunctionDeclaration); ok {
			if !hasTag(ff.Tags, "@export") || ff.Name == nil {
				continue
			}
			start := ff.Name.Token.Position
			end := start + len(ff.Name.Token.Literal)
			out = append(out, Symbol{
				Name:       ff.Name.Value,
				Kind:       "function",
				Start:      start,
				End:        end,
				ScopeDepth: 0,
				Callables: []Callable{{
					Name:       ff.Name.Value,
					TypeParams: append([]string(nil), ff.TypeParams...),
					Parameters: append([]*ast.FunctionParameter(nil), ff.Parameters...),
					ReturnType: strings.TrimSpace(ff.ReturnType),
					Signature:  ff.Signature,
				}},
			})
			continue
		}

		es, ok := st.(*ast.ExpressionStatement)
		if !ok || es.Expression == nil {
			continue
		}

		switch e := es.Expression.(type) {
		case *ast.ValExpression:
			if !hasTag(e.Tags, "@export") {
				continue
			}
			kind := "constant"
			callables := []Callable{}
			if fn, ok := e.Value.(*ast.FunctionLiteral); ok {
				kind = "function"
				callables = append(callables, Callable{
					Name:       topLevelPatternName(e.Pattern),
					TypeParams: append([]string(nil), fn.TypeParams...),
					Parameters: append([]*ast.FunctionParameter(nil), fn.Parameters...),
					ReturnType: strings.TrimSpace(fn.ReturnType),
					Signature:  fn.Signature,
				})
			}
			for _, n := range topLevelPatternNames(e.Pattern) {
				sym := Symbol{
					Name:       n.Name,
					Kind:       kind,
					Start:      n.Start,
					End:        n.End,
					ScopeDepth: 0,
				}
				if len(callables) > 0 {
					sym.Callables = append(sym.Callables, callables...)
				}
				out = append(out, sym)
			}
		case *ast.VarExpression:
			if !hasTag(e.Tags, "@export") {
				continue
			}
			kind := "variable"
			callables := []Callable{}
			if fn, ok := e.Value.(*ast.FunctionLiteral); ok {
				kind = "function"
				callables = append(callables, Callable{
					Name:       topLevelPatternName(e.Pattern),
					TypeParams: append([]string(nil), fn.TypeParams...),
					Parameters: append([]*ast.FunctionParameter(nil), fn.Parameters...),
					ReturnType: strings.TrimSpace(fn.ReturnType),
					Signature:  fn.Signature,
				})
			}
			for _, n := range topLevelPatternNames(e.Pattern) {
				sym := Symbol{
					Name:       n.Name,
					Kind:       kind,
					Start:      n.Start,
					End:        n.End,
					ScopeDepth: 0,
				}
				if len(callables) > 0 {
					sym.Callables = append(sym.Callables, callables...)
				}
				out = append(out, sym)
			}
		}
	}
	return out
}

func hasTag(tags []*ast.Tag, name string) bool {
	for _, t := range tags {
		if t != nil && t.Name == name {
			return true
		}
	}
	return false
}

type patternNameRef struct {
	Name  string
	Start int
	End   int
}

func topLevelPatternNames(pat ast.MatchPattern) []patternNameRef {
	out := []patternNameRef{}
	switch p := pat.(type) {
	case *ast.IdentifierPattern:
		if p != nil && p.Value != nil {
			start := p.Value.Token.Position
			end := start + len(p.Value.Token.Literal)
			out = append(out, patternNameRef{Name: p.Value.Value, Start: start, End: end})
		}
	case *ast.BindingPattern:
		if p != nil && p.Name != nil {
			start := p.Name.Token.Position
			end := start + len(p.Name.Token.Literal)
			out = append(out, patternNameRef{Name: p.Name.Value, Start: start, End: end})
		}
	}
	return out
}

func topLevelPatternName(pat ast.MatchPattern) string {
	names := topLevelPatternNames(pat)
	if len(names) == 0 {
		return ""
	}
	return names[0].Name
}

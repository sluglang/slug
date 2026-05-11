package vm

import (
	"bytes"
	"fmt"
	"slug/internal/ast"
	"slug/internal/object"
	"strings"
)

type VMFunction struct {
	Name       string
	Tags       map[string]object.List
	Params     []VMParam
	ParamIndex map[string]int
	Chunk      *Chunk
	Closure    *object.Environment
	Signature  ast.FSig
	Parameters []*ast.FunctionParameter
}

type VMParam struct {
	Name       string
	IsVariadic bool
	Default    *Chunk
	Tags       []*ast.Tag
}

func (f *VMFunction) Type() object.ObjectType { return object.FUNCTION_OBJ }

func (f *VMFunction) Inspect() string {
	var out bytes.Buffer
	out.WriteString("fn(")
	parts := make([]string, 0, len(f.Params))
	for _, p := range f.Params {
		name := p.Name
		if p.IsVariadic {
			name = "..." + name
		}
		if p.Default != nil {
			name += "=<default>"
		}
		parts = append(parts, name)
	}
	out.WriteString(strings.Join(parts, ", "))
	out.WriteString(") { <vm bytecode> }")
	if f.Name != "" {
		return fmt.Sprintf("%s %s", f.Name, out.String())
	}
	return out.String()
}

func (f *VMFunction) HasTag(tag string) bool {
	_, ok := f.Tags[tag]
	return ok
}

func (f *VMFunction) GetTagParams(tag string) (object.List, bool) {
	v, ok := f.Tags[tag]
	return v, ok
}

func (f *VMFunction) GetTags() map[string]object.List {
	if f.Tags == nil {
		f.Tags = map[string]object.List{}
	}
	return f.Tags
}

func (f *VMFunction) SetTag(tag string, params object.List) {
	if f.Tags == nil {
		f.Tags = map[string]object.List{}
	}
	f.Tags[tag] = params
}

func (f *VMFunction) GetSignature() ast.FSig {
	return f.Signature
}

func (f *VMFunction) GetParameters() []*ast.FunctionParameter {
	return f.Parameters
}

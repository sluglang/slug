package vm

import (
	"bytes"
	"fmt"
	"slug/internal/ast"
	"slug/internal/object"
	"strings"
)

type VMFunction struct {
	Name    string
	Params  []VMParam
	Chunk   *Chunk
	Closure *object.Environment
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

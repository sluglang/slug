package vm

import (
	"bytes"
	"fmt"
	"slug/internal/object"
	"strings"
)

type VMFunction struct {
	Name    string
	Params  []string
	Chunk   *Chunk
	Closure *object.Environment
}

func (f *VMFunction) Type() object.ObjectType { return object.FUNCTION_OBJ }

func (f *VMFunction) Inspect() string {
	var out bytes.Buffer
	out.WriteString("fn(")
	out.WriteString(strings.Join(f.Params, ", "))
	out.WriteString(") { <vm bytecode> }")
	if f.Name != "" {
		return fmt.Sprintf("%s %s", f.Name, out.String())
	}
	return out.String()
}

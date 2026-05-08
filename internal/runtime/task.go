package runtime

import (
	"slug/internal/ast"
	"slug/internal/vm"
)

type VMCallContext = vm.VMCallContext

type NurseryScope = vm.NurseryScope

func buildParamIndex(params []*ast.FunctionParameter) map[string]int {
	index := make(map[string]int, len(params))
	for i, param := range params {
		index[param.Name.Value] = i
	}
	return index
}

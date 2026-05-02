package runtime

import (
	"fmt"
	"slug/internal/object"
)

// FindMainEntrypoint returns the unique function tagged with @main in a module env.
// It returns (nil, nil) when the module has no @main.
func FindMainEntrypoint(env *object.Environment) (object.Object, error) {
	var found object.Object

	for _, binding := range env.Bindings {
		if binding == nil || binding.Value == nil {
			continue
		}
		if binding.Meta.IsImport {
			continue
		}
		tagged := taggedMainFunctions(binding.Value)
		for _, fn := range tagged {
			if found != nil && found != fn {
				return nil, fmt.Errorf("a module may define at most one @main function")
			}
			found = fn
		}
	}

	return found, nil
}

func taggedMainFunctions(value object.Object) []object.Object {
	var result []object.Object

	switch v := value.(type) {
	case *object.FunctionGroup:
		for _, fn := range v.Functions {
			if hasMainTag(fn) {
				result = append(result, fn)
			}
		}
	case *object.Function, *object.Foreign:
		if hasMainTag(v) {
			result = append(result, v)
		}
	default:
		if value != nil && value.Type() == object.FUNCTION_OBJ && hasMainTag(value) {
			result = append(result, value)
		}
	}

	return result
}

func hasMainTag(value object.Object) bool {
	taggable, ok := value.(object.Taggable)
	if !ok {
		return false
	}
	return taggable.HasTag(object.MAIN_TAG)
}

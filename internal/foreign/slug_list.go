package foreign

import (
	"slug/internal/dec64"
	"slug/internal/object"
	"sort"
)

func fnListSortWithComparator() *object.Foreign {
	return &object.Foreign{
		Name: "sortWithComparator",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			// Check if there are exactly two arguments: the list and the comparator.
			if len(args) != 2 {
				return ctx.NewError("wrong number of arguments. got=%d, want=2", len(args))
			}

			// Ensure the first argument is a LIST_OBJ (the list to sort).
			listObj, ok := args[0].(*object.List)
			if !ok {
				return ctx.NewError("first argument to `sortWithComparator` must be a LIST, got=%s", args[0].Type())
			}

			comparator := args[1]
			switch comparator.Type() {
			case object.FUNCTION_OBJ, object.FUNCTION_GROUP_OBJ, object.FOREIGN_OBJ:
				// accepted callable types
			default:
				return ctx.NewError("second argument to `sortWithComparator` must be callable, got=%s", comparator.Type())
			}

			call := func(args []object.Object) object.Object {
				return ctx.ApplyFunction(0, "", comparator, args, nil)
			}

			// Sorting logic using the custom comparator.
			elements := listObj.Elements
			sortedElements := make([]object.Object, len(elements))
			copy(sortedElements, elements)

			// Use Go's sort.Slice with a custom comparison using the provided comparator.
			sort.Slice(sortedElements, func(i, j int) bool {
				// Apply the comparator function to the pair of elements.
				args := []object.Object{sortedElements[i], sortedElements[j]}
				callResult := call(args)

				// Ensure the comparator result is a NUMBER_OBJ.
				resultObj, ok := callResult.(*object.Number)
				if !ok {
					return false
				}

				// A negative number indicates a < b, zero indicates a == b, and positive indicates a > b.
				return resultObj.Value.Lt(dec64.ZERO)
			})

			// Return a new sorted List object.
			return &object.List{Elements: sortedElements}
		},
	}
}

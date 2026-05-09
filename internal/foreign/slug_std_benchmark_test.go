package foreign

import (
	"testing"

	"slug/internal/dec64"
	"slug/internal/object"
	"slug/internal/util"
)

func BenchmarkStdMapPersistentOps(b *testing.B) {
	ctx := testRuntimeContext{}
	putFn := fnStdPut().Fn
	removeFn := fnStdRemove().Fn

	keys := make([]*object.Number, 220)
	for i := range keys {
		keys[i] = &object.Number{Value: dec64.FromInt(i + 1)}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := &object.Map{}
		for _, k := range keys {
			out := putFn(ctx, m, k, k)
			next, ok := out.(*object.Map)
			if !ok {
				b.Fatalf("put returned non-map: %T", out)
			}
			m = next
		}
		for _, k := range keys {
			out := removeFn(ctx, m, k)
			next, ok := out.(*object.Map)
			if !ok {
				b.Fatalf("remove returned non-map: %T", out)
			}
			m = next
		}
		if m.Len() != 0 {
			b.Fatalf("expected empty map, got len=%d", m.Len())
		}
	}
}

type sortBenchContext struct{}

func (sortBenchContext) CurrentEnv() *object.Environment { return nil }
func (sortBenchContext) ApplyFunction(pos int, fnName string, fnObj object.Object, positional []object.Object, named map[string]object.Object) object.Object {
	l, lok := positional[0].(*object.Number)
	r, rok := positional[1].(*object.Number)
	if !lok || !rok {
		return &object.Number{Value: dec64.ZERO}
	}
	return &object.Number{Value: l.Value.Sub(r.Value)}
}
func (sortBenchContext) NewError(message string, a ...interface{}) *object.Error {
	return &object.Error{Message: "bench error"}
}
func (sortBenchContext) Nil() *object.Nil { return object.NIL }
func (sortBenchContext) NativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return object.TRUE
	}
	return object.FALSE
}
func (sortBenchContext) LoadModule(pathParts string) (*object.Module, error) { return nil, nil }
func (sortBenchContext) GetConfiguration() util.Configuration {
	return util.Configuration{Version: "bench"}
}
func (sortBenchContext) NextHandleID() int64 { return 0 }

func benchmarkNumbersList(n int) *object.List {
	els := make([]object.Object, n)
	for i := 0; i < n; i++ {
		// descending sequence to keep sort work stable
		els[i] = &object.Number{Value: dec64.FromInt(n - i)}
	}
	return &object.List{Elements: els}
}

func BenchmarkStdUpdatePersistentOps(b *testing.B) {
	ctx := testRuntimeContext{}
	updateFn := fnStdUpdate().Fn
	base := benchmarkNumbersList(220)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lst := base
		for j := 0; j < 220; j++ {
			idx := &object.Number{Value: dec64.FromInt(j)}
			val := &object.Number{Value: dec64.FromInt(j + 1)}
			out := updateFn(ctx, lst, idx, val)
			next, ok := out.(*object.List)
			if !ok {
				b.Fatalf("update returned non-list: %T", out)
			}
			lst = next
		}
	}
}

func BenchmarkStdSwapPersistentOps(b *testing.B) {
	ctx := testRuntimeContext{}
	swapFn := fnStdSwap().Fn
	base := benchmarkNumbersList(220)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lst := base
		for j := 0; j < 220; j++ {
			i1 := &object.Number{Value: dec64.FromInt(j)}
			i2 := &object.Number{Value: dec64.FromInt(219 - j)}
			out := swapFn(ctx, lst, i1, i2)
			next, ok := out.(*object.List)
			if !ok {
				b.Fatalf("swap returned non-list: %T", out)
			}
			lst = next
		}
	}
}

func BenchmarkListSortWithComparator(b *testing.B) {
	ctx := sortBenchContext{}
	sortFn := fnListSortWithComparator().Fn
	comparator := &object.Foreign{Name: "cmp"}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lst := benchmarkNumbersList(220)
		out := sortFn(ctx, lst, comparator)
		next, ok := out.(*object.List)
		if !ok {
			b.Fatalf("sortWithComparator returned non-list: %T", out)
		}
		if len(next.Elements) != 220 {
			b.Fatalf("unexpected sorted len=%d", len(next.Elements))
		}
	}
}

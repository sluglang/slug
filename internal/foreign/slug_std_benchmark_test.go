package foreign

import (
	"testing"

	"slug/internal/dec64"
	"slug/internal/object"
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

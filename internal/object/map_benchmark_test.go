package object

import (
	"fmt"
	"testing"

	"slug/internal/dec64"
)

func BenchmarkMapBackendsPersistentUpdateChain(b *testing.B) {
	keys := benchmarkSymbolKeys(512)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := &Map{}
		for j := 0; j < len(keys); j++ {
			m = persistentPutForBench(m, keys[j], &Number{Value: dec64.FromInt(j)})
		}
		if m.Len() != len(keys) {
			b.Fatalf("unexpected map size: got=%d want=%d", m.Len(), len(keys))
		}
	}
}

func BenchmarkMapBackendsPersistentBranchFanout(b *testing.B) {
	baseKeys := benchmarkSymbolKeys(256)
	branchKeys := benchmarkSymbolKeys(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base := &Map{}
		for j := 0; j < len(baseKeys); j++ {
			base = persistentPutForBench(base, baseKeys[j], &Number{Value: dec64.FromInt(j)})
		}
		branches := make([]*Map, len(branchKeys))
		for j := 0; j < len(branchKeys); j++ {
			branches[j] = persistentPutForBench(base, branchKeys[j], &Number{Value: dec64.FromInt(1000 + j)})
		}
		total := 0
		for _, br := range branches {
			total += br.Len()
		}
		if total == 0 {
			b.Fatal("unexpected empty benchmark result")
		}
	}
}

func benchmarkSymbolKeys(n int) []*Symbol {
	keys := make([]*Symbol, n)
	for i := 0; i < n; i++ {
		keys[i] = InternSymbol(fmt.Sprintf("k%d", i))
	}
	return keys
}

func persistentPutForBench(base *Map, key Hashable, value Object) *Map {
	next := &Map{
		storage: base.ensureStorage(),
	}
	next.Put(key, value)
	return next
}

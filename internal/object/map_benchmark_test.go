package object

import (
	"fmt"
	"testing"

	"slug/internal/dec64"
)

func BenchmarkMapBackendsPersistentUpdateChain(b *testing.B) {
	runMapBackendBenchmark(b, func(b *testing.B) {
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
	})
}

func BenchmarkMapBackendsPersistentBranchFanout(b *testing.B) {
	runMapBackendBenchmark(b, func(b *testing.B) {
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
	})
}

func runMapBackendBenchmark(b *testing.B, fn func(*testing.B)) {
	backends := []string{"native", "hamt"}
	for _, backend := range backends {
		backend := backend
		b.Run(backend, func(b *testing.B) {
			prev := defaultMapBackend
			defer func() {
				defaultMapBackend = prev
			}()
			SetDefaultMapBackend(backend)
			fn(b)
		})
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
	switch defaultMapBackend {
	case mapBackendHAMT:
		next := &Map{
			storage: base.ensureStorage(),
		}
		next.Put(key, value)
		return next
	default:
		next := &Map{
			Pairs: make(map[MapKey]MapPair, base.Len()+1),
		}
		base.ForEach(func(k MapKey, p MapPair) bool {
			next.Pairs[k] = p
			return true
		})
		next.Put(key, value)
		return next
	}
}

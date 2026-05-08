package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"slug/internal/object"
)

var runtimeExecutionSupportedFixtures = map[string]struct{}{
	"arithmetic.slug":              {},
	"builtin-len.slug":             {},
	"closure.slug":                 {},
	"function-default-params.slug": {},
	"function-variadic.slug":       {},
	"list-nested-map-pattern.slug": {},
	"map-dot.slug":                 {},
	"match-binding-pattern.slug":   {},
	"match-expression.slug":        {},
	"match-guard.slug":             {},
	"named-and-spread.slug":        {},
	"oob-and-negative-index.slug":  {},
	"slicing.slug":                 {},
	"struct-binding-pattern.slug":  {},
}

var runtimeExecutionErrorFixtures = map[string]struct{}{
	"const-reassign.slug":               {},
	"invalid-call-target.slug":          {},
	"map-key-type.slug":                 {},
	"missing-required-arg.slug":         {},
	"non-hashable-map-literal-key.slug": {},
	"unknown-identifier.slug":           {},
}

func TestRuntimeExecutionSupportedFixtures(t *testing.T) {
	root := repoRoot(t)
	supportedDir := filepath.Join(root, "tests", "vm-conformance", "supported")
	entries, err := os.ReadDir(supportedDir)
	if err != nil {
		t.Fatalf("read fixtures dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".slug") {
			continue
		}
		name := entry.Name()
		if _, keep := runtimeExecutionSupportedFixtures[name]; !keep {
			continue
		}
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(supportedDir, name)
			sourceBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture %s: %v", name, err)
			}
			source := string(sourceBytes)

			run := runProgramForRuntimeIntegration(t, path, source)
			if run.result.Type() == object.ERROR_OBJ {
				t.Fatalf("supported runtime execution fixture must succeed, got error:\n%s", run.result.Inspect())
			}
		})
	}
}

func TestRuntimeExecutionExpectedErrorFixtures(t *testing.T) {
	root := repoRoot(t)
	errorDir := filepath.Join(root, "tests", "vm-conformance", "error-parity")
	entries, err := os.ReadDir(errorDir)
	if err != nil {
		t.Fatalf("read error fixtures dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".slug") {
			continue
		}
		name := entry.Name()
		if _, keep := runtimeExecutionErrorFixtures[name]; !keep {
			continue
		}
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(errorDir, name)
			sourceBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture %s: %v", name, err)
			}
			source := string(sourceBytes)

			run := runProgramForRuntimeIntegration(t, path, source)
			if run.result.Type() != object.ERROR_OBJ {
				t.Fatalf("runtime execution error fixture must fail, got %T (%s)", run.result, run.result.Inspect())
			}
		})
	}
}

func TestRuntimeFixturePartitionIsCompleteAndDisjoint(t *testing.T) {
	root := repoRoot(t)
	allSupported := readFixtureSet(t, filepath.Join(root, "tests", "vm-conformance", "supported"))
	allErrors := readFixtureSet(t, filepath.Join(root, "tests", "vm-conformance", "error-parity"))

	assertDisjoint(t, runtimeBoundarySupportedFixtures, runtimeExecutionSupportedFixtures, "supported")
	assertDisjoint(t, runtimeBoundaryErrorFixtures, runtimeExecutionErrorFixtures, "error")

	assertComplete(t, allSupported, runtimeBoundarySupportedFixtures, runtimeExecutionSupportedFixtures, "supported")
	assertComplete(t, allErrors, runtimeBoundaryErrorFixtures, runtimeExecutionErrorFixtures, "error")
}

func readFixtureSet(t *testing.T, dir string) map[string]struct{} {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures dir %s: %v", dir, err)
	}
	out := make(map[string]struct{})
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".slug") {
			continue
		}
		out[e.Name()] = struct{}{}
	}
	return out
}

func assertDisjoint(t *testing.T, a, b map[string]struct{}, kind string) {
	t.Helper()
	for name := range a {
		if _, ok := b[name]; ok {
			t.Fatalf("%s fixture %q is assigned to both boundary and execution suites", kind, name)
		}
	}
}

func assertComplete(t *testing.T, all, a, b map[string]struct{}, kind string) {
	t.Helper()
	for name := range all {
		if _, ok := a[name]; ok {
			continue
		}
		if _, ok := b[name]; ok {
			continue
		}
		t.Fatalf("%s fixture %q is not assigned to any runtime suite", kind, name)
	}
}

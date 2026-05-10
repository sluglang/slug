package semantic_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slug/internal/lexer"
	"slug/internal/parser"
	"slug/internal/semantic"
	"strings"
	"testing"
)

type semanticSnapshot struct {
	Errors      int            `json:"errors"`
	Warnings    int            `json:"warnings"`
	TraceEvents map[string]int `json:"trace_events"`
}

func TestSemanticConformanceSnapshots(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	modules := []string{
		filepath.Join("lib", "slug", "std.slug"),
		filepath.Join("lib", "slug", "mustache.slug"),
		filepath.Join("lib", "slug", "json.slug"),
		filepath.Join("lib", "slug", "web", "response.slug"),
		filepath.Join("lib", "slug", "crypto.slug"),
	}

	got := map[string]semanticSnapshot{}
	for _, rel := range modules {
		abs := filepath.Join(repoRoot, rel)
		src, err := os.ReadFile(abs)
		if err != nil {
			t.Fatalf("failed to read %s: %v", rel, err)
		}
		l := lexer.New(string(src))
		p := parser.New(l, abs, string(src))
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("unexpected parser errors in %s: %v", rel, p.Errors())
		}
		var trace bytes.Buffer
		errs, warns := semantic.AnalyzeWithOptions(abs, string(src), program, semantic.AnalyzeOptions{
			EnableTypeCheck: true,
			TypeCheckTrace:  true,
			TraceWriter:     &trace,
		})
		got[rel] = semanticSnapshot{
			Errors:      len(errs),
			Warnings:    len(warns),
			TraceEvents: countTraceEvents(trace.String()),
		}
	}

	goldenPath := filepath.Join("testdata", "semantic_conformance_golden.json")
	if os.Getenv("UPDATE_SEMANTIC_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		b, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal golden: %v", err)
		}
		if err := os.WriteFile(goldenPath, append(b, '\n'), 0o644); err != nil {
			t.Fatalf("failed to write golden: %v", err)
		}
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}
	want := map[string]semanticSnapshot{}
	if err := json.Unmarshal(wantBytes, &want); err != nil {
		t.Fatalf("failed to unmarshal golden file %s: %v", goldenPath, err)
	}

	if !reflect.DeepEqual(want, got) {
		gotBytes, _ := json.MarshalIndent(got, "", "  ")
		t.Fatalf("semantic conformance snapshot mismatch\nwant:\n%s\n\ngot:\n%s\n\nre-run with UPDATE_SEMANTIC_GOLDEN=1 to update", string(wantBytes), string(gotBytes))
	}
}

func countTraceEvents(trace string) map[string]int {
	out := map[string]int{}
	for _, line := range strings.Split(trace, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "TypeTrace: ") {
			continue
		}
		rest := strings.TrimPrefix(line, "TypeTrace: ")
		parts := strings.Fields(rest)
		if len(parts) == 0 {
			continue
		}
		event := strings.TrimSpace(parts[0])
		if event == "" {
			continue
		}
		out[event]++
	}
	return out
}

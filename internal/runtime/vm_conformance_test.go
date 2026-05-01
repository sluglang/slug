package runtime

import (
	"os"
	"path/filepath"
	"regexp"
	"slug/internal/lexer"
	"slug/internal/object"
	"slug/internal/parser"
	"slug/internal/util"
	"strings"
	"testing"
)

func TestVMConformanceFixtures(t *testing.T) {
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
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(supportedDir, name)
			sourceBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture %s: %v", name, err)
			}
			source := string(sourceBytes)

			treewalk := runProgramForConformance(t, RuntimeTreewalk, path, source)
			vm := runProgramForConformance(t, RuntimeVM, path, source)

			if treewalk.Type() == object.ERROR_OBJ || vm.Type() == object.ERROR_OBJ {
				t.Fatalf("supported fixture must succeed in both runtimes treewalk=%T vm=%T\n--- treewalk ---\n%s\n--- vm ---\n%s", treewalk, vm, treewalk.Inspect(), vm.Inspect())
			}
			if treewalk.Inspect() != vm.Inspect() {
				t.Fatalf("inspect mismatch\n--- treewalk ---\n%s\n--- vm ---\n%s", treewalk.Inspect(), vm.Inspect())
			}
		})
	}
}

func TestVMKnownUnsupportedFixtures(t *testing.T) {
	root := repoRoot(t)
	unsupportedDir := filepath.Join(root, "tests", "vm-conformance", "known-unsupported")
	entries, err := os.ReadDir(unsupportedDir)
	if err != nil {
		t.Fatalf("read unsupported fixtures dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".slug") {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(unsupportedDir, name)
			sourceBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture %s: %v", name, err)
			}
			source := string(sourceBytes)

			treewalk := runProgramForConformance(t, RuntimeTreewalk, path, source)
			if treewalk.Type() == object.ERROR_OBJ {
				t.Fatalf("unsupported fixture must succeed on treewalk, got error:\n%s", treewalk.Inspect())
			}

			vm := runProgramForConformance(t, RuntimeVM, path, source)
			if vm.Type() != object.ERROR_OBJ {
				t.Fatalf("unsupported fixture expected VM error, got %T (%s)", vm, vm.Inspect())
			}
		})
	}
}

func TestVMConformanceErrorParityFixtures(t *testing.T) {
	root := repoRoot(t)
	errorDir := filepath.Join(root, "tests", "vm-conformance", "error-parity")
	entries, err := os.ReadDir(errorDir)
	if err != nil {
		t.Fatalf("read error-parity fixtures dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".slug") {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(errorDir, name)
			sourceBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture %s: %v", name, err)
			}
			source := string(sourceBytes)

			treewalk := runProgramForConformance(t, RuntimeTreewalk, path, source)
			vm := runProgramForConformance(t, RuntimeVM, path, source)

			if treewalk.Type() != object.ERROR_OBJ {
				t.Fatalf("error-parity fixture must fail on treewalk, got %T (%s)", treewalk, treewalk.Inspect())
			}
			if vm.Type() != object.ERROR_OBJ {
				t.Fatalf("error-parity fixture must fail on vm, got %T (%s)", vm, vm.Inspect())
			}

			sigTreewalk := canonicalErrorSignature(treewalk.Inspect())
			sigVM := canonicalErrorSignature(vm.Inspect())
			if sigTreewalk != sigVM {
				t.Fatalf("canonical error mismatch\n--- treewalk ---\n%s\n--- vm ---\n%s", sigTreewalk, sigVM)
			}
		})
	}
}

var (
	reLineCol = regexp.MustCompile(`\[[ ]*\d+:[ ]*\d+\]`)
	rePathSep = regexp.MustCompile(`/[^ \n]+\.slug`)
	reCodeCtx = regexp.MustCompile(`^\d+\s+\|`)
)

func normalizeErrorInspect(in string) string {
	lines := strings.Split(in, "\n")
	out := make([]string, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		// Remove unstable stack/context lines.
		if strings.HasPrefix(line, "-->") || strings.HasPrefix(line, ">") {
			continue
		}
		if strings.Contains(line, "^ unexpected here") {
			continue
		}
		if reCodeCtx.MatchString(line) {
			continue
		}
		if strings.Contains(line, "Stacktrace:") || strings.HasPrefix(line, "at [") {
			continue
		}
		line = reLineCol.ReplaceAllString(line, "[L:C]")
		line = rePathSep.ReplaceAllString(line, "/<path>.slug")
		line = strings.TrimPrefix(line, "Error: ")
		if strings.HasPrefix(line, "vm runtime error at pos ") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				line = parts[1]
			}
		} else if strings.HasPrefix(line, "vm runtime error: ") {
			line = strings.TrimPrefix(line, "vm runtime error: ")
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func canonicalErrorSignature(inspect string) string {
	normalized := normalizeErrorInspect(inspect)
	lower := strings.ToLower(normalized)

	category := "other"
	switch {
	case strings.Contains(lower, "failed to import"),
		strings.Contains(lower, "could not load module"),
		strings.Contains(lower, "parse errors in module"):
		category = "import-load"
	case strings.Contains(lower, "identifier not found"):
		category = "identifier-not-found"
	case strings.Contains(lower, "failed to assign to val"):
		category = "assign-to-val"
	case strings.Contains(lower, "missing required parameter"),
		strings.Contains(lower, "no suitable function"):
		category = "call-arity"
	case strings.Contains(lower, "unusable as map key"):
		category = "map-key-type"
	}

	switch category {
	case "identifier-not-found", "assign-to-val", "call-arity", "map-key-type", "import-load":
		return category
	default:
		return category + "::" + normalized
	}
}

func runProgramForConformance(t *testing.T, mode, scriptPath, source string) object.Object {
	t.Helper()

	l := lexer.New(source)
	p := parser.New(l, scriptPath, source)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors in %s: %v", scriptPath, p.Errors())
	}

	cfg := util.Configuration{
		RootPath:     filepath.Dir(scriptPath),
		ProjectRoot:  filepath.Dir(scriptPath),
		Cwd:          filepath.Dir(scriptPath),
		SlugHome:     repoRoot(t),
		DefaultLimit: 4,
		MainModule:   strings.TrimSuffix(filepath.Base(scriptPath), ".slug"),
		RuntimeMode:  mode,
	}
	rt := NewRuntime(cfg)
	if rt.Modules == nil {
		rt.Modules = make(map[string]*object.Module)
	}

	env := object.NewRootEnvironment(cfg.DefaultLimit)
	env.Path = scriptPath
	env.LibRoot = cfg.ProjectRoot
	env.Src = source
	env.ModuleFqn = cfg.MainModule

	rt.Modules[cfg.MainModule] = &object.Module{
		Name:    cfg.MainModule,
		Path:    scriptPath,
		Src:     source,
		Program: program,
		Env:     env,
		Doc:     program.ModuleDoc,
		HasDoc:  program.HasModuleDoc,
	}

	return ExecuteProgram(mode, rt, env, program)
}

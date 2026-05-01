package runtime

import (
	"os"
	"path/filepath"
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
		})
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

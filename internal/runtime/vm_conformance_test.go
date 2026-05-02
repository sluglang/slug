package runtime

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slug/internal/lexer"
	"slug/internal/object"
	"slug/internal/parser"
	"slug/internal/util"
	"strings"
	"testing"
)

type conformanceRun struct {
	result object.Object
	stdout string
	stderr string
}

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

			vm := runProgramForConformance(t, path, source)
			if vm.result.Type() == object.ERROR_OBJ {
				t.Fatalf("supported fixture must succeed on vm, got error:\n%s", vm.result.Inspect())
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

			vm := runProgramForConformance(t, path, source)
			if vm.result.Type() != object.ERROR_OBJ {
				t.Fatalf("unsupported fixture expected VM error, got %T (%s)", vm.result, vm.result.Inspect())
			}
		})
	}
}

func TestVMConformanceExpectedErrorFixtures(t *testing.T) {
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

			vm := runProgramForConformance(t, path, source)
			if vm.result.Type() != object.ERROR_OBJ {
				t.Fatalf("error-parity fixture must fail on vm, got %T (%s)", vm.result, vm.result.Inspect())
			}
		})
	}
}

func runProgramForConformance(t *testing.T, scriptPath, source string) conformanceRun {
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
		RuntimeMode:  RuntimeVM,
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

	result, stdout, stderr := captureExecutionOutput(func() object.Object {
		prevLogger := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		defer slog.SetDefault(prevLogger)
		return ExecuteProgram(rt, env, program)
	})

	return conformanceRun{
		result: result,
		stdout: normalizeOutput(stdout),
		stderr: normalizeOutput(stderr),
	}
}

func captureExecutionOutput(run func() object.Object) (object.Object, string, string) {
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	doneOut := make(chan struct{})
	doneErr := make(chan struct{})
	go func() {
		_, _ = io.Copy(&outBuf, stdoutR)
		close(doneOut)
	}()
	go func() {
		_, _ = io.Copy(&errBuf, stderrR)
		close(doneErr)
	}()

	result := run()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	<-doneOut
	<-doneErr
	_ = stdoutR.Close()
	_ = stderrR.Close()

	return result, outBuf.String(), errBuf.String()
}

func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimSpace(s)
}

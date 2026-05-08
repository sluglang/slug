package runtime

import (
	"os"
	"path/filepath"
	"slug/internal/object"
	"slug/internal/util"
	"slug/internal/vm"
	"testing"
	"time"
)

func TestStdinReadLinesSingletonAndClose(t *testing.T) {
	oldStdin := os.Stdin
	t.Cleanup(func() { os.Stdin = oldStdin })

	stdinFile, err := os.CreateTemp(t.TempDir(), "stdin-readLines-*")
	if err != nil {
		t.Fatalf("create temp stdin file: %v", err)
	}
	t.Cleanup(func() { _ = stdinFile.Close() })

	// Includes CRLF, an empty line, and a final partial line without newline.
	if _, err := stdinFile.WriteString("a\r\n\nb"); err != nil {
		t.Fatalf("write temp stdin file: %v", err)
	}
	if _, err := stdinFile.Seek(0, 0); err != nil {
		t.Fatalf("rewind temp stdin file: %v", err)
	}
	os.Stdin = stdinFile

	cfg := util.Configuration{
		RootPath:     t.TempDir(),
		SlugHome:     repoRoot(t),
		DefaultLimit: 4,
	}
	rt := NewRuntime(cfg)
	task := vm.NewVMCallContext(vm.VMCallContextDeps{
		Config:       rt.Config,
		LoadModule:   rt.LoadModule,
		NextHandleID: rt.NextHandleID,
		BridgeFactory: func(callEnv *object.Environment) func(pos int, callee object.Object, positional []object.Object, named map[string]object.Object) object.Object {
			return makeVMCallBridge(rt, callEnv)
		},
	})
	task.PushNurseryScope(&NurseryScope{Limit: make(chan struct{}, cfg.DefaultLimit)})
	task.PushEnv(object.NewRootEnvironment(cfg.DefaultLimit))

	fn, ok := rt.LookupForeign("slug.io.stdin.readLines")
	if !ok {
		t.Fatal("foreign `slug.io.stdin.readLines` not registered")
	}

	first := fn.Fn(task)
	ch1, ok := first.(*object.Channel)
	if !ok {
		t.Fatalf("readLines() first call returned %T, want *object.Channel", first)
	}

	second := fn.Fn(task)
	ch2, ok := second.(*object.Channel)
	if !ok {
		t.Fatalf("readLines() second call returned %T, want *object.Channel", second)
	}

	if ch1 != ch2 {
		t.Fatal("readLines() did not return a singleton channel")
	}

	recv := func() (object.Object, bool) {
		t.Helper()
		select {
		case v, ok := <-ch1.GoChan():
			return v, ok
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for stdin event")
			return nil, false
		}
	}

	assertString := func(v object.Object, ok bool, expected string) {
		t.Helper()
		if !ok {
			t.Fatalf("channel closed before receiving %q", expected)
		}
		s, ok := v.(*object.String)
		if !ok {
			t.Fatalf("got %T, want *object.String(%q)", v, expected)
		}
		if s.Value != expected {
			t.Fatalf("got string %q, want %q", s.Value, expected)
		}
	}

	v1, ok1 := recv()
	assertString(v1, ok1, "a")

	v2, ok2 := recv()
	assertString(v2, ok2, "")

	v3, ok3 := recv()
	assertString(v3, ok3, "b")

	v4, ok4 := recv()
	if ok4 {
		t.Fatalf("expected closed channel, got extra value %s", v4.Inspect())
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("failed to find repo root (go.mod)")
		}
		dir = parent
	}
}

package runtime

import (
	"slug/internal/util"
	"testing"
)

func TestUseVMModuleLoaderDefaultsFalse(t *testing.T) {
	t.Setenv("SLUG_VM_MODULE_LOADER", "")
	rt := NewRuntime(util.Configuration{
		RootPath:    ".",
		ProjectRoot: ".",
		Cwd:         ".",
		SlugHome:    "",
		MainModule:  "test",
	})
	if rt.useVMModuleLoader() {
		t.Fatal("expected VM module loader to be disabled by default")
	}
}

func TestUseVMModuleLoaderEnabledByConfigStore(t *testing.T) {
	t.Setenv("SLUG_VM_MODULE_LOADER", "")
	cfg := util.Configuration{
		RootPath:    ".",
		ProjectRoot: ".",
		Cwd:         ".",
		SlugHome:    "",
		MainModule:  "test",
	}
	cfg.Store = &util.ConfigStore{Values: map[string]interface{}{
		"runtime.vm-module-loader": "true",
	}}
	rt := NewRuntime(cfg)
	rt.Config.Store = cfg.Store
	if !rt.useVMModuleLoader() {
		t.Fatal("expected VM module loader to be enabled by runtime.vm-module-loader")
	}
}

func TestUseVMModuleLoaderEnvOverride(t *testing.T) {
	t.Setenv("SLUG_VM_MODULE_LOADER", "true")
	cfg := util.Configuration{
		RootPath:    ".",
		ProjectRoot: ".",
		Cwd:         ".",
		SlugHome:    "",
		MainModule:  "test",
	}
	cfg.Store = &util.ConfigStore{Values: map[string]interface{}{
		"runtime.vm-module-loader": "false",
	}}
	rt := NewRuntime(cfg)
	rt.Config.Store = cfg.Store
	if !rt.useVMModuleLoader() {
		t.Fatal("expected env override to enable VM module loader")
	}
}

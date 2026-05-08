# Spec Change Log

## 2026-05-07

### Import collision diagnostics use stderr

- Updated import collision warnings in `internal/runtime/slug_fn_builtin.go` to write to `stderr` instead of `stdout`.
- Rationale: collision warnings should not pollute normal program output streams.
- Updated `internal/runtime/slug_fn_builtin_import_test.go` to capture `stderr` for warning assertions.
- Validation performed:
  - `go test ./internal/runtime -run Import -count=1`
  - `make test`

## 2026-05-08

### Runtime/VM refactor Phase 0 seam coverage

- Added `internal/runtime/vm_runtime_seam_test.go` to lock behavior before moving VM execution code out of `internal/runtime`.
- New coverage includes:
  - `prepareProgramForVM` binding registered foreign functions and stripping `foreign` declarations from executable statements.
  - `applyForeignTagsForVM` evaluating tag arguments through the VM execution path.
  - `makeVMCallBridge` handling named arguments, default values, and named variadic list flattening for foreign calls.
- Validation performed:
  - `go test ./internal/runtime -run 'TestPrepareProgramForVMBindsForeignAndStripsDeclaration|TestApplyForeignTagsForVMEvaluatesTagArgs|TestMakeVMCallBridgeBindsNamedAndDefaultArgsForForeign|TestMakeVMCallBridgeFlattensNamedVariadicListForForeign' -count=1`
  - `go test ./internal/runtime ./internal/vm -count=1`

### Runtime/VM refactor Phase 1 prep/tag move

- Moved VM program preparation and tag-expression evaluation implementations into `internal/vm/program_prepare.go`.
- Added VM-owned APIs:
  - `vm.PrepareProgram`
  - `vm.ApplyForeignTags`
  - `vm.EvalTagArgs`
  - `vm.EvalExpr`
- Kept compatibility wrappers in runtime:
  - `internal/runtime/vm_prepare.go`
  - `internal/runtime/vm_tag_eval.go`
- Rationale: shift VM execution concerns out of `internal/runtime` while preserving current call sites and behavior during incremental refactoring.
- Validation performed:
  - `go test ./internal/runtime ./internal/vm -count=1`
  - `go test ./cmd/app ./internal/runtime ./internal/vm -count=1`

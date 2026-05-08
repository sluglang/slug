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

### Runtime/VM refactor Phase 2 call-context/nursery move

- Moved VM call-bridge execution context and nursery implementation from `internal/runtime` into VM-owned files:
  - `internal/vm/call_context.go`
  - `internal/vm/nursery_scope.go`
- Introduced dependency injection for VM call context via `vm.VMCallContextDeps`:
  - runtime configuration
  - module loader
  - handle ID generator
  - bridge factory for nested calls
- Updated runtime execution wiring to construct contexts with `vm.NewVMCallContext(...)` in:
  - `internal/runtime/execute.go`
  - `internal/runtime/slug_io_stdin_test.go`
- Kept compatibility aliases in `internal/runtime/task.go`:
  - `type VMCallContext = vm.VMCallContext`
  - `type NurseryScope = vm.NurseryScope`
- Validation performed:
  - `go test ./internal/runtime ./internal/vm -count=1`
  - `go test ./cmd/app ./internal/runtime ./internal/vm -count=1`

### Runtime/VM refactor Phase 3 runtime shim removal

- Removed transitional runtime wrappers and aliases now that VM-owned APIs are wired directly.
- Updated runtime execution/module-loading call sites to use VM APIs directly:
  - `vm.PrepareProgram`
  - `vm.ApplyForeignTags`
  - `vm.EvalTagArgs`
- Replaced runtime alias usage with direct VM types (`vm.NurseryScope`) and introduced a runtime-local helper for function param indexing used by bridge adaptation.
- Deleted obsolete runtime files:
  - `internal/runtime/vm_prepare.go`
  - `internal/runtime/vm_tag_eval.go`
  - `internal/runtime/task.go`
  - `internal/runtime/nursery.go`
- Updated seam tests to target VM APIs while preserving behavior assertions.
- Validation performed:
  - `go test ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Runtime/VM refactor Phase 4 boundary cleanup

- Tightened VM API surface by making expression-evaluation helper private inside VM prep code:
  - `internal/vm/program_prepare.go`: `EvalExpr` -> `evalExpr`
  - Updated VM-internal call sites accordingly (`internal/vm/call_context.go`).
- Moved VM ownership/boundary notes into a project-level technical section in `README.md` instead of a package-local README.
- Removed `internal/vm/README.md`.
- Validation performed:
  - `go test ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Parser/Semantic split Phase 1 extraction

- Added a new semantic analysis package:
  - `internal/semantic/analyzer.go`
- Moved semantic responsibilities out of parse-time function construction into semantic analysis:
  - recur tail-position validation
  - tailcall flag annotation (`CallExpression.IsTailCall`, `FunctionLiteral.HasTailCall`)
  - `@main` usage validation
  - struct schema placement validation
- Updated parser orchestration in `internal/parser/parser.go`:
  - `ParseProgram` now delegates semantic checks via `semantic.Analyze(...)`
  - Removed parser-local tailcall/recur analysis call from `parseFunctionLiteral`
- Added project pipeline note in `README.md` documenting Parse -> Semantic -> Runtime/VM phases.
- Validation performed:
  - `go test ./internal/parser ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Parser/Semantic split Phase 2 cleanup and semantic tests

- Removed duplicated semantic implementation from `internal/parser/parser.go`:
  - tailcall annotation helpers
  - recur tail-position validation helpers
  - struct schema placement validation helpers
  - `@main` validation helpers
- Kept parser focused on syntax/AST construction and semantic delegation.
- Added focused semantic tests in `internal/semantic/analyzer_test.go`:
  - tailcall flag tagging
  - non-tail `recur` rejection
  - `@main` non-function rejection
  - struct schema placement rejection outside binding RHS
- Validation performed:
  - `go test ./internal/semantic ./internal/parser ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Parser/Semantic split Phase 3 parser-test boundary cleanup

- Moved `@main` semantic behavior assertions out of parser tests and into semantic tests.
- Removed semantic-focused tests from `internal/parser/parser_test.go`:
  - `TestMainTagAllowsDefaultedParameters`
  - `TestMainTagRejectsNonFunctionTarget`
  - `TestMainTagRejectsNonZeroArity`
  - `TestMainTagRejectsMultipleDeclarations`
- Added/expanded equivalent semantic coverage in `internal/semantic/analyzer_test.go`:
  - `TestSemanticAllowsMainWithDefaultedParameters`
  - `TestSemanticRejectsMainOnNonFunction`
  - `TestSemanticRejectsMainNonZeroArity`
  - `TestSemanticRejectsMultipleMainDeclarations`
- Validation performed:
  - `go test ./internal/semantic ./internal/parser ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Parser/Semantic split Phase 4 parser-only error boundary

- Completed parser/semantic boundary separation:
  - `internal/parser.ParseProgram` no longer runs semantic analysis.
  - Parser now reports syntax/lexing errors only.
- Semantic phase is now invoked explicitly at execution/analysis entry points:
  - `internal/runtime/runtime.go` module load path
  - `internal/runtime/execute.go` program execution path
  - VM direct test/benchmark parse helpers (`internal/vm/executor_test.go`, `internal/vm/benchmark_test.go`)
  - Semantic package tests explicitly invoke `semantic.Analyze(...)`.
- Added parser guard test:
  - `TestParserReportsSyntaxOnlyForSemanticInputs` in `internal/parser/parser_test.go`
  - Asserts semantic-only invalid programs do not produce parser errors.
- Validation performed:
  - `go test ./internal/parser ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Runtime conformance retargeting to integration ownership

- Renamed runtime conformance test file from `internal/runtime/vm_conformance_test.go` to `internal/runtime/runtime_conformance_test.go`.
- Reframed suite purpose from VM parity to runtime integration behavior.
- Renamed tests:
  - `TestVMConformanceFixtures` -> `TestRuntimeIntegrationSupportedFixtures`
  - `TestVMConformanceExpectedErrorFixtures` -> `TestRuntimeIntegrationExpectedErrorFixtures`
- Scoped fixture execution to runtime-owned concerns via allowlists:
  - Supported fixtures: `import-bridge`, `stdout-output`, `stderr-output`, `await-expression`, `select-expression`, `spawn-expression`
  - Error fixtures: `import-missing-module`, `import-parse-error`
- Updated helper naming to reflect integration scope:
  - `conformanceRun` -> `runtimeIntegrationRun`
  - `runProgramForConformance` -> `runProgramForRuntimeIntegration`
- Validation performed:
  - `go test ./internal/runtime ./internal/vm ./internal/semantic ./internal/parser ./cmd/app -count=1`

### Runtime fixture suite split and partition guard

- Renamed runtime integration subset suite to explicit boundary naming:
  - `TestRuntimeBoundarySupportedFixtures`
  - `TestRuntimeBoundaryExpectedErrorFixtures`
  - Fixture ownership maps renamed to `runtimeBoundary*`.
- Added second runtime fixture suite for the remaining execution-focused fixtures:
  - `internal/runtime/runtime_execution_fixtures_test.go`
  - `TestRuntimeExecutionSupportedFixtures`
  - `TestRuntimeExecutionExpectedErrorFixtures`
  - Fixture ownership maps `runtimeExecution*`.
- Added fixture partition guard:
  - `TestRuntimeFixturePartitionIsCompleteAndDisjoint`
  - Fails if any fixture is unassigned or assigned to both suites.
- Validation performed:
  - `go test ./internal/runtime ./internal/parser ./internal/semantic ./internal/vm ./cmd/app -count=1`

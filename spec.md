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

### Parser context keyword cleanup for defer modes

- Removed `onsuccess` and `onerror` from global keyword tokenization in `internal/token/token.go`.
- Updated parser to treat defer modes contextually after `defer`:
  - `internal/parser/parser.go`
  - `parseDeferStatement` now recognizes `onsuccess` and `onerror` via identifier literals (`IDENT`), similar to contextual handling patterns used elsewhere.
- Removed `ONSUCCESS`/`ONERROR` token-type usage from parser symbol-name handling.
- Added parser coverage in `internal/parser/parser_test.go`:
  - `TestDeferParsesContextualOnSuccessAndOnError`
  - `TestOnSuccessAndOnErrorAreRegularIdentifiersOutsideDefer`
- Validation performed:
  - `go test ./internal/parser ./internal/runtime ./internal/vm ./internal/semantic ./cmd/app -count=1`

### VM performance optimization batch 1 (latency-first)

- Added VM call-path and select hot-path optimizations:
  - `internal/vm/function.go`: added `VMFunction.ParamIndex` cache for named-parameter lookup.
  - `internal/vm/compiler.go`: populate `ParamIndex` during function compilation.
  - `internal/vm/executor.go`:
    - preserve cached `ParamIndex` in closure binding.
    - `popCallArguments` fast-path for positional-only calls (avoids intermediate argument reshaping overhead).
    - reuse cached `ParamIndex` in `bindVMArgumentsInto` to avoid rebuilding name-to-index maps on named calls.
    - reduce select receive-result allocation churn by reusing package-level `Full`/`Empty` schema singletons.
- Expanded benchmark workloads in `internal/vm/benchmark_test.go` to capture additional hot paths:
  - `call_dispatch`
  - `named_args`
  - `select_await`
  - `spawn_storm`
- Validation performed:
  - `go test ./internal/vm -run '^$' -bench BenchmarkVMExecuteOnly -benchtime=1x`
  - `go test ./internal/parser ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

#### Benchmark summary (post-change sample)

- Ran: `go test ./internal/vm -run '^$' -bench 'BenchmarkVMExecuteOnly/(struct_copy|call_dispatch|named_args|select_await|spawn_storm)$' -benchmem -count=8 -benchtime=300ms`
- Approximate median `ns/op` from post-change runs:
  - `struct_copy`: ~0.66ms
  - `call_dispatch`: ~0.93ms
  - `named_args`: ~0.83ms
  - `select_await`: ~3.8µs
  - `spawn_storm`: ~0.45ms (new benchmark)
- Compared with earlier pre-change smoke readings, observed latency gains:
  - `struct_copy`: ~37% faster
  - `call_dispatch`: ~23% faster
  - `named_args`: ~17% faster
- Note: allocation counts in the call-heavy paths are largely unchanged in this batch; gains are primarily dispatch/binding latency improvements.

### Immutable map storage backend abstraction and HAMT prototype

- Added pluggable map storage backend support in `internal/object`:
  - new `mapStorage` interface with `put/get/len/forEach`.
  - native backend (`nativeMapStorage`) and persistent HAMT backend (`hamtMapStorage`).
  - backend switcher: `object.SetDefaultMapBackend("native"|"hamt")`.
- Added HAMT prototype implementation:
  - `internal/object/hamt.go` with persistent trie insert/get/iteration.
  - deterministic entry traversal reused by `Map.Inspect()` and `Entries()`.
- Refactored `object.Map` to route operations through backend-safe methods:
  - `Put`, `PutPair`, `Get`, `GetPair`, `Len`, `Entries`, `ForEach`.
  - lazy migration path preserves legacy `Pairs` map compatibility.
- Removed direct runtime/VM dependency on `Map.Pairs` for object maps:
  - `internal/runtime/execute.go`
  - `internal/runtime/slug_fn_builtin.go`
  - `internal/vm/executor.go`
  - `internal/vm/task.go`
- Added runtime config plumbing for backend selection:
  - `util.Configuration.MapBackend`
  - CLI flag `-map-backend` in `cmd/app/main.go`
  - backend initialization in `runtime.NewRuntime`.
- Added object tests to validate backend parity and legacy migration:
  - `TestMapSupportsNativeAndHAMTBackends`
  - `TestMapMigratesLegacyPairsWhenHAMTEnabled`
- Validation performed:
  - `go test ./internal/object ./internal/vm ./internal/runtime ./internal/parser ./internal/semantic ./cmd/app -count=1`

### Immutable map allocation-churn follow-up (persistent delete + std map ops)

- Refactored stdlib map primitives to backend-aware persistent operations:
  - `internal/foreign/slug_std.go`
  - `put` now uses `Map.Clone()` + `PutPair(...)`
  - `remove` now uses `Map.Clone()` + `DeleteKey(...)`
- Added persistent delete support to map storage:
  - `mapStorage.del(...)` in `internal/object/map_storage.go`
  - `hamtDelete(...)` persistent path-copy delete in `internal/object/hamt.go`
  - `Map.DeleteKey(...)` and backend-safe `Map.Clone()` in `internal/object/object.go`
- Updated foreign map helper utilities to avoid direct `Pairs` mutation/iteration:
  - `internal/foreign/util.go`
  - `internal/foreign/slug_io_http.go`
- Added tests:
  - `TestMapCloneIsIndependentAcrossBackends`
  - `TestMapDeleteKeyAcrossBackends`
  - in `internal/object/object_test.go`
- Added focused benchmark for std persistent map ops:
  - `BenchmarkStdMapPersistentOps` in `internal/foreign/slug_std_benchmark_test.go`
  - workload: 220 `put` + 220 `remove` immutable chain.

#### Benchmark comparison (before vs after persistent delete)

- Ran: `go test ./internal/foreign -run '^$' -bench 'BenchmarkStdMapPersistentOps' -benchmem -count=5`
- HAMT before (pre-delete implementation): ~9.22ms/op, ~17.97MB/op, ~223,889 allocs/op
- HAMT after: ~0.158ms/op, ~322KB/op, ~4,508 allocs/op
- Approximate HAMT improvement:
  - latency: ~58x faster
  - bytes/op: ~56x lower
  - allocs/op: ~49.7x lower

### HAMT-only map runtime (native backend removal)

- Removed native map backend implementation and backend-switch plumbing.
- `object.Map` is now backed by HAMT storage only:
  - removed backend toggles from `internal/object/map_storage.go`
  - `ensureStorage()` in `internal/object/object.go` always migrates legacy `Pairs` into HAMT storage.
- Removed runtime/config/CLI backend selection:
  - removed `Configuration.MapBackend` from `internal/util/config.go`
  - removed `-map-backend` flag from `cmd/app/main.go`
  - removed map-backend initialization in `internal/runtime/runtime.go`
- Simplified map tests/benchmarks to a single HAMT path:
  - `internal/object/object_test.go`
  - `internal/object/map_benchmark_test.go`
  - `internal/foreign/slug_std_benchmark_test.go`
- Validation performed:
  - `go test ./internal/object ./internal/foreign ./internal/runtime ./internal/vm ./internal/parser ./internal/semantic ./cmd/app -count=1`

## 2026-05-09

### Import collision warnings include importing module context

- Updated import collision warnings in `internal/runtime/slug_fn_builtin.go` to include the current importing module (`ctx.CurrentEnv().ModuleFqn`) in warning text.
- Applies to both warning forms:
    - name collision warnings (`WARNING: import name collision ...`)
    - duplicate function signature collision warnings (`WARNING: import collision in ...`)
- Rationale: helps users identify which module triggered the warning when diagnosing import collisions.
- Updated `internal/runtime/slug_fn_builtin_import_test.go` to assert warnings include importing module context.
- Validation performed:
    - `go test ./internal/runtime -run Import -count=1`

### VM string indexing cache for CSV-heavy workloads

- Added lazy rune caching on `object.String` in `internal/object/object.go`:
    - `Runes()`
    - `RuneCount()`
    - `RuneAt(...)`
- Updated VM string index/slice execution in `internal/vm/executor.go` to reuse cached runes instead of rebuilding `[]rune` per index operation.
- Updated `len(@str)` builtin in `internal/runtime/slug_fn_builtin.go` to use `(*object.String).RuneCount()`.
- Added coverage for rune helper behavior in `internal/object/object_test.go`.
- Added `string_index_scan` execution benchmark case in `internal/vm/benchmark_test.go`.
- Rationale: `playground.slug` CSV parsing performs heavy repeated string indexing (`s[i]`, `s[i+1]`), and full rune conversion per access was the dominant cost.
- Validation performed:
    - `go test ./internal/object ./internal/runtime ./internal/vm ./cmd/app -count=1`
    - `go test ./internal/vm -run '^$' -bench BenchmarkVMExecuteOnly/string_index_scan -benchmem -count=1`
    - `SLUG_HOME=$(pwd) make run ARGS="playground.slug"` (before/after timing)

## 2026-05-09

### VM immutable-list allocation reduction pass (hot paths)

- Optimized list-heavy VM executor paths in `internal/vm/executor.go` to reduce allocation overhead while preserving immutable semantics:
  - `OpListPrepend` (`+:`) now uses fixed-size allocation and `copy`.
  - `OpListAppend` (`:+`) now uses fixed-size allocation and `copy`.
  - `OpAdd` list concatenation now uses a single fixed-size destination with two `copy` operations.
  - `OpMatchListHeadTail` tail binding now uses fixed-size copy rather than append-based cloning.
  - `OpMatchSeqTail` list branch now uses fixed-size copy for extracted tails.
  - List slicing in `evalIndex` now preallocates capacity based on computed slice cardinality via new helper `sequenceSliceLen`.
- Added list-focused VM execute benchmarks in `internal/vm/benchmark_test.go`:
  - `list_append_chain`
  - `list_prepend_chain`
  - `list_concat_chain`
  - `list_match_tail`
- Validation performed:
  - `go test ./internal/vm -count=1`
  - `go test ./internal/vm -run '^$' -bench 'BenchmarkVMExecuteOnly/(match_nested|list_append_chain|list_prepend_chain|list_concat_chain|list_match_tail)$' -benchmem`

### VM immutable-list structural sharing (tail/slice)

- Added structural-sharing behavior for immutable list tail/slice paths in `internal/vm/executor.go`:
  - `OpMatchListHeadTail` now binds `...tail` via shared backing slice (`lst.Elements[1:]`) instead of copying.
  - `OpMatchSeqTail` list branch now returns shared tail slice (`s.Elements[start:]`) instead of copying.
  - `evalIndex` list slicing now returns shared contiguous sub-slices when `step == 1`.
- Rationale: immutable list semantics permit backing-slice sharing; this removes repeated tail-copy allocations in match-heavy recursion.
- Validation performed:
  - `go test ./internal/vm ./internal/runtime ./cmd/app -count=1`
  - `go test ./internal/vm -run '^$' -bench 'BenchmarkVMExecuteOnly/(match_nested|list_append_chain|list_prepend_chain|list_concat_chain|list_match_tail)$' -benchmem`

### VM list concat peephole and structural-sharing follow-up

- Added VM compiler peephole in `internal/vm/compiler.go`:
  - Rewrites `list + [x]` to `list :+ x`.
  - Rewrites `[x] + list` to `x +: list`.
  - Skips spread-element list literals to preserve semantics.
- This removes temporary single-element list allocations from common concat patterns and makes concat chains use append/prepend fast paths.
- Combined with prior tail/slice structural sharing in executor, this improves both concat-heavy and match-tail workloads.
- Validation performed:
  - `go test ./internal/vm -count=1`
  - `go test ./internal/runtime ./internal/vm ./cmd/app -count=1`
  - `go test ./internal/vm -run '^$' -bench 'BenchmarkVMExecuteOnly/(match_nested|list_append_chain|list_prepend_chain|list_concat_chain|list_match_tail)$' -benchmem`

### Standard library list transform optimization pass

- Fixed a boundary condition in `lib/slug/list.slug`:
  - `indexOf` now stops at `idx >= len(list)` (previously `idx > len(list)`).

### Foreign list-manipulation optimization pass

- Optimized list-manipulation foreign functions in Go (not `.slug` stdlib code):
  - `internal/foreign/slug_list.go` (`sortWithComparator`): removed per-compare argument-slice allocation by reusing a fixed two-slot argument buffer.
  - `internal/foreign/slug_std.go`:
    - `update`: early-return original list when replacement value is identical to existing element.
    - `swap`: early-return original list when swapping the same index (`i1 == i2`).
- Added targeted foreign benchmarks in `internal/foreign/slug_std_benchmark_test.go`:
  - `BenchmarkStdUpdatePersistentOps`
  - `BenchmarkStdSwapPersistentOps`
  - `BenchmarkListSortWithComparator`
- Validation performed:
  - `go test ./internal/foreign -count=1`
  - `go test ./internal/foreign -run '^$' -bench 'BenchmarkStd(UpdatePersistentOps|SwapPersistentOps)|BenchmarkListSortWithComparator' -benchmem`
  - `go test ./internal/foreign ./internal/runtime ./internal/vm ./cmd/app -count=1`

#### Benchmark snapshot

- `BenchmarkStdUpdatePersistentOps`: `165540 ns/op`, `925766 B/op`, `1100 allocs/op`
- `BenchmarkStdSwapPersistentOps`: `124083 ns/op`, `925763 B/op`, `1100 allocs/op`
- `BenchmarkListSortWithComparator`: `19528 ns/op`, `15624 B/op`, `460 allocs/op`

### Regression fix: list slice fast-path bounds guard

- Fixed a VM regression introduced by list slice structural sharing in `internal/vm/executor.go`.
- Root cause: step-1 fast path sliced `l.Elements[start:end]` without guarding `start >= end`, which can occur after slice index normalization and caused `slice bounds out of range` panics.
- Fix: return an empty list when `start >= end` before taking the shared sub-slice.
- Validation performed:
  - `go test ./internal/vm ./internal/runtime ./cmd/app -count=1`
  - `SLUG_HOME=$(pwd) go run ./cmd/app/main.go -log-level error --root ./tests tests/lists-slicing.slug`
  - `make test`

## 2026-05-09

### ADR-036: unified `copy` semantics for structs and maps

- Extended VM `copy` execution to support both struct and map sources in `internal/vm/executor.go`.
  - Struct behavior remains unchanged: schema-preserving updates with unknown-field rejection and hint validation.
  - Map behavior now performs immutable shallow merge by cloning the source map and applying RHS map entries.
  - Runtime errors now report unsupported sources as `copy expects a struct or map value, got <type>`.
- Added VM unit coverage in `internal/vm/executor_test.go`:
  - `TestExecutorMapCopyShallowMerge`
  - `TestExecutorCopyRejectsUnsupportedSource`
  - `TestExecutorCopyRejectsNonMapPayload`
- Added language-level map-copy assertions in `tests/maps.slug` to verify:
  - copied map gets updated keys
  - original map remains unchanged
- Aligned ADR examples with implemented syntax in `docs/_adr/ADR-036.md`:
  - examples now use infix `value copy { ... }` form.
- Validation performed:
  - `go test ./internal/vm ./internal/runtime ./cmd/app -count=1`
  - `make test`

## 2026-05-09

### ADR-037: channel receive returns payload-or-nil

- Updated VM channel receive/select semantics in `internal/vm/executor.go`:
  - `recv`/`select recv` now return payload values directly.
  - closed-and-drained channel receive now returns `nil`.
  - removed `Full`/`Empty` struct wrapping path from select receive.
- Enforced ADR rule that channel sends must not send `nil`:
  - select send path now returns runtime error `send expects a non-nil payload` for `nil` payloads.
  - applies to both `send` and `trySend` since both lower to select send.
- Removed `slug.channel` runtime prebinding of legacy receive schemas in `internal/runtime/runtime.go`:
  - removed `FullSchema`/`EmptySchema` runtime state.
  - removed module-load injection of `Full` and `Empty` constants.
- Updated channel/receive consumers to new contract:
  - `test-suites/channels.slug` switched `match recv(...)` and `select recv` handlers from `Full/Empty` patterns to value-or-`nil` patterns.
  - `tests/vm-conformance/supported/select-expression.slug` migrated to value-or-`nil` match handling.
- Updated stdlib/docs surface and generated manifest:
  - `lib/slug/io/stdin.slug` now documents and returns `@chan(@str)` and `readLine` now returns direct `recv` result.
  - `docs/_libraries/slug.channel.md` removed `Full`/`Empty` struct sections and documented payload-or-`nil` receives plus nil-send restriction.
  - `docs/_libraries/slug.io.stdin.md` updated `readLines` signature/event semantics to payload-or-`nil` contract.
  - `lib/MANIFEST.ai` removed `slug.channel#Full`/`slug.channel#Empty` struct entries and updated `slug.io.stdin#readLines()` signature.
- Validation performed:
  - `go test ./internal/vm ./internal/runtime ./cmd/app -count=1`
  - `make test`

## 2026-05-09

### ADR-038: structural brace disambiguation for maps and blocks

- Updated parser brace handling in `internal/parser/parser.go` to structurally disambiguate `{...}` in expression positions:
  - `{}` parses as empty map.
  - map-shaped entries (e.g. `{k: v}`, `{"k": v}`, `{:k: v}`, `{[expr]: v}`) parse as map literals.
  - non-map brace bodies parse as block expressions.
- Updated match arm parsing in `parseMatchCase` to allow direct map-return bodies (`=> {k: v}`) while preserving multiline statement block behavior for non-map brace bodies.
- Kept legacy `{{...}}` behavior working (block containing map), while migrating canonical fixture usage to single-brace map return style.
- Added parser coverage in `internal/parser/parser_test.go`:
  - `TestMatchCaseBraceBodyParsesMapLiteral`
  - `TestMatchCaseDoubleBraceStillParses`
  - `TestParsingBlockExpressionLiteral`
- Updated integration fixtures:
  - `test-suites/match.slug` uses direct map-return braces and includes a legacy double-brace compatibility test.
  - `test-suites/parsing.slug` uses direct map-return braces and includes a legacy double-brace compatibility test.
- Validation performed:
  - `go test ./internal/parser ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`
  - `make test`

## 2026-05-09

### Documentation and syntax grammar alignment pass (post ADR-036/037/038)

- Updated editor syntax grammars to align with current language behavior:
  - `extras/Slug.sublime-package/Slug.sublime-syntax`
    - `onsuccess`/`onerror` are now highlighted contextually after `defer` instead of as global keywords.
  - `extras/Slug.tmbundle/Syntaxes/Slug.tmLanguage.json`
    - `onsuccess`/`onerror` are now contextual defer-mode keywords via defer lookbehind.
    - annotation token set expanded to include current `@chan` and `@sym` forms.
- Updated Sublime syntax assertions in `extras/Slug.sublime-package/syntax_test_slug.slug`:
  - added checks for `@sym` and `@chan` annotations.
  - added contextual keyword test for `defer onsuccess(...)`.
  - added out-of-context identifier test for `onsuccess`.
- Updated docs site lexer parity in `docs/_plugins/slug_lexer.rb`:
  - removed global keyword classification for `onsuccess`/`onerror`.
  - added contextual defer lookbehind highlighting for both.
- Added historical supersession notes to older accepted ADRs whose examples predate current channel semantics:
  - `docs/_adr/ADR-025.md` now notes supersession by ADR-037 for receive shape.
  - `docs/_adr/ADR-031.md` now notes supersession by ADR-037 for stdin stream value shape.
- Validation performed:
  - `go test ./internal/parser ./internal/runtime ./internal/vm ./cmd/app -count=1`
  - `make test`


### `slug.ebnf` alignment with parser behavior

- Updated `extras/slug.ebnf` to match current parser semantics and lexical behavior:
    - Added `doc_comment_stmt = DOC_COMMENT` to reflect parser support for top-level doc comment tokens.
    - Clarified declaration pattern acceptance with shared behavior:
        - `var` and `val` both use full `match_pattern`, including exact-map `{| ... |}` patterns.
    - Updated lexical identifier production to Unicode-aware form (`UNICODE_LETTER`/`UNICODE_MARK` and `UNICODE_DIGIT`) to match lexer behavior.
    - Expanded lexical comment note to include block comments (`/* ... */`) and doc comments (`/** ... */`) in addition to line comments.

### `val`/`var` destructuring entry parity

- Updated parser `val` entry-token validation to match `var` by allowing exact-map pattern starts (`{| ... |}`).
- Added parser test coverage:
    - `TestValAcceptsExactMapPattern` in `internal/parser/parser_test.go`.

### TextMate/Sublime syntax parity and lexer-alignment pass

- Updated `extras/Slug.tmbundle/Syntaxes/Slug.tmLanguage.json`:
    - Added builtin function highlighting parity (`len`, `import`, `print`, `println`, `stacktrace`, `argv`, `argm`, `cfg`).
    - Added explicit `{|` and `|}` exact-map delimiter punctuation scopes.
    - Tightened bytes-literal regex to disallow internal whitespace in `0x"..."`.
    - Upgraded tag/symbol/function-call/identifier patterns to Unicode-aware identifier classes.
- Updated `extras/Slug.sublime-package/Slug.sublime-syntax` with parity changes:
    - Unicode-aware tag/symbol/function-call/identifier regexes.
    - Tightened bytes-literal regex to disallow internal whitespace.

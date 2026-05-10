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

### Developers guide full refresh (tutorial style + language alignment)

- Rewrote all modules in `docs/_developers-guide` in a friendlier, concise tutorial voice.
- Updated examples and explanations to align with current language behavior and recent ADR outcomes, including:
  - structural brace disambiguation for map literals vs block expressions,
  - contextual defer modes (`defer onsuccess`, `defer onerror(err)`),
  - exact-map pattern usage (`{| ... |}`),
  - current pipeline and pattern-matching idioms,
  - canonical `await(handle)` and `await(handle, timeoutMs)` concurrency forms.
- Replaced outdated reference content:
  - removed stale conditional `?:` precedence entry,
  - added current operator precedence ordering including `/>`, `:+`, and `+:`.
- Normalized module pedagogy across the guide with runnable examples and `Try it` prompts.

### Developers guide beginner-depth expansion pass

- Expanded refreshed `docs/_developers-guide` modules for new learners with step-by-step pedagogy.
- Added explicit `Mental model`, `Expected output`, and `Common mistake` sections across core lessons.
- Reworked mini project/testing/concurrency modules into guided walkthroughs rather than compact snippets.
- Kept language examples aligned with current parser/runtime behavior (including contextual `defer` modes and canonical `await(handle[, timeoutMs])` forms).

### Developers guide voice-consistency polish pass

- Normalized tutorial section labeling across `docs/_developers-guide` for smoother scanning:
  - standardized `Common mistakes:` callouts,
  - standardized `### Mental model` headings where previously inline.
- Kept all technical content and syntax semantics unchanged; this pass was editorial consistency only.

### Semantic inferred type checking (constraint-first, initial warn-gated rollout)

- Added inferred type checking to semantic analysis with a constraint-based engine in `internal/semantic/typecheck.go`.
- Added semantic API options:
  - `semantic.AnalyzeWithOptions(path, src, program, AnalyzeOptions)` returning `(errors, warnings)`.
  - Historical rollout shape at this stage: `AnalyzeOptions{EnableTypeCheck, StrictTypeCheck}`.
- Kept `semantic.Analyze(...)` compatibility by returning errors only (no behavior change for existing callers).
- Type-checking v1 coverage includes:
  - declaration inference (`val`/`var`),
  - call arity and argument compatibility,
  - operator compatibility constraints,
  - hard tag constraints for known type tags (e.g., `@num`, `@str`, `@list`, `@map`, `@task`, `@chan`, `@struct`).
- Added warn-gated rollout plumbing through runtime/CLI config (historical at this stage):
  - `Configuration.EnableTypeCheck` and `Configuration.StrictTypeCheck`.
  - CLI flags `-type-check` (default `true`) and `-type-check-strict` (default `false`).
  - Runtime execution/module loading now consumes semantic warnings separately from errors.
- Added semantic tests for warn vs strict behavior:
  - `TestSemanticTypeCheckWarnsByDefaultForTypeMismatch`
  - `TestSemanticTypeCheckStrictPromotesMismatchToError`
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic inferred type checking expansion: match narrowing + struct-aware checks

- Extended semantic type inference with case-local match narrowing:
  - match scrutinee constraints are now applied to each case pattern in a scoped context,
  - pattern-bound names are inferred and available in case body type checks.
- Added struct-schema-aware field checking for struct initialization:
  - schema declarations captured from `val/var Name = struct { ... }`,
  - struct init expressions `Name { field: value }` now enforce field tag/default-derived type constraints.
- Added semantic strict-mode tests:
  - `TestSemanticTypeCheckStrictChecksStructFieldTagsInInit`
  - `TestSemanticTypeCheckStrictChecksMatchPatternNarrowing`
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic inferred type diagnostics quality pass

- Improved inferred type diagnostic readability by rendering context-specific mismatch messages in semantic solver output.
- Added targeted message forms for common failure contexts:
  - assignment mismatch,
  - call argument mismatch,
  - if/match-guard boolean mismatch,
  - numeric/logical/prefix operator mismatches,
  - comparison and select-await mismatches.
- Kept inference behavior unchanged; this pass improves message quality only.
- Added semantic test coverage for call argument mismatch messaging.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing nil-compatibility alignment

- Updated semantic type unification to allow `nil` compatibility with concrete types, aligning inferred checks with runtime/tag dispatch behavior.
- Fixes false-positive diagnostics for typed defaults like `@str msg = nil`.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsTaggedDefaultNil` in `internal/semantic/analyzer_test.go`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing branch-join fix for throw paths

- Fixed false-positive `if` branch type mismatch diagnostics when one branch throws.
- Semantic inference now models `throw` statements as non-returning/any-compatible for branch result unification.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsIfBranchWithThrowElse`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing arity fix for default parameters

- Updated inferred call arity checking to respect function signature min/max arity (default parameters) instead of raw parameter count.
- Function type nodes now retain inferred arity bounds derived from parser signature metadata.
- Fixes false-positive arity diagnostics for calls that omit defaulted trailing arguments.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsCallsUsingDefaultArity`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for heterogeneous list literals in match-body functions

- Fixed false-positive list element unification errors in inferred type checking.
- Root cause: list literals were inferred as homogeneous; this conflicted with parser-injected match scrutinee lists for `fn(...) match { ... }` (which can contain heterogeneous parameter types).
- Updated list literal inference to treat list element type as `any`-compatible while still inferring each element expression.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsMatchBodyFunctionWithHeterogeneousParams`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for generic @fn constraint unification

- Fixed false-positive function argument mismatch diagnostics where both sides were function-typed (`expected fn, got fn`).
- Root cause: unifier required concrete function shape/arity equivalence even for generic `@fn` tag constraints.
- Updated function unification to treat generic `@fn` constraints as function-kind compatibility (not exact arity shape).
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsGenericFnTagAcrossArities`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for multi-pattern alternative narrowing

- Fixed false-positive type diagnostics for comma-separated match pattern alternatives (OR semantics).
- Root cause: multi-pattern narrowing applied constraints from all alternatives (intersection), over-constraining scrutinee types.
- Updated narrowing to treat alternatives safely for v1 by narrowing from the first alternative only.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsMultiPatternLiteralAlternatives`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for match-case branch type leakage

- Fixed false-positive diagnostics caused by branch inference leaking across `match` cases.
- Introduced isolated case scopes seeded from visible bindings so per-case constraints do not over-constrain sibling cases.
- This resolves nested type-dispatch + inner-literal-match scenarios (e.g., `match v /> type()` with string-literal alternatives in a `^STRING_TYPE` branch).
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsNestedMatchAfterTypeDispatch`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing alignment for bytes as sequence-like in list patterns

- Updated pattern typing so list patterns accept bytes as sequence-like values, matching runtime behavior.
- Added sequence-aware pattern helpers in semantic typing:
  - list-pattern narrowing/binding now supports `list` and `bytes`,
  - bytes element pattern bindings infer numeric element type.
- Added clearer diagnostic for obvious non-sequence list-pattern subjects.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsListPatternMatchOnBytes`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing alignment for bytes append/prepend operators

- Fixed false-positive type diagnostics for bytes sequence update operators:
  - `bytes :+ num` (append byte)
  - `num +: bytes` (prepend byte)
- Updated inferred operator rules for `:+` / `+:` to align with runtime behavior:
  - list append/prepend remains supported,
  - bytes append/prepend now explicitly modeled with numeric byte elements.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsBytesAppendAndPrependOperators`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing alignment for bytes bitwise operators

- Fixed false-positive numeric-only diagnostics for bytes bitwise expressions.
- Updated inferred operator rules for `&`, `|`, `^` to support runtime-compatible modes:
  - `num <op> num` -> `num`
  - `bytes <op> bytes` -> `bytes`
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsBytesBitwiseOperators`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing alignment for bytes bitwise-not prefix (`~`)

- Fixed false-positive prefix numeric diagnostics for bytes bitwise complement expressions.
- Updated inferred prefix operator rules so `~` supports runtime-compatible modes:
  - `~num` -> `num`
  - `~bytes` -> `bytes`
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsBytesBitNotPrefixOperator`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing alignment for mixed bytes/number bitwise operators

- Fixed false-positive diagnostics for mixed notation bitwise expressions involving bytes and numbers.
- Updated inferred rules for `&`, `|`, `^` to match runtime mixed-mode behavior:
  - `bytes <op> num` -> `bytes`
  - `num <op> bytes` -> `bytes`
  - `num` side constrained as numeric (runtime handles byte-range conversion).
- Expanded regression coverage in bytes bitwise test to include mixed forms.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for heterogeneous map-pattern arms and case scrutinee isolation

- Fixed over-constrained map pattern inference and cross-case scrutinee leakage in match analysis.
- Updated match case typing to clone scrutinee type per case, preventing sibling case constraints from intersecting.
- Updated map-pattern narrowing/binding to avoid forcing homogeneous map value types per arm.
- Added regression test based on heterogeneous map arm sample:
  - `TestSemanticTypeCheckStrictAllowsHeterogeneousMapPatternArms`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for call-site monomorphization across heterogeneous map shapes

- Fixed false-positive call argument mismatches caused by early call-site monomorphization of function parameter types.
- Updated call argument checking to use deferred compatibility validation after constraint solving, instead of eagerly constraining parameter nodes from each call site.
- Preserves inferred function requirements while allowing valid heterogeneous argument shapes across different calls when function logic supports them.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsMultipleMapValueShapesAcrossCalls`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for repeated function binding overload sets

- Fixed false-positive call mismatch diagnostics for repeated `var/val name = fn(...)` overload patterns.
- Root cause: semantic scope binding replaced earlier function types with the latest one, unlike runtime function-group merge behavior.
- Updated semantic binding behavior to merge repeated function bindings into a generic function-group-like function type for call-site checking.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsFunctionTagOverloadSetCalls`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for heterogeneous map literals

- Fixed false-positive map value/key homogeneity diagnostics in map literals.
- Root cause: map literals were inferred as single key type + single value type, conflicting with runtime-allowed heterogeneous maps.
- Updated map literal inference to retain heterogeneous key/value compatibility while still inferring nested expression constraints.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsHeterogeneousMapLiteralValues`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for string concatenation/interpolation with non-string values

- Fixed false-positive `+` operator diagnostics in string interpolation/concatenation flows involving numbers.
- Updated `+` typing to use deferred runtime-aligned compatibility checks:
  - `string + any` and `any + string` are accepted,
  - non-string `+` remains constrained to valid homogeneous modes (`num`, `list`, `bytes`).
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsStringInterpolationWithNumbers`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for chained string repetition (`*`) inference

- Fixed false-positive numeric-operator diagnostics for chained string repetition expressions like `" " * spaces * depth`.
- Updated `*` typing to deferred runtime-aligned compatibility checks (similar to `+`), preventing eager intermediate-type misclassification.
- Runtime-compatible `*` modes enforced by deferred checks:
  - `num * num`
  - `str * num`
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsChainedStringTimesNumberTimesNumber`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for dynamic untagged struct fields

- Fixed false-positive struct field mismatch diagnostics for intentionally dynamic untagged struct fields (e.g., `ParseResult.value` in JSON parser).
- Updated struct schema registration behavior:
  - untagged fields with no informative default (or `nil` default) are inferred as `any`,
  - tagged fields and non-`nil` defaults continue to constrain field type.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsDynamicUntaggedStructField`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for mixed-value if branches used for control flow

- Fixed false-positive branch mismatch diagnostics in `if` expressions that are used for side effects/control flow, where branch bodies may produce unrelated value types.
- Updated `if` expression inference to:
  - keep strict boolean checking for the condition,
  - infer both branches for internal constraints,
  - avoid forcing branch result type unification at expression level (result is treated as dynamic).
- This aligns semantic checks with existing language/runtime behavior and prevents errors like `num vs bool` from branch-local assignments.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsIfUsedForSideEffectsWithMixedBranchValues`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for heterogeneous `match` case result values

- Fixed false-positive `match case result` type mismatches when different `match` arms intentionally return different runtime shapes (for example, different struct schemas such as `SectionNode` vs `ParentNode`).
- Updated `match` inference to:
  - preserve case-local checks (pattern narrowing, guard boolean validation, inner expression constraints),
  - avoid forcing all case bodies to unify to one static result type,
  - treat the overall `match` expression result as dynamic.
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsMatchWithHeterogeneousStructCaseResults`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing fix for spread arguments and fixed-arity calls

- Fixed false-positive call arity mismatch diagnostics for spread-based calls such as `rgbStyle(...v)` where `v` is list-like and the target function has fixed arity.
- Updated call inference to be spread-aware:
  - parse call arguments as spread vs non-spread,
  - avoid strict exact-arity checks when spread is present (only fail when non-spread arguments already exceed max arity),
  - propagate spread element typing into remaining positional parameter checks (`list` -> element type, `bytes` -> `num`, `str` -> `str`).
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsSpreadToSatisfyFixedArity`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`
  - `SLUG_HOME=$(pwd) go run ./cmd/app/main.go -log-level error --root ./lib lib/slug/term/colour.slug` (passes)

### Semantic typing fix for deferred bitwise-mode inference

- Fixed false-positive `+` mismatch cascades caused by eager numeric pinning of bitwise expressions (`&`, `|`, `^`) before operand kinds were fully inferred.
- Updated bitwise inference to defer mode validation until post-unification, aligned with runtime-supported bitwise modes:
  - `num <op> num`
  - `bytes <op> bytes`
  - `bytes <op> num` and `num <op> bytes`
- Added regression test:
  - `TestSemanticTypeCheckStrictAllowsDeferredBitwiseBytesModeInConcatFlow`.
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing enhancement: union-based branch joins for `if`/`match`

- Implemented lightweight union typing in inferred semantic checks to preserve branch result alternatives instead of collapsing to `any`.
- Added new inferred type kind: `union<...>` and union flatten/dedup behavior.
- Updated `if` expression inference to return a union of `then`/`else` result types.
- Updated `match` expression inference to return a union across case body result types.
- Made compatibility logic union-aware for:
  - general call/type compatibility checks,
  - deferred operator checks (`+`, `*`, bitwise).
- Result: flow-sensitive branch typing is retained for downstream checks (for example, passing `if`/`match` result unions into typed function parameters now reports precise mismatches when not all alternatives are valid).
- Added regression tests:
  - `TestSemanticTypeCheckStrictTracksIfBranchUnionForCalls`
  - `TestSemanticTypeCheckStrictTracksMatchCaseUnionForCalls`
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing enhancement: guard-based flow narrowing in `if` and `match`

- Added branch-local flow refinements driven by boolean guards, applied during semantic inference.
- New guard narrowing support:
  - `x == nil` / `x != nil` (including negated/else polarity handling),
  - `type(x) == <TYPE_CONST>` / `!=` for common type constants (`NIL_TYPE`, `BOOLEAN_TYPE`, `NUMBER_TYPE`, `STRING_TYPE`, etc.),
  - conservative composition over `&&`, `||`, and `!`.
- Refinements are applied only within branch/case scopes:
  - `if` true/false branches receive separate narrowed bindings,
  - `match` guard-true path narrows case-local bindings before body inference.
- Added strict concrete-type include/exclude helpers for refinement set operations, avoiding nil-compatibility overreach in union subtraction.
- Added regression tests:
  - `TestSemanticTypeCheckStrictNarrowsIfTypeGuardTrueBranch`
  - `TestSemanticTypeCheckStrictNarrowsIfNilGuardElseBranch`
  - `TestSemanticTypeCheckStrictNarrowsMatchGuardTypePredicate`
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing enhancement: predicate-based guard narrowing

- Extended flow-sensitive guard refinement to support predicate functions in `if` and `match` guards.
- Added predicate narrowing support for common runtime shape/type predicates:
  - `isList(x)`, `isMap(x)`, `isStruct(x)`, `isFn(x)`, `isBytes(x)`
  - plus scalar aliases: `isStr/isString`, `isNum/isNumber`, `isBool/isBoolean`
- Added conservative `len(...)` comparison shape narrowing for guard analysis:
  - recognizes forms like `len(x) > 0`, `len(x) == 0`, and reversed operand variants,
  - narrows guarded variable to sequence/map-like union (`list | map | bytes | str`) where condition implies a valid length-bearing shape.
- Guard refinements remain branch-local and integrate with existing union-based branch typing.
- Added regression tests:
  - `TestSemanticTypeCheckStrictNarrowsIfPredicateTrueBranch`
  - `TestSemanticTypeCheckStrictNarrowsIfPredicateElseBranch`
  - `TestSemanticTypeCheckStrictNarrowsMatchGuardPredicate`
  - `TestSemanticTypeCheckStrictNarrowsLenGuardShape`
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing enhancement: contradiction-aware guard narrowing and unreachable diagnostics

- Added contradiction detection to flow-sensitive guard refinement.
- Introduced internal `never` inferred type to represent impossible narrowed states.
- Refinement engine now composes cumulatively within a single guard (`&&`/`||` paths) using current in-guard narrowed bindings.
- When branch/case refinements become contradictory, semantic checks now emit focused diagnostics:
  - `unreachable if-branch: guard refinements are contradictory`
  - `unreachable else-branch: guard refinements are contradictory`
  - `unreachable match case: guard refinements are contradictory`
- Unreachable refined branches short-circuit local inference to avoid noisy follow-on mismatch errors.
- Added refinement-shape matching for generic predicate targets (e.g. `list<any>`, `map<any,any>`) so shape narrowing remains practical with concrete inferred element/key/value types.
- Added regression tests:
  - `TestSemanticTypeCheckStrictReportsUnreachableIfBranchOnContradictoryGuard`
  - `TestSemanticTypeCheckStrictReportsUnreachableMatchGuardOnContradiction`
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing enhancement: staged mutable reassignment tracking (`var`)

- Added lightweight reassignment-aware typing for identifier assignments (`=`) to better model mutable `var` evolution.
- On assignment to an existing binding, semantic inference now widens the binding type to include assigned value shapes (`old ∪ new`) instead of freezing to initializer-only type.
- This improves handling of mutable values across iterative/control-flow-heavy code without requiring a full SSA/control-flow graph.
- Scope of this stage:
  - supports identifier reassignments,
  - intentionally conservative/path-insensitive widening at assignment sites,
  - does not yet model precise execution-order/path dominance across branches.
- Added regression tests:
  - `TestSemanticTypeCheckStrictTracksVarReassignmentWidening`
  - `TestSemanticTypeCheckStrictTracksVarReassignmentAcrossIfBranches`
- Validation performed:
  - `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing enhancement: staged block-local path sensitivity for mutable vars

Implemented staged precision improvements for reassigned `var` handling without full SSA/CFG.

- Added per-block override typing environment:
  - new block-local override stack tracked alongside scope stack.
- Lookup precedence updated:
  - identifier lookup now prefers current block override, then lexical scope bindings.
- Assignment behavior updated (`=` on identifiers):
  - keeps existing conservative scope-level widening fallback (`old ∪ new`),
  - also writes a dominating block-local override to assigned RHS type for path-sensitive same-block use.
- `if` merge behavior added:
  - branch blocks infer with isolated overrides,
  - post-`if` overrides are merged with union per outer-visible binding,
  - if no `else`, merge includes unchanged outer path.
- Existing widening fallback remains in place as safety net for paths not yet modeled precisely.

Regression coverage added/updated:
- `TestSemanticTypeCheckStrictTracksVarReassignmentWidening` (now includes `acc = acc + 2`)
- `TestSemanticTypeCheckStrictTracksVarReassignmentAcrossIfBranches` (numeric reassignment in both branches + numeric use after join)

Validation performed:
- `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`
- `go run ./cmd/app/main.go -log-level error --root ./lib lib/slug/web/response.slug`

### Semantic typing enhancement: bounded recur fixed-point stabilization

- Added bounded fixed-point inference for functions that contain `recur`.
- Function-body inference now runs iterative stabilization passes (bounded) for recursive flows:
  - each pass infers body with current parameter assumptions,
  - pass results feed back into parameter types via union widening,
  - inference stops when parameter type signatures stabilize or the iteration cap is reached.
- Added diagnostic deduplication in semantic checker to prevent repeated messages across iterative passes.
- This improves type precision and reduces false positives in loop-like recursive functions without requiring full CFG/SSA.

Regression coverage added:
- `TestSemanticTypeCheckStrictStabilizesRecurParameterTypes`

Validation performed:
- `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing enhancement: union simplification and dominance pruning

- Added union normalization pipeline to reduce union explosion and improve diagnostics.
- Union construction now includes:
  - dominance pruning (`any` dominates all members, `never` removed),
  - structural deduplication by normalized type signature,
  - safe family merging:
    - `list<t1> | list<t2> -> list<t1|t2>`
    - `map<k1,v1> | map<k2,v2> -> map<k1|k2, v1|v2>`
- Added union member cap (`maxUnionOptions`) with widening fallback to `any` when unions grow beyond cap.
- Improved union diagnostic readability:
  - stable sorted union member rendering in `describe(...)`.

Regression coverage added:
- `TestSemanticTypeCheckStrictSimplifiesUnionListFamilies`

Validation performed:
- `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic typing enhancement: debug trace mode (`-type-check-trace`)

- Added non-invasive semantic type-check trace mode to explain inference decisions without changing normal diagnostics behavior.
- New CLI flag:
  - `-type-check-trace` (disabled by default)
- Config/runtime plumbing:
  - added `TypeCheckTrace` to runtime configuration,
  - semantic analysis options now carry trace enablement + writer.
- Trace emission format:
  - `TypeTrace: <event> @ <path>:<line>:<col> | <details>`
  - location omitted when event is global (non-positioned).
- Implemented trace event categories (initial coverage):
  - `refine` (guard narrowing include/exclude)
  - `merge` (branch merge unioning)
  - `widen` (assignment/accumulator widening)
  - `contradiction` (refinement collapses to `never`)
  - `union-normalize` (normalization and cap widening decisions)
- Added regression coverage:
  - `TestSemanticTypeCheckTraceEmitsEvents`.

Validation performed:
- `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Semantic conformance snapshot suite for stdlib modules

- Added semantic conformance golden test harness to guard behavior-level regressions across key stdlib modules.
- New test:
  - `internal/semantic/semantic_conformance_test.go`
- Snapshot coverage includes:
  - `lib/slug/std.slug`
  - `lib/slug/mustache.slug`
  - `lib/slug/json.slug`
  - `lib/slug/web/response.slug`
  - `lib/slug/crypto.slug`
- For each module, snapshot captures:
  - strict semantic error count,
  - strict semantic warning count,
  - type-check trace event counts by event name.
- Golden file:
  - `internal/semantic/testdata/semantic_conformance_golden.json`
- Update flow:
  - run with `UPDATE_SEMANTIC_GOLDEN=1` to refresh snapshots after intentional behavior changes.

Validation performed:
- `UPDATE_SEMANTIC_GOLDEN=1 go test ./internal/semantic -run TestSemanticConformanceSnapshots -count=1`
- `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

### Type-check flag simplification: single strict `-type-check`

- Simplified semantic type-check control to a single flag:
  - `-type-check` now enables strict semantic inferred type checking (errors) by default.
  - `-type-check=false` disables semantic inferred type checking.
- Removed legacy non-strict/warn-mode path and removed `-type-check-strict` flag.
- API/config cleanup:
  - removed `StrictTypeCheck` from semantic `AnalyzeOptions`, runtime config, and runtime/analyzer call sites.
- Updated tests to reflect new behavior:
  - added/updated coverage for disabled mode (`EnableTypeCheck: false`) and default error behavior when enabled.

Validation performed:
- `go test ./internal/semantic ./internal/runtime ./internal/vm ./cmd/app -count=1`

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

### LSP mode Phase 1: diagnostics-only stdio language server

- Added CLI language server mode to `slug`:
  - `-lsp`
  - `--language-server` (alias)
- LSP mode behavior:
  - does not require positional script argument,
  - runs over stdio using standard LSP framing (`Content-Length`),
  - bypasses normal script execution path.
- Implemented Go-native LSP server core in `internal/lsp` with lifecycle + diagnostics scope:
  - requests/notifications: `initialize`, `initialized`, `shutdown`, `exit`,
  - text sync (full): `textDocument/didOpen`, `textDocument/didChange`, `textDocument/didClose`,
  - publishes `textDocument/publishDiagnostics` notifications.
- Diagnostics pipeline reuses existing toolchain:
  - parser diagnostics from `lexer/parser`,
  - semantic/type-check diagnostics via `semantic.AnalyzeWithOptions` (honors `-type-check`, `-type-check-trace`).
- Added protocol handling behavior:
  - unknown request methods return JSON-RPC method-not-found,
  - unknown notifications are ignored.

Tests added:
- `internal/lsp/server_test.go`
  - initialize/shutdown/exit flow,
  - diagnostics publishing on didOpen/didChange,
  - diagnostic location parsing.

Validation performed:
- `go test ./internal/lsp ./cmd/app ./internal/semantic ./internal/runtime ./internal/vm -count=1`
- `go test ./... -count=1`

### LSP mode Phase 2: diagnostics hardening and protocol-state correctness

- Hardened LSP server behavior in `internal/lsp/server.go` with protocol and diagnostics quality improvements.

Protocol/state handling improvements:
- Added explicit pre-initialize request handling:
  - requests before `initialize` now return JSON-RPC error `-32002` (server not initialized),
  - pre-initialize notifications are ignored safely.
- Added post-shutdown request guard:
  - requests after `shutdown` (except `exit`) now return JSON-RPC `-32600` invalid request.
- Added shutdown/exit sequencing guard:
  - `exit` before `shutdown` is treated as invalid lifecycle and returns runtime error from server loop.
- Kept unknown notification behavior safe (ignored), unknown request methods return `-32601` method not found.

Document/URI handling improvements:
- Added URI normalization and local path derivation for `file://` documents.
- Normalized URI is used for internal document store keys and diagnostic publication URI.
- Local normalized path is provided to parser/semantic analyzer where available.
- Added explicit invalid transition handling for `didChange` before `didOpen` (ignored safely).

Diagnostics quality improvements:
- Improved diagnostic position parsing with robust location regex extraction.
- Improved diagnostic message normalization (ParseError/TypeWarning prefix cleanup).
- Added diagnostic deduplication to prevent repeated identical diagnostics in publish payloads.
- Maintained close behavior to clear diagnostics (`didClose` publishes empty diagnostics).

Tests expanded in `internal/lsp/server_test.go`:
- initialize/shutdown/exit success path,
- exit-before-shutdown failure,
- unknown method request returns `-32601`,
- requests rejected after shutdown (`-32600`),
- didChange-before-open ignored,
- diagnostics publish on open/change,
- duplicate diagnostics deduped,
- URI normalization coverage,
- diagnostic range parsing coverage.

Validation performed:
- `go test ./internal/lsp ./cmd/app ./internal/semantic ./internal/runtime ./internal/vm -count=1`
- `go test ./... -count=1`

### LSP mode Phase 3: performance guardrails and extensibility scaffolding

- Completed Phase 3 for LSP server with debounce/coalescing, dispatch refactor, and cancellation scaffold.

Dispatch and extensibility:
- Refactored request handling to method-dispatch map (`handlers`) with dedicated handler methods.
- This enables straightforward addition of future methods (`hover`, `completion`, etc.) without growing a single switch block.

Performance and coalescing:
- Added diagnostics debounce window (`diagnosticsDebounceWindow`).
- `didChange` now coalesces rapid bursts per document:
  - immediate publish if outside debounce window,
  - otherwise mark document dirty and defer diagnostics publish.
- Added deferred dirty-doc flush on:
  - non-change message boundaries,
  - shutdown,
  - loop exit/EOF.

Cancellation scaffold:
- Added `$/cancelRequest` notification handler.
- Cancelled request IDs are recorded internally (`canceledReqs`) for future long-running request methods.

Phase 2 behavior retained:
- protocol state guards (initialize/shutdown/exit ordering),
- URI normalization + local path derivation,
- diagnostic deduplication and message/range normalization.

Tests expanded in `internal/lsp/server_test.go`:
- cancel-request acceptance,
- deferred diagnostics flush behavior,
- debounce window sanity,
- existing lifecycle/protocol/diagnostics hardening tests remain green.

Validation performed:
- `go test ./internal/lsp ./cmd/app ./internal/semantic ./internal/runtime ./internal/vm -count=1`
- `go test ./... -count=1`

### LSP mode Phase 4: read-only navigation APIs (hover, symbols, definition)

- Added initial read-only language intelligence endpoints in `internal/lsp/server.go`:
  - `textDocument/hover`
  - `textDocument/documentSymbol`
  - `textDocument/definition`

Implementation details:
- Extended handler dispatch map with the three new methods.
- Added lightweight source indexing over open document text:
  - top-level symbol collection (`val` and `fn`) with source ranges,
  - local declaration lookup for definition resolution,
  - identifier extraction at cursor offsets with UTF-8 safe position/offset conversion.
- Hover now returns concise symbol details (kind/name/signature-style summary) when cursor is on a known symbol.
- Document symbols now return top-level symbols with LSP symbol kinds and selection ranges.
- Definition now resolves to the nearest in-file declaration location for the symbol under cursor.

Tests added in `internal/lsp/server_test.go`:
- `TestServerHoverReturnsSymbolInfo`
- `TestServerDocumentSymbolReturnsTopLevel`
- `TestServerDefinitionReturnsLocalDeclaration`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP fix: prevent over-broad Cmd/Ctrl-click link behavior in IntelliJ

- Tightened identifier resolution for navigation requests (`hover` / `definition`) in `internal/lsp/server.go`:
  - only treat a position as symbol-bearing when the cursor is directly on an identifier rune,
  - avoid resolving symbols from adjacent whitespace/punctuation positions.
- Expanded initialize capabilities to explicitly advertise supported read-only features:
  - `hoverProvider: true`
  - `definitionProvider: true`
  - `documentSymbolProvider: true`
- Added regression test in `internal/lsp/server_test.go`:
  - `TestServerDefinitionReturnsNilOutsideIdentifier`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP fix: avoid false definition links in non-identifier text

- Refined symbol-at-cursor resolution in `internal/lsp/server.go` to use lexer token spans:
  - `textDocument/hover` and `textDocument/definition` now resolve only when cursor is on an `IDENT` token.
  - Added end-of-word fallback (`offset-1`) to support clients that send boundary positions.
- This prevents definition links from activating inside string literals/comments/other non-identifier text while preserving expected navigation on real symbols.

Tests added in `internal/lsp/server_test.go`:
- `TestServerDefinitionReturnsNilInsideStringLiteral`
- Existing `TestServerDefinitionReturnsNilOutsideIdentifier` remains as guard.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP diagnostics aid: debug tracing for hover/definition requests

- Added debug-level structured logs in `internal/lsp/server.go` for:
  - `textDocument/hover` request outcomes: no symbol, unresolved symbol, resolved symbol.
  - `textDocument/definition` request outcomes: no symbol, unresolved symbol, resolved location.
- Logged fields include URI, request line/character/offset, symbol name/kind (when available), resolved definition range (when available), and request id.
- Logging uses existing `slog` pipeline and is enabled via CLI log level (e.g. `-log-level debug`).

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP logging update: protocol-level inbound/outbound tracing only

- Replaced method-specific hover/definition debug logs with protocol-level JSON-RPC tracing in `internal/lsp/server.go`.
- Added debug logs for every framed message:
  - inbound: `lsp.rpc.inbound` with raw JSON payload
  - outbound: `lsp.rpc.outbound` with raw JSON payload
- Removed granular symbol-resolution debug logging for `hover`/`definition` to reduce noise and keep transport-level visibility focused.
- Removed now-unused identifier helper (`isIdentRune`) after logging refactor.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 5: textDocument/completion (keywords + in-scope symbols)

- Added `textDocument/completion` support in `internal/lsp/server.go`.
- Advertised completion capability during initialize:
  - `completionProvider` with `resolveProvider: false`.
- Implemented completion candidate set:
  - Slug language keywords.
  - In-scope symbols discovered from current document symbol collection.
- Implemented prefix filtering based on cursor position:
  - completion suggestions are filtered to labels matching current identifier prefix.
  - supports end-of-token cursor positions through existing identifier offset logic.
- Added completion item kind mapping for common symbol categories.

Tests added/updated in `internal/lsp/server_test.go`:
- `TestServerInitializeShutdownExit` now asserts `completionProvider` capability is present.
- `TestServerCompletionReturnsKeywordsAndSymbols` validates:
  - symbol completion (e.g. `answer`) is returned for matching prefix,
  - keyword completion (e.g. `val`, `var`) is returned for matching prefix,
  - non-matching symbol is excluded by prefix filter.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 6: completionItem/resolve (lazy completion enrichment)

- Added `completionItem/resolve` handler in `internal/lsp/server.go`.
- Updated initialize capabilities to advertise resolve support:
  - `completionProvider.resolveProvider: true`
- Extended completion items to include:
  - `data` payload (`uri`, `label`, `kind`) for round-tripping into resolve requests,
  - optional `documentation` field for lazy docs.
- Resolve behavior:
  - looks up current open document by item `data.uri`,
  - resolves symbol metadata by `data.label`,
  - enriches item `kind`, `detail`, and markdown `documentation`.
  - falls back gracefully if symbol/document is unavailable.

Tests added/updated in `internal/lsp/server_test.go`:
- `TestServerInitializeShutdownExit` now asserts `completionProvider.resolveProvider == true`.
- `TestServerCompletionResolveEnrichesItem` validates that resolve populates detail + documentation for a symbol completion item.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 7: textDocument/documentHighlight

- Added `textDocument/documentHighlight` support in `internal/lsp/server.go`.
- Advertised `documentHighlightProvider: true` in initialize capabilities.
- Implemented document highlights using lexer-backed identifier matching:
  - resolve symbol at cursor via existing identifier resolution,
  - return all `IDENT` token occurrences with matching label in the current document,
  - return empty result when cursor is not on an identifier.
- Highlight result shape uses `DocumentHighlight` ranges with `kind: Text` (`1`).

Tests added/updated in `internal/lsp/server_test.go`:
- `TestServerInitializeShutdownExit` now asserts `documentHighlightProvider == true`.
- `TestServerDocumentHighlightReturnsAllIdentifierMatches` validates expected highlight count for repeated symbol occurrences.
- `TestServerDocumentHighlightReturnsEmptyOutsideIdentifier` validates empty results when cursor is not on a symbol.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 8: textDocument/references (in-document, scope-aware)

- Added `textDocument/references` handler in `internal/lsp/server.go`.
- Advertised `referencesProvider: true` in initialize capabilities.
- Implemented in-document references with scope-aware filtering:
  - resolve the symbol under cursor using existing symbol resolution,
  - scan identifier tokens with matching label,
  - keep only occurrences that resolve to the same declaration symbol,
  - respect `context.includeDeclaration` to include/exclude declaration location.
- References are returned as LSP `Location[]` with normalized document URI and precise ranges.

Tests added/updated in `internal/lsp/server_test.go`:
- `TestServerInitializeShutdownExit` now asserts `referencesProvider == true`.
- `TestServerReferencesIncludeDeclaration` validates declaration-inclusive results.
- `TestServerReferencesExcludeDeclaration` validates declaration-excluded results.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 9: prepareRename + textDocument/rename (scope-aware)

- Added rename capabilities in `internal/lsp/server.go`:
  - `renameProvider` with `prepareProvider: true`.
  - request handlers:
    - `textDocument/prepareRename`
    - `textDocument/rename`
- Implemented `prepareRename` behavior:
  - validates cursor is on a renameable identifier symbol,
  - returns precise target range and placeholder symbol name.
- Implemented scope-aware rename behavior:
  - validates `newName` as a legal identifier,
  - resolves symbol under cursor,
  - computes scoped references bound to the same declaration,
  - returns `WorkspaceEdit.changes` text edits for all matching occurrences,
  - supports declaration-inclusive rename edits.
- Added identifier validation helper for rename names.

Tests added/updated in `internal/lsp/server_test.go`:
- `TestServerInitializeShutdownExit` now asserts `renameProvider.prepareProvider == true`.
- `TestServerPrepareRenameReturnsRangeAndPlaceholder`.
- `TestServerRenameReturnsScopedWorkspaceEdits`.
- `TestServerRenameRejectsInvalidIdentifier`.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 10: cross-open-document references and rename (top-level)

- Extended references/rename from single-document to cross-open-document behavior in `internal/lsp/server.go`.
- New cross-document strategy (conservative):
  - always include scoped references in the origin document,
  - for symbols resolved at top-level scope (`scopeDepth == 0`), also scan other currently open documents,
  - include only occurrences resolving to top-level symbols with the same name and symbol kind,
  - continue respecting `includeDeclaration` for references.
- `textDocument/rename` now produces `WorkspaceEdit.changes` grouped by URI across affected open documents.

Implementation notes:
- Added helper `collectReferencesAcrossOpenDocs` for orchestration.
- Added helper `collectTopLevelReferenceLocations` for per-document top-level reference collection.

Tests added in `internal/lsp/server_test.go`:
- `TestServerReferencesIncludeOpenDocumentsForTopLevelSymbol`.
- `TestServerRenameAppliesEditsAcrossOpenDocumentsForTopLevelSymbol`.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 11: module-aware cross-file references/rename (open-doc index)

- Upgraded cross-file reference/rename logic from name+kind heuristics to a module-aware identity model in `internal/lsp/server.go`.

Implemented module-aware identity/indexing:
- Added module symbol identity type (`module`, `name`, `kind`).
- Added import binding extraction from top-level destructured imports:
  - supports patterns like `val { x } = import("mod")` / `var { x } = import("mod")`.
- Added exported-symbol extraction from `@export` top-level declarations (`val`, `var`, `foreign`).
- Added URI->module-name derivation helper based on `.slug` path with `/lib/` normalization.

Behavior changes:
- `textDocument/references` and `textDocument/rename` now use module-aware identity for cross-open-document resolution:
  - same-module top-level references are resolved via declaration binding,
  - importer aliases are resolved via parsed import bindings (`import("module")` destructuring),
  - local scope references in the origin doc remain scope-safe as before.
- `textDocument/rename` now emits multi-URI workspace edits based on module-aware matches.

Safety/constraints:
- Cross-document propagation is conservative and limited to currently open documents.
- Cross-document local-scope symbols are not propagated; only top-level/module-identity paths are considered.

New helpers introduced:
- `resolveModuleSymbolIdentity`
- `inferExportKindFromOpenDocs`
- `collectExportedTopLevelSymbols`
- `collectImportBindingsForModule`
- `collectReferencesAcrossOpenDocs` (module-aware orchestration)
- `dedupeLocations`

Tests updated/added in `internal/lsp/server_test.go`:
- Updated cross-open-document tests to module-aware importer/exporter scenario.
- `TestServerReferencesIncludeOpenDocumentsForTopLevelSymbol`
- `TestServerRenameAppliesEditsAcrossOpenDocumentsForTopLevelSymbol`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 12: module-object import member indexing (`m = import("mod"); m.name`)

- Extended module-aware cross-file references/rename to support module-object imports in `internal/lsp/server.go`.

New capabilities added:
- Parse top-level module-object bindings:
  - `val m = import("module")` / `var m = import("module")` (single-module import call).
- Resolve cursor-on-member identity for dot lookups:
  - `m.answer` where `m` is a module-object import alias.
- Include module-object member usages in cross-open-document identity reference collection.

Behavior changes:
- `textDocument/references`:
  - if cursor is on a module-object member token (`m.foo`), resolve to module identity and return module-aware references (including declaration when requested).
- `textDocument/prepareRename`:
  - supports module-object member tokens with proper rename range/placeholder.
- `textDocument/rename`:
  - supports initiating rename from module-object member usage and applies edits across exporter/importer usages.

New helpers introduced:
- `collectImportObjectBindings`
- `resolveModuleMemberIdentityAtOffset`
- `collectMemberReferencesForAliases`
- `collectReferencesForIdentityAcrossOpenDocs`

Tests added in `internal/lsp/server_test.go`:
- `TestServerReferencesIncludeModuleObjectMemberUsages`
- `TestServerRenameFromModuleObjectMemberEditsExporterAndUsage`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 13: inline/chained import-member indexing (`import("mod").name`)

- Extended module-aware indexing to support inline/chained import-member expressions in `internal/lsp/server.go`.

Added support for patterns like:
- `import("slug.log").logger(...)`
- `import("mod").answer`

Implementation details:
- Added module resolution for dot-lookup left operand when left side is:
  - module-object alias identifier (`m.name`) or
  - inline import call (`import("mod").name`).
- Added inline member reference collector across open docs:
  - `collectInlineImportMemberReferences`.
- Updated cross-doc identity collection pipeline to include:
  - destructured imports,
  - module-object alias member uses,
  - inline/chained import-member uses.
- Updated cursor identity resolution for rename/references/prepareRename to detect inline import-member tokens.

Bugfix during phase:
- Removed premature early-return in member identity resolver when no module-object aliases exist, which incorrectly blocked inline `import("mod").name` cases.

Tests added in `internal/lsp/server_test.go`:
- `TestServerReferencesIncludeInlineImportMemberUsages`
- `TestServerRenameFromInlineImportMemberEditsExporterAndUsage`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 14: multi-module inline import support + collision-safe rename

- Extended inline/chained import-member handling to support multi-module calls:
  - `import("a", "b").name`
- Added module-candidate resolution for dot lookups:
  - left operand now maps to one or more module candidates (alias or inline import args).
- Added disambiguation against open-document export index:
  - if exactly one candidate module exports the member name, use that module identity.
  - if multiple candidates export the same member, mark as ambiguous.

Safety behavior:
- `textDocument/rename` now fails safely for ambiguous multi-module member targets:
  - returns JSON-RPC invalid params (`-32602`) with clear ambiguity reason.
- `textDocument/prepareRename` returns `nil` for ambiguous targets.
- `textDocument/references` only returns results when member identity is resolvable.

Implementation notes:
- Refactored member identity extraction to collect candidate-module hits at cursor offset.
- Added helpers:
  - `collectModuleMemberHitsAtOffset`
  - `modulesForDotLookupLeft`
  - `modulesExportingName` / `moduleExportsName`
  - `containsString`

Tests added in `internal/lsp/server_test.go`:
- `TestServerReferencesFromMultiModuleInlineImportResolvesWhenUnambiguous`
- `TestServerRenameFromAmbiguousMultiModuleInlineImportFailsSafely`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP compatibility fix: definition response payload shape

- Adjusted `textDocument/definition` response shape in `internal/lsp/server.go` from `Location[]` to a single `Location` (or `null`).
- Motivation: improve client compatibility for editors that are strict/fragile on union payload decoding for definition responses.
- Updated `internal/lsp/server_test.go` definition test to assert object-shaped `Location` response.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP protocol fix: always include `result` in success responses

- Fixed JSON-RPC response encoding in `internal/lsp/server.go` for success replies that return `nil`.
- Previous bug could emit payloads like `{"jsonrpc":"2.0","id":N}` (missing both `result` and `error`), which is invalid JSON-RPC and breaks strict clients.
- Updated writer paths to use explicit response envelopes:
  - success responses always include `result` (including `null`),
  - error responses include `error` only.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP definition enhancement: import-aware module jump (including wildcard imports)

- Enhanced `textDocument/definition` in `internal/lsp/server.go` to resolve imported symbols to source module exports.

New behavior:
- If cursor is on module-member usage (`m.foo` / `import("mod").foo`) and identity resolves, definition jumps to exported symbol in source module.
- If cursor is on a top-level local alias imported via destructuring (`val { foo } = import("mod")`), definition jumps to exported source symbol.
- If local symbol is unresolved and module is imported via wildcard (`var {*} = import("mod")`), definition now resolves against module exports and jumps to source file.

Implementation notes:
- Added wildcard import module extraction: `collectWildcardImportModules`.
- Added module export location resolver with open-doc preference and filesystem fallback:
  - `resolveModuleExportLocation`
  - `modulePathFromURI`
  - `findOpenDocURIByModule`

Tests added in `internal/lsp/server_test.go`:
- `TestServerDefinitionResolvesWildcardImportedExportFromModuleFile`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP alignment: module resolution now follows runtime import search order

- Updated LSP module export resolution in `internal/lsp/server.go` to match runtime import search priority used by `LoadModule`:
  1. local root (directory of current file) + `<module path>.slug`
  2. `$SLUG_HOME/lib/<module path>.slug`
- Replaced previous path heuristic (`/lib/` URI slicing) with explicit candidate path generation:
  - `modulePathCandidatesFromURI`
- `resolveModuleExportLocation` now tries candidates in runtime-consistent order and keeps open-document preference when module file is already open.

Test update:
- `TestServerDefinitionResolvesWildcardImportedExportFromModuleFile` now sets `SLUG_HOME` via `t.Setenv` to validate fallback behavior consistently.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 15: codeAction quick-fix for unresolved imports

- Added initial `textDocument/codeAction` support in `internal/lsp/server.go`.
- Advertised `codeActionProvider: true` in initialize capabilities.

Implemented quick-fix behavior:
- For unresolved identifier under selected range, LSP now suggests import actions:
  - `Add import for '<symbol>' from '<module>'`
- Actions return `WorkspaceEdit` inserting:
  - `val { <symbol> } = import("<module>")`

Module candidate discovery for quick-fix:
- open documents with `@export` symbols,
- `$SLUG_HOME/lib` recursive scan for exported top-level symbols.

Insertion behavior:
- inserts import near top import block via `importInsertionPosition` heuristic.

Tests added/updated in `internal/lsp/server_test.go`:
- Initialize capability assertion includes `codeActionProvider == true`.
- `TestServerCodeActionSuggestsImportForUnresolvedSymbolFromSlugHome` validates quick-fix generation using `SLUG_HOME/lib` export discovery.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP codeAction enhancement: prefer extending existing import bindings

- Improved unresolved-import quick-fix generation in `internal/lsp/server.go`.
- New priority for import edits:
  1. extend an existing same-module destructured import when present (e.g. `val { map } = import("slug.std")` -> add `, reduce`),
  2. otherwise insert a new import line.

Implementation notes:
- Added import edit planner:
  - `buildImportEditPlan`
  - `extendExistingImportBinding`
- Added AST offset helper for precise insertion in existing import pattern:
  - `offsetAfterNode`

Tests added/updated in `internal/lsp/server_test.go`:
- `TestServerCodeActionExtendsExistingDestructuredImport` validates in-place extension (`", reduce"`) instead of new-line insertion.
- Updated module fixture sources to use real newlines for robust export parsing in codeAction tests.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP codeAction enhancement: qualify unresolved symbols with existing import alias

- Added a second unresolved-symbol quick-fix path in `internal/lsp/server.go`:
  - when a module-object alias import exists (e.g. `val std = import("slug.std")`) and module exports the unresolved symbol,
  - offer `Qualify with 'std.<symbol>'`.
- Action applies an in-place text replacement of the unresolved identifier range (e.g. `reduce` -> `std.reduce`).

Behavior notes:
- Qualification quick-fixes are generated before import-insertion actions.
- Module export verification uses open-doc export index first, then runtime-aligned module path candidates (including `SLUG_HOME`).

Implementation additions:
- `moduleMayExportName` helper.
- `handleCodeAction` now emits both qualification and import actions when applicable.

Tests added in `internal/lsp/server_test.go`:
- `TestServerCodeActionSuggestsQualifyWithExistingImportAlias`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP codeAction ranking/dedup: deterministic quick-fix ordering

- Added deterministic ranking + dedup for code actions in `internal/lsp/server.go`.
- Quick-fix rank order (least disruptive first):
  1. qualify with existing alias (`std.reduce`)
  2. extend existing import binding (`, reduce`)
  3. insert new import line (`val { reduce } = import("...")`)
- Added stable secondary ordering by title for deterministic UX.

Implementation additions:
- `rankAndDedupeCodeActions`
- `actionRank`
- `actionDedupKey`
- `importEditPlan.Mode` to classify import edit strategy (`extend` vs `insert`).

Test update:
- `TestServerCodeActionSuggestsQualifyWithExistingImportAlias` now asserts ranked first action is the qualification fix.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP codeAction preference refinement: favor import-extension over alias qualification

- Refined quick-fix ranking in `internal/lsp/server.go` to prefer import-style consistency.
- Updated ranking order when multiple fixes are available:
  1. extend existing import binding (`extend`)
  2. qualify with alias (`qualify`)
  3. insert new import line (`insert`)

Implementation details:
- Added internal code-action metadata field `RankGroup` (non-serialized) to carry strategy classification.
- `handleCodeAction` now tags actions with `RankGroup`:
  - `extend`, `qualify`, `insert`.
- `actionRank` now uses `RankGroup` primarily for deterministic style-aware ordering.

Test added in `internal/lsp/server_test.go`:
- `TestServerCodeActionPrefersExtendImportOverQualifyWhenBothAvailable`
  - verifies first ranked action is import extension (`", reduce"`) when both extension and qualification fixes are possible.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 16: textDocument/signatureHelp (initial function parameter help)

- Added `textDocument/signatureHelp` support in `internal/lsp/server.go`.
- Advertised `signatureHelpProvider` capabilities with trigger characters:
  - `(`
  - `,`

Implemented behavior:
- Detect enclosing call context and active parameter index from cursor position.
- Resolve function signatures from:
  - local function bindings (`val/var name = fn(...)`),
  - imported destructured bindings,
  - wildcard import module exports,
  - module-member forms (`alias.fn(...)`, `import("mod").fn(...)`).
- Returns LSP `SignatureHelp` payload with:
  - signature label,
  - parameter labels,
  - active parameter index.

Implementation helpers added:
- `collectFunctionSignatures`
- `resolveFunctionAt`
- `findCallContext`
- `resolveSignatureForCallee`
- `resolveModuleExportSignature`
- inline module parse helpers (`parseInlineImportModules`, `moduleForAliasInDoc`).

Tests added/updated in `internal/lsp/server_test.go`:
- initialize capabilities now assert `signatureHelpProvider` exists.
- `TestServerSignatureHelpLocalFunctionAndActiveParam`
- `TestServerSignatureHelpImportedFunction`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 17: signatureHelp call-context robustness

- Hardened call-context scanning in `internal/lsp/server.go` (`findCallContext`) to improve active-parameter inference in real code.
- Backward scan now safely ignores parentheses inside string literals and handles escape sequences.
- Forward argument scan now tracks nested `()`, `[]`, `{}` and string literals, so commas only increment active parameter at the top level of the current call.
- This reduces false/missing `signatureHelp` responses in nested calls and string-heavy argument lists.

Tests added:
- `internal/lsp/server_test.go`:
  - `TestServerSignatureHelpIgnoresCommasInsideStringsAndNestedCalls`
  - Verifies outer-call `activeParameter` remains correct when prior arguments contain:
    - commas inside string literals,
    - nested function calls.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 18: signatureHelp delimiter regression (map/list literals)

- Added a regression test for signature help active-parameter tracking when earlier arguments contain mixed map/list literals with internal commas.
- Ensures delimiters inside `{...}` and `[...]` do not increment the outer-call argument index.

Tests added:
- `internal/lsp/server_test.go`:
  - `TestServerSignatureHelpIgnoresCommasInsideMapAndListLiterals`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 19: call-context escape handling hardening

- Fixed backward call-context scanning in `findCallContext` to correctly handle escaped quote characters while traversing string literals.
- Added `isEscapedAt` helper (counts preceding backslashes) and used it when deciding whether a quote closes the current string during backward scan.
- This prevents false call-context loss around arguments like `"...\",..."` that include escaped delimiters.

Tests added/updated:
- `internal/lsp/server_test.go`
  - Added `TestFindCallContextHandlesEscapedQuotesInStringArgs`.
  - Directly validates `findCallContext` returns callee `sum` and `activeParameter=2` for `sum("a\",b", 2, 3)`.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 20: multiline signatureHelp position-mapping regression

- Added an end-to-end `textDocument/signatureHelp` regression for multiline call expressions with nested map/list literals.
- Verifies that LSP line/character position mapping and call-context argument counting remain aligned when commas appear inside nested literals across lines.

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpMultilineNestedLiteralCall`
  - Asserts `activeParameter=2` for a multiline `sum(...)` call where previous args are nested literals.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 21: signatureHelp nil-response regression outside call context

- Added end-to-end regression coverage for `textDocument/signatureHelp` when cursor is outside a callable argument context.
- Ensures server returns `null` result (not an invalid payload or empty structure), matching expected LSP response behavior.

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpReturnsNilOutsideCallContext`
  - Verifies `result == nil` for:
    - cursor at start of a non-call line,
    - cursor after a completed call expression.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 22: signatureHelp trigger-sequence stability regression

- Added end-to-end regression coverage to ensure repeated `textDocument/signatureHelp` requests remain stable as cursor advances through a call expression.
- Covers expected `activeParameter` progression across canonical trigger points around `(` and `,`.

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpStableAcrossTriggerSequence`
  - Verifies active parameter transitions for `sum(1, 2, 3)`:
    - at first argument: `0`
    - at second argument: `1`
    - at third argument: `2`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 23: signatureHelp protocol-shape regression (activeSignature)

- Strengthened signature-help regression coverage to lock response shape for future multi-signature support.
- Existing trigger-sequence test now also asserts:
  - `activeSignature` is present and equals `0`,
  - `signatures` payload is present and non-empty.

Tests updated:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpStableAcrossTriggerSequence`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 24: canceled-request response suppression

- Enforced `$/cancelRequest` behavior for request responses by suppressing outbound success payloads for canceled request IDs.
- Added cancel-id normalization helper and one-shot consume semantics so cancellation entries are cleared when applied.
- This prevents stale `signatureHelp` responses from being emitted for requests canceled by the client.

Implementation updates:
- `internal/lsp/server.go`
  - `writeResult` now checks cancellation state before sending.
  - Added:
    - `isCanceledID(id json.RawMessage) bool`
    - `cancelIDKey(id json.RawMessage) string`

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpCanceledRequestSuppressesResponse`
  - Verifies a canceled `textDocument/signatureHelp` request ID yields no response with that ID.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 25: canceled-request suppression for error responses

- Completed cancellation handling so both success and error responses are suppressed for canceled request IDs.
- `writeError` now checks cancellation state before sending outbound payloads.
- Added generic cancel-ID normalization for interface IDs (numeric and string) so `$/cancelRequest` matching is consistent across response paths.

Implementation updates:
- `internal/lsp/server.go`
  - `writeError` now suppresses canceled responses.
  - Added:
    - `isCanceledAnyID(id interface{}) bool`
    - `cancelAnyIDKey(id interface{}) string`

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerCanceledUnknownRequestSuppressesErrorResponse`
  - Cancels request id `9` before sending unknown method request and verifies no response with id `9` is emitted.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 26: cancellation regression coverage for string request IDs

- Added cancellation regressions for string-based request IDs to lock ID normalization behavior across clients that do not use numeric IDs.
- Covers both success-response suppression and error-response suppression paths.

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpCanceledStringIDSuppressesResponse`
  - `TestServerCanceledUnknownRequestStringIDSuppressesErrorResponse`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 27: leverage existing Slug doc comments for hover and signature help

- Reused existing parser doc metadata (no language doc format changes) to enrich LSP UX.
- Hover now includes top-level symbol doc comments when present.
- Signature help now includes function doc comments and per-parameter docs parsed from existing `@param <name> ...` lines inside the same doc block.

Implementation updates:
- `internal/lsp/server.go`
  - `lspParameterInformation` now supports `documentation`.
  - `functionSignature` now carries aligned `ParamDocs`.
  - `collectSymbols` now propagates top-level `val/var` docs into symbol detail.
  - `collectFunctionSignatures` now propagates `val/var/foreign` docs and extracts per-parameter docs via `parseParamDocs`.
  - `handleSignatureHelp` now emits per-parameter documentation payloads.
  - Added helper: `parseParamDocs(doc string) map[string]string`.

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerHoverIncludesDocComment`
  - `TestServerSignatureHelpIncludesDocAndParamDocs`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 28: completion resolve doc enrichment for imported module symbols

- Extended `completionItem/resolve` to enrich symbols imported from modules (including modules not currently open in the editor).
- Resolution now follows language import semantics via existing module path resolution (`SLUG_HOME` and module candidate search), then loads module source and extracts symbol docs/kinds.
- Aligned with `slug.doc.markdown` behavior by propagating full doc-comment text (trimmed, uncollapsed) into markdown documentation payloads.

Implementation updates:
- `internal/lsp/server.go`
  - `handleCompletionResolve` now attempts import-based enrichment when local symbol detail is missing/empty.
  - Added:
    - `resolveCompletionImportedSymbol(originURI, src, localName)`
    - `resolveModuleExportSymbolInfo(originURI, module, name)`
  - Module symbol extraction now uses rich symbol/signature collectors to preserve doc comments and infer function kind at top level.

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerCompletionResolveEnrichesImportedItemFromModuleDocs`
  - Verifies completion resolve for imported `reduce` reads function kind and doc text from module file under `SLUG_HOME`.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 29: import-aware hover doc enrichment

- Extended hover to use the same import-aware symbol enrichment strategy as completion resolve.
- When hovering an imported alias with no local doc detail, hover now resolves module export metadata (kind + docs) and renders it in markdown.
- This keeps hover behavior consistent with completion docs and aligned with full-doc rendering conventions from `slug.doc.markdown`.

Implementation updates:
- `internal/lsp/server.go`
  - `handleHover` now attempts import-based enrichment via existing `resolveCompletionImportedSymbol` when local symbol detail is empty or kind is generic `variable`.

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerHoverIncludesImportedAliasDocComment`
  - Verifies imported `reduce` hover shows:
    - function kind,
    - doc text loaded from module source under `SLUG_HOME`.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 30: signatureHelp imported-doc coverage for wildcard and member calls

- Added end-to-end regressions to confirm signature-help doc enrichment for imported calls loaded from disk modules (`SLUG_HOME`) across additional import forms.
- Confirms docs and parameter docs are available for:
  - wildcard import calls (`var {*} = import("slug.std")`),
  - module member calls (`val std = import("slug.std"); std.reduce(...)`).

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpImportedWildcardIncludesDocsFromDiskModule`
  - `TestServerSignatureHelpImportedMemberCallIncludesDocsFromDiskModule`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 31: include `@testWith` examples in markdown docs

- Added shared LSP doc markdown enrichment for function symbols/signatures to include rendered `@testWith` examples.
- Output format aligns with `slug.doc.markdown` style:
  - `#### Examples` heading
  - fenced `slug` code block
  - lines like `fnName(args...)  // => expected`

Implementation updates:
- `internal/lsp/server.go`
  - Function doc assembly now uses `buildFunctionDocMarkdown(name, doc, hasDoc, tags)`.
  - Added example render helpers:
    - `renderTestWithExamplesMarkdown(name, tags)`
    - `tagArgs(tags, "@testWith")`
    - `renderTestWithValue(expr)`
  - Applied during symbol/signature extraction for:
    - top-level `val/var` function bindings,
    - top-level foreign functions.

Tests updated:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpIncludesDocAndParamDocs` now asserts `@testWith` examples appear in signature markdown documentation.
  - `TestServerCompletionResolveEnrichesImportedItemFromModuleDocs` now asserts imported completion docs include an examples section from `@testWith`.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 32: `@testWith` string example quoting fix

- Fixed `@testWith` markdown rendering so string literals are always emitted with quotes.
- This avoids ambiguous example output where string inputs could be mistaken for identifiers or numeric literals.

Implementation update:
- `internal/lsp/server.go`
  - `renderTestWithValue` now renders `*ast.StringLiteral` via `strconv.Quote(x.Value)`.

Tests updated:
- `internal/lsp/server_test.go`
  - `TestServerSignatureHelpIncludesDocAndParamDocs` now includes a string-input `@testWith` case and asserts rendered output contains:
    - `add("1")  // => 1`

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 33: honor `codeAction.context.only` for source-only requests

- Fixed `textDocument/codeAction` behavior for clients requesting only source actions (e.g., Sublime sending `context.only: ["source"]`).
- Previous behavior always emitted `kind: "quickfix"`, which source-only clients could ignore/filter.
- New behavior:
  - parses `context.only`,
  - when source-only is requested, import actions are emitted as `kind: "source.organizeImports"`,
  - quick-fix qualify actions are omitted unless quickfix kinds are allowed by `context.only`.

Implementation updates:
- `internal/lsp/server.go`
  - Extended `codeActionParams` with `context.only`.
  - Updated `handleCodeAction` kind selection/filtering.
  - Added helpers:
    - `containsCodeActionKind(only, kind)`
    - `prefersSourceActionsOnly(only)`

Tests added:
- `internal/lsp/server_test.go`
  - `TestServerCodeActionSourceOnlyReturnsSourceKindImportAction`
  - Verifies `only:["source"]` returns import action kind `source.organizeImports`.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### LSP Phase 34: classify `val` as constant and `var` as variable

- Updated LSP symbol classification to better match language semantics:
  - `val` bindings now surface as `constant`.
  - `var` bindings surface as `variable`.
  - function bindings (`... = fn(...)`) surface as `function`.

Implementation updates:
- `internal/lsp/server.go`
  - `collectSymbols`:
    - `val` function bindings -> `function`.
    - `val` non-function bindings -> `constant`.
    - `var` function bindings -> `function`.
  - `collectTopLevelSymbols`: top-level non-function `val` now `constant`.
  - `collectExportedTopLevelSymbols`: exported non-function `val` now `constant`.
  - LSP kind maps:
    - `constant` -> `DocumentSymbolKind.Constant` (14)
    - `constant` -> `CompletionItemKind.Constant` (21)

Compatibility follow-up fixes:
- import-aware enrichment now upgrades both `variable` and `constant` placeholders to imported kinds (e.g., function).
- module symbol identity inference now treats `constant` similar to previous `variable` fallback behavior.
- module-member identity kind filtering relaxed (`kind:""`) to avoid over-filtering during cross-module rename/reference resolution.

Tests added/updated:
- `internal/lsp/server_test.go`
  - Added `TestServerHoverDistinguishesValConstantAndVarVariable`.
- Existing hover/completion/rename tests continue to pass after resolver adjustments.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### Language: core colon type-annotation support (non-breaking rollout)

- Implemented first-class optional colon type annotations in core syntax while preserving existing tag-based typing for compatibility.
- Added parser + AST + semantic type-check integration for:
  - variable/constant declarations: `var a:num = 1`, `val s:str = ""`
  - function parameters: `fn(x:num, y:list<num>)`
  - function return types: `fn(x:num):num { ... }`
  - declared type forms in checker: scalars, unions (`num|str`), composites (`list<T>`, `map<K,V>`, `fn<A,B,R>`), and struct refs (`struct<User>` / bare identifiers as struct refs).

Implementation updates:
- `internal/ast/ast.go`
  - Added optional type fields:
    - `VarExpression.Type`
    - `ValExpression.Type`
    - `FunctionParameter.Type`
    - `FunctionLiteral.ReturnType`
  - String renderers now include these annotations when present.

- `internal/parser/parser.go`
  - Added annotation parsing after identifier bindings/params and after function parameter list for returns.
  - Added `parseTypeAnnotationLiteral(...)` token-collector with nested generic depth handling.
  - Preserved compatibility with keyword-shaped and underscore parameter names.

- `internal/semantic/typecheck.go`
  - Added declared-type parser and constraint bridge:
    - `parseDeclaredType(...)`
    - `splitTypeTopLevel(...)`
    - `isSimpleTypeIdent(...)`
  - Enforced declaration/parameter/return annotations as type constraints in inference.

Tests added/updated:
- `internal/parser/parser_test.go`
  - Expanded parameter parsing cases for colon syntax.
  - Added `TestDeclarationAndReturnTypeAnnotations`.

- `internal/semantic/analyzer_test.go`
  - Added `TestSemanticTypeCheckSupportsColonTypeAnnotations`.

Validation performed:
- `go test ./... -count=1`

### Language: declared nilability enforcement for colon annotations

- Enforced explicit nilability for declared colon types:
  - `var/val x:num = nil` now fails type checking.
  - `var/val x:num|nil = nil` is accepted.
  - Reassignment also honors declared nilability (`x = nil` only allowed when declared type includes `nil`).

Implementation updates:
- `internal/semantic/typecheck.go`
  - Added declared-type scope tracking to bind identifier declarations to parsed declared types across lexical scopes.
  - Added nilability checks for:
    - declaration bindings (`var` / `val`)
    - identifier reassignment paths
  - Added helper logic for nilability evaluation (`typeAllowsNil`, `typeMayBeNil`) and pattern binding for declared types.

Tests added:
- `internal/semantic/analyzer_test.go`
  - `TestSemanticTypeCheckEnforcesDeclaredNilabilityOnBinding`
  - `TestSemanticTypeCheckEnforcesDeclaredNilabilityOnAssignment`

Validation performed:
- `go test ./internal/semantic -count=1`
- `go test ./... -count=1`

### Language: stdlib/test migration from `@tag` param hints to colon types

- Migrated existing Slug source signatures from legacy parameter tags to colon type syntax.
  - Example: `fn(@num n, @fn f)` -> `fn(n:num, f:fn)`.
  - Applied across `lib/**/*.slug` and relevant `tests/**/*.slug` modules.
- Preserved semantic/decorator tags (`@export`, `@main`, `@returns`, `@throws`, `@testWith`, etc.).
- Kept struct field typing in current supported form (`@type field`) because struct-field colon syntax is not yet supported by parser.
- Updated a few struct-ref parameters to colon form with nominal refs (e.g. `m:Migration`, `v:User`).

Validation performed:
- `go test ./... -count=1`

### Migration compatibility: colon-typed params preserve overload dispatch

- Fixed overload dispatch regression after migrating function parameter hints from `@tag` to `name:type` syntax.
- Parser now infers runtime overload tags from colon parameter types when explicit parameter tags are absent.
  - Examples:
    - `name:str` -> `@str`
    - `b:fn` -> `@fn`
    - `m:User` -> `@struct(User)`
- This keeps existing function overloading behavior compatible during migration while retaining colon syntax.

Validation performed:
- `go run ./cmd/app/main.go --root ./tests tests/functions-type-tags.slug`
- `go test ./... -count=1`

### Language break: struct field types now use colon syntax only

- Introduced hard-break struct schema field syntax:
  - New: `field:type` (for example `age:num`, `type:str = "Error"`)
  - Removed: legacy tag-style struct field hints (`@num age`, `@str name`)
- Parser now rejects legacy struct-field tag syntax with an explicit error.
- Updated AST model for struct fields to store declared field type strings.
- Updated semantic struct schema typing to read declared field types directly.
- Migrated stdlib and tests to colon struct field syntax.

Implementation updates:
- `internal/ast/ast.go`
  - `StructField` now stores `Type string` (removed `Tags`).
- `internal/parser/parser.go`
  - `parseStructSchemaField` now parses `name:type` and errors on `@type name`.
- `internal/semantic/typecheck.go`
  - struct schema registration now uses declared field type parsing rather than field tags.
- `internal/vm/compiler.go`
  - struct schema compilation now stores field declared type and derives runtime hint tags from that type.
- `internal/object/object.go`
  - `StructSchemaField` includes `Type string`; struct inspect output renders colon field types.
- `internal/parser/debug_ast_text.go`, `internal/parser/debug_ast_json.go`
  - debug renderers updated for struct field type representation.

Validation performed:
- `go test ./... -count=1`

### LSP: signature help supports colon type syntax

- Updated LSP signature help rendering to include colon-declared parameter types in labels.
  - Example: `sum(a:num, b:str)` instead of `sum(a, b)` when types are declared.
- Uses parser AST parameter `Type` metadata directly for local and imported function signatures.
- Added regression test for typed signature labels:
  - `internal/lsp/server_test.go` `TestServerSignatureHelpShowsColonTypedParameters`.

Validation performed:
- `go test ./internal/lsp -count=1`
- `go test ./... -count=1`

### Documentation migration: colon type syntax in generated docs

- Updated documentation generators to render type annotations in colon-style syntax.
  - Function params now render as `name:type` instead of `@type name`.
  - Struct fields now render as `field:type`.
  - Return/throw types are normalized for display (`@str` -> `str`, `@struct(User)` -> `User`).
- Updated developer guide examples to use colon syntax for typed functions/structs.
- Regenerated docs artifacts:
  - `lib/MANIFEST.ai`
  - `docs/_libraries/*.md`

Implementation updates:
- `lib/slug/doc/manifest.slug`
  - Type normalization + colon rendering in params/returns/errors/field lines.
- `lib/slug/doc/markdown.slug`
  - Type normalization + colon rendering in signatures/param tables/throws blocks.
- `docs/_developers-guide/*.md`
  - Example syntax refresh to colon types.

Validation performed:
- `make manifest`
- `make generate-docs`
- `go test ./... -count=1`

### Docs generator: benchmark function descriptions and formal return types

- Fixed missing function prose for `slug.benchmark` docs by adding explicit `@desc('...')` tags to exported benchmark functions and using generator fallback logic when runtime `describe(...).docs` is empty.
- Updated benchmark exports with explicit return annotations for doc output consistency:
  - `printResult(res):nil`
  - `printCompareReport(report):nil`
  - `compare(...):map`
- Updated markdown generator to render docs with fallback order:
  - doc comment text
  - `@desc` tag (if doc comment metadata is empty)
- Updated manifest generator to use the same `@desc` fallback for `desc:` lines.
- Regenerated artifacts:
  - `docs/_libraries/slug.benchmark.md`
  - `lib/MANIFEST.ai`

Validation performed:
- `SLUG_HOME=$(pwd) go run ./cmd/app/main.go doc --dir ./lib --out ./lib/MANIFEST.ai manifest`
- `SLUG_HOME=$(pwd) go run ./cmd/app/main.go doc --dir ./lib --moduleToc --multiPage --out ./docs/_libraries markdown`

### Type display update: use `any` instead of `?` in generated signatures

- Introduced `any` as the canonical unknown/unspecified type name in documentation outputs.
- Updated doc generators so missing function return annotations render as `any` instead of `?`.
- Updated type normalization so explicit `@returns('?')` also renders as `any`.
- Regenerated generated artifacts:
  - `lib/MANIFEST.ai`
  - `docs/_libraries/*.md`

Implementation updates:
- `lib/slug/doc/markdown.slug`
  - default return type fallback `?` -> `any`
  - normalize type expression maps `?` -> `any`
- `lib/slug/doc/manifest.slug`
  - default return type fallback `:?` -> `:any`
  - normalize type expression maps `?` -> `any`

Validation performed:
- `SLUG_HOME=$(pwd) go run ./cmd/app/main.go doc --dir ./lib --out ./lib/MANIFEST.ai manifest`
- `SLUG_HOME=$(pwd) go run ./cmd/app/main.go doc --dir ./lib --moduleToc --multiPage --out ./docs/_libraries markdown`

### Doc comment metadata: remove tag workaround and restore binding-sourced docs

- Removed temporary function-level tag workaround from `slug.benchmark` (`@desc(...)` lines removed).
- Restored docs generation to rely on canonical Slug doc comments (`/** ... */`) instead of fallback tag text.
- Fixed parser separator handling so pending doc comments are not dropped on top-level `;` separators before a tagged declaration.
- Fixed VM doc metadata application to always apply `OpSetDoc` in current environment (previously gated by `env.Outer == nil`).
- Added `slug.meta.describeSymbol(module, symbol)` to expose symbol metadata directly from module bindings, preserving top-level doc comments for imported symbols.
- Updated generators to use `describeSymbol` for per-symbol reflection:
  - `lib/slug/doc/manifest.slug`
  - `lib/slug/doc/markdown.slug`
- Unified callable metadata handling so VM functions participate in function-group bindings and reflection paths:
  - callable detection/merge in `internal/object/environment.go`
  - callable signature/params accessors on object/vm function types.

Validation performed:
- `go test ./internal/foreign ./internal/object ./internal/vm ./internal/parser -count=1`
- `SLUG_HOME=$(pwd) go run ./cmd/app/main.go doc --dir ./lib --out ./lib/MANIFEST.ai manifest`
- `SLUG_HOME=$(pwd) go run ./cmd/app/main.go doc --dir ./lib --moduleToc --multiPage --out ./docs/_libraries markdown`

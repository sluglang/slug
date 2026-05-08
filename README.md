Slug
===

A small, opinionated programming language with a functional core, explicit control flow, and structured concurrency.

[![build-test-tag-and-release](https://github.com/sluglang/slug/actions/workflows/build-test-tag-and-release.yml/badge.svg)](https://github.com/sluglang/slug/actions/workflows/build-test-tag-and-release.yml)

[![GitHub watchers](https://img.shields.io/github/watchers/babyman/slug-lang.svg?style=social&label=Watch)](https://github.com/sluglang/slug/watchers)
[![GitHub forks](https://img.shields.io/github/forks/babyman/slug-lang.svg?style=social&label=Fork)](https://github.com/sluglang/slug/fork)
[![GitHub stars](https://img.shields.io/github/stars/babyman/slug-lang?style=social&label=Star)](https://github.com/sluglang/slug/stargazers)

## Overview

Slug focuses on readability, explicitness, and predictable behavior. It combines a compact syntax with a small standard
library and an emphasis on data-first pipelines.

Key ideas:

- Functional-first style with first-class functions and pattern matching.
- Structured concurrency with explicit `spawn` and `await`.
- Simple modules and live imports for predictable initialization.
- Small builtins and clear defaults over implicit magic.

## Example

```slug
var {*} = import("slug.std")

val numbers = [1, 2, 3, 4, 5, 6]

val result = numbers
    /> map(fn(x) { x * x })
    /> filter(fn(x) { x % 2 == 0 })
    /> reduce(0, fn(acc, x) { acc + x })

println("Result:", result)
```

## Getting started

- Installation and setup: http://www.sluglang.org
- Developers guide: `docs/developers-guide.md`
- Setup guide: `docs/setup.md`
- Cookbook: `docs/cookbook.md`
- Architectural Decision Records: `docs/adr.md`

## Running Slug

```shell
slug --root [path to module root] script[.slug] [args...]
```

You can also run a file directly:

```shell
slug hello.slug
```

## Environment variables

- `SLUG_HOME`: directory where Slug searches for libraries and modules.

## Status

Slug is an active work in progress. It is used in real projects and evolves quickly; breaking changes may occur while
features stabilize.

## Development

Build from source with Go installed:

```shell
git clone https://github.com/sluglang/slug.git
cd slug-lang
make build
```

## Technical Notes

### VM and runtime ownership

- `internal/vm` owns execution behavior:
  - bytecode compiler and instruction model (`compiler.go`, `bytecode.go`)
  - bytecode executor and runtime semantics (`executor.go`)
  - VM task handles (`task.go`)
  - VM call bridge context and nursery lifecycle (`call_context.go`, `nursery_scope.go`)
  - VM-side module prep and tag evaluation helpers (`program_prepare.go`)
- `internal/runtime` owns orchestration and environment setup:
  - module discovery/loading and parse lifecycle
  - builtin and foreign registry construction
  - top-level module metadata/export wiring

### Runtime to VM API boundary

`internal/runtime` calls into VM through:

- `vm.PrepareProgram`
- `vm.ApplyForeignTags`
- `vm.EvalTagArgs`
- `vm.NewExecutorWithBridgeFactory`
- `vm.NewVMCallContext`

### Dependency direction rules

- Keep dependency direction one-way: `runtime -> vm`.
- Avoid introducing `vm -> runtime` imports.
- Keep bridge dependencies injected via `VMCallContextDeps` instead of concrete runtime types.
- Add VM behavior tests in `internal/vm` unless behavior is explicitly runtime orchestration.

## License

See `LICENSE.md`.

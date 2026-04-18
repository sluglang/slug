# AGENTS.md

This file is the operating guide for agents (human or AI) working in this repository.

## Purpose

Slug is a small, opinionated programming language implemented in Go with its standard library and tests stored in this repo.

Primary goals when making changes:

- Preserve language behaviour and runtime stability.
- Keep parser/runtime changes covered by tests.
- Prefer small, focused diffs over broad refactors.

## Environment

- Required Go version: `1.25` (see `go.mod`).
- Set `SLUG_HOME` to the repository root when running CLI commands directly:
  - `export SLUG_HOME=$(pwd)`
- Main entrypoint: `./cmd/app/main.go`

## Key reference docs

- `SLUG.md`: AI-friendly language-level reference for Slug syntax and behaviour.
- `lib/MANIFEST.md`: AI-friendly reference for standard library modules and functions.
- In this repository, these are currently stored as `SLUG.ai` and `lib/MANIFEST.ai`; treat them as canonical when the `.md` variants are not present.

## Repository map

- `cmd/app/`: CLI entrypoint.
- `internal/`: language implementation (lexer, parser, runtime, objects, ast, token, util).
- `lib/`: Slug standard library modules.
- `tests/`: positive `.slug` execution tests.
- `tests-negative/`: negative tests expected to fail.
- `test-suites/`: built-in test runner suites (including DB-related suites).
- `docs/`: user/developer docs and generated library docs.
- `extras/`: editor tooling and examples.

## Canonical commands

Use `make` targets where possible.

- Build: `make build`
- Run tests (Go + Slug integration suites): `make test`
- Run locally: `make run ARGS='--root ./tests ./tests/boolean-logic.slug'`
- Regenerate docs/manifests: `make generate-docs`
- Stress run with race detector: `make stress`

Notes:

- CI runs `make test` on pushes affecting `VERSION`, `*.go`, or `*.slug`.
- macOS build may codesign the local binary in `make build`/`make release`.

## Change guidelines

- Keep changes scoped to the task; avoid unrelated cleanup.
- For language semantics changes, update/add:
  - relevant files in `internal/`
  - `.slug` tests in `tests/` or `tests-negative/`
  - docs when user-facing behaviour changes
- Do not commit generated artifacts unless they are intentionally updated (e.g. docs output requested).
- Preserve existing style and naming; run `gofmt` on modified Go files.

## Testing expectations

Before finishing large changes:

1. Run targeted checks for touched areas when practical.
2. Run `make test` for full validation.
3. If full validation is skipped, clearly state what was not run and why.

For parser/runtime behaviour changes, include at least one regression test in Slug test suites.

## Safety and review checklist

- Verify imports and module roots still resolve with `--root` semantics.
- Confirm negative tests still fail for the expected reason.
- Avoid breaking existing CLI flags or command behaviour unless explicitly requested.
- Prefer deterministic tests; avoid introducing time/network dependencies into suites.

## PR/commit guidance

- Use concise commit messages with intent and scope.
- In PR descriptions, include:
  - behaviour changed
  - test coverage added/updated
  - any follow-up work or known limitations

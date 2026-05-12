Slug is a small, opinionated programming language implemented in Go with its standard library and tests stored in this repo.

## Project Preferences
- Required Go version: `1.25` (see `go.mod`).
- Set `SLUG_HOME` to the repository root when running CLI commands directly:
    - `export SLUG_HOME=$(pwd)`
- Main entrypoint: `./cmd/app/main.go`
- `make test` runs all tests.
- `make manifest` to generate updated MANIFEST.ai

## Interaction Rules
* Ask clarifying questions if input is unclear.
* Explain why and suggest alternatives if task is not feasible.
* Use structured, readable formatting (headings, lists, code blocks).
* Follow instructions closely and explain clearly what you have done.
* Don't modify code unrelated to the current task.
* Try always to match the style of the code you are touching.

## Coding Standards
* Write meaningful tests with assertions for all code.
* Avoid duplicated test assertions.
* Maintain evolving test coverage.
* Apply Four Rules of Simple Design:
    1. Code works (passes tests).
    2. Reveals intent.
    3. No duplication.
    4. Minimal elements.
* Prefer functional style:
* Use explicit parameters.
* Prefer immutability.
* Prefer declarative over imperative.
* Minimize state.

## Architecture
* Modularize by concern, not by technical layer.
* One responsibility per module.
* Low inter-module coupling.
* Short functions, no overengineering.

## Workflow
* Read `spec.md` before coding.
* Append changes to the bottom of `spec.md` after task (log changes).
* Write and pass tests before finalizing.
* Keep a `README.md` with setup/run info.
* Ensure language changes are reflected in `typecheck.go`, `slug.ebnf`, `SLUG.ai`, Slug.sublime-package and Slug.tmLanguage.
* Store all docs/specs in Markdown.

## Commit Strategy
* One prompt = one commit.
* Each commit:
* Self-contained.
* Includes tests.
* Uses 50/70 commit message format.
* Commit messages MUST use Semantic Commit Messages with one of the following types (feat, refactor, fix, chore)

### Commit message format:
```text
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

## Safe Practices
* Do not change test assertions during refactoring.
* Do not skip failing tests.
* Do not invent unknown APIs; ask if you are unsure.

## Goal
Produce consistent, safe, testable, and maintainable code.
Stick to the rules---no shortcuts.

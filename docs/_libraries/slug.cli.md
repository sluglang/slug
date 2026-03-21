---
title: cli (slug)
---

## slug.cli

slug.cli — command-line interface utilities

Provides `opt` for simple multi-source option coercion and `argsWith`
for full declarative CLI argument parsing, validation, and help generation.

## Quick start

```slug
val { argsWith } = import("slug.cli")

val spec = {
  name:    "backup",
  version: "1.2.0",
  summary: "Backup files to a remote location",
  strict: false,
  alias:    { a: "all", u: "user" },
  required: ["user"],
  defaults: { all: false },
  types:    { all: "bool", user: "str" },
  desc:     { user: "User account to run as", all: "Process all files" },
  positional: { min: 1, max: 1, name: "file" },
}

val result = argm() /> argsWith(spec)

match result {
  { ok: true, options, positional } => run(options, positional)
  { action: :help,    text }        => { println(text); import("slug.sys").exit(0) }
  { action: :version, text }        => { println(text); import("slug.sys").exit(0) }
  { errors }                        => { errors /> map(fn(e) { println(e.msg) }); import("slug.sys").exit(1) }
}
```

## Result shapes

`argsWith` always returns a map. The `ok` field is the primary discriminator:

**Success:** `{ ok: true, action: :none, options: @map, positional: @list, errors: [] }`

**Help/version:** `{ ok: false, action: :help | :version, text: @str, options: {}, positional: [], errors: [] }`

**Validation failure:** `{ ok: false, action: :error, options: {}, positional: [], errors: @list }`

## Spec fields

| Field        | Type    | Description                                                        |
|--------------|---------|--------------------------------------------------------------------|
| `name`       | `@str`  | Program name; used in help and version output                      |
| `version`    | `@str`  | Version string; enables `--version` / `-v`                         |
| `summary`    | `@str`  | One-line description shown in help                                 |
| `strict`     | `@bool` | Allow unrecognized parameters, true to treat as an error           |
| `alias`      | `@map`  | Short name → canonical name; e.g. `{ u: "user" }`                  |
| `required`   | `@list` | Canonical option names that must be present                        |
| `defaults`   | `@map`  | Default values by canonical name                                   |
| `types`      | `@map`  | Type hints: `"bool"`, `"str"`, `"num"`, `"list"` by canonical name |
| `desc`       | `@map`  | Option descriptions by canonical name                              |
| `allow`      | `@list` | Explicit allowlist; inferred from other fields if omitted          |
| `positional` | `@map`  | `{ min, max?, name }` positional argument constraints              |

## Error shapes

Each entry in `errors` is a map with a `type` and `msg` field:

| `type`               | Extra fields                       |
|----------------------|------------------------------------|
| `"missing-required"` | `option`                           |
| `"unknown-option"`   | `option`                           |
| `"invalid-type"`     | `option`, `expected`, `value`      |
| `"positional-min"`   | `min`, `actual`                    |
| `"positional-max"`   | `max`, `actual`                    |

### TOC

- [`argsWith(raw, spec)`](#argswithraw-spec)
- [`helpText(spec)`](#helptextspec)
- [`opt(args)`](#optargs)
- [`versionText(spec)`](#versiontextspec)

### Functions

#### `argsWith(raw, spec)`
```slug
fn slug.cli#argsWith(raw, spec) -> ?
```


parses, normalizes, validates, and coerces CLI arguments against a spec.

Receives the map returned by `argm()` and returns a structured result.
Never prints or exits — the caller decides how to handle each outcome.

`--help` / `-h` are handled automatically and always take precedence.
`--version` / `-v` are handled automatically when `spec.version` is set.

Processing order:
1. Check for `--help` / `-h` → return help action
2. Check for `--version` / `-v` → return version action
3. Resolve aliases to canonical option names
4. Apply defaults for missing options
5. Coerce option values to declared types
6. Validate: unknown options, required options, positional bounds
7. Return success or error result

| Parameter | Type | Default |
| --- | --- | --- |
| `raw` |  | — |
| `spec` |  | — |

---

#### `helpText(spec)`
```slug
fn slug.cli#helpText(spec) -> ?
```


renders the help text for a spec as a string.

Produces the same output that `argsWith` returns under `action: :help`.
Useful for displaying help outside the normal argument parsing flow.

| Parameter | Type | Default |
| --- | --- | --- |
| `spec` |  | — |

---

#### `opt(args)`
```slug
fn slug.cli#opt(...args) -> ?
```


returns the first non-nil value from `args`, coerced to the type of the last argument.

The last argument acts as both the default value and the type hint.
If all arguments are nil, returns nil. Throws `TypeError` if coercion
is not possible.

```slug
val port = opt(args.options.port, args.options.p, 8080)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `args` |  | — |

**Throws:** `@struct(Error{type:TypeError})`

---

#### `versionText(spec)`
```slug
fn slug.cli#versionText(spec) -> ?
```


renders the version text for a spec as a string.

Produces the same output that `argsWith` returns under `action: :version`.

| Parameter | Type | Default |
| --- | --- | --- |
| `spec` |  | — |
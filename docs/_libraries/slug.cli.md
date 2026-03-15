---
title: cli (slug)
---

## slug.cli

slug.cli — command-line option helpers

Utilities for working with parsed CLI options from `argm()`.

The primary use case is normalising option values that may come from
multiple sources (long flag, short flag, environment variable, default)
into a single typed value. `opt` picks the first non-nil value from a
list of candidates and coerces it to the type of the last argument.

## Example

```slug
val { opt } = import("slug.cli")
val args = argm()

// accept --port or -p, default to 8080
val port = opt(args.options.port, args.options.p, 8080)
// => 8080 if neither flag was passed, coerced to @num

// accept --verbose or -v, default to false
val verbose = opt(args.options.verbose, args.options.v, false)
// => false if neither flag was passed, coerced to @bool
```

### Functions

#### `opt(args)`
```slug
fn slug.cli#opt(...args) -> ?
```


returns the first non-nil value from `args`, coerced to the type of the last argument.

The last argument acts as both the default value and the type hint.
If all arguments are nil, returns nil.

Coercion uses the same rules as `toBoolean`, `toNumber`, and `toString`.
Throws `TypeError` if coercion is not possible.

| Parameter | Type | Default |
| --- | --- | --- |
| `args` |  | — |

**Throws:** `@struct(Error{type:TypeError})`
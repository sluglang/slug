---
title: stderr (slug.io)
---

## slug.io.stderr

slug.io.stderr — write to standard error

Provides `print` and `println` equivalents that write to stderr
instead of stdout. Useful for diagnostic output, error messages,
and logging that should not mix with a program's regular output.

## Important: import style

`print` and `println` are builtins that take precedence over imported
names, so `val {*} = import('slug.io.stderr')` will not shadow them.
Always import this module under a name:

```slug
val err = import('slug.io.stderr')

err.print("warning: ")
err.println("retrying after timeout")
```

@effects('io')

### TOC

- [`print(args)`](#printargs)
- [`println(args)`](#printlnargs)

### Functions

#### `print(args)`
```slug
fn slug.io.stderr#print(...args):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `args` |  | — |

---

#### `println(args)`
```slug
fn slug.io.stderr#println(...args):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `args` |  | — |
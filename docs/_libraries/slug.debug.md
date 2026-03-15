---
title: debug (slug)
---

## slug.debug

slug.debug — runtime debugging utilities

Low-level inspection tools for development and troubleshooting.
Not intended for production use.

### TOC

- [`ident(value)`](#identvalue)

### Functions

#### `ident(value)`
```slug
fn slug.debug#ident(value) -> @str
```


returns a stable string identity key for any runtime value.

Useful for distinguishing two values that print identically but are
distinct runtime objects, or for using values as map keys in contexts
where their string representation would collide.

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
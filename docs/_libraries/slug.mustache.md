---
title: mustache (slug)
---

## slug.mustache

slug.mustache
A native Mustache renderer for Slug.

Intended to satisfy the official Mustache spec v1.4.3 (core suites + inheritance).

### Functions

#### `parse(template)`
```slug
fn slug.mustache#parse(@str template) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `template` | @str  | — |

---

#### `render(templateOrAst, data, partials)`
```slug
fn slug.mustache#render(templateOrAst, data, partials = nil) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `templateOrAst` |  | — |
| `data` |  | — |
| `partials` |  | `nil` |

**Throws:** `@struct(Error{type:MustacheError})`

---

#### `renderCached(templateOrAst, data, partials, cache)`
```slug
fn slug.mustache#renderCached(templateOrAst, data, partials = nil, cache = {}) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `templateOrAst` |  | — |
| `data` |  | — |
| `partials` |  | `nil` |
| `cache` |  | `{}` |

**Throws:** `@struct(Error{type:MustacheError})`
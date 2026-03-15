---
title: meta (slug)
---

## slug.meta

slug.meta — runtime reflection and metadata

Provides access to runtime metadata about values, tags, bindings, and
modules. Used by tooling such as `slug.doc.manifest` and `slug.doc.markdown`
to generate documentation from live module state.

## describe shape

`describe(value)` returns a map whose shape depends on the value type:

```
// function
{ type: :fn, docs: @str, tags: @map, details: { params: @list } }

// overloaded function group
{ type: :grp, details: { groups: @list } }

// struct schema
{ type: :struct, docs: @str, tags: @map, details: { fields: @list } }

// other (str, num, sym, map, ...)
{ type: :str | :num | :sym | :map | ... }
```

Each param/field in `details.params` or `details.fields` is a map:
```
{ name: @str, tags: @map, vargs: @bool, default: ? }
```

## Tag format

Tags are stored as a map of tag name → list of values.
Use `getTag(value, "@returns")` to retrieve a tag's values.
A tag with no arguments is stored as `[""]`.

## Binding constants

`BINDING_NAME`, `BINDING_VALUE`, and `BINDING_PARAMS` are index
constants for working with raw binding tuples returned by
`searchScopeTags`.

### Constants

#### `BINDING_NAME`

```slug
num slug.meta#BINDING_NAME
```

#### `BINDING_PARAMS`

```slug
num slug.meta#BINDING_PARAMS
```

#### `BINDING_VALUE`

```slug
num slug.meta#BINDING_VALUE
```

### Functions

#### `describe(value)`
```slug
fn slug.meta#describe(value) -> @map
```


returns a metadata map describing the given value.

See module doc for the full shape. The `type` field indicates the
kind of value. Functions expose their parameter list and doc comment.
Struct schemas expose their field list. Overloaded functions are
represented as a `:grp` with a `groups` list of individual function
descriptors.

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |

---

#### `getTag(value, tag)`
```slug
fn slug.meta#getTag(value, @str tag) -> @list
```


returns the list of arguments for a tag on `value`.

Returns an empty list if the tag is not present.
Returns `[""]` for tags with no arguments (e.g. `@export`).

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `tag` | @str  | — |

---

#### `hasTag(value, tag)`
```slug
fn slug.meta#hasTag(value, @str tag) -> @bool
```


returns `true` if `value` has the given tag annotation.

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `tag` | @str  | — |

---

#### `moduleDocs(module)`
```slug
fn slug.meta#moduleDocs(@str module) -> @str
```


returns the module-level doc comment for the given module name.

Returns the raw comment text with markers stripped, or `nil` if the
module has no doc comment. The first paragraph is suitable for use
as a short description; subsequent paragraphs provide extended docs.

| Parameter | Type | Default |
| --- | --- | --- |
| `module` | @str  | — |

---

#### `searchModuleTags(module, tag, includePrivate)`
```slug
fn slug.meta#searchModuleTags(@str module, @str tag, @bool includePrivate = false) -> @map
```


searches all exported bindings in a module for a given tag.

Returns a map of binding name → binding value for all exported
symbols that have the specified tag. Pass `includePrivate: true`
to include non-exported bindings.

| Parameter | Type | Default |
| --- | --- | --- |
| `module` | @str  | — |
| `tag` | @str  | — |
| `includePrivate` | @bool  | `false` |

---

#### `searchScopeTags(tag)`
```slug
fn slug.meta#searchScopeTags(@str tag) -> ?
```


searches all bindings in the current scope for a given tag.

Returns a list of binding tuples. Use `BINDING_NAME`, `BINDING_VALUE`,
and `BINDING_PARAMS` as indices to access tuple elements.

| Parameter | Type | Default |
| --- | --- | --- |
| `tag` | @str  | — |
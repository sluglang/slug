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

### TOC

- [BINDING_NAME](#binding_name)
- [BINDING_PARAMS](#binding_params)
- [BINDING_VALUE](#binding_value)
- [`describe(value)`](#describevalue)
- [`getTag(value, tag)`](#gettagvalue-tag)
- [`hasTag(value, tag)`](#hastagvalue-tag)
- [`moduleDocs(module)`](#moduledocsmodule)
- [`searchModuleTags(module, tag, includePrivate)`](#searchmoduletagsmodule-tag-includeprivate)
- [`searchScopeTags(tag)`](#searchscopetagstag)

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
fn slug.meta#describe(value):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |

---

#### `getTag(value, tag)`
```slug
fn slug.meta#getTag(value, tag:str):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `tag` | str | — |

---

#### `hasTag(value, tag)`
```slug
fn slug.meta#hasTag(value, tag:str):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `tag` | str | — |

---

#### `moduleDocs(module)`
```slug
fn slug.meta#moduleDocs(module:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `module` | str | — |

---

#### `searchModuleTags(module, tag, includePrivate)`
```slug
fn slug.meta#searchModuleTags(module:str, tag:str, includePrivate:bool = false):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `module` | str | — |
| `tag` | str | — |
| `includePrivate` | bool | `false` |

---

#### `searchScopeTags(tag)`
```slug
fn slug.meta#searchScopeTags(tag:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `tag` | str | — |
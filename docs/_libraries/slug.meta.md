---
title: meta (slug)
---

## slug.meta

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

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |

---

#### `getTag(value, tag)`
```slug
fn slug.meta#getTag(value, @str tag) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `tag` | @str  | — |

---

#### `hasTag(value, tag)`
```slug
fn slug.meta#hasTag(value, @str tag) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `tag` | @str  | — |

---

#### `moduleDocs(module)`
```slug
fn slug.meta#moduleDocs(@str module) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `module` | @str  | — |

---

#### `searchModuleTags(module, tag, includePrivate)`
```slug
fn slug.meta#searchModuleTags(@str module, @str tag, @bool includePrivate = false) -> @map
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `tag` | @str  | — |
---
title: path (slug)
---

## slug.path

slug.path - path utilities

Small helpers for path composition and normalization.

### TOC

- [`abs(path)`](#abspath)
- [`cwd()`](#cwd)
- [`join(parts)`](#joinparts)
- [`libRoot()`](#libroot)
- [`localize(path)`](#localizepath)
- [`moduleDir()`](#moduledir)
- [`projectRoot()`](#projectroot)

### Functions

#### `abs(path)`
```slug
fn slug.path#abs(path:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `cwd()`
```slug
fn slug.path#cwd():str
```

---

#### `join(parts)`
```slug
fn slug.path#join(...parts):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `parts` |  | — |

---

#### `libRoot()`
```slug
fn slug.path#libRoot():str
```

---

#### `localize(path)`
```slug
fn slug.path#localize(path:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `moduleDir()`
```slug
fn slug.path#moduleDir():str
```

---

#### `projectRoot()`
```slug
fn slug.path#projectRoot():str
```
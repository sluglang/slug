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


Returns a normalized absolute path.

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `cwd()`
```slug
fn slug.path#cwd():str
```


Returns the current process working directory.

---

#### `join(parts)`
```slug
fn slug.path#join(...parts):str
```


Joins path elements using the platform-specific separator.

| Parameter | Type | Default |
| --- | --- | --- |
| `parts` |  | — |

---

#### `libRoot()`
```slug
fn slug.path#libRoot():str
```


Returns the library root of the current module namespace.

---

#### `localize(path)`
```slug
fn slug.path#localize(path:str|nil):str|nil
```


If path starts with '__project__/', '__library__/' or '__module__/' returns path new path with substitution,
useful for config.

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str \| nil | — |

---

#### `moduleDir()`
```slug
fn slug.path#moduleDir():str
```


Returns the directory of the current module file.

---

#### `projectRoot()`
```slug
fn slug.path#projectRoot():str
```


Returns the entry module directory.
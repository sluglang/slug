---
title: repo (slug.db)
---

## slug.db.repo

file-based SQL query loader with named parameter extraction and schema mapping.

### Functions

#### `into(nil)`
```slug
fn slug.db.repo#into(nil) -> @list
```


maps a list of result rows onto a struct or map; normalises snake_case keys to camelCase by
default; accepts custom key transform fn
nil

---

#### `loadQueries(base)`
```slug
fn slug.db.repo#loadQueries(@str base = DefaultBase) -> @map
```


scans base directory recursively for .sql files; builds a nested map mirroring folder structure;
each leaf is a fn(conn, args) ready to call

| Parameter | Type | Default |
| --- | --- | --- |
| `base` | @str  | `DefaultBase` |
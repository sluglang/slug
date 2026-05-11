---
title: repo (slug.db)
---

## slug.db.repo

slug.db.repo — file-based SQL query loader with named parameter extraction and schema mapping

Scans a directory of `.sql` files and builds a nested map of ready-to-call
query functions that mirrors the folder structure. Named parameters in SQL
(`:param`) are extracted and rewritten to `?` placeholders automatically.

## Quick start

```slug
val { loadQueries, into } = import("slug.db.repo")
val { connect, SQLITE_DRIVER } = import("slug.io.db")

val conn  = connect("file:app.db", SQLITE_DRIVER)
val repos = loadQueries()

// db/queries/users/byAge.sql: SELECT * FROM users WHERE age > :age
val users = conn /> repos.users.byAge({ age: 25 }) /> into(User)
```

## SQL file layout

Files are discovered recursively under the base directory. The path
relative to the base becomes the nested map key:
```
db/queries/users/byAge.sql  =>  repos.users.byAge(conn, args)
db/queries/posts/recent.sql =>  repos.posts.recent(conn, args)
```

## Named parameters

Use `:paramName` in SQL to define named parameters. They are rewritten
to `?` placeholders and extracted in order. Parameters inside string
literals and comments are safely ignored.

Pass args as a map `{ age: 25 }`, a list `[25]`, or a struct.

## Query type detection

`SELECT`/`WITH` statements call `query()` (returns `@list`).
All other statements call `exec()` (returns `@map`).

@effects('fs db')

### TOC

- [`into(data, toType, columnToFieldName)`](#intodata-totype-columntofieldname)
- [`loadQueries(base)`](#loadqueriesbase)

### Functions

#### `into(data, toType, columnToFieldName)`
```slug
fn slug.db.repo#into(data:list, toType, columnToFieldName:fn = fn((s)) {camelCase(s, _)}):list
fn slug.db.repo#into(data:map, target, columnToFieldName:fn = fn((s)) {camelCase(s, _)}):list
```


maps a list of db result rows onto a struct or map target.

Delegates each row to the `@map` variant. Column names are normalised
from `snake_case` to `camelCase` by default.

| Parameter | Type | Default |
| --- | --- | --- |
| `data` | list | — |
| `toType` |  | — |
| `columnToFieldName` | fn | `fn((s)) {camelCase(s, _)}` |

---

#### `loadQueries(base)`
```slug
fn slug.db.repo#loadQueries(base:str = DefaultBase):map
```


scans `base` recursively for `.sql` files and returns a nested map of query functions.

Each leaf is a `fn(conn, args)` that executes the query against `conn`.
The map structure mirrors the directory structure under `base`.
The default base directory is `db/queries` or the `base-directory` config key.

| Parameter | Type | Default |
| --- | --- | --- |
| `base` | str | `DefaultBase` |
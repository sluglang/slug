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

- [`into(nil)`](#intonil)
- [`loadQueries(base)`](#loadqueriesbase)

### Functions

#### `into(nil)`
```slug
fn slug.db.repo#into(nil):list
```
nil

---

#### `loadQueries(base)`
```slug
fn slug.db.repo#loadQueries(base:str = DefaultBase):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `base` | str | `DefaultBase` |
---
title: db (slug.io)
---

## slug.io.db

### Constants

#### `MYSQL_DRIVER`

```slug
str slug.io.db#MYSQL_DRIVER
```

var con = "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True"
  /> connect(MYSQL_DRIVER)

#### `PGSQL_DRIVER`

```slug
str slug.io.db#PGSQL_DRIVER
```

var con = "postgres://user:password@localhost/dbname?sslmode=disable"
  /> connect(PGSQL_DRIVER)

#### `SQLITE_DRIVER`

```slug
str slug.io.db#SQLITE_DRIVER
```

var con = ":memory:"
  /> connect(SQLITE_DRIVER)

open a file with optional parameters (e.g. `?cache=shared&mode=rwc`)

var con = "file:filename.db"
  /> connect(SQLITE_DRIVER)

### Functions

#### `begin(connection)`
```slug
fn slug.io.db#begin(@num connection) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | @num  | — |

---

#### `close(connection)`
```slug
fn slug.io.db#close(@num connection) -> nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | @num  | — |

---

#### `commit(connection)`
```slug
fn slug.io.db#commit(@num connection) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | @num  | — |

---

#### `connect(connectionString, driver)`
```slug
fn slug.io.db#connect(@str connectionString, @str driver) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connectionString` | @str  | — |
| `driver` | @str  | — |

---

#### `exec(connection, sql, params)`
```slug
fn slug.io.db#exec(@num connection, @str sql, ...params) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | @num  | — |
| `sql` | @str  | — |
| `params` |  | — |

---

#### `query(connection, sql, params)`
```slug
fn slug.io.db#query(@num connection, @str sql, ...params) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | @num  | — |
| `sql` | @str  | — |
| `params` |  | — |

---

#### `rollback(connection)`
```slug
fn slug.io.db#rollback(@num connection) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | @num  | — |
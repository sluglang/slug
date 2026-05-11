---
title: db (slug.io)
---

## slug.io.db

slug.io.db — relational database access

Connect to SQLite, MySQL, or PostgreSQL databases and execute queries
and statements. Connection handles are numeric values returned by `connect`
and passed to all subsequent operations.

Always close connections with `defer close(conn)` or explicit `close(conn)`.

## SQLite examples

```slug
val { connect, query, exec, close, SQLITE_DRIVER } = import("slug.io.db")

// in-memory database
val conn = connect(":memory:", SQLITE_DRIVER)
defer close(conn)

// file database with options
val conn = connect("file:app.db?cache=shared&mode=rwc", SQLITE_DRIVER)
defer close(conn)
```

## MySQL example

```slug
val conn = connect("user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True", MYSQL_DRIVER)
defer close(conn)
```

## PostgreSQL example

```slug
val conn = connect("postgres://user:password@localhost/dbname?sslmode=disable", PGSQL_DRIVER)
defer close(conn)
```

## Transactions

```slug
val tx = begin(conn)
exec(tx, "INSERT INTO ...")
commit(tx)
```

@effects('db')

### TOC

- [MYSQL_DRIVER](#mysql_driver)
- [PGSQL_DRIVER](#pgsql_driver)
- [SQLITE_DRIVER](#sqlite_driver)
- [`begin(connection)`](#beginconnection)
- [`close(connection)`](#closeconnection)
- [`commit(connection)`](#commitconnection)
- [`connect(connectionString, driver)`](#connectconnectionstring-driver)
- [`exec(connection, sql, params)`](#execconnection-sql-params)
- [`query(connection, sql, params)`](#queryconnection-sql-params)
- [`rollback(connection)`](#rollbackconnection)

### Constants

#### `MYSQL_DRIVER`

```slug
str slug.io.db#MYSQL_DRIVER
```

#### `PGSQL_DRIVER`

```slug
str slug.io.db#PGSQL_DRIVER
```

#### `SQLITE_DRIVER`

```slug
str slug.io.db#SQLITE_DRIVER
```

### Functions

#### `begin(connection)`
```slug
fn slug.io.db#begin(connection:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |

---

#### `close(connection)`
```slug
fn slug.io.db#close(connection:num):nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |

---

#### `commit(connection)`
```slug
fn slug.io.db#commit(connection:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |

---

#### `connect(connectionString, driver)`
```slug
fn slug.io.db#connect(connectionString:str, driver:str):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connectionString` | str | — |
| `driver` | str | — |

---

#### `exec(connection, sql, params)`
```slug
fn slug.io.db#exec(connection:num, sql:str, ...params):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |
| `sql` | str | — |
| `params` |  | — |

---

#### `query(connection, sql, params)`
```slug
fn slug.io.db#query(connection:num, sql:str, ...params):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |
| `sql` | str | — |
| `params` |  | — |

---

#### `rollback(connection)`
```slug
fn slug.io.db#rollback(connection:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |
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

driver string for MySQL connections.

#### `PGSQL_DRIVER`

```slug
str slug.io.db#PGSQL_DRIVER
```

driver string for PostgreSQL connections.

#### `SQLITE_DRIVER`

```slug
str slug.io.db#SQLITE_DRIVER
```

driver string for SQLite3 connections.

### Functions

#### `begin(connection)`
```slug
fn slug.io.db#begin(connection:num):num
```


begins a transaction and returns a transaction handle.

Pass the transaction handle to `query`, `exec`, `commit`, or `rollback`
in place of a connection handle.

@effects('db')

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |

---

#### `close(connection)`
```slug
fn slug.io.db#close(connection:num):nil
```


closes a database connection or transaction handle.

Always close connections when done. Use `defer close(conn)` for safety.

@effects('db')

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |

---

#### `commit(connection)`
```slug
fn slug.io.db#commit(connection:num):num
```


commits a transaction.

@effects('db')

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |

---

#### `connect(connectionString, driver)`
```slug
fn slug.io.db#connect(connectionString:str, driver:str):num
```


opens a database connection and returns a connection handle.

`connectionString` format depends on the driver — see module doc for examples.
Use `defer close(conn)` to ensure the connection is released.

@effects('db')

| Parameter | Type | Default |
| --- | --- | --- |
| `connectionString` | str | — |
| `driver` | str | — |

---

#### `exec(connection, sql, params)`
```slug
fn slug.io.db#exec(connection:num, sql:str, ...params):map
```


executes a SQL statement and returns a result map.

The result map contains `rowsAffected` and `lastInsertId` where supported
by the driver. Use `...params` to pass positional `?` parameter values.

@effects('db')

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


executes a SQL query and returns the result rows as a list of string-keyed maps.

Use `...params` to pass positional `?` parameter values.
For named parameters, use `slug.db.repo`.

@effects('db')

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


rolls back a transaction.

@effects('db')

| Parameter | Type | Default |
| --- | --- | --- |
| `connection` | num | — |
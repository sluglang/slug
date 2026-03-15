---
title: tcp (slug.io)
---

## slug.io.tcp

slug.io.tcp — low-level TCP networking

Raw TCP socket operations for building servers and clients. Connection
and listener handles are numeric values — always close them when done.

## TCP server example

```slug
val { bind, accept, read, write, close } = import("slug.io.tcp")

val listener = bind("0.0.0.0", 8080)
defer close(listener)

nursery {
  spawn {
    val conn = accept(listener)
    defer close(conn)
    val data = read(conn, 4096)
    write(conn, "HTTP/1.1 200 OK\r\n\r\nHello!")
  }
}
```

## TCP client example

```slug
val conn = connect("example.com", 80)
defer close(conn)
write(conn, "GET / HTTP/1.0\r\n\r\n")
val response = read(conn, 4096)
```

For HTTP use cases, prefer `slug.io.http`. For web servers, prefer
`slug.web.server`.

@effects('io net')

### Functions

#### `accept(listener)`
```slug
fn slug.io.tcp#accept(@num listener) -> @num
```


waits for an incoming connection on `listener` and returns a connection handle.

Blocks until a client connects. Use in a `spawn` or `nursery` block
for concurrent connection handling.

@effects('io net')

| Parameter | Type | Default |
| --- | --- | --- |
| `listener` | @num  | — |

---

#### `bind(addr, port)`
```slug
fn slug.io.tcp#bind(@str addr, @num port) -> @num
```


binds a TCP listener to `addr:port` and returns a listener handle.

Use `accept` to wait for incoming connections.
Use `defer close(listener)` to release the port when done.

@effects('io net')

| Parameter | Type | Default |
| --- | --- | --- |
| `addr` | @str  | — |
| `port` | @num  | — |

---

#### `close(handle)`
```slug
fn slug.io.tcp#close(@num handle) -> @num
```


closes a TCP listener or connection handle.

@effects('io net')

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |

---

#### `connect(addr, port)`
```slug
fn slug.io.tcp#connect(@str addr, @num port) -> @num
```


connects to a TCP server at `addr:port` and returns a connection handle.

Use `defer close(conn)` to release the connection.

@effects('io net')

| Parameter | Type | Default |
| --- | --- | --- |
| `addr` | @str  | — |
| `port` | @num  | — |

---

#### `read(stream, maxBytes)`
```slug
fn slug.io.tcp#read(@num stream, @num maxBytes) -> @str
```


reads up to `maxBytes` from a TCP stream.

Returns the data read as a string, or `nil` on EOF.

@effects('io net')

| Parameter | Type | Default |
| --- | --- | --- |
| `stream` | @num  | — |
| `maxBytes` | @num  | — |

---

#### `write(stream, data)`
```slug
fn slug.io.tcp#write(@num stream, @str data) -> @num
```


sends `data` over a TCP stream.

Returns the number of bytes written.

@effects('io net')

| Parameter | Type | Default |
| --- | --- | --- |
| `stream` | @num  | — |
| `data` | @str  | — |
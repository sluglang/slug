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

### TOC

- [`accept(listener)`](#acceptlistener)
- [`bind(addr, port)`](#bindaddr-port)
- [`close(handle)`](#closehandle)
- [`connect(addr, port)`](#connectaddr-port)
- [`read(stream, maxBytes)`](#readstream-maxbytes)
- [`write(stream, data)`](#writestream-data)

### Functions

#### `accept(listener)`
```slug
fn slug.io.tcp#accept(listener:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `listener` | num | — |

---

#### `bind(addr, port)`
```slug
fn slug.io.tcp#bind(addr:str, port:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `addr` | str | — |
| `port` | num | — |

---

#### `close(handle)`
```slug
fn slug.io.tcp#close(handle:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |

---

#### `connect(addr, port)`
```slug
fn slug.io.tcp#connect(addr:str, port:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `addr` | str | — |
| `port` | num | — |

---

#### `read(stream, maxBytes)`
```slug
fn slug.io.tcp#read(stream:num, maxBytes:num):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `stream` | num | — |
| `maxBytes` | num | — |

---

#### `write(stream, data)`
```slug
fn slug.io.tcp#write(stream:num, data:str):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `stream` | num | — |
| `data` | str | — |
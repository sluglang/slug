---
title: server (slug.web)
---

## slug.web.server

slug.web.server — HTTP/1.1 server

A pure-Slug HTTP/1.1 server built on `slug.io.tcp` and structured
concurrency. Each connection is handled in a spawned task within a
nursery with a connection limit.

## Quick start

```slug
val { serve }  = import("slug.web.server")
val { router, get, withLog } = import("slug.web.routes")
val { html } = import("slug.web.response")

val r = router()
/> get("/", fn(req) { html("<h1>Hello</h1>") })

serve(r /> withLog /> handle)
```

## Configuration

`addr` and `port` default to `cfg("address", "0.0.0.0")` and
`cfg("port", 8080)`. Override via config or pass directly:

```slug
serve(app, "127.0.0.1", 3000)
```

## Connection handling

- Up to 1000 concurrent connections (nursery limit)
- Up to 100 requests per connection (keep-alive)
- 10 second idle timeout per connection
- Request bodies up to 2MB are read into memory as strings
- Larger bodies throw `ServerError` — use `withMaxBody` middleware

## App function contract

`app` must be a `fn(request) -> Response`. Use `slug.web.routes#handle`
to dispatch to a router, or write a plain function for simple cases.

@effects('io net')

### Functions

#### `serve(app, addr, port)`
```slug
fn slug.web.server#serve(app:fn, addr:str = cfg(address, 0.0.0.0), port:num = cfg(port, 8080)):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `app` | fn | — |
| `addr` | str | `cfg(address, 0.0.0.0)` |
| `port` | num | `cfg(port, 8080)` |

**Throws:** `Error{type:ServerError}`
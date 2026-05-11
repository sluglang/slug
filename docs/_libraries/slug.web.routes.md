---
title: routes (slug.web)
---

## slug.web.routes

slug.web.routes — HTTP router and middleware

Pattern-based HTTP routing with composable middleware wrappers.
Routes are registered on a `Router` and dispatched via `handle`.

## Quick start

```slug
val { router, get, post, handle, withLog, withRecover } = import("slug.web.routes")
val { html, jsonOk, notFound } = import("slug.web.response")

val r = router()
/> get("/",           fn(req) { html("<h1>Hello</h1>") })
/> get("/users/:id",  fn(req) { jsonOk({ id: req.params["id"] }) })
/> post("/users",     fn(req) { jsonOk(req.body) })

val app = r /> handle /> withLog /> withRecover
```

## Route patterns

- Exact match: `"/users"` matches only `/users`
- Named param: `"/users/:id"` captures the segment as `req.params["id"]`
- Wildcard mount: `"/static/*"` matches any path starting with `/static/`
  and captures the remainder as `req.params["*"]`

## Middleware

Middleware wraps a handler function `fn(req) -> Response` and returns
a new handler with the same signature. Chain middleware with `/>`:

```slug
val app = fn(req) { handle(req, r) }
/> withLog
/> withRecover
/> withMaxBody(512_000)
/> withRequestId
```

## Subrouters

Mount a child router at a prefix with `mountRouter`. The child router
sees paths relative to the mount point:

```slug
val api = router()
/> get("/users",    usersHandler)
/> get("/users/:id", userHandler)

val r = router()
/> mountRouter("/api", api)
```

## Static files

```slug
val r = router()
/> mount("/static", static("public"))
```

### TOC

- [`get(nil)`](#getnil)
- [`handle(request, r)`](#handlerequest-r)
- [`head(r, pattern, handler)`](#headr-pattern-handler)
- [`isRouter(x)`](#isrouterx)
- [`mount(r, prefix, handler)`](#mountr-prefix-handler)
- [`mountRouter(r, prefix, childRouter)`](#mountrouterr-prefix-childrouter)
- [`post(r, pattern, handler)`](#postr-pattern-handler)
- [`router()`](#router)
- [`static(dir, cacheTimeSeconds)`](#staticdir-cachetimeseconds)
- [`subrouter(router)`](#subrouterrouter)
- [`withHeader(handler, header, value)`](#withheaderhandler-header-value)
- [`withLog(handler)`](#withloghandler)
- [`withMaxBody(h, maxBytes)`](#withmaxbodyh-maxbytes)
- [`withRecover(h)`](#withrecoverh)
- [`withRequestId(handler, newRequestId)`](#withrequestidhandler-newrequestid)
- [`withTimeout(h, ms)`](#withtimeouth-ms)
- [`withTraceContext(handler, newTraceId, newSpanId)`](#withtracecontexthandler-newtraceid-newspanid)

### Functions

#### `get(nil)`
```slug
fn slug.web.routes#get(nil):Router
```
nil


#### Examples

```slug
get({}, :k)  // => nil
get({:k: 1}, :k)  // => 1
```

---

#### `handle(request, r)`
```slug
fn slug.web.routes#handle(request, r):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `request` |  | — |
| `r` |  | — |

---

#### `head(r, pattern, handler)`
```slug
fn slug.web.routes#head(r, pattern:str, handler:fn):Router
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` |  | — |
| `pattern` | str | — |
| `handler` | fn | — |

---

#### `isRouter(x)`
```slug
fn slug.web.routes#isRouter(x):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `mount(r, prefix, handler)`
```slug
fn slug.web.routes#mount(r, prefix:str, handler:fn):Router
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` |  | — |
| `prefix` | str | — |
| `handler` | fn | — |

---

#### `mountRouter(r, prefix, childRouter)`
```slug
fn slug.web.routes#mountRouter(r, prefix:str, childRouter):Router
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` |  | — |
| `prefix` | str | — |
| `childRouter` |  | — |

---

#### `post(r, pattern, handler)`
```slug
fn slug.web.routes#post(r, pattern:str, handler:fn):Router
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` |  | — |
| `pattern` | str | — |
| `handler` | fn | — |

---

#### `router()`
```slug
fn slug.web.routes#router():Router
```

---

#### `static(dir, cacheTimeSeconds)`
```slug
fn slug.web.routes#static(dir, cacheTimeSeconds = 3600):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `dir` |  | — |
| `cacheTimeSeconds` |  | `3600` |

---

#### `subrouter(router)`
```slug
fn slug.web.routes#subrouter(router):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `router` |  | — |

---

#### `withHeader(handler, header, value)`
```slug
fn slug.web.routes#withHeader(handler:fn, header:str, value:str):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` | fn | — |
| `header` | str | — |
| `value` | str | — |

---

#### `withLog(handler)`
```slug
fn slug.web.routes#withLog(handler):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` |  | — |

---

#### `withMaxBody(h, maxBytes)`
```slug
fn slug.web.routes#withMaxBody(h, maxBytes = 1048576):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |
| `maxBytes` |  | `1048576` |

---

#### `withRecover(h)`
```slug
fn slug.web.routes#withRecover(h):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |

---

#### `withRequestId(handler, newRequestId)`
```slug
fn slug.web.routes#withRequestId(handler:fn, newRequestId:fn = fn() {randomHexString(32)}):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` | fn | — |
| `newRequestId` | fn | `fn() {randomHexString(32)}` |

---

#### `withTimeout(h, ms)`
```slug
fn slug.web.routes#withTimeout(h, ms):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |
| `ms` |  | — |

---

#### `withTraceContext(handler, newTraceId, newSpanId)`
```slug
fn slug.web.routes#withTraceContext(handler, newTraceId:fn = fn() {randomHexString(32)}, newSpanId:fn = fn() {randomHexString(16)}):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` |  | — |
| `newTraceId` | fn | `fn() {randomHexString(32)}` |
| `newSpanId` | fn | `fn() {randomHexString(16)}` |
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

- [`get(map, key)`](#getmap-key)
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

#### `get(map, key)`
```slug
fn slug.web.routes#get(map:map, key):Router
fn slug.web.routes#get(r, pattern:str, handler:fn):Router
```


registers a `GET` route on `r` and returns the updated router.

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | map | — |
| `key` |  | — |


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


dispatches `request` through the router's routes and returns a `Response`.

Returns a 404 Not Found response if no route matches.
Route handlers receive the request with `params` populated from URL pattern segments.

| Parameter | Type | Default |
| --- | --- | --- |
| `request` |  | — |
| `r` |  | — |

---

#### `head(r, pattern, handler)`
```slug
fn slug.web.routes#head(r, pattern:str, handler:fn):Router
```


registers a `HEAD` route on `r` and returns the updated router.

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


returns true if `x` is a `Router` struct instance.

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `mount(r, prefix, handler)`
```slug
fn slug.web.routes#mount(r, prefix:str, handler:fn):Router
```


mounts `handler` at `prefix/*`, matching any method.

The path remainder after the prefix is available as `req.params["*"]`.

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


mounts a child `Router` at `prefix`, handling paths relative to the prefix.

Equivalent to `mount(r, prefix, subrouter(childRouter))`.

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


registers a `POST` route on `r` and returns the updated router.

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


creates an empty `Router`.

---

#### `static(dir, cacheTimeSeconds)`
```slug
fn slug.web.routes#static(dir, cacheTimeSeconds = 3600):fn
```


returns a handler that serves static files from `dir`.

Intended for use with `mount`: `mount("/static", static("public"))`.
Files are served with the appropriate `Content-Type` based on extension
and cached for `cacheTimeSeconds` seconds (default 1 hour).

Returns 404 for missing files, path traversal attempts (`..`), and
non-GET/HEAD requests. `HEAD` requests return headers without a body.

| Parameter | Type | Default |
| --- | --- | --- |
| `dir` |  | — |
| `cacheTimeSeconds` |  | `3600` |

---

#### `subrouter(router)`
```slug
fn slug.web.routes#subrouter(router):fn
```


wraps a `Router` as a handler function for use with `mount`.

Strips the mount prefix from the path and clears the `"*"` param
before dispatching to the inner router.

| Parameter | Type | Default |
| --- | --- | --- |
| `router` |  | — |

---

#### `withHeader(handler, header, value)`
```slug
fn slug.web.routes#withHeader(handler:fn, header:str, value:str):fn
fn slug.web.routes#withHeader(res, key:str, value:str):fn
```


wraps `handler` to add a fixed response header to every response.

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


wraps `handler` with request/response logging to stdout.

Logs timestamp, method, elapsed time (ms), status code, path, and
any request ID / trace ID / span ID present on the request.

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` |  | — |

---

#### `withMaxBody(h, maxBytes)`
```slug
fn slug.web.routes#withMaxBody(h, maxBytes = 1048576):fn
```


wraps `handler` to enforce a maximum request body size.

Checks `Content-Length` for `POST`, `PUT`, and `PATCH` requests.
Returns 413 Payload Too Large if the declared size exceeds `maxBytes`.
Default limit is 1MB (1_048_576 bytes).

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |
| `maxBytes` |  | `1048576` |

---

#### `withRecover(h)`
```slug
fn slug.web.routes#withRecover(h):fn
```


wraps `handler` so that any thrown error returns a 500 Server Error response.

Use as the outermost middleware layer to prevent unhandled errors from
crashing the connection.

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |

---

#### `withRequestId(handler, newRequestId)`
```slug
fn slug.web.routes#withRequestId(handler:fn, newRequestId:fn = fn() {randomHexString(32)}):fn
```


wraps `handler` to attach a request ID to every request and response.

Uses the incoming `X-Request-Id` header if present, otherwise generates
a new random hex ID. The ID is attached to `req.requestId` and returned
as the `X-Request-Id` response header.

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` | fn | — |
| `newRequestId` | fn | `fn() {randomHexString(32)}` |

---

#### `withTimeout(h, ms)`
```slug
fn slug.web.routes#withTimeout(h, ms):fn
```


wraps `handler` with a timeout of `ms` milliseconds.

If the handler does not return within the timeout, a `TimeoutError` is thrown.

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |
| `ms` |  | — |

---

#### `withTraceContext(handler, newTraceId, newSpanId)`
```slug
fn slug.web.routes#withTraceContext(handler, newTraceId:fn = fn() {randomHexString(32)}, newSpanId:fn = fn() {randomHexString(16)}):fn
```


wraps `handler` to propagate W3C `traceparent` trace context.

Extracts or generates a `traceId` and creates a new `spanId` for each
request. Attaches both to `req.traceId` and `req.spanId`. Returns the
`traceparent` header on the response for downstream propagation.

Format: `00-<traceId>-<spanId>-01`

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` |  | — |
| `newTraceId` | fn | `fn() {randomHexString(32)}` |
| `newSpanId` | fn | `fn() {randomHexString(16)}` |
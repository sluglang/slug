---
title: routes (slug.web)
---

## slug.web.routes

### Functions

#### `get(nil)`
```slug
fn slug.web.routes#get(nil) -> @struct(Router)
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
fn slug.web.routes#handle(request, r) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `request` |  | — |
| `r` |  | — |

---

#### `head(r, pattern, handler)`
```slug
fn slug.web.routes#head(r, @str pattern, @fn handler) -> @struct(Router)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` |  | — |
| `pattern` | @str  | — |
| `handler` | @fn  | — |

---

#### `isRouter(x)`
```slug
fn slug.web.routes#isRouter(x) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `mount(r, prefix, handler)`
```slug
fn slug.web.routes#mount(r, @str prefix, @fn handler) -> @struct(Router)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` |  | — |
| `prefix` | @str  | — |
| `handler` | @fn  | — |

---

#### `mountRouter(r, prefix, childRouter)`
```slug
fn slug.web.routes#mountRouter(r, @str prefix, childRouter) -> @struct(Router)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` |  | — |
| `prefix` | @str  | — |
| `childRouter` |  | — |

---

#### `post(r, pattern, handler)`
```slug
fn slug.web.routes#post(r, @str pattern, @fn handler) -> @struct(Router)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` |  | — |
| `pattern` | @str  | — |
| `handler` | @fn  | — |

---

#### `router()`
```slug
fn slug.web.routes#router() -> @struct(Router)
```

---

#### `static(dir, cacheTimeSeconds)`
```slug
fn slug.web.routes#static(dir, cacheTimeSeconds = 3600) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `dir` |  | — |
| `cacheTimeSeconds` |  | `3600` |

---

#### `subrouter(router)`
```slug
fn slug.web.routes#subrouter(router) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `router` |  | — |

---

#### `withHeader(nil)`
```slug
fn slug.web.routes#withHeader(nil) -> @fn
```
nil

---

#### `withLog(handler)`
```slug
fn slug.web.routes#withLog(handler) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` |  | — |

---

#### `withMaxBody(h, maxBytes)`
```slug
fn slug.web.routes#withMaxBody(h, maxBytes = 1048576) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |
| `maxBytes` |  | `1048576` |

---

#### `withRecover(h)`
```slug
fn slug.web.routes#withRecover(h) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |

---

#### `withRequestId(handler, newRequestId)`
```slug
fn slug.web.routes#withRequestId(@fn handler, @fn newRequestId = requestId) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` | @fn  | — |
| `newRequestId` | @fn  | `requestId` |

---

#### `withTimeout(h, ms)`
```slug
fn slug.web.routes#withTimeout(h, ms) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `h` |  | — |
| `ms` |  | — |

---

#### `withTraceContext(handler, newTraceId, newSpanId)`
```slug
fn slug.web.routes#withTraceContext(handler, @fn newTraceId = traceId, @fn newSpanId = spanId) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handler` |  | — |
| `newTraceId` | @fn  | `traceId` |
| `newSpanId` | @fn  | `spanId` |
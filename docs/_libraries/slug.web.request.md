---
title: request (slug.web)
---

## slug.web.request

### Functions

#### `isRequest(x)`
```slug
fn slug.web.request#isRequest(x) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `parseRequestHeaders(buf)`
```slug
fn slug.web.request#parseRequestHeaders(buf) -> [@struct(Request), @str]
```

| Parameter | Type | Default |
| --- | --- | --- |
| `buf` |  | — |

**Throws:** `@struct(Error{type:RequestError})`

---

#### `request(method, path, version, headers, body)`
```slug
fn slug.web.request#request(method, path, version = "HTTP/1.1", headers = {}, body = "") -> @struct(Request)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `method` |  | — |
| `path` |  | — |
| `version` |  | `"HTTP/1.1"` |
| `headers` |  | `{}` |
| `body` |  | `""` |

---

#### `shouldKeepAlive(req)`
```slug
fn slug.web.request#shouldKeepAlive(req) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `req` |  | — |

---

#### `withBody(req, body)`
```slug
fn slug.web.request#withBody(req, body) -> @struct(Request)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `req` |  | — |
| `body` |  | — |

---

#### `withParams(request, params)`
```slug
fn slug.web.request#withParams(request, @map params) -> @struct(Request)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `request` |  | — |
| `params` | @map  | — |

---

#### `withPath(req, path)`
```slug
fn slug.web.request#withPath(req, path) -> @struct(Request)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `req` |  | — |
| `path` |  | — |

---

#### `withQuery(req, query)`
```slug
fn slug.web.request#withQuery(req, query) -> @struct(Request)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `req` |  | — |
| `query` |  | — |

---

#### `withoutParam(request, param)`
```slug
fn slug.web.request#withoutParam(request, param) -> @struct(Request)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `request` |  | — |
| `param` |  | — |
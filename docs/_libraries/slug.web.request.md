---
title: request (slug.web)
---

## slug.web.request

slug.web.request — HTTP request representation and parsing

Defines the `Request` struct and provides parsing, construction, and
accessor functions for working with incoming HTTP requests.

## Request struct fields

| Field       | Type    | Description                                      |
|-------------|---------|--------------------------------------------------|
| `method`    | `@str`  | HTTP method (`"GET"`, `"POST"`, etc.)            |
| `path`      | `@str`  | URL path (without query string)                  |
| `version`   | `@str`  | HTTP version (`"HTTP/1.1"`)                      |
| `headers`   | `@map`  | Lowercase header names → values                  |
| `body`      | any     | Request body (string for small bodies)           |
| `query`     | `@map`  | Parsed query string parameters                   |
| `form`      | `@map`  | Parsed `application/x-www-form-urlencoded` body  |
| `files`     | `@list` | Uploaded files (multipart, when supported)       |
| `params`    | `@map`  | URL path parameters (populated by the router)    |
| `requestId` | any     | Request ID (set by `withRequestId` middleware)   |
| `traceId`   | any     | Trace ID (set by `withTraceContext` middleware)   |
| `spanId`    | any     | Span ID (set by `withTraceContext` middleware)    |

Header names are always normalised to lowercase. Access headers as
`req.headers["content-type"]` not `req.headers["Content-Type"]`.

### TOC

- [Request](#request)
- [`isRequest(x)`](#isrequestx)
- [`parseRequestHeaders(buf)`](#parserequestheadersbuf)
- [`request(method, path, version, headers, body)`](#requestmethod-path-version-headers-body)
- [`shouldKeepAlive(req)`](#shouldkeepalivereq)
- [`withBody(req, body)`](#withbodyreq-body)
- [`withParams(request, params)`](#withparamsrequest-params)
- [`withPath(req, path)`](#withpathreq-path)
- [`withQuery(req, query)`](#withqueryreq-query)
- [`withoutParam(request, param)`](#withoutparamrequest-param)

### Structs

#### `Request`
```slug
struct slug.web.request#Request{method:str, path:str, version:str, headers:map, body, query:map, form:map, files:list, params:map, requestId, traceId, spanId}
```

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `method` | str | — |  |
| `path` | str | — |  |
| `version` | str | — |  |
| `headers` | map | — |  |
| `body` |  | — |  |
| `query` | map | — |  |
| `form` | map | — |  |
| `files` | list | — |  |
| `params` | map | — |  |
| `requestId` |  | — |  |
| `traceId` |  | — |  |
| `spanId` |  | — |  |

### Functions

#### `isRequest(x)`
```slug
fn slug.web.request#isRequest(x):bool
```


returns true if `x` is a `Request` struct instance.

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `parseRequestHeaders(buf)`
```slug
fn slug.web.request#parseRequestHeaders(buf):[Request, str]
```


parses the headers section of a raw HTTP request buffer.

Returns `[Request, remainingBuffer]` where `remainingBuffer` is the
bytes after the `\r\n\r\n` header/body separator. The body is not
read — the caller (the server) is responsible for reading it based
on `content-length` and then calling `withBody`.

Header names are normalised to lowercase. Query string parameters
are parsed and attached to `req.query`.

Throws `RequestError` if headers are incomplete or the request line is malformed.

| Parameter | Type | Default |
| --- | --- | --- |
| `buf` |  | — |

**Throws:** `Error{type:RequestError}`

---

#### `request(method, path, version, headers, body)`
```slug
fn slug.web.request#request(method, path, version = "HTTP/1.1", headers = {}, body = ""):Request
```


constructs a new `Request` with the given method, path, and optional fields.

All other fields default to empty/nil. This is the canonical way to
create a request for testing or manual construction.

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
fn slug.web.request#shouldKeepAlive(req):bool
```


returns true if the connection should be kept alive after this request.

For HTTP/1.0, keep-alive requires an explicit `Connection: keep-alive` header.
For HTTP/1.1, keep-alive is the default unless `Connection: close` is set.

| Parameter | Type | Default |
| --- | --- | --- |
| `req` |  | — |

---

#### `withBody(req, body)`
```slug
fn slug.web.request#withBody(req, body):Request
```


attaches a body to a request and parses form data if applicable.

If the `content-type` is `application/x-www-form-urlencoded`, the body
is parsed and the result is available as `req.form`.

| Parameter | Type | Default |
| --- | --- | --- |
| `req` |  | — |
| `body` |  | — |

---

#### `withParams(request, params)`
```slug
fn slug.web.request#withParams(request, params:map):Request
```


returns a new request with path params map replaced.

Used by the router to inject matched URL parameters (e.g. `:id`).

| Parameter | Type | Default |
| --- | --- | --- |
| `request` |  | — |
| `params` | map | — |

---

#### `withPath(req, path)`
```slug
fn slug.web.request#withPath(req, path):Request
```


returns a new request with `path` replaced.

Used by the router to strip path prefixes for subrouters.

| Parameter | Type | Default |
| --- | --- | --- |
| `req` |  | — |
| `path` |  | — |

---

#### `withQuery(req, query)`
```slug
fn slug.web.request#withQuery(req, query):Request
```


returns a new request with `query` replaced.

| Parameter | Type | Default |
| --- | --- | --- |
| `req` |  | — |
| `query` |  | — |

---

#### `withoutParam(request, param)`
```slug
fn slug.web.request#withoutParam(request, param):Request
```


returns a new request with a single param key removed from `params`.

Used by subrouters to remove the `"*"` wildcard param after consuming it.

| Parameter | Type | Default |
| --- | --- | --- |
| `request` |  | — |
| `param` |  | — |
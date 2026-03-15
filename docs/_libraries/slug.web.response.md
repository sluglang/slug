---
title: response (slug.web)
---

## slug.web.response

slug.web.response — HTTP response construction and formatting

Provides the `Response` and `Reply` structs, constructor helpers for
all common HTTP status codes, header management, cookie support,
HTMX helpers, and HTTP wire-format serialisation.

## Quick start

```slug
val { ok, html, jsonOk, notFound, redirect, withHeader } = import("slug.web.response")

// simple text response
ok("Hello, World!")

// HTML response
html("<h1>Hello</h1>")

// JSON response
jsonOk({ name: "Alice", age: 30 })

// chaining modifiers
ok("done")
/> withHeader("x-custom", "value")
/> noCache
```

## Response struct

`Response{ status, headers, body }` — header names are always lowercase.
Use `withHeader`, `withHeaders`, `addHeader`, and `withoutHeader` to
manage headers rather than constructing the map directly.

## Reply struct

`Reply` is a higher-level descriptor used with the views layer.
It carries a view name, optional HTMX fragment name, template data,
status code, and extra headers. Pass a `Reply` to `renderReply` in
`slug.web.views` to produce a `Response`.

## Cookie options

The `cookie` function accepts an `opts` map with optional keys:
`path`, `domain`, `maxAge`, `expires`, `secure` (@bool), `httpOnly` (@bool),
`sameSite` (`"Lax"` | `"Strict"` | `"None"`).

## HTMX

`hxTrigger`, `hxRedirect`, and `hxRetarget` add the corresponding
HTMX response headers (`HX-Trigger`, `HX-Redirect`, `HX-Retarget`).

### TOC

- [Reply](#reply)
- [Response](#response)
- [`accepted(body)`](#acceptedbody)
- [`addHeader(res, key, value)`](#addheaderres-key-value)
- [`badRequest(body)`](#badrequestbody)
- [`body(res)`](#bodyres)
- [`cacheSeconds(res, seconds)`](#cachesecondsres-seconds)
- [`clearCookie(res, name, opts)`](#clearcookieres-name-opts)
- [`conflict(body)`](#conflictbody)
- [`cookie(name, value, opts)`](#cookiename-value-opts)
- [`created(body)`](#createdbody)
- [`forbidden(body)`](#forbiddenbody)
- [`formatHead(res)`](#formatheadres)
- [`formatResponse(res)`](#formatresponseres)
- [`hasHeader(res, key)`](#hasheaderres-key)
- [`headers(res)`](#headersres)
- [`html(markup)`](#htmlmarkup)
- [`hxRedirect(res, location)`](#hxredirectres-location)
- [`hxRetarget(res, selector)`](#hxretargetres-selector)
- [`hxTrigger(res, eventName)`](#hxtriggerres-eventname)
- [`isFormattedResponse(x)`](#isformattedresponsex)
- [`isResponse(x)`](#isresponsex)
- [`jsonOk(value)`](#jsonokvalue)
- [`noCache(res)`](#nocacheres)
- [`noContent()`](#nocontent)
- [`notFound(body)`](#notfoundbody)
- [`ok(body)`](#okbody)
- [`payloadTooLarge(body)`](#payloadtoolargebody)
- [`redirect(location)`](#redirectlocation)
- [`redirectPermanent(location)`](#redirectpermanentlocation)
- [`renderResponse(res)`](#renderresponseres)
- [`response(status, body)`](#responsestatus-body)
- [`serverError(body)`](#servererrorbody)
- [`setCookie(res, name, value, opts)`](#setcookieres-name-value-opts)
- [`status(res)`](#statusres)
- [`text(txt)`](#texttxt)
- [`unauthorized(body)`](#unauthorizedbody)
- [`withBody(res, body)`](#withbodyres-body)
- [`withConnClose(res)`](#withconncloseres)
- [`withConnKeepAlive(res)`](#withconnkeepaliveres)
- [`withContentType(res, ct)`](#withcontenttyperes-ct)
- [`withHeader(res, key, value)`](#withheaderres-key-value)
- [`withHeaders(res, headersMap)`](#withheadersres-headersmap)
- [`withStatus(res, status)`](#withstatusres-status)
- [`withoutHeader(res, key)`](#withoutheaderres-key)

### Structs

#### `Reply`
```slug
struct slug.web.response#Reply{@str view, @str fragment, @map data = {}, @num status = 200, @map headers = {}}
```


higher-level view descriptor for use with `slug.web.views#renderReply`.

Set `fragment` to render an HTMX partial instead of the full page.
Extra `headers` are merged into the final response.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `view` | @str  | — |  |
| `fragment` | @str  | — |  |
| `data` | @map  | `{}` |  |
| `status` | @num  | `200` |  |
| `headers` | @map  | `{}` |  |

#### `Response`
```slug
struct slug.web.response#Response{@num status, @map headers, body}
```


raw HTTP response value. Construct via the helper functions, not directly.

Header names are always normalised to lowercase.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `status` | @num  | — |  |
| `headers` | @map  | — |  |
| `body` |  | — |  |

### Functions

#### `accepted(body)`
```slug
fn slug.web.response#accepted(@str body) -> @struct(Response)
```


202 Accepted response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `addHeader(res, key, value)`
```slug
fn slug.web.response#addHeader(res, @str key, @str value) -> @struct(Response)
```


adds a header value without replacing existing values for the same key.

Upgrades a single string value to a list when a second value is added.
This is the correct way to set multiple `Set-Cookie` headers.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` | @str  | — |
| `value` | @str  | — |

---

#### `badRequest(body)`
```slug
fn slug.web.response#badRequest(@str body) -> @struct(Response)
```


400 Bad Request response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `body(res)`
```slug
fn slug.web.response#body(res) -> ?
```


returns the body of a response.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `cacheSeconds(res, seconds)`
```slug
fn slug.web.response#cacheSeconds(res, @num seconds) -> @struct(Response)
```


sets `Cache-Control: public, max-age=<seconds>`.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `seconds` | @num  | — |

---

#### `clearCookie(res, name, opts)`
```slug
fn slug.web.response#clearCookie(res, @str name, @map opts) -> @struct(Response)
```


expires a cookie immediately by setting `Max-Age=0`.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `name` | @str  | — |
| `opts` | @map  | — |

---

#### `conflict(body)`
```slug
fn slug.web.response#conflict(@str body) -> @struct(Response)
```


409 Conflict response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `cookie(name, value, opts)`
```slug
fn slug.web.response#cookie(@str name, @str value, @map opts) -> @str
```


builds a `Set-Cookie` header value string.

`opts` may contain: `path`, `domain`, `maxAge`, `expires`,
`secure` (@bool), `httpOnly` (@bool), `sameSite` (`"Lax"|"Strict"|"None"`).

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | @str  | — |
| `value` | @str  | — |
| `opts` | @map  | — |

---

#### `created(body)`
```slug
fn slug.web.response#created(@str body) -> @struct(Response)
```


201 Created response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `forbidden(body)`
```slug
fn slug.web.response#forbidden(@str body) -> @struct(Response)
```


403 Forbidden response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `formatHead(res)`
```slug
fn slug.web.response#formatHead(res) -> @str
```


formats the response head (status line + headers) as an HTTP/1.1 string.

Automatically adds `Content-Length` if not already present.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `formatResponse(res)`
```slug
fn slug.web.response#formatResponse(res) -> @struct(FormattedResponse)
```


formats a response into a `FormattedResponse{ head, body }`.

Automatically adds `Content-Length` if not already present.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `hasHeader(res, key)`
```slug
fn slug.web.response#hasHeader(res, key) -> @bool
```


returns true if the response has a header with the given key.

Key comparison is case-insensitive.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` |  | — |

---

#### `headers(res)`
```slug
fn slug.web.response#headers(res) -> @map
```


returns the headers map of a response.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `html(markup)`
```slug
fn slug.web.response#html(@str markup) -> @struct(Response)
```


200 OK response with `Content-Type: text/html; charset=utf-8`.

| Parameter | Type | Default |
| --- | --- | --- |
| `markup` | @str  | — |

---

#### `hxRedirect(res, location)`
```slug
fn slug.web.response#hxRedirect(res, @str location) -> @struct(Response)
```


sets the `HX-Redirect` response header for a client-side redirect.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `location` | @str  | — |

---

#### `hxRetarget(res, selector)`
```slug
fn slug.web.response#hxRetarget(res, @str selector) -> @struct(Response)
```


sets the `HX-Retarget` response header to override the HTMX swap target.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `selector` | @str  | — |

---

#### `hxTrigger(res, eventName)`
```slug
fn slug.web.response#hxTrigger(res, @str eventName) -> @struct(Response)
```


sets the `HX-Trigger` response header to trigger a client-side event.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `eventName` | @str  | — |

---

#### `isFormattedResponse(x)`
```slug
fn slug.web.response#isFormattedResponse(x) -> @bool
```


returns true if `x` is a `FormattedResponse` struct instance.

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `isResponse(x)`
```slug
fn slug.web.response#isResponse(x) -> @bool
```


returns true if `x` is a `Response` struct instance.

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `jsonOk(value)`
```slug
fn slug.web.response#jsonOk(value) -> @struct(Response)
```


200 OK response with JSON-encoded `value` and `Content-Type: application/json`.

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |

---

#### `noCache(res)`
```slug
fn slug.web.response#noCache(res) -> @struct(Response)
```


sets headers to prevent all caching.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `noContent()`
```slug
fn slug.web.response#noContent() -> @struct(Response)
```


204 No Content response.

---

#### `notFound(body)`
```slug
fn slug.web.response#notFound(@str body = "404 Not Found") -> @struct(Response)
```


404 Not Found response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | `"404 Not Found"` |

---

#### `ok(body)`
```slug
fn slug.web.response#ok(@str body) -> @struct(Response)
```


200 OK response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `payloadTooLarge(body)`
```slug
fn slug.web.response#payloadTooLarge(@str body = "413 Payload Too Large") -> @struct(Response)
```


413 Payload Too Large response. Also sets `Connection: close`.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | `"413 Payload Too Large"` |

---

#### `redirect(location)`
```slug
fn slug.web.response#redirect(@str location) -> @struct(Response)
```


302 Found redirect to `location`.

| Parameter | Type | Default |
| --- | --- | --- |
| `location` | @str  | — |

---

#### `redirectPermanent(location)`
```slug
fn slug.web.response#redirectPermanent(@str location) -> @struct(Response)
```


301 Moved Permanently redirect to `location`.

| Parameter | Type | Default |
| --- | --- | --- |
| `location` | @str  | — |

---

#### `renderResponse(res)`
```slug
fn slug.web.response#renderResponse(res) -> @str
```


renders a response to a complete HTTP/1.1 wire-format string (head + body).

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `response(status, body)`
```slug
fn slug.web.response#response(@num status, @str body) -> @struct(Response)
```


constructs a `Response` with the given status and body.

| Parameter | Type | Default |
| --- | --- | --- |
| `status` | @num  | — |
| `body` | @str  | — |

---

#### `serverError(body)`
```slug
fn slug.web.response#serverError(@str body = "500 Server Error") -> @struct(Response)
```


500 Internal Server Error response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | `"500 Server Error"` |

---

#### `setCookie(res, name, value, opts)`
```slug
fn slug.web.response#setCookie(res, @str name, @str value, @map opts) -> @struct(Response)
```


adds a `Set-Cookie` header to `res` using `cookie(name, value, opts)`.

Multiple cookies can be set by calling `setCookie` repeatedly —
`addHeader` semantics ensure values are not clobbered.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `name` | @str  | — |
| `value` | @str  | — |
| `opts` | @map  | — |

---

#### `status(res)`
```slug
fn slug.web.response#status(res) -> @num
```


returns the status code of a response.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `text(txt)`
```slug
fn slug.web.response#text(@str txt) -> @struct(Response)
```


200 OK response with `Content-Type: text/plain; charset=utf-8`.

| Parameter | Type | Default |
| --- | --- | --- |
| `txt` | @str  | — |

---

#### `unauthorized(body)`
```slug
fn slug.web.response#unauthorized(@str body) -> @struct(Response)
```


401 Unauthorized response.

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `withBody(res, body)`
```slug
fn slug.web.response#withBody(res, @str body) -> @struct(Response)
```


returns a new response with `body` replaced.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `body` | @str  | — |

---

#### `withConnClose(res)`
```slug
fn slug.web.response#withConnClose(res) -> @struct(Response)
```


sets `Connection: close`.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `withConnKeepAlive(res)`
```slug
fn slug.web.response#withConnKeepAlive(res) -> @struct(Response)
```


sets `Connection: keep-alive`.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `withContentType(res, ct)`
```slug
fn slug.web.response#withContentType(res, @str ct) -> @struct(Response)
```


sets the `content-type` header.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `ct` | @str  | — |

---

#### `withHeader(res, key, value)`
```slug
fn slug.web.response#withHeader(res, @str key, @str value) -> @struct(Response)
```


returns a new response with `key` set to `value`, replacing any existing value.

Header keys are normalised to lowercase.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` | @str  | — |
| `value` | @str  | — |

---

#### `withHeaders(res, headersMap)`
```slug
fn slug.web.response#withHeaders(res, @map headersMap) -> @struct(Response)
```


returns a new response with all headers from `headersMap` applied.

Each key is normalised to lowercase. Existing headers with the same
key are replaced.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `headersMap` | @map  | — |

---

#### `withStatus(res, status)`
```slug
fn slug.web.response#withStatus(res, @num status) -> @struct(Response)
```


returns a new response with `status` replaced.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `status` | @num  | — |

---

#### `withoutHeader(res, key)`
```slug
fn slug.web.response#withoutHeader(res, @str key) -> @struct(Response)
```


returns a new response with `key` removed from headers.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` | @str  | — |
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
struct slug.web.response#Reply{view:str, fragment:str, data:map = {}, status:num = 200, headers:map = {}}
```

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `view` | str | — |  |
| `fragment` | str | — |  |
| `data` | map | `{}` |  |
| `status` | num | `200` |  |
| `headers` | map | `{}` |  |

#### `Response`
```slug
struct slug.web.response#Response{status:num, headers:map, body}
```

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `status` | num | — |  |
| `headers` | map | — |  |
| `body` |  | — |  |

### Functions

#### `accepted(body)`
```slug
fn slug.web.response#accepted(body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |

---

#### `addHeader(res, key, value)`
```slug
fn slug.web.response#addHeader(res, key:str, value:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` | str | — |
| `value` | str | — |

---

#### `badRequest(body)`
```slug
fn slug.web.response#badRequest(body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |

---

#### `body(res)`
```slug
fn slug.web.response#body(res):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `cacheSeconds(res, seconds)`
```slug
fn slug.web.response#cacheSeconds(res, seconds:num):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `seconds` | num | — |

---

#### `clearCookie(res, name, opts)`
```slug
fn slug.web.response#clearCookie(res, name:str, opts:map):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `name` | str | — |
| `opts` | map | — |

---

#### `conflict(body)`
```slug
fn slug.web.response#conflict(body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |

---

#### `cookie(name, value, opts)`
```slug
fn slug.web.response#cookie(name:str, value:str, opts:map):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | str | — |
| `value` | str | — |
| `opts` | map | — |

---

#### `created(body)`
```slug
fn slug.web.response#created(body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |

---

#### `forbidden(body)`
```slug
fn slug.web.response#forbidden(body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |

---

#### `formatHead(res)`
```slug
fn slug.web.response#formatHead(res):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `formatResponse(res)`
```slug
fn slug.web.response#formatResponse(res):FormattedResponse
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `hasHeader(res, key)`
```slug
fn slug.web.response#hasHeader(res, key):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` |  | — |

---

#### `headers(res)`
```slug
fn slug.web.response#headers(res):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `html(markup)`
```slug
fn slug.web.response#html(markup:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `markup` | str | — |

---

#### `hxRedirect(res, location)`
```slug
fn slug.web.response#hxRedirect(res, location:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `location` | str | — |

---

#### `hxRetarget(res, selector)`
```slug
fn slug.web.response#hxRetarget(res, selector:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `selector` | str | — |

---

#### `hxTrigger(res, eventName)`
```slug
fn slug.web.response#hxTrigger(res, eventName:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `eventName` | str | — |

---

#### `isFormattedResponse(x)`
```slug
fn slug.web.response#isFormattedResponse(x):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `isResponse(x)`
```slug
fn slug.web.response#isResponse(x):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `jsonOk(value)`
```slug
fn slug.web.response#jsonOk(value):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |

---

#### `noCache(res)`
```slug
fn slug.web.response#noCache(res):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `noContent()`
```slug
fn slug.web.response#noContent():Response
```

---

#### `notFound(body)`
```slug
fn slug.web.response#notFound(body:str = "404 Not Found"):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | `"404 Not Found"` |

---

#### `ok(body)`
```slug
fn slug.web.response#ok(body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |

---

#### `payloadTooLarge(body)`
```slug
fn slug.web.response#payloadTooLarge(body:str = "413 Payload Too Large"):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | `"413 Payload Too Large"` |

---

#### `redirect(location)`
```slug
fn slug.web.response#redirect(location:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `location` | str | — |

---

#### `redirectPermanent(location)`
```slug
fn slug.web.response#redirectPermanent(location:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `location` | str | — |

---

#### `renderResponse(res)`
```slug
fn slug.web.response#renderResponse(res):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `response(status, body)`
```slug
fn slug.web.response#response(status:num, body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `status` | num | — |
| `body` | str | — |

---

#### `serverError(body)`
```slug
fn slug.web.response#serverError(body:str = "500 Server Error"):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | `"500 Server Error"` |

---

#### `setCookie(res, name, value, opts)`
```slug
fn slug.web.response#setCookie(res, name:str, value:str, opts:map):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `name` | str | — |
| `value` | str | — |
| `opts` | map | — |

---

#### `status(res)`
```slug
fn slug.web.response#status(res):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `text(txt)`
```slug
fn slug.web.response#text(txt:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `txt` | str | — |

---

#### `unauthorized(body)`
```slug
fn slug.web.response#unauthorized(body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |

---

#### `withBody(res, body)`
```slug
fn slug.web.response#withBody(res, body:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `body` | str | — |

---

#### `withConnClose(res)`
```slug
fn slug.web.response#withConnClose(res):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `withConnKeepAlive(res)`
```slug
fn slug.web.response#withConnKeepAlive(res):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `withContentType(res, ct)`
```slug
fn slug.web.response#withContentType(res, ct:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `ct` | str | — |

---

#### `withHeader(res, key, value)`
```slug
fn slug.web.response#withHeader(res, key:str, value:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` | str | — |
| `value` | str | — |

---

#### `withHeaders(res, headersMap)`
```slug
fn slug.web.response#withHeaders(res, headersMap:map):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `headersMap` | map | — |

---

#### `withStatus(res, status)`
```slug
fn slug.web.response#withStatus(res, status:num):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `status` | num | — |

---

#### `withoutHeader(res, key)`
```slug
fn slug.web.response#withoutHeader(res, key:str):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` | str | — |
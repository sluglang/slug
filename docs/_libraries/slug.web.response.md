---
title: response (slug.web)
---

## slug.web.response

### Structs

#### `Reply`
```slug
struct slug.web.response#Reply{@str view, @str fragment, @map data = {}, @num status = 200, @map headers = {}}
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `addHeader(res, key, value)`
```slug
fn slug.web.response#addHeader(res, @str key, @str value) -> @struct(Response)
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `body(res)`
```slug
fn slug.web.response#body(res) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `cacheSeconds(res, seconds)`
```slug
fn slug.web.response#cacheSeconds(res, @num seconds) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `seconds` | @num  | — |

---

#### `clearCookie(res, name, opts)`
```slug
fn slug.web.response#clearCookie(res, @str name, @map opts) -> @struct(Response)
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `cookie(name, value, opts)`
```slug
fn slug.web.response#cookie(@str name, @str value, @map opts) -> @str
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `forbidden(body)`
```slug
fn slug.web.response#forbidden(@str body) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `formatHead(res)`
```slug
fn slug.web.response#formatHead(res) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `formatResponse(res)`
```slug
fn slug.web.response#formatResponse(res) -> @struct(FormattedResponse)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `hasHeader(res, key)`
```slug
fn slug.web.response#hasHeader(res, key) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` |  | — |

---

#### `headers(res)`
```slug
fn slug.web.response#headers(res) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `html(markup)`
```slug
fn slug.web.response#html(@str markup) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `markup` | @str  | — |

---

#### `hxRedirect(res, location)`
```slug
fn slug.web.response#hxRedirect(res, @str location) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `location` | @str  | — |

---

#### `hxRetarget(res, selector)`
```slug
fn slug.web.response#hxRetarget(res, @str selector) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `selector` | @str  | — |

---

#### `hxTrigger(res, eventName)`
```slug
fn slug.web.response#hxTrigger(res, @str eventName) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `eventName` | @str  | — |

---

#### `isFormattedResponse(x)`
```slug
fn slug.web.response#isFormattedResponse(x) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `isResponse(x)`
```slug
fn slug.web.response#isResponse(x) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `x` |  | — |

---

#### `jsonOk(value)`
```slug
fn slug.web.response#jsonOk(value) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |

---

#### `noCache(res)`
```slug
fn slug.web.response#noCache(res) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `noContent()`
```slug
fn slug.web.response#noContent() -> @struct(Response)
```

---

#### `notFound(body)`
```slug
fn slug.web.response#notFound(@str body = "404 Not Found") -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | `"404 Not Found"` |

---

#### `ok(body)`
```slug
fn slug.web.response#ok(@str body) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `payloadTooLarge(body)`
```slug
fn slug.web.response#payloadTooLarge(@str body = "413 Payload Too Large") -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | `"413 Payload Too Large"` |

---

#### `redirect(location)`
```slug
fn slug.web.response#redirect(@str location) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `location` | @str  | — |

---

#### `redirectPermanent(location)`
```slug
fn slug.web.response#redirectPermanent(@str location) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `location` | @str  | — |

---

#### `renderResponse(res)`
```slug
fn slug.web.response#renderResponse(res) -> @struct(FormattedResponse)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `response(status, body)`
```slug
fn slug.web.response#response(@num status, @str body) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `status` | @num  | — |
| `body` | @str  | — |

---

#### `serverError(body)`
```slug
fn slug.web.response#serverError(@str body = "500 Server Error") -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | `"500 Server Error"` |

---

#### `setCookie(res, name, value, opts)`
```slug
fn slug.web.response#setCookie(res, @str name, @str value, @map opts) -> @struct(Response)
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `text(txt)`
```slug
fn slug.web.response#text(@str txt) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `txt` | @str  | — |

---

#### `unauthorized(body)`
```slug
fn slug.web.response#unauthorized(@str body) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |

---

#### `withBody(res, body)`
```slug
fn slug.web.response#withBody(res, @str body) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `body` | @str  | — |

---

#### `withConnClose(res)`
```slug
fn slug.web.response#withConnClose(res) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `withConnKeepAlive(res)`
```slug
fn slug.web.response#withConnKeepAlive(res) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |

---

#### `withContentType(res, ct)`
```slug
fn slug.web.response#withContentType(res, @str ct) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `ct` | @str  | — |

---

#### `withHeader(res, key, value)`
```slug
fn slug.web.response#withHeader(res, @str key, @str value) -> @struct(Response)
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `headersMap` | @map  | — |

---

#### `withStatus(res, status)`
```slug
fn slug.web.response#withStatus(res, @num status) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `status` | @num  | — |

---

#### `withoutHeader(res, key)`
```slug
fn slug.web.response#withoutHeader(res, @str key) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
| `key` | @str  | — |
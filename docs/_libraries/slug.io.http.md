---
title: http (slug.io)
---

## slug.io.http

slug.io.http — HTTP client

Makes outbound HTTP requests and returns `[@num status, @str body]`.
All methods are convenience wrappers around `request`.

## Example

```slug
val { get, post } = import("slug.io.http")
val { decode }    = import("slug.json")

val [status, body] = get("https://api.example.com/users")
match status {
  200 => decode(body) /> println
  _   => println("error: " + status)
}

val [status, body] = post(
  encode({ name: "Alice" }),
  "https://api.example.com/users",
  { "Content-Type": "application/json" }
)
```

Note: `post`, `put`, and `patch` take the body as the **first** argument,
before the URL. This is designed for call-chain usage:
```slug
encode(payload) /> post("https://api.example.com/data")
```

@effects('net')

### Functions

#### `delete(url, headers)`
```slug
fn slug.io.http#delete(@str url, @map headers = {}) -> [@num, @str]
```


makes a DELETE request and returns `[status, body]`.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `url` | @str  | — |
| `headers` | @map  | `{}` |

---

#### `get(url, headers)`
```slug
fn slug.io.http#get(@str url, @map headers = {}) -> [@num, @str]
```


makes a GET request and returns `[status, body]`.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `url` | @str  | — |
| `headers` | @map  | `{}` |

---

#### `patch(body, url, headers)`
```slug
fn slug.io.http#patch(@str body, @str url, @map headers = {}) -> [@num, @str]
```


makes a PATCH request and returns `[status, body]`.

Note: `body` is the first argument to support call-chain style.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |
| `url` | @str  | — |
| `headers` | @map  | `{}` |

---

#### `post(body, url, headers)`
```slug
fn slug.io.http#post(@str body, @str url, @map headers = {}) -> [@num, @str]
```


makes a POST request and returns `[status, body]`.

Note: `body` is the first argument to support call-chain style:
`encode(payload) /> post("https://...")`

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |
| `url` | @str  | — |
| `headers` | @map  | `{}` |

---

#### `put(body, url, headers)`
```slug
fn slug.io.http#put(@str body, @str url, @map headers = {}) -> [@num, @str]
```


makes a PUT request and returns `[status, body]`.

Note: `body` is the first argument to support call-chain style.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | @str  | — |
| `url` | @str  | — |
| `headers` | @map  | `{}` |

---

#### `request(method, url, body, headers)`
```slug
fn slug.io.http#request(@str method, @str url, @str body = "", @map headers = {}) -> [@num, @str]
```


makes an HTTP request and returns `[status, body]`.

The low-level entry point — prefer the method-specific helpers for
common cases.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `method` | @str  | — |
| `url` | @str  | — |
| `body` | @str  | `""` |
| `headers` | @map  | `{}` |

---

#### `urlDecode(str)`
```slug
fn slug.io.http#urlDecode(@str str) -> ?
```


decodes a URL-encoded string.

Note: not yet implemented — returns the input string unchanged.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
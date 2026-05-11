---
title: http (slug.io)
---

## slug.io.http

slug.io.http — HTTP client

Makes outbound HTTP requests and returns `[status:num, body:str]`.
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

### TOC

- [`delete(url, headers)`](#deleteurl-headers)
- [`get(url, headers)`](#geturl-headers)
- [`patch(body, url, headers)`](#patchbody-url-headers)
- [`post(body, url, headers)`](#postbody-url-headers)
- [`put(body, url, headers)`](#putbody-url-headers)
- [`request(method, url, body, headers, timeout)`](#requestmethod-url-body-headers-timeout)
- [`urlDecode(str)`](#urldecodestr)

### Functions

#### `delete(url, headers)`
```slug
fn slug.io.http#delete(url:str, headers:map = {}):[@num, @str]
```


makes a DELETE request and returns `[status, body]`.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `url` | str | — |
| `headers` | map | `{}` |

---

#### `get(url, headers)`
```slug
fn slug.io.http#get(url:str, headers:map = {}):[@num, @str]
```


makes a GET request and returns `[status, body]`.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `url` | str | — |
| `headers` | map | `{}` |

---

#### `patch(body, url, headers)`
```slug
fn slug.io.http#patch(body:str, url:str, headers:map = {}):[@num, @str]
```


makes a PATCH request and returns `[status, body]`.

Note: `body` is the first argument to support call-chain style.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |
| `url` | str | — |
| `headers` | map | `{}` |

---

#### `post(body, url, headers)`
```slug
fn slug.io.http#post(body:str, url:str, headers:map = {}):[@num, @str]
```


makes a POST request and returns `[status, body]`.

Note: `body` is the first argument to support call-chain style:
`encode(payload) /> post("https://...")`

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |
| `url` | str | — |
| `headers` | map | `{}` |

---

#### `put(body, url, headers)`
```slug
fn slug.io.http#put(body:str, url:str, headers:map = {}):[@num, @str]
```


makes a PUT request and returns `[status, body]`.

Note: `body` is the first argument to support call-chain style.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `body` | str | — |
| `url` | str | — |
| `headers` | map | `{}` |

---

#### `request(method, url, body, headers, timeout)`
```slug
fn slug.io.http#request(method:str, url:str, body:str = "", headers:map = {}, timeout:num = 30000):[@num, @str]
```


makes an HTTP request and returns `[status, body]`.

The low-level entry point — prefer the method-specific helpers for
common cases.

@effects('net')

| Parameter | Type | Default |
| --- | --- | --- |
| `method` | str | — |
| `url` | str | — |
| `body` | str | `""` |
| `headers` | map | `{}` |
| `timeout` | num | `30000` |

---

#### `urlDecode(str)`
```slug
fn slug.io.http#urlDecode(str:str):any
```


decodes a URL-encoded string.

Note: not yet implemented — returns the input string unchanged.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
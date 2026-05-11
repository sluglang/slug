---
title: jsonrpc (slug)
---

## slug.jsonrpc

slug.jsonrpc — JSON-RPC 2.0 helpers

Builders, validators, and parser/encoder utilities for JSON-RPC 2.0
messages as defined by the specification:
https://www.jsonrpc.org/specification

### TOC

- [INTERNAL_ERROR](#internal_error)
- [INVALID_PARAMS](#invalid_params)
- [INVALID_REQUEST](#invalid_request)
- [METHOD_NOT_FOUND](#method_not_found)
- [PARSE_ERROR](#parse_error)
- [VERSION](#version)
- [`encode(message)`](#encodemessage)
- [`errorObject(code, message, data)`](#errorobjectcode-message-data)
- [`failure(id, code, message, data)`](#failureid-code-message-data)
- [`isBatch(v)`](#isbatchv)
- [`isErrorResponse(v)`](#iserrorresponsev)
- [`isMessage(v)`](#ismessagev)
- [`isNotification(v)`](#isnotificationv)
- [`isRequest(v)`](#isrequestv)
- [`isResponse(v)`](#isresponsev)
- [`isSuccessResponse(v)`](#issuccessresponsev)
- [`notification(method, params)`](#notificationmethod-params)
- [`parse(payload)`](#parsepayload)
- [`request(id, method, params)`](#requestid-method-params)
- [`success(id, result)`](#successid-result)
- [`validate(message)`](#validatemessage)

### Constants

#### `INTERNAL_ERROR`

```slug
num slug.jsonrpc#INTERNAL_ERROR
```

#### `INVALID_PARAMS`

```slug
num slug.jsonrpc#INVALID_PARAMS
```

#### `INVALID_REQUEST`

```slug
num slug.jsonrpc#INVALID_REQUEST
```

#### `METHOD_NOT_FOUND`

```slug
num slug.jsonrpc#METHOD_NOT_FOUND
```

#### `PARSE_ERROR`

```slug
num slug.jsonrpc#PARSE_ERROR
```

#### `VERSION`

```slug
str slug.jsonrpc#VERSION
```

### Functions

#### `encode(message)`
```slug
fn slug.jsonrpc#encode(message):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` |  | — |

---

#### `errorObject(code, message, data)`
```slug
fn slug.jsonrpc#errorObject(code:num, message:str, data = nil):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | num | — |
| `message` | str | — |
| `data` |  | `nil` |

---

#### `failure(id, code, message, data)`
```slug
fn slug.jsonrpc#failure(id, code:num, message:str, data = nil):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `id` |  | — |
| `code` | num | — |
| `message` | str | — |
| `data` |  | `nil` |

---

#### `isBatch(v)`
```slug
fn slug.jsonrpc#isBatch(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isErrorResponse(v)`
```slug
fn slug.jsonrpc#isErrorResponse(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isMessage(v)`
```slug
fn slug.jsonrpc#isMessage(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isNotification(v)`
```slug
fn slug.jsonrpc#isNotification(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isRequest(v)`
```slug
fn slug.jsonrpc#isRequest(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isResponse(v)`
```slug
fn slug.jsonrpc#isResponse(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isSuccessResponse(v)`
```slug
fn slug.jsonrpc#isSuccessResponse(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `notification(method, params)`
```slug
fn slug.jsonrpc#notification(method:str, params = nil):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `method` | str | — |
| `params` |  | `nil` |

---

#### `parse(payload)`
```slug
fn slug.jsonrpc#parse(payload:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `payload` | str | — |

**Throws:** `Error{type:JsonRpcError}`

---

#### `request(id, method, params)`
```slug
fn slug.jsonrpc#request(id, method:str, params = nil):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `id` |  | — |
| `method` | str | — |
| `params` |  | `nil` |

---

#### `success(id, result)`
```slug
fn slug.jsonrpc#success(id, result = nil):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `id` |  | — |
| `result` |  | `nil` |

---

#### `validate(message)`
```slug
fn slug.jsonrpc#validate(message):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` |  | — |

**Throws:** `Error{type:JsonRpcError}`
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
fn slug.jsonrpc#encode(message:list|map):str
```


encodes a JSON-RPC message or batch to JSON.

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | list \| map | — |

---

#### `errorObject(code, message, data)`
```slug
fn slug.jsonrpc#errorObject(code:num, message:str, data:any = nil):map<str,any>
```


builds a JSON-RPC error object.

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | num | — |
| `message` | str | — |
| `data` | any | `nil` |

---

#### `failure(id, code, message, data)`
```slug
fn slug.jsonrpc#failure(id:str|num|nil, code:num, message:str, data:any = nil):map<str,any>
```


builds a JSON-RPC error response object.

| Parameter | Type | Default |
| --- | --- | --- |
| `id` | str \| num \| nil | — |
| `code` | num | — |
| `message` | str | — |
| `data` | any | `nil` |

---

#### `isBatch(v)`
```slug
fn slug.jsonrpc#isBatch(v):bool
```


returns true when `v` is a JSON-RPC batch payload.

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isErrorResponse(v)`
```slug
fn slug.jsonrpc#isErrorResponse(v):bool
```


returns true when `v` is an error response.

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isMessage(v)`
```slug
fn slug.jsonrpc#isMessage(v):bool
```


returns true when `v` is a valid single JSON-RPC message object.

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isNotification(v)`
```slug
fn slug.jsonrpc#isNotification(v):bool
```


returns true when `v` is a notification object.

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isRequest(v)`
```slug
fn slug.jsonrpc#isRequest(v):bool
```


returns true when `v` is a request object (includes notifications).

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isResponse(v)`
```slug
fn slug.jsonrpc#isResponse(v):bool
```


returns true when `v` is a response object.

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `isSuccessResponse(v)`
```slug
fn slug.jsonrpc#isSuccessResponse(v):bool
```


returns true when `v` is a success response.

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `notification(method, params)`
```slug
fn slug.jsonrpc#notification(method:str, params:list|map|nil = nil):map<str,any>
```


builds a JSON-RPC notification object (request without `id`).

| Parameter | Type | Default |
| --- | --- | --- |
| `method` | str | — |
| `params` | list \| map \| nil | `nil` |

---

#### `parse(payload)`
```slug
fn slug.jsonrpc#parse(payload:str):any
```


decodes JSON and validates it as JSON-RPC 2.0.

Throws `JsonRpcError` with code:
- `-32700` for JSON parse errors
- `-32600` for structurally invalid JSON-RPC payloads

| Parameter | Type | Default |
| --- | --- | --- |
| `payload` | str | — |

**Throws:** `Error{type:JsonRpcError}`

---

#### `request(id, method, params)`
```slug
fn slug.jsonrpc#request(id:str|num|nil, method:str, params:list|map|nil = nil):map<str,any>
```


builds a JSON-RPC request object.

| Parameter | Type | Default |
| --- | --- | --- |
| `id` | str \| num \| nil | — |
| `method` | str | — |
| `params` | list \| map \| nil | `nil` |

---

#### `success(id, result)`
```slug
fn slug.jsonrpc#success(id:str|num|nil, result:any = nil):map<str,any>
```


builds a JSON-RPC success response object.

The `result` member is always present (it may be `nil`).

| Parameter | Type | Default |
| --- | --- | --- |
| `id` | str \| num \| nil | — |
| `result` | any | `nil` |

---

#### `validate(message)`
```slug
fn slug.jsonrpc#validate(message:any):map|list
```


validates a JSON-RPC message or batch.

Returns the input when valid, otherwise throws `JsonRpcError`
with `code = -32600` (Invalid Request).

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | any | — |

**Throws:** `Error{type:JsonRpcError}`
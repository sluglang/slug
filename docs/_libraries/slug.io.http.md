---
title: http (slug.io)
---

## slug.io.http

### Functions

#### `delete(url, headers)`
```slug
fn slug.io.http#delete(@str url, @map headers = {}) -> [@num, @str]
```

| Parameter | Type | Default |
| --- | --- | --- |
| `url` | @str  | — |
| `headers` | @map  | `{}` |

---

#### `get(url, headers)`
```slug
fn slug.io.http#get(@str url, @map headers = {}) -> [@num, @str]
```

| Parameter | Type | Default |
| --- | --- | --- |
| `url` | @str  | — |
| `headers` | @map  | `{}` |

---

#### `patch(body, url, headers)`
```slug
fn slug.io.http#patch(@str body, @str url, @map headers = {}) -> [@num, @str]
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
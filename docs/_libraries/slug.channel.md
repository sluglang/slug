---
title: channel (slug)
---

## slug.channel

### Structs

#### `Empty`
```slug
struct slug.channel#Empty{}
```

#### `Full`
```slug
struct slug.channel#Full{value}
```

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `value` |  | — |  |

### Functions

#### `await(handle, timeout)`
```slug
fn slug.channel#await(@task handle, @num timeout = 0) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @task  | — |
| `timeout` | @num  | `0` |

**Throws:** `@struct(Error{type:TimeoutError})`

---

#### `chan(capacity)`
```slug
fn slug.channel#chan(@num capacity = 0) -> @chan
```

| Parameter | Type | Default |
| --- | --- | --- |
| `capacity` | @num  | `0` |

---

#### `close(channel)`
```slug
fn slug.channel#close(@chan channel) -> nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |

---

#### `recv(channel, timeout)`
```slug
fn slug.channel#recv(@chan channel, @num timeout = 0) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |
| `timeout` | @num  | `0` |

**Throws:** `@struct(Error{type:TimeoutError})`

---

#### `send(channel, payload)`
```slug
fn slug.channel#send(@chan channel, payload) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |
| `payload` |  | — |

---

#### `tryRecv(channel)`
```slug
fn slug.channel#tryRecv(@chan channel) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |

---

#### `trySend(channel, payload)`
```slug
fn slug.channel#trySend(@chan channel, payload) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `channel` | @chan  | — |
| `payload` |  | — |
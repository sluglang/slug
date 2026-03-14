---
title: tcp (slug.io)
---

## slug.io.tcp

### Functions

#### `accept(listener)`
```slug
fn slug.io.tcp#accept(@num listener) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `listener` | @num  | — |

---

#### `bind(addr, port)`
```slug
fn slug.io.tcp#bind(@str addr, @num port) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `addr` | @str  | — |
| `port` | @num  | — |

---

#### `close(handle)`
```slug
fn slug.io.tcp#close(@num handle) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |

---

#### `connect(addr, port)`
```slug
fn slug.io.tcp#connect(@str addr, @num port) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `addr` | @str  | — |
| `port` | @num  | — |

---

#### `read(stream, maxBytes)`
```slug
fn slug.io.tcp#read(@num stream, @num maxBytes) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `stream` | @num  | — |
| `maxBytes` | @num  | — |

---

#### `write(stream, data)`
```slug
fn slug.io.tcp#write(@num stream, @str data) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `stream` | @num  | — |
| `data` | @str  | — |
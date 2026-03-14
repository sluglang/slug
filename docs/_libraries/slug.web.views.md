---
title: views (slug.web)
---

## slug.web.views

### Functions

#### `new(root)`
```slug
fn slug.web.views#new(@str root = "views") -> @struct(Views)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `root` | @str  | `"views"` |

---

#### `render(v, name, data)`
```slug
fn slug.web.views#render(v, name, data = {}) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |
| `name` |  | — |
| `data` |  | `{}` |

---

#### `renderReply(vx, req, reply)`
```slug
fn slug.web.views#renderReply(vx, req, reply) -> @struct(Response)
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vx` |  | — |
| `req` |  | — |
| `reply` |  | — |
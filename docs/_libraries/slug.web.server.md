---
title: server (slug.web)
---

## slug.web.server

### Functions

#### `serve(app, addr, port)`
```slug
fn slug.web.server#serve(@fn app, @str addr = cfg(address, 0.0.0.0), @num port = cfg(port, 8080)) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `app` | @fn  | — |
| `addr` | @str  | `cfg(address, 0.0.0.0)` |
| `port` | @num  | `cfg(port, 8080)` |

**Throws:** `@struct(Error{type:ServerError})`
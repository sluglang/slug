---
title: migration (slug.db)
---

## slug.db.migration

### Structs

#### `Migration`
```slug
struct slug.db.migration#Migration{@num id, @str version, @str environment, @str filename, @str checksum, @bool success, @str appliedAt}
```

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `id` | @num  | — |  |
| `version` | @str  | — |  |
| `environment` | @str  | — |  |
| `filename` | @str  | — |  |
| `checksum` | @str  | — |  |
| `success` | @bool  | — |  |
| `appliedAt` | @str  | — |  |

### Functions

#### `down(conn, step, base)`
```slug
fn slug.db.migration#down(@num conn, @num step = 1, @str base = DefaultBase) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `conn` | @num  | — |
| `step` | @num  | `1` |
| `base` | @str  | `DefaultBase` |

---

#### `status(conn, env, base)`
```slug
fn slug.db.migration#status(@num conn, @list env = DefaultEnv, @str base = DefaultBase) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `conn` | @num  | — |
| `env` | @list  | `DefaultEnv` |
| `base` | @str  | `DefaultBase` |

**Throws:** `@struct(Error{type:MigrationError})`

---

#### `up(conn, env, base)`
```slug
fn slug.db.migration#up(@num conn, @list env = DefaultEnv, @str base = DefaultBase) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `conn` | @num  | — |
| `env` | @list  | `DefaultEnv` |
| `base` | @str  | `DefaultBase` |

**Throws:** `@struct(Error{type:MigrationError})`
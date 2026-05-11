---
title: migration (slug.db)
---

## slug.db.migration

slug.db.migration — SQL schema migration runner

Applies, rolls back, and reports on versioned SQL migration scripts
stored as `.sql` files on the filesystem.

## File naming convention

Migration files must follow the pattern:
```
<version>__<description>.sql
<version>__<environment>__<description>.sql
```

Examples:
```
db/migrations/001__create_users.sql
db/migrations/002__production__add_index.sql
```

Files are sorted and applied in ascending version order.

## Script sections

Each `.sql` file may contain `-- @up` and `-- @down` section markers.
The `up` section is applied on `up`, the `down` section on `down`.
If no markers are present, the entire file is treated as the `up` script.

```sql
-- @up
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);

-- @down
DROP TABLE users;
```

## Tracking table

Applied migrations are recorded in a `slug_schema_migrations` table
which is created automatically on first run. Each entry stores the
version, environment, filename, SHA-256 checksum, and success status.

## Configuration

Default base directory and environment can be set via `cfg`:
```
base-directory = db/migrations
environment    = production
```

@effects('fs db')

### TOC

- [Migration](#migration)
- [`down(conn, step, base)`](#downconn-step-base)
- [`status(conn, env, base)`](#statusconn-env-base)
- [`up(conn, env, base)`](#upconn-env-base)

### Structs

#### `Migration`
```slug
struct slug.db.migration#Migration{id:num, version:str, environment:str, filename:str, checksum:str, success:bool, appliedAt:str}
```

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `id` | num | — |  |
| `version` | str | — |  |
| `environment` | str | — |  |
| `filename` | str | — |  |
| `checksum` | str | — |  |
| `success` | bool | — |  |
| `appliedAt` | str | — |  |

### Functions

#### `down(conn, step, base)`
```slug
fn slug.db.migration#down(conn:num, step:num = 1, base:str = DefaultBase):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `conn` | num | — |
| `step` | num | `1` |
| `base` | str | `DefaultBase` |

---

#### `status(conn, env, base)`
```slug
fn slug.db.migration#status(conn:num, env:list = DefaultEnv, base:str = DefaultBase):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `conn` | num | — |
| `env` | list | `DefaultEnv` |
| `base` | str | `DefaultBase` |

**Throws:** `Error{type:MigrationError}`

---

#### `up(conn, env, base)`
```slug
fn slug.db.migration#up(conn:num, env:list = DefaultEnv, base:str = DefaultBase):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `conn` | num | — |
| `env` | list | `DefaultEnv` |
| `base` | str | `DefaultBase` |

**Throws:** `Error{type:MigrationError}`
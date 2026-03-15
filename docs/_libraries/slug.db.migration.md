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

### Structs

#### `Migration`
```slug
struct slug.db.migration#Migration{@num id, @str version, @str environment, @str filename, @str checksum, @bool success, @str appliedAt}
```


represents a migration record as stored in the tracking table.

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


rolls back the most recently applied migrations.

Applies the `-- @down` section of each rolled-back migration script
and removes its entry from the tracking table. `step` controls how
many migrations to roll back (default 1).

Returns the connection handle for chaining.

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


prints the status of all migration files against the database.

Each migration is reported as APPLIED, MISSING, or INELIGIBLE (for
environments not in `env`). Applied migrations whose checksum no longer
matches the file will throw `MigrationError`.

Returns the connection handle for chaining.

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


applies all pending eligible migrations in ascending version order.

Skips already-applied migrations and migrations for ineligible environments.
On failure, records the migration as failed and rethrows the error.

Returns the connection handle for chaining.

| Parameter | Type | Default |
| --- | --- | --- |
| `conn` | @num  | — |
| `env` | @list  | `DefaultEnv` |
| `base` | @str  | `DefaultBase` |

**Throws:** `@struct(Error{type:MigrationError})`
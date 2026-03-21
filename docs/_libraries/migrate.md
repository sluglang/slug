---
title: migrate
---

## migrate

Slug Migrate — database schema migration tool

Applies, rolls back, and reports on versioned SQL migration scripts.

Usage:
```bash
slug migrate [options] [up|down|status]
```

Commands:
- `up`       apply all pending migrations (default)
- `down`     roll back the most recent migration(s)
- `status`   show applied / pending / ineligible migrations

Options:
- `-h, --help`      show this screen
- `-v, --version`   show version
- `--db`            database connection string (required)
- `--driver`        database driver: sqlite3 | mysql | postgres (default: sqlite3)
- `--env`           environment tag(s) to include; may be passed multiple times
- `--base`          migrations directory (default: db/migrations)
- `--step`          number of migrations to roll back with `down` (default: 1)
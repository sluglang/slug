---
title: log (slug)
---

## slug.log

slug.log — structured logging

Provides levelled, coloured, optionally file-backed logging for Slug
programs. Supports both module-level convenience functions and named
per-source loggers for multi-module applications.

## Log levels

Levels are ordered: `trace` < `debug` < `info` < `warn` < `error` < `none`.
A message is emitted only when its level is >= the configured log level.
Setting the level to `none` suppresses all output.

## Configuration

| cfg key              | Default  | Description                                    |
|----------------------|----------|------------------------------------------------|
| `level`              | `"info"` | Global minimum log level                       |
| `level__<src>`       |          | Per-source override (dots replaced with `_`)   |
| `log-file`           | `nil`    | Write to this file path instead of stderr      |
| `colour`             | `true`   | Write to stderr with ANSI color codes          |

## Quick start

```slug
val log = import("slug.log")

log.info("server started on port {}", 8080)
log.warn("retrying after {} ms", 500)
```

## Named loggers

```slug
val log = import("slug.log").logger("slug.io.http")

log.debug("GET {} {}", path, status)

// chainable — passes the value through unchanged
val result = fetchData()
  /> log.cInfo("fetched {} records")
  /> process
```

## Chainable variants

The `c*` functions (`cTrace`, `cDebug`, `cInfo`, `cWarn`, `cError`) log a
message and return their first argument unchanged. This lets you insert
logging into a pipeline without breaking the data flow:

```slug
items
  /> filter(fn(x) { x.active })
  /> log.cDebug("active items: {}")
  /> map(transform)
```

## File output

When `log-file` is set, log entries are appended to that file with no
colour codes. When writing to stderr, ANSI colour codes are applied per
level.

@effects('io')

### TOC

- [Debug](#debug)
- [Error](#error)
- [Info](#info)
- [None](#none)
- [Trace](#trace)
- [Warn](#warn)
- [`cDebug(arg, message, args)`](#cdebugarg-message-args)
- [`cError(arg, message, args)`](#cerrorarg-message-args)
- [`cInfo(arg, message, args)`](#cinfoarg-message-args)
- [`cTrace(arg, message, args)`](#ctracearg-message-args)
- [`cWarn(arg, message, args)`](#cwarnarg-message-args)
- [`debug(message, args)`](#debugmessage-args)
- [`error(message, args)`](#errormessage-args)
- [`info(nil)`](#infonil)
- [`logger(src)`](#loggersrc)
- [`none(message, args)`](#nonemessage-args)
- [`trace(message, args)`](#tracemessage-args)
- [`warn(message, args)`](#warnmessage-args)

### Constants

#### `Debug`

```slug
num slug.log#Debug
```

#### `Error`

```slug
num slug.log#Error
```

#### `Info`

```slug
num slug.log#Info
```

#### `None`

```slug
num slug.log#None
```

#### `Trace`

```slug
num slug.log#Trace
```

#### `Warn`

```slug
num slug.log#Warn
```

### Functions

#### `cDebug(arg, message, args)`
```slug
fn slug.log#cDebug(arg, @str message, ...args) -> ?
```


logs a debug-level message and returns the first argument unchanged.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | @str  | — |
| `args` |  | — |

---

#### `cError(arg, message, args)`
```slug
fn slug.log#cError(arg, @str message, ...args) -> ?
```


logs an error-level message and returns the first argument unchanged.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | @str  | — |
| `args` |  | — |

---

#### `cInfo(arg, message, args)`
```slug
fn slug.log#cInfo(arg, @str message, ...args) -> ?
```


logs an info-level message and returns the first argument unchanged.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | @str  | — |
| `args` |  | — |

---

#### `cTrace(arg, message, args)`
```slug
fn slug.log#cTrace(arg, @str message, ...args) -> ?
```


logs a trace-level message and returns the first argument unchanged.

Designed for use in call chains — the value flows through unmodified:

```slug
val result = fetchItems()
  /> cTrace("fetched: {}")
  /> process
```

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | @str  | — |
| `args` |  | — |

---

#### `cWarn(arg, message, args)`
```slug
fn slug.log#cWarn(arg, @str message, ...args) -> ?
```


logs a warn-level message and returns the first argument unchanged.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | @str  | — |
| `args` |  | — |

---

#### `debug(message, args)`
```slug
fn slug.log#debug(@str message, ...args) -> @str|nil
```


logs a debug-level message using the global log level.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | @str  | — |
| `args` |  | — |

---

#### `error(message, args)`
```slug
fn slug.log#error(@str message, ...args) -> @str|nil
```


logs an error-level message using the global log level.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | @str  | — |
| `args` |  | — |

---

#### `info(nil)`
```slug
fn slug.log#info(nil) -> @str|nil
```


logs an info-level message using the global log level.

@effects('io')
nil

---

#### `logger(src)`
```slug
fn slug.log#logger(@str src) -> @map
```


creates a named logger for a specific source module.

The source name is prepended to every log message, making it easy
to filter output by module in multi-component applications.

The log level can be overridden per-source via cfg:
dots in the source name are replaced with underscores.
e.g. `logger("slug.io.http")` reads `cfg("level__slug_io_http")`.

Returns a map of logging functions: `trace`, `debug`, `info`, `warn`,
`error` and their chainable counterparts `cTrace`…`cError`.

```slug
val log = import("slug.log").logger("myapp.worker")
log.info("started {} workers", count)
val result = compute() /> log.cWarn("slow result: {}")
```

| Parameter | Type | Default |
| --- | --- | --- |
| `src` | @str  | — |

---

#### `none(message, args)`
```slug
fn slug.log#none(@str message, ...args) -> nil
```


logs at the `none` level.

Since `none` is above all other levels, this message is only emitted
when the configured level is also `none` — effectively never. Useful
as a placeholder or for temporarily silencing a log call.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | @str  | — |
| `args` |  | — |

---

#### `trace(message, args)`
```slug
fn slug.log#trace(@str message, ...args) -> @str|nil
```


logs a trace-level message using the global log level.

`message` is a `fmt`-style format string; `args` are its arguments.
Emits nothing if the configured level is above `trace`.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | @str  | — |
| `args` |  | — |

---

#### `warn(message, args)`
```slug
fn slug.log#warn(@str message, ...args) -> @str|nil
```


logs a warn-level message using the global log level.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | @str  | — |
| `args` |  | — |
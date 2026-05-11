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
fn slug.log#cDebug(arg, message:str, ...args):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | str | — |
| `args` |  | — |

---

#### `cError(arg, message, args)`
```slug
fn slug.log#cError(arg, message:str, ...args):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | str | — |
| `args` |  | — |

---

#### `cInfo(arg, message, args)`
```slug
fn slug.log#cInfo(arg, message:str, ...args):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | str | — |
| `args` |  | — |

---

#### `cTrace(arg, message, args)`
```slug
fn slug.log#cTrace(arg, message:str, ...args):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | str | — |
| `args` |  | — |

---

#### `cWarn(arg, message, args)`
```slug
fn slug.log#cWarn(arg, message:str, ...args):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `arg` |  | — |
| `message` | str | — |
| `args` |  | — |

---

#### `debug(message, args)`
```slug
fn slug.log#debug(message:str, ...args):str|nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | str | — |
| `args` |  | — |

---

#### `error(message, args)`
```slug
fn slug.log#error(message:str, ...args):str|nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | str | — |
| `args` |  | — |

---

#### `info(nil)`
```slug
fn slug.log#info(nil):str|nil
```
nil

---

#### `logger(src)`
```slug
fn slug.log#logger(src:str):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `src` | str | — |

---

#### `none(message, args)`
```slug
fn slug.log#none(message:str, ...args):nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | str | — |
| `args` |  | — |

---

#### `trace(message, args)`
```slug
fn slug.log#trace(message:str, ...args):str|nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | str | — |
| `args` |  | — |

---

#### `warn(message, args)`
```slug
fn slug.log#warn(message:str, ...args):str|nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | str | — |
| `args` |  | — |
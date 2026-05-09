---
title: stdin (slug.io)
---

## slug.io.stdin

slug.io.stdin — process standard input as a line stream

Exposes process stdin as a shared channel stream for use with
channels and `select`.

### TOC

- [`confirm(message, default)`](#confirmmessage-default)
- [`prompt(prompt)`](#promptprompt)
- [`readLines()`](#readlines)

### Functions

#### `confirm(message, default)`
```slug
fn slug.io.stdin#confirm(@str message, @bool default = false) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | @str  | — |
| `default` | @bool  | `false` |

---

#### `prompt(prompt)`
```slug
fn slug.io.stdin#prompt(@str prompt) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `prompt` | @str  | — |

---

#### `readLines()`
```slug
fn slug.io.stdin#readLines() -> @chan(@str)
```


Returns a shared singleton channel of stdin line events.

Event stream:
- each line is emitted as `@str` (without trailing newline)
- empty lines are preserved as `""`
- a final partial line (without trailing newline) is emitted
- when input ends, the channel closes

With `slug.channel#recv`, closure is observed as `nil`.

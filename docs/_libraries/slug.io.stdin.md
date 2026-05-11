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
- [`readLine()`](#readline)
- [`readLines()`](#readlines)

### Functions

#### `confirm(message, default)`
```slug
fn slug.io.stdin#confirm(message:str, default:bool = false):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `message` | str | — |
| `default` | bool | `false` |

---

#### `prompt(prompt)`
```slug
fn slug.io.stdin#prompt(prompt:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `prompt` | str | — |

---

#### `readLine()`
```slug
fn slug.io.stdin#readLine():any
```

---

#### `readLines()`
```slug
fn slug.io.stdin#readLines():chan(@str)
```


Returns a shared singleton channel of stdin line events.

Event stream:
- each line is emitted as `@str` (without trailing newline)
- empty lines are preserved as `""`
- a final partial line (without trailing newline) is emitted
- when input ends, the channel closes

With `slug.channel#recv`, closure is observed as `nil`.
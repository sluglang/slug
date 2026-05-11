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
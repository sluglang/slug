---
title: stdin (slug.io)
---

## slug.io.stdin

slug.io.stdin — process standard input as a line stream

Exposes process stdin as a shared channel stream for use with
channels and `select`.

### Functions

#### `readLines()`
```slug
fn slug.io.stdin#readLines() -> @chan(@struct(Full)|@struct(Empty))
```


Returns a shared singleton channel of stdin line events.

Event stream:
- each line is emitted as `Full{value: @str}` (without trailing newline)
- empty lines are preserved as `""`
- a final partial line (without trailing newline) is emitted
- when input ends, the channel closes

With `slug.channel#recv`, closure is observed as `Empty`.
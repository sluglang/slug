---
title: control (slug.term)
---

## slug.term.control

slug.term.control — terminal control sequences

Utilities for clearing the terminal display. These functions print
ANSI escape sequences directly to stdout via `print`.

### TOC

- [`clear()`](#clear)
- [`clearAll()`](#clearall)

### Functions

#### `clear()`
```slug
fn slug.term.control#clear():any
```


clears the terminal screen and moves the cursor to the top-left.

Equivalent to the `clear` shell command. Does not clear the scrollback buffer.

---

#### `clearAll()`
```slug
fn slug.term.control#clearAll():any
```


clears the terminal screen, scrollback buffer, and moves the cursor to the top-left.

More thorough than `clear` — also wipes the scrollback history visible
when scrolling up in the terminal.
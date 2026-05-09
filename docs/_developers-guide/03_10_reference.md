# Module 9: Reference

Use this as a quick syntax and precedence lookup.

## Operator precedence (lowest to highest)

| Level | Operators | Associativity |
|---|---|---|
| 1 | `=` | Right |
| 2 | `\|\|` | Left |
| 3 | `&&` | Left |
| 4 | `==` `!=` | Left |
| 5 | `<` `<=` `>` `>=` | Left |
| 6 | `\|` | Left |
| 7 | `^` | Left |
| 8 | `&` | Left |
| 9 | `<<` `>>` | Left |
| 10 | `+` `-` | Left |
| 11 | `*` `/` `%` | Left |
| 12 | `:+` `+:` | Left |
| 13 | prefix `!` `-` `~` | Right |
| 14 | `/>` | Left |
| 15 | calls, indexing, dot access, struct init/copy | Left |

## Quick syntax reminders

- Symbols: `:ok`, `:"content-type"`
- Exact map patterns: `{| :id: id |}`
- Map literals vs blocks are structurally disambiguated in expression position.
- Defer modes are contextual after `defer`: `onsuccess`, `onerror(err)`.

## Minimal examples

```slug
val x = 10 /> fn(v) { v * 2 }
val y = {name: "slug"}.name
val z = [1, 2, 3][1]
```

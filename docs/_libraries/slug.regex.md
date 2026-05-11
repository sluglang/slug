---
title: regex (slug)
---

## slug.regex

slug.regex — regular expression matching and extraction

All functions accept standard regex syntax. Patterns are not compiled
or cached — for performance-sensitive code with repeated matches on
the same pattern, consider pre-processing outside hot paths.

Patterns use standard Go/RE2 syntax (no lookaheads or backreferences).

### TOC

- [`findAll(str, pattern)`](#findallstr-pattern)
- [`findAllGroups(str, pattern)`](#findallgroupsstr-pattern)
- [`indexOf(str, pattern, index)`](#indexofstr-pattern-index)
- [`matches(str, pattern)`](#matchesstr-pattern)
- [`replaceAll(str, pattern, repl)`](#replaceallstr-pattern-repl)
- [`split(str, pattern)`](#splitstr-pattern)

### Functions

#### `findAll(str, pattern)`
```slug
fn slug.regex#findAll(str:str, pattern:str):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `pattern` | str | — |


#### Examples

```slug
findAll("1|2|3", "\d+")  // => ["1", "2", "3"]
findAll("(foo)", "[a-z]+")  // => ["foo"]
```

---

#### `findAllGroups(str, pattern)`
```slug
fn slug.regex#findAllGroups(str:str, pattern:str):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `pattern` | str | — |


#### Examples

```slug
findAllGroups("<a href="foo">bar</a>", "<a href="(.*?)">(.*?)</a>")  // => [["<a href="foo">bar</a>", "foo", "bar"]]
```

---

#### `indexOf(str, pattern, index)`
```slug
fn slug.regex#indexOf(str:str, pattern:str, index:num = 0):[@list, @list]
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `pattern` | str | — |
| `index` | num | `0` |

---

#### `matches(str, pattern)`
```slug
fn slug.regex#matches(str:str, pattern:str):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `pattern` | str | — |


#### Examples

```slug
matches("test", "\d+")  // => false
matches("aaabbb", "b+")  // => true
matches("123", "\d+")  // => true
```

---

#### `replaceAll(str, pattern, repl)`
```slug
fn slug.regex#replaceAll(str:str, pattern:str, repl:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `pattern` | str | — |
| `repl` | str | — |


#### Examples

```slug
replaceAll("1|2|3", "\d+", "x")  // => "x|x|x"
```

---

#### `split(str, pattern)`
```slug
fn slug.regex#split(str:str, pattern:str):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `pattern` | str | — |


#### Examples

```slug
split("1|2|3", "\|")  // => ["1", "2", "3"]
```
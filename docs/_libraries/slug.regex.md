---
title: regex (slug)
---

## slug.regex

slug.regex — regular expression matching and extraction

All functions accept standard regex syntax. Patterns are not compiled
or cached — for performance-sensitive code with repeated matches on
the same pattern, consider pre-processing outside hot paths.

Patterns use standard Go/RE2 syntax (no lookaheads or backreferences).

### Functions

#### `findAll(str, pattern)`
```slug
fn slug.regex#findAll(@str str, @str pattern) -> @list
```


returns a list of all non-overlapping matches of `pattern` in `str`.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `pattern` | @str  | — |


#### Examples

```slug
findAll("1|2|3", "\d+")  // => ["1", "2", "3"]
findAll("(foo)", "[a-z]+")  // => ["foo"]
```

---

#### `findAllGroups(str, pattern)`
```slug
fn slug.regex#findAllGroups(@str str, @str pattern) -> @list
```


returns a list of all matches including capture groups.

Each entry is a list where the first element is the full match and
subsequent elements are the capture group values.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `pattern` | @str  | — |


#### Examples

```slug
findAllGroups("<a href="foo">bar</a>", "<a href="(.*?)">(.*?)</a>")  // => [["<a href="foo">bar</a>", "foo", "bar"]]
```

---

#### `indexOf(str, pattern, index)`
```slug
fn slug.regex#indexOf(@str str, @str pattern, @num index = 0) -> [@list, @list]
```


returns the byte index range of the first match of `pattern` in `str`
starting from `index`, as `[@list start_end, @list group_ranges]`.

Returns `nil` if no match is found. The first list contains the start
and end byte indices of the full match. Subsequent lists contain indices
for capture groups.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `pattern` | @str  | — |
| `index` | @num  | `0` |

---

#### `matches(str, pattern)`
```slug
fn slug.regex#matches(@str str, @str pattern) -> @bool
```


returns true if the pattern matches anywhere in the string.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `pattern` | @str  | — |


#### Examples

```slug
matches("test", "\d+")  // => false
matches("aaabbb", "b+")  // => true
matches("123", "\d+")  // => true
```

---

#### `replaceAll(str, pattern, repl)`
```slug
fn slug.regex#replaceAll(@str str, @str pattern, @str repl) -> @str
```


replaces all matches of `pattern` in `str` with `repl`.

Use `$1`, `$2` etc. in `repl` to reference capture groups.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `pattern` | @str  | — |
| `repl` | @str  | — |


#### Examples

```slug
replaceAll("1|2|3", "\d+", "x")  // => "x|x|x"
```

---

#### `split(str, pattern)`
```slug
fn slug.regex#split(@str str, @str pattern) -> @list
```


splits `str` into a list of substrings divided by matches of `pattern`.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `pattern` | @str  | — |


#### Examples

```slug
split("1|2|3", "\|")  // => ["1", "2", "3"]
```
---
title: regex (slug)
---

## slug.regex

### Functions

#### `findAll(str, pattern)`
```slug
fn slug.regex#findAll(@str str, @str pattern) -> @list
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `pattern` | @str  | — |


#### Examples

```slug
split("1|2|3", "\|")  // => ["1", "2", "3"]
```
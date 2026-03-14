---
title: string (slug)
---

## slug.string

### Functions

#### `camelCase(s, sep)`
```slug
fn slug.string#camelCase(@str s, @str sep = " ") -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s` | @str  | — |
| `sep` | @str  | `" "` |


#### Examples

```slug
camelCase("LAST NAME")  // => "lastName"
camelCase("last name")  // => "lastName"
camelCase("l")  // => "l"
camelCase("l n")  // => "lN"
camelCase("l   n")  // => "lN"
```

---

#### `contains(str, seq)`
```slug
fn slug.string#contains(@str str, @str seq) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `seq` | @str  | — |


#### Examples

```slug
contains("hello slug", "slug")  // => true
contains("hello slug", "snail")  // => false
```

---

#### `endsWith(str, end)`
```slug
fn slug.string#endsWith(@str str, @str end) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `end` | @str  | — |


#### Examples

```slug
endsWith("hello slug", "slug")  // => true
endsWith("hello slug", "hello")  // => false
```

---

#### `fromCodePoint(codePoint)`
```slug
fn slug.string#fromCodePoint(@num codePoint) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `codePoint` | @num  | — |


#### Examples

```slug
fromCodePoint(65)  // => "A"
fromCodePoint(129315)  // => "🤣"
```

---

#### `indexOf(str, seq, index)`
```slug
fn slug.string#indexOf(@str str, @str seq, @num index = 0) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `seq` | @str  | — |
| `index` | @num  | `0` |


#### Examples

```slug
indexOf([1, 2], 2)  // => 1
indexOf([1, 2], 1)  // => 0
indexOf([1, 2], 9)  // => -1
```

---

#### `isLower(str)`
```slug
fn slug.string#isLower(@str str) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |


#### Examples

```slug
isLower(nil)  // => false
isLower("")  // => false
isLower("slug")  // => true
isLower("SLUG")  // => false
```

---

#### `isUpper(str)`
```slug
fn slug.string#isUpper(@str str) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |


#### Examples

```slug
isUpper(nil)  // => false
isUpper("")  // => false
isUpper("slug")  // => false
isUpper("SLUG")  // => true
```

---

#### `join(strs, delimiter, str)`
```slug
fn slug.string#join(@list strs, @str delimiter = "", @str str = nil) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `strs` | @list  | — |
| `delimiter` | @str  | `""` |
| `str` | @str  | `nil` |


#### Examples

```slug
join([], ".")  // => ""
join(["slug"], ".")  // => "slug"
join(["slug", "test"], ".")  // => "slug.test"
```

---

#### `kebabCase(s, sep)`
```slug
fn slug.string#kebabCase(@str s, @str sep = " ") -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s` | @str  | — |
| `sep` | @str  | `" "` |


#### Examples

```slug
kebabCase("LAST NAME")  // => "LAST-NAME"
kebabCase("last name")  // => "last-name"
kebabCase("l")  // => "l"
kebabCase("l n")  // => "l-n"
kebabCase("l   n")  // => "l---n"
```

---

#### `lastIndexOf(str, seq, index, prev)`
```slug
fn slug.string#lastIndexOf(@str str, @str seq, @num index = 0, prev = (-1)) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `seq` | @str  | — |
| `index` | @num  | `0` |
| `prev` |  | `(-1)` |


#### Examples

```slug
lastIndexOf("hello slug", "l")  // => 7
lastIndexOf("hello slug", "h")  // => 0
lastIndexOf("hello slug", "g")  // => 9
```

---

#### `padLeft(str, with, length)`
```slug
fn slug.string#padLeft(@str str, @str with, @num length) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `with` | @str  | — |
| `length` | @num  | — |

---

#### `padRight(str, with, length)`
```slug
fn slug.string#padRight(@str str, @str with, @num length) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `with` | @str  | — |
| `length` | @num  | — |

---

#### `pascalCase(s, sep)`
```slug
fn slug.string#pascalCase(@str s, @str sep = " ") -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s` | @str  | — |
| `sep` | @str  | `" "` |


#### Examples

```slug
pascalCase("LAST NAME")  // => "LastName"
pascalCase("last name")  // => "LastName"
pascalCase("l")  // => "L"
pascalCase("l n")  // => "LN"
pascalCase("l   n")  // => "LN"
```

---

#### `randomHexString(length)`
```slug
fn slug.string#randomHexString(@num length) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `length` | @num  | — |

---

#### `randomString(length, chars, acc)`
```slug
fn slug.string#randomString(@num length, @str chars, @str acc = "") -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `length` | @num  | — |
| `chars` | @str  | — |
| `acc` | @str  | `""` |

---

#### `replaceAll(str, replace, with)`
```slug
fn slug.string#replaceAll(@str str, @str replace, @str with) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `replace` | @str  | — |
| `with` | @str  | — |


#### Examples

```slug
replaceAll(nil, "/", ".")  // => nil
replaceAll("slug/test", "/", ".")  // => "slug.test"
replaceAll("slug", "/", ".")  // => "slug"
replaceAll("E &amp; S", "&amp;", "&")  // => "E & S"
```

---

#### `snakeCase(s, sep, screaming)`
```slug
fn slug.string#snakeCase(@str s, @str sep = " ", screaming = false) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s` | @str  | — |
| `sep` | @str  | `" "` |
| `screaming` |  | `false` |


#### Examples

```slug
snakeCase("LAST NAME", " ", true)  // => "LAST_NAME"
snakeCase("last name", " ", true)  // => "LAST_NAME"
snakeCase("LAST NAME")  // => "LAST_NAME"
snakeCase("last name")  // => "last_name"
snakeCase("l")  // => "l"
snakeCase("l n")  // => "l_n"
snakeCase("l   n")  // => "l___n"
```

---

#### `split(str, delimiter, max, count, strs)`
```slug
fn slug.string#split(@str str, @str delimiter, @num max = (-1), @num count = 1, @list strs = []) -> @list
```


split splits a string into a list of substrings based on a delimiter

Parameters:
- str: The input string to split
- delimiter: The delimiter string to split on
- max: Maximum number of splits to perform (-1 for unlimited)
- count: Internal counter for number of splits performed
- strs: Internal accumulator for storing split strings

Returns:
- Array of substrings split by the delimiter

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `delimiter` | @str  | — |
| `max` | @num  | `(-1)` |
| `count` | @num  | `1` |
| `strs` | @list  | `[]` |


#### Examples

```slug
split("slug/test", "/")  // => ["slug", "test"]
split("éé|éé", "|")  // => ["éé", "éé"]
split("a|b|c", "|", 2)  // => ["a", "b|c"]
```

---

#### `startsWith(str, start)`
```slug
fn slug.string#startsWith(@str str, @str start) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `start` | @str  | — |


#### Examples

```slug
startsWith("hello slug", "slug")  // => false
startsWith("hello slug", "hello")  // => true
```

---

#### `toLower(str)`
```slug
fn slug.string#toLower(@str str) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |


#### Examples

```slug
toLower(nil)  // => nil
toLower("")  // => ""
toLower("SLUG")  // => "slug"
```

---

#### `toUpper(str)`
```slug
fn slug.string#toUpper(@str str) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |


#### Examples

```slug
toUpper(nil)  // => nil
toUpper("")  // => ""
toUpper("slug")  // => "SLUG"
```

---

#### `trim(nil)`
```slug
fn slug.string#trim(nil) -> @str
```
nil


#### Examples

```slug
trim("  a  ", " ")  // => "a"
trim("xxaxxx", "x")  // => "a"
trim("xxa", "x")  // => "a"
trim("axxx", "x")  // => "a"
trim("abxxx", "xx")  // => "abx"
```
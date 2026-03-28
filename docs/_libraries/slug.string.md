---
title: string (slug)
---

## slug.string

slug.string — string manipulation utilities

Trimming, splitting, joining, searching, replacing, case conversion,
padding, case formatting, and random string generation.

All functions treat strings as sequences of Unicode code points where
relevant (e.g. `indexOf` is Unicode-aware).

### TOC

- [`camelCase(s, sep)`](#camelcases-sep)
- [`contains(str, seq)`](#containsstr-seq)
- [`containsAny(str, seq)`](#containsanystr-seq)
- [`endsWith(str, end)`](#endswithstr-end)
- [`fromCodePoint(codePoint)`](#fromcodepointcodepoint)
- [`indexOf(str, seq, index)`](#indexofstr-seq-index)
- [`isLower(str)`](#islowerstr)
- [`isUpper(str)`](#isupperstr)
- [`join(strs, delimiter, str)`](#joinstrs-delimiter-str)
- [`kebabCase(s, sep)`](#kebabcases-sep)
- [`lastIndexOf(str, seq, index, prev)`](#lastindexofstr-seq-index-prev)
- [`padLeft(str, with, length)`](#padleftstr-with-length)
- [`padRight(str, with, length)`](#padrightstr-with-length)
- [`pascalCase(s, sep)`](#pascalcases-sep)
- [`randomHexString(length)`](#randomhexstringlength)
- [`randomString(length, chars, acc)`](#randomstringlength-chars-acc)
- [`replaceAll(str, replace, with)`](#replaceallstr-replace-with)
- [`snakeCase(s, sep, screaming)`](#snakecases-sep-screaming)
- [`split(str, delimiter, max, count, strs)`](#splitstr-delimiter-max-count-strs)
- [`startsWith(str, start)`](#startswithstr-start)
- [`toLower(str)`](#tolowerstr)
- [`toUpper(str)`](#toupperstr)
- [`trim(nil)`](#trimnil)

### Functions

#### `camelCase(s, sep)`
```slug
fn slug.string#camelCase(@str s, @str sep = " ") -> @str
```


converts a string to `camelCase`, capitalising each word except the first.

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


returns true if `seq` appears anywhere in `str`.

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

#### `containsAny(str, seq)`
```slug
fn slug.string#containsAny(@str str, ...seq) -> @bool
```


returns true if any on the `seq` strings appear in `str`.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `seq` |  | — |


#### Examples

```slug
containsAny("hello slug", "slug")  // => true
containsAny("hello slug", "snail", "fish")  // => false
```

---

#### `endsWith(str, end)`
```slug
fn slug.string#endsWith(@str str, @str end) -> @bool
```


returns true if `str` ends with `end`.

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


returns the string corresponding to a Unicode code point.

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


returns the byte index of the first occurrence of `seq` in `str` at or after `index`.

Returns `-1` if not found. Unicode-aware.

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


returns true if `str` is non-empty and entirely lowercase.

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


returns true if `str` is non-empty and entirely uppercase.

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


joins a list of strings with `delimiter` between each element.

Returns an empty string for an empty list.

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


converts a string to `kebab-case` by replacing `sep` with `-`.

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


returns the index of the last occurrence of `seq` in `str`.

Returns `-1` if not found.

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


pads `str` on the left with `with` until it reaches `length`.

Returns `str` unchanged if already at or above `length`.

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


pads `str` on the right with `with` until it reaches `length`.

Returns `str` unchanged if already at or above `length`.

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


converts a string to `PascalCase`, capitalising the first letter of each word.

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


generates a random lowercase hex string of `length` characters.

@effects('random')

| Parameter | Type | Default |
| --- | --- | --- |
| `length` | @num  | — |

**Effects:** `random`

---

#### `randomString(length, chars, acc)`
```slug
fn slug.string#randomString(@num length, @str chars, @str acc = "") -> @str
```


generates a random string of `length` characters drawn from `chars`.

@effects('random')

| Parameter | Type | Default |
| --- | --- | --- |
| `length` | @num  | — |
| `chars` | @str  | — |
| `acc` | @str  | `""` |

**Effects:** `random`

---

#### `replaceAll(str, replace, with)`
```slug
fn slug.string#replaceAll(@str str, @str replace, @str with) -> @str
```


replaces all occurrences of `replace` in `str` with `with`.

Returns `nil` for `nil` input.

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


converts a string to `snake_case` by replacing `sep` with `_`.

Pass `screaming: true` for `SCREAMING_SNAKE_CASE`.

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


splits `str` on `delimiter`, returning a list of substrings.

`max` limits the number of splits performed; `-1` means unlimited.
The remaining unsplit portion is always included as the last element.

`count` and `strs` are internal accumulators — do not pass them directly.

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


returns true if `str` begins with `start`.

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


converts a string to lowercase.

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


converts a string to uppercase.

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


removes all leading and trailing occurrences of `prefix` from `s`.

Repeatedly strips the prefix/suffix until neither end matches.
Note: `prefix` is stripped from both ends, not just the left.
nil


#### Examples

```slug
trim("  a  ", " ")  // => "a"
trim("xxaxxx", "x")  // => "a"
trim("xxa", "x")  // => "a"
trim("axxx", "x")  // => "a"
trim("abxxx", "xx")  // => "abx"
```
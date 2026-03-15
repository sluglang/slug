---
title: list (slug)
---

## slug.list

slug.list — list utilities

Sorting, searching, flattening, shuffling, and character conversion
helpers for `@list` values. Complements the core list operations in
`slug.std` (`map`, `filter`, `reduce`, `find`, `flatMap`, etc.).

### Functions

#### `asList(chars, i, acc)`
```slug
fn slug.list#asList(@str chars, @num i = 0, @list acc = []) -> @list
```


converts a string into a list of single-character strings.

Returns an empty list for `nil` or an empty string.

| Parameter | Type | Default |
| --- | --- | --- |
| `chars` | @str  | — |
| `i` | @num  | `0` |
| `acc` | @list  | `[]` |


#### Examples

```slug
asList(nil)  // => []
asList("")  // => []
asList("123")  // => ["1", "2", "3"]
```

---

#### `flatten(lsts)`
```slug
fn slug.list#flatten(@list lsts) -> @list
```


flattens a list of lists into a single list.

Only one level deep — nested lists within the inner lists are not
recursively flattened.

| Parameter | Type | Default |
| --- | --- | --- |
| `lsts` | @list  | — |


#### Examples

```slug
flatten([])  // => []
flatten([1])  // => [1]
flatten([[1, 2], [3]])  // => [1, 2, 3]
flatten([[1, 2], [3], [4, 5]])  // => [1, 2, 3, 4, 5]
```

---

#### `indexOf(list, value, idx)`
```slug
fn slug.list#indexOf(@list list, value, @num idx = 0) -> @num
```


returns the index of the first occurrence of `value` in `list`.

Returns `-1` if the value is not found. Uses value equality (`==`).

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | @list  | — |
| `value` |  | — |
| `idx` | @num  | `0` |


#### Examples

```slug
indexOf("hello slug", "lu")  // => 7
indexOf("hello slug", "l")  // => 2
indexOf("hello slug", "l", 3)  // => 3
indexOf("hello slug", "l", 4)  // => 7
indexOf("éé|éé", "|")  // => 2
```

---

#### `removeValue(list, value)`
```slug
fn slug.list#removeValue(@list list, value) -> @list
```


returns a new list with the first occurrence of `value` removed.

Returns the original list unchanged if the value is not present.

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | @list  | — |
| `value` |  | — |


#### Examples

```slug
removeValue([1, 2, 3], 2)  // => [1, 3]
removeValue([1, 2, 3], 5)  // => [1, 2, 3]
```

---

#### `shuffle(list)`
```slug
fn slug.list#shuffle(@list list) -> @list
```


returns a new list with elements shuffled in random order.

Uses the Fisher-Yates algorithm. The original list is not modified.

@effects('random')

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | @list  | — |

**Effects:** `random`

---

#### `sort(lst)`
```slug
fn slug.list#sort(@list lst) -> @list
```


sorts a list in natural ascending order.

Uses `slug.std#compare` for ordering. Works correctly for lists of
numbers or lists of strings. For mixed or custom types, use
`sortWithComparator` instead.

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | @list  | — |

---

#### `sortWithComparator(lst, comparator)`
```slug
fn slug.list#sortWithComparator(@list lst, @fn comparator) -> @list
```


sorts a list using a custom comparator function.

The comparator receives two elements and must return a negative number
if the first should come before the second, zero if equal, or a positive
number if the first should come after.

Use `slug.std#compare` as the comparator for natural ordering of
numbers and strings.

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | @list  | — |
| `comparator` | @fn  | — |


#### Examples

```slug
sortWithComparator([3, 1, 2], function group: [{|| 2 2 false} => fn((a), (b)) {
{(a - b)}
}])  // => [1, 2, 3]
sortWithComparator(["c", "a", "b"], function group: [{|| 2 2 false} => fn((a), (b)) {
{if(a < b) {(-1)} else {if(a > b) {1} else {0}}}
}])  // => ["a", "b", "c"]
```
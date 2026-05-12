---
title: list (slug)
---

## slug.list

slug.list — list utilities

Sorting, searching, flattening, shuffling, and character conversion
helpers for `list` values. Complements the core list operations in
`slug.std` (`map`, `filter`, `reduce`, `find`, `flatMap`, etc.).

### TOC

- [`asList(chars, i, acc)`](#aslistchars-i-acc)
- [`flatten<T>(lsts)`](#flattentlsts)
- [`indexOf<T>(list, value, idx)`](#indexoftlist-value-idx)
- [`removeValue<T>(list, value)`](#removevaluetlist-value)
- [`shuffle<T>(list)`](#shuffletlist)
- [`sort<T>(lst)`](#sorttlst)
- [`sortWithComparator(lst, comparator)`](#sortwithcomparatorlst-comparator)

### Functions

#### `asList(chars, i, acc)`
```slug
fn slug.list#asList(chars:str, i:num = 0, acc:list<str> = []):list<str>
```


converts a string into a list of single-character strings.

Returns an empty list for `nil` or an empty string.

| Parameter | Type | Default |
| --- | --- | --- |
| `chars` | str | — |
| `i` | num | `0` |
| `acc` | list<str> | `[]` |


#### Examples

```slug
asList(nil)  // => []
asList("")  // => []
asList("123")  // => ["1", "2", "3"]
```

---

#### `flatten<T>(lsts)`
```slug
fn slug.list#flatten<T>(lsts:list<list<T>|T>):list<T>
```


flattens a list of lists into a single list.

Only one level deep — nested lists within the inner lists are not
recursively flattened.

| Parameter | Type | Default |
| --- | --- | --- |
| `lsts` | list<list<T> \| T> | — |


#### Examples

```slug
flatten([])  // => []
flatten([1])  // => [1]
flatten([[1, 2], [3]])  // => [1, 2, 3]
flatten([[1, 2], [3], [4, 5]])  // => [1, 2, 3, 4, 5]
```

---

#### `indexOf<T>(list, value, idx)`
```slug
fn slug.list#indexOf<T>(list:list<T>, value:T, idx:num = 0):num
```


returns the index of the first occurrence of `value` in `list`.

Returns `-1` if the value is not found. Uses value equality (`==`).

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | list<T> | — |
| `value` | T | — |
| `idx` | num | `0` |


#### Examples

```slug
indexOf([1, 2], 2)  // => 1
indexOf([1, 2], 1)  // => 0
indexOf([1, 2], 9)  // => -1
```

---

#### `removeValue<T>(list, value)`
```slug
fn slug.list#removeValue<T>(list:list<T>, value:T):list<T>
```


returns a new list with the first occurrence of `value` removed.

Returns the original list unchanged if the value is not present.

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | list<T> | — |
| `value` | T | — |


#### Examples

```slug
removeValue([1, 2, 3], 2)  // => [1, 3]
removeValue([1, 2, 3], 5)  // => [1, 2, 3]
```

---

#### `shuffle<T>(list)`
```slug
fn slug.list#shuffle<T>(list:list<T>):list<T>
```


returns a new list with elements shuffled in random order.

Uses the Fisher-Yates algorithm. The original list is not modified.

@effects('random')

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | list<T> | — |

**Effects:** `random`

---

#### `sort<T>(lst)`
```slug
fn slug.list#sort<T>(lst:list<T>):list<T>
```


sorts a list in natural ascending order.

Uses `slug.std#compare` for ordering. Works correctly for lists of
numbers or lists of strings. For mixed or custom types, use
`sortWithComparator` instead.

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | list<T> | — |

---

#### `sortWithComparator(lst, comparator)`
```slug
fn slug.list#sortWithComparator(lst:list<T>, comparator:fn<num, T>):list<T>
```


sorts a list using a custom comparator function.

The comparator receives two elements and must return a negative number
if the first should come before the second, zero if equal, or a positive
number if the first should come after.

Use `slug.std#compare` as the comparator for natural ordering of
numbers and strings.

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | list<T> | — |
| `comparator` | fn<num, T> | — |


#### Examples

```slug
sortWithComparator([3, 1, 2], function group: [{|| 2 2 false} => fn(a, b) { <vm bytecode> }])  // => [1, 2, 3]
sortWithComparator(["c", "a", "b"], function group: [{|| 2 2 false} => fn(a, b) { <vm bytecode> }])  // => ["a", "b", "c"]
```
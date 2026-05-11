---
title: list (slug)
---

## slug.list

slug.list — list utilities

Sorting, searching, flattening, shuffling, and character conversion
helpers for `@list` values. Complements the core list operations in
`slug.std` (`map`, `filter`, `reduce`, `find`, `flatMap`, etc.).

### TOC

- [`asList(chars, i, acc)`](#aslistchars-i-acc)
- [`flatten(lsts)`](#flattenlsts)
- [`indexOf(list, value, idx)`](#indexoflist-value-idx)
- [`removeValue(list, value)`](#removevaluelist-value)
- [`shuffle(list)`](#shufflelist)
- [`sort(lst)`](#sortlst)
- [`sortWithComparator(lst, comparator)`](#sortwithcomparatorlst-comparator)

### Functions

#### `asList(chars, i, acc)`
```slug
fn slug.list#asList(chars:str, i:num = 0, acc:list = []):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `chars` | str | — |
| `i` | num | `0` |
| `acc` | list | `[]` |


#### Examples

```slug
asList(nil)  // => []
asList("")  // => []
asList("123")  // => ["1", "2", "3"]
```

---

#### `flatten(lsts)`
```slug
fn slug.list#flatten(lsts:list):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lsts` | list | — |


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
fn slug.list#indexOf(list:list, value, idx:num = 0):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | list | — |
| `value` |  | — |
| `idx` | num | `0` |


#### Examples

```slug
indexOf([1, 2], 2)  // => 1
indexOf([1, 2], 1)  // => 0
indexOf([1, 2], 9)  // => -1
```

---

#### `removeValue(list, value)`
```slug
fn slug.list#removeValue(list:list, value):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | list | — |
| `value` |  | — |


#### Examples

```slug
removeValue([1, 2, 3], 2)  // => [1, 3]
removeValue([1, 2, 3], 5)  // => [1, 2, 3]
```

---

#### `shuffle(list)`
```slug
fn slug.list#shuffle(list:list):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | list | — |

**Effects:** `random`

---

#### `sort(lst)`
```slug
fn slug.list#sort(lst:list):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | list | — |

---

#### `sortWithComparator(lst, comparator)`
```slug
fn slug.list#sortWithComparator(lst:list, comparator:fn):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | list | — |
| `comparator` | fn | — |


#### Examples

```slug
sortWithComparator([3, 1, 2], fn(a, b) { <vm bytecode> })  // => [1, 2, 3]
sortWithComparator(["c", "a", "b"], fn(a, b) { <vm bytecode> })  // => ["a", "b", "c"]
```
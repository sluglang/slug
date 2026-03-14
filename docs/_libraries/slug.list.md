---
title: list (slug)
---

## slug.list

### Functions

#### `asList(chars, i, acc)`
```slug
fn slug.list#asList(@str chars, @num i = 0, @list acc = []) -> @list
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | @list  | — |

---

#### `sort(lst)`
```slug
fn slug.list#sort(@list lst) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | @list  | — |

---

#### `sortWithComparator(lst, comparator)`
```slug
fn slug.list#sortWithComparator(@list lst, @fn comparator) -> @list
```

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
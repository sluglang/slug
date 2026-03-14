---
title: std (slug)
---

## slug.std

the slug standard library

### Constants

#### `BOOLEAN_TYPE`

```slug
sym slug.std#BOOLEAN_TYPE
```

#### `BYTES_TYPE`

```slug
sym slug.std#BYTES_TYPE
```

#### `ERROR_TYPE`

```slug
sym slug.std#ERROR_TYPE
```

#### `FUNCTION_TYPE`

```slug
sym slug.std#FUNCTION_TYPE
```

#### `LIST_TYPE`

```slug
sym slug.std#LIST_TYPE
```

#### `MAP_TYPE`

```slug
sym slug.std#MAP_TYPE
```

#### `NIL_TYPE`

```slug
sym slug.std#NIL_TYPE
```

Constant for NIL_TYPE, `type`

#### `NUMBER_TYPE`

```slug
sym slug.std#NUMBER_TYPE
```

#### `STRING_TYPE`

```slug
sym slug.std#STRING_TYPE
```

#### `STRUCT_TYPE`

```slug
sym slug.std#STRUCT_TYPE
```

#### `SYMBOL_TYPE`

```slug
sym slug.std#SYMBOL_TYPE
```

#### `TASK_TYPE`

```slug
sym slug.std#TASK_TYPE
```

### Structs

#### `Error`
```slug
struct slug.std#Error{@str type = "Error", @str msg, code = nil, data = nil, cause = nil}
```


Standard error payload shape for idiomatic error handling

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `type` | @str  | `"Error"` |  |
| `msg` | @str  | — |  |
| `code` |  | `nil` |  |
| `data` |  | `nil` |  |
| `cause` |  | `nil` |  |

### Functions

#### `compare(a, b)`
```slug
fn slug.std#compare(a, b) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `b` |  | — |

---

#### `compute(map, key, f)`
```slug
fn slug.std#compute(@map map, key, f) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | @map  | — |
| `key` |  | — |
| `f` |  | — |


#### Examples

```slug
compute({:k: 1}, :k, function group: [{|| 2 2 false} => fn((k), (v)) {
{(v + 1)}
}])  // => {:k: 2}
compute({}, :k, function group: [{|| 2 2 false} => fn((k), (v)) {
{(v == nil)}
}])  // => {:k: true}
```

---

#### `counter(start)`
```slug
fn slug.std#counter(@num start = 0) -> @fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `start` | @num  | `0` |

---

#### `equals(nil)`
```slug
fn slug.std#equals(nil) -> @bool
```
nil


#### Examples

```slug
equals(nil, nil)  // => false
equals({}, nil)  // => false
equals(nil, {})  // => false
equals({}, {})  // => true
equals({:k1: 1}, {:k1: 1})  // => true
equals({:k1: 1}, {})  // => false
equals({:k1: 1}, {:k2: 2})  // => false
```

---

#### `filter(vs, f, acc)`
```slug
fn slug.std#filter(@list vs, @fn f, @list acc = []) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vs` | @list  | — |
| `f` | @fn  | — |
| `acc` | @list  | `[]` |


#### Examples

```slug
filter([1, 2, 3, 4], function group: [{| 1 1 false} => fn((v)) {
{((v % 2) == 0)}
}])  // => [2, 4]
```

---

#### `find(xs, f)`
```slug
fn slug.std#find(@list xs, @fn f) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` | @list  | — |
| `f` | @fn  | — |

---

#### `flatMap(vs, f)`
```slug
fn slug.std#flatMap(@list vs, @fn f) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vs` | @list  | — |
| `f` | @fn  | — |


#### Examples

```slug
flatMap([1, 2, 3], function group: [{| 1 1 false} => fn((n)) {
{[n, n]}
}])  // => [1, 1, 2, 2, 3, 3]
flatMap([1, 2, 3], function group: [{| 1 1 false} => fn((n)) {
{if((n % 2) == 0) {[n]} else {[]}}
}])  // => [2]
```

---

#### `fmt(str, args)`
```slug
fn slug.std#fmt(@str str, ...args) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `args` |  | — |


#### Examples

```slug
fmt("Hello {}", "Slug")  // => "Hello Slug"
fmt("{1} then {0}", "A", "B")  // => "B then A"
fmt("x={:.2f} y={:.2f}", 1.2, 3.4)  // => "x=1.20 y=3.40"
fmt("{:d}", 12.5)  // => "12"
fmt("{:,}", 1234567)  // => "1,234,567"
fmt("{:.1%}", 0.1234)  // => "12.3%"
fmt("value=\{x\}")  // => "value={x}"
fmt("|{:>8}|", 12.3)  // => "|    12.3|"
fmt("|{:<10s}|", "Slug")  // => "|Slug      |"
fmt("|{:^10s}|", "Slug")  // => "|   Slug   |"
```

---

#### `get(map, key)`
```slug
fn slug.std#get(@map map, key) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | @map  | — |
| `key` |  | — |


#### Examples

```slug
get({}, :k)  // => nil
get({:k: 1}, :k)  // => 1
```

---

#### `ifNil(nil)`
```slug
fn slug.std#ifNil(nil) -> ?
```
nil

---

#### `isDefined(varName)`
```slug
fn slug.std#isDefined(@str varName) -> @bool
```


Determine if a variable is defined in the current scope.

| Parameter | Type | Default |
| --- | --- | --- |
| `varName` | @str  | — |


#### Examples

```slug
isDefined("type")  // => true
isDefined("__not_defined__")  // => false
```

---

#### `isStructInstance(v)`
```slug
fn slug.std#isStructInstance(v) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `keys(map)`
```slug
fn slug.std#keys(map) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` |  | — |


#### Examples

```slug
keys({})  // => []
keys({:k: 1})  // => [:k]
```

---

#### `label(symbol)`
```slug
fn slug.std#label(symbol) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `symbol` |  | — |


#### Examples

```slug
label(:foo)  // => "foo"
label(:"a b")  // => "a b"
```

---

#### `map(vs, f, acc)`
```slug
fn slug.std#map(@list vs, @fn f, acc = []) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vs` | @list  | — |
| `f` | @fn  | — |
| `acc` |  | `[]` |


#### Examples

```slug
map([1, 2], function group: [{| 1 1 false} => fn((n)) {
{(n * 2)}
}])  // => [2, 4]
```

---

#### `nonNil(v, f, default)`
```slug
fn slug.std#nonNil(v, @fn f, default = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |
| `f` | @fn  | — |
| `default` |  | `nil` |

---

#### `parseNumber(value)`
```slug
fn slug.std#parseNumber(@str value) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` | @str  | — |


#### Examples

```slug
parseNumber("1")  // => 1
parseNumber("1.1")  // => 1.1
```

---

#### `put(map, key, value)`
```slug
fn slug.std#put(@map map, key, value) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | @map  | — |
| `key` |  | — |
| `value` |  | — |


#### Examples

```slug
put({}, :k, "v")  // => {:k: v}
```

---

#### `range(start, end, step, acc)`
```slug
fn slug.std#range(@num start, @num end, @num step = 1, @list acc = []) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `start` | @num  | — |
| `end` | @num  | — |
| `step` | @num  | `1` |
| `acc` | @list  | `[]` |


#### Examples

```slug
range(0, 0)  // => []
range(0, 2)  // => [0, 1]
range(0, 6, 2)  // => [0, 2, 4]
range(0, 6, -2)  // => []
range(6, 0, -2)  // => [6, 4, 2]
```

---

#### `reduce(vs, v, f)`
```slug
fn slug.std#reduce(@list vs, v, @fn f) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vs` | @list  | — |
| `v` |  | — |
| `f` | @fn  | — |


#### Examples

```slug
reduce([], 0, function group: [{|| 2 2 false} => fn((a), (b)) {
{(a + b)}
}])  // => 0
reduce([1, 2, 3], 0, function group: [{|| 2 2 false} => fn((a), (b)) {
{(a + b)}
}])  // => 6
reduce([1, 2, 3], 9, function group: [{|| 2 2 false} => fn((a), (b)) {
{(a + b)}
}])  // => 15
```

---

#### `remove(map, key)`
```slug
fn slug.std#remove(@map map, key) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | @map  | — |
| `key` |  | — |


#### Examples

```slug
remove({}, :k)  // => {}
remove({:k: 1}, :k)  // => {}
remove({:k: 1}, :j)  // => {:k: 1}
```

---

#### `structEquals(m1, m2)`
```slug
fn slug.std#structEquals(m1, m2) -> @bool
```


deep equals check for struct types and struct instances

| Parameter | Type | Default |
| --- | --- | --- |
| `m1` |  | — |
| `m2` |  | — |

---

#### `swap(list, index1, index2)`
```slug
fn slug.std#swap(@list list, @num index1, @num index2) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | @list  | — |
| `index1` | @num  | — |
| `index2` | @num  | — |


#### Examples

```slug
swap([1, 2], 0, 1)  // => [2, 1]
```

---

#### `sym(name)`
```slug
fn slug.std#sym(@str name) -> @sym
```

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | @str  | — |


#### Examples

```slug
sym("foo")  // => :foo
sym("foo bar")  // => :"foo bar"
sym("Content-Type")  // => :"Content-Type"
```

---

#### `then(m, f)`
```slug
fn slug.std#then(m, @fn f) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `m` |  | — |
| `f` | @fn  | — |


#### Examples

```slug
then(1, function group: [{| 1 1 false} => fn((v)) {
{(v * 2)}
}])  // => 2
then({:n: Slug}, function group: [{| 1 1 false} => fn((v)) {
{(v[:n])}
}])  // => "Slug"
```

---

#### `toBoolean(v)`
```slug
fn slug.std#toBoolean(v) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

**Throws:** `@struct(Error{type:TypeError})`


#### Examples

```slug
toBoolean(nil)  // => false
toBoolean(0)  // => false
toBoolean(1)  // => true
toBoolean(true)  // => true
toBoolean(false)  // => false
toBoolean("true")  // => true
toBoolean("yes")  // => true
toBoolean("1")  // => true
toBoolean("false")  // => false
toBoolean("no")  // => false
toBoolean("0")  // => false
```

---

#### `toNumber(v)`
```slug
fn slug.std#toNumber(v) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

**Throws:** `@struct(Error{type:TypeError})`


#### Examples

```slug
toNumber(nil)  // => nil
toNumber(1)  // => 1
toNumber(1.1)  // => 1.1
toNumber("1")  // => 1
toNumber("1.1")  // => 1.1
```

---

#### `toString(v)`
```slug
fn slug.std#toString(v) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |


#### Examples

```slug
toString(nil)  // => nil
toString("str")  // => "str"
toString(1.1)  // => "1.1"
toString(true)  // => "true"
```

---

#### `type(val)`
```slug
fn slug.std#type(val) -> @sym
```


return a string value indicating the type of `val`

| Parameter | Type | Default |
| --- | --- | --- |
| `val` |  | — |


#### Examples

```slug
type(nil)  // => :nil
type(true)  // => :bool
type(1)  // => :number
type(1.1)  // => :number
type("Hello Slug!")  // => :string
type([1, 2])  // => :list
type({:key: value})  // => :map
type(0x"ff")  // => :bytes
type(function group: [{| 1 1 false} => fn((a)) {
{a}
}])  // => :function
```

---

#### `update(list, index, value)`
```slug
fn slug.std#update(@list list, @num index, value) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | @list  | — |
| `index` | @num  | — |
| `value` |  | — |


#### Examples

```slug
update([1, 2, 3], 1, 99)  // => [1, 99, 3]
```

---

#### `zeroIfAbove(a, b)`
```slug
fn slug.std#zeroIfAbove(@num a, @num b) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` | @num  | — |
| `b` | @num  | — |


#### Examples

```slug
zeroIfAbove(1, 1)  // => 0
zeroIfAbove(1, 2)  // => 1
zeroIfAbove(2, 1)  // => 0
```

---

#### `zip(lst1, lst2, acc)`
```slug
fn slug.std#zip(@list lst1, @list lst2, acc = []) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst1` | @list  | — |
| `lst2` | @list  | — |
| `acc` |  | `[]` |


#### Examples

```slug
zip([], [])  // => []
zip([1], [])  // => []
zip([], [1])  // => []
zip([1], [2])  // => [[1, 2]]
```

---

#### `zipWith(lst, f)`
```slug
fn slug.std#zipWith(@list lst, f) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | @list  | — |
| `f` |  | — |


#### Examples

```slug
zipWith(["a", "b"], function group: [{ 0 0 false} => fn() {
{1}
}])  // => [["a", 1], ["b", 1]]
```

---

#### `zipWithIndex(lst)`
```slug
fn slug.std#zipWithIndex(@list lst) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | @list  | — |


#### Examples

```slug
zipWithIndex([])  // => []
zipWithIndex(["a", "b"])  // => [["a", 0], ["b", 1]]
```
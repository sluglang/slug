---
title: std (slug)
---

## slug.std

slug.std — the Slug standard library

Core utilities available to all Slug programs. Provides type constants,
the canonical `Error` struct, collection transforms, type conversions,
string formatting, and common functional helpers.

## Type constants

Type symbols returned by `type()` are available as named constants:
`NIL_TYPE`, `BOOLEAN_TYPE`, `NUMBER_TYPE`, `STRING_TYPE`, `LIST_TYPE`,
`MAP_TYPE`, `BYTES_TYPE`, `FUNCTION_TYPE`, `TASK_TYPE`, `STRUCT_TYPE`,
`SYMBOL_TYPE`, `ERROR_TYPE`.

Use these with pinned patterns for readable type dispatch:
```slug
match type(v) {
  ^STRING_TYPE => ...
  ^LIST_TYPE   => ...
  _            => ...
}
```

## Error handling

The canonical error shape is `Error{ type, msg, code, data, cause }`.
Construct errors with `Error{ type: "MyError", msg: "something went wrong" }`
and throw them with `throw`. There is no try/catch — handle errors via
`match` on return values or `defer onerror`.

### TOC

- [BOOLEAN_TYPE](#boolean_type)
- [BYTES_TYPE](#bytes_type)
- [ERROR_TYPE](#error_type)
- [FUNCTION_TYPE](#function_type)
- [LIST_TYPE](#list_type)
- [MAP_TYPE](#map_type)
- [NIL_TYPE](#nil_type)
- [NUMBER_TYPE](#number_type)
- [STRING_TYPE](#string_type)
- [STRUCT_TYPE](#struct_type)
- [SYMBOL_TYPE](#symbol_type)
- [TASK_TYPE](#task_type)
- [Error](#error)
- [`compare(a, b)`](#comparea-b)
- [`compute(map, key, f)`](#computemap-key-f)
- [`counter(start)`](#counterstart)
- [`equals(nil)`](#equalsnil)
- [`filter(vs, f, acc)`](#filtervs-f-acc)
- [`find(xs, f)`](#findxs-f)
- [`flatMap(vs, f)`](#flatmapvs-f)
- [`fmt(str, args)`](#fmtstr-args)
- [`get(map, key)`](#getmap-key)
- [`ifNil(nil)`](#ifnilnil)
- [`isDefined(varName)`](#isdefinedvarname)
- [`isStructInstance(v)`](#isstructinstancev)
- [`keys(map)`](#keysmap)
- [`label(symbol)`](#labelsymbol)
- [`map(vs, f, acc)`](#mapvs-f-acc)
- [`moduleName()`](#modulename)
- [`nonNil(v, f, default)`](#nonnilv-f-default)
- [`parseNumber(value)`](#parsenumbervalue)
- [`put(map, key, value)`](#putmap-key-value)
- [`range(start, end, step, acc)`](#rangestart-end-step-acc)
- [`reduce(vs, v, f)`](#reducevs-v-f)
- [`remove(map, key)`](#removemap-key)
- [`structEquals(m1, m2)`](#structequalsm1-m2)
- [`swap(list, index1, index2)`](#swaplist-index1-index2)
- [`sym(name)`](#symname)
- [`then(m, f)`](#thenm-f)
- [`toBoolean(v)`](#tobooleanv)
- [`toNumber(v)`](#tonumberv)
- [`toString(v)`](#tostringv)
- [`type(val)`](#typeval)
- [`update(list, index, value)`](#updatelist-index-value)
- [`zeroIfAbove(a, b)`](#zeroifabovea-b)
- [`zip(lst1, lst2, acc)`](#ziplst1-lst2-acc)
- [`zipWith(lst, f)`](#zipwithlst-f)
- [`zipWithIndex(lst)`](#zipwithindexlst)

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
struct slug.std#Error{type:str = "Error", msg:str, code = nil, data = nil, cause = nil}
```

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `type` | str | `"Error"` |  |
| `msg` | str | — |  |
| `code` |  | `nil` |  |
| `data` |  | `nil` |  |
| `cause` |  | `nil` |  |

### Functions

#### `compare(a, b)`
```slug
fn slug.std#compare(a, b):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `b` |  | — |

---

#### `compute(map, key, f)`
```slug
fn slug.std#compute(map:map, key, f):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | map | — |
| `key` |  | — |
| `f` |  | — |


#### Examples

```slug
compute({:k: 1}, :k, fn(k, v) { <vm bytecode> })  // => {:k: 2}
compute({}, :k, fn(k, v) { <vm bytecode> })  // => {:k: true}
```

---

#### `counter(start)`
```slug
fn slug.std#counter(start:num = 0):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `start` | num | `0` |

---

#### `equals(nil)`
```slug
fn slug.std#equals(nil):bool
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
fn slug.std#filter(vs:list, f:fn, acc:list = []):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vs` | list | — |
| `f` | fn | — |
| `acc` | list | `[]` |


#### Examples

```slug
filter([1, 2, 3, 4], fn(v) { <vm bytecode> })  // => [2, 4]
```

---

#### `find(xs, f)`
```slug
fn slug.std#find(xs:list, f:fn):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` | list | — |
| `f` | fn | — |

---

#### `flatMap(vs, f)`
```slug
fn slug.std#flatMap(vs:list, f:fn):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vs` | list | — |
| `f` | fn | — |


#### Examples

```slug
flatMap([1, 2, 3], fn(n) { <vm bytecode> })  // => [1, 1, 2, 2, 3, 3]
flatMap([1, 2, 3], fn(n) { <vm bytecode> })  // => [2]
```

---

#### `fmt(str, args)`
```slug
fn slug.std#fmt(str:str, ...args):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
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
fn slug.std#get(map:map, key):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | map | — |
| `key` |  | — |


#### Examples

```slug
get({}, :k)  // => nil
get({:k: 1}, :k)  // => 1
```

---

#### `ifNil(nil)`
```slug
fn slug.std#ifNil(nil):any
```
nil

---

#### `isDefined(varName)`
```slug
fn slug.std#isDefined(varName:str):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `varName` | str | — |


#### Examples

```slug
isDefined("type")  // => true
isDefined("__not_defined__")  // => false
```

---

#### `isStructInstance(v)`
```slug
fn slug.std#isStructInstance(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `keys(map)`
```slug
fn slug.std#keys(map):list
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
fn slug.std#label(symbol):str
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
fn slug.std#map(vs:list, f:fn, acc = []):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vs` | list | — |
| `f` | fn | — |
| `acc` |  | `[]` |


#### Examples

```slug
map([1, 2], fn(n) { <vm bytecode> })  // => [2, 4]
```

---

#### `moduleName()`
```slug
fn slug.std#moduleName():str
```

---

#### `nonNil(v, f, default)`
```slug
fn slug.std#nonNil(v, f:fn, default = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |
| `f` | fn | — |
| `default` |  | `nil` |

---

#### `parseNumber(value)`
```slug
fn slug.std#parseNumber(value:str):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` | str | — |


#### Examples

```slug
parseNumber("1")  // => 1
parseNumber("1.1")  // => 1.1
```

---

#### `put(map, key, value)`
```slug
fn slug.std#put(map:map, key, value):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | map | — |
| `key` |  | — |
| `value` |  | — |


#### Examples

```slug
put({}, :k, "v")  // => {:k: v}
```

---

#### `range(start, end, step, acc)`
```slug
fn slug.std#range(start:num, end:num, step:num = 1, acc:list = []):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `start` | num | — |
| `end` | num | — |
| `step` | num | `1` |
| `acc` | list | `[]` |


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
fn slug.std#reduce(vs:list, v, f:fn):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vs` | list | — |
| `v` |  | — |
| `f` | fn | — |


#### Examples

```slug
reduce([], 0, fn(a, b) { <vm bytecode> })  // => 0
reduce([1, 2, 3], 0, fn(a, b) { <vm bytecode> })  // => 6
reduce([1, 2, 3], 9, fn(a, b) { <vm bytecode> })  // => 15
```

---

#### `remove(map, key)`
```slug
fn slug.std#remove(map:map, key):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | map | — |
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
fn slug.std#structEquals(m1, m2):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `m1` |  | — |
| `m2` |  | — |

---

#### `swap(list, index1, index2)`
```slug
fn slug.std#swap(list:list, index1:num, index2:num):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | list | — |
| `index1` | num | — |
| `index2` | num | — |


#### Examples

```slug
swap([1, 2], 0, 1)  // => [2, 1]
```

---

#### `sym(name)`
```slug
fn slug.std#sym(name:str):sym
```

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | str | — |


#### Examples

```slug
sym("foo")  // => :foo
sym("foo bar")  // => :"foo bar"
sym("Content-Type")  // => :"Content-Type"
```

---

#### `then(m, f)`
```slug
fn slug.std#then(m, f:fn):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `m` |  | — |
| `f` | fn | — |


#### Examples

```slug
then(1, fn(v) { <vm bytecode> })  // => 2
then({:n: Slug}, fn(v) { <vm bytecode> })  // => "Slug"
```

---

#### `toBoolean(v)`
```slug
fn slug.std#toBoolean(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

**Throws:** `Error{type:TypeError}`


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
fn slug.std#toNumber(v):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

**Throws:** `Error{type:TypeError}`


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
fn slug.std#toString(v):str
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
fn slug.std#type(val):sym
```

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
type(fn(a) { <vm bytecode> })  // => :function
```

---

#### `update(list, index, value)`
```slug
fn slug.std#update(list:list, index:num, value):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `list` | list | — |
| `index` | num | — |
| `value` |  | — |


#### Examples

```slug
update([1, 2, 3], 1, 99)  // => [1, 99, 3]
```

---

#### `zeroIfAbove(a, b)`
```slug
fn slug.std#zeroIfAbove(a:num, b:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` | num | — |
| `b` | num | — |


#### Examples

```slug
zeroIfAbove(1, 1)  // => 0
zeroIfAbove(1, 2)  // => 1
zeroIfAbove(2, 1)  // => 0
```

---

#### `zip(lst1, lst2, acc)`
```slug
fn slug.std#zip(lst1:list, lst2:list, acc = []):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst1` | list | — |
| `lst2` | list | — |
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
fn slug.std#zipWith(lst:list, f):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | list | — |
| `f` |  | — |


#### Examples

```slug
zipWith(["a", "b"], fn() { <vm bytecode> })  // => [["a", 1], ["b", 1]]
```

---

#### `zipWithIndex(lst)`
```slug
fn slug.std#zipWithIndex(lst:list):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `lst` | list | — |


#### Examples

```slug
zipWithIndex([])  // => []
zipWithIndex(["a", "b"])  // => [["a", 0], ["b", 1]]
```
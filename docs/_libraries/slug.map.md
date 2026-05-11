---
title: map (slug)
---

## slug.map

slug.map — persistent map utilities

Higher-level map operations built on the core `put` and `remove`
primitives from `slug.std`. All functions are pure and return new
maps without modifying their inputs.

For basic key operations (`put`, `remove`, `update`, `get`, `keys`)
see `slug.std`.

### TOC

- [`difference(s1, s2)`](#differences1-s2)
- [`intersect(s1, s2)`](#intersects1-s2)
- [`merge(base, patch)`](#mergebase-patch)
- [`patch(base, patchData)`](#patchbase-patchdata)
- [`putNested(map, keys, value)`](#putnestedmap-keys-value)
- [`union(s1, s2)`](#unions1-s2)

### Functions

#### `difference(s1, s2)`
```slug
fn slug.map#difference(s1:map, s2:map):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s1` | map | — |
| `s2` | map | — |


#### Examples

```slug
difference({:k1: 1}, {:k1: 2})  // => {}
difference({:k1: 1, :k2: 1}, {:k2: 2})  // => {:k1: 1}
```

---

#### `intersect(s1, s2)`
```slug
fn slug.map#intersect(s1:map, s2:map):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s1` | map | — |
| `s2` | map | — |


#### Examples

```slug
intersect({:k1: 1}, {:k1: 2})  // => {:k1: 1}
intersect({:k1: 1}, {:k2: 2})  // => {}
intersect({}, {:k2: 2})  // => {}
intersect({:k1: 1}, {})  // => {}
```

---

#### `merge(base, patch)`
```slug
fn slug.map#merge(base:map, patch:map):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `base` | map | — |
| `patch` | map | — |


#### Examples

```slug
merge({:a: 1, :b: 2}, {:b: 3, :c: 4})  // => {:a: 1, :b: 3, :c: 4}
merge({:a: 1}, {})  // => {:a: 1}
merge({}, {:a: 1})  // => {:a: 1}
```

---

#### `patch(base, patchData)`
```slug
fn slug.map#patch(base:map, patchData:map):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `base` | map | — |
| `patchData` | map | — |


#### Examples

```slug
patch({:a: {:x: 1}}, {:a: {:y: 2}})  // => {:a: {:x: 1, :y: 2}}
patch({:a: 1}, {:a: {:b: 2}})  // => {:a: {:b: 2}}
patch({:a: {:b: 2}}, {:a: 1})  // => {:a: 1}
```

---

#### `putNested(map, keys, value)`
```slug
fn slug.map#putNested(map:map, keys:list, value):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `map` | map | — |
| `keys` | list | — |
| `value` |  | — |


#### Examples

```slug
putNested({}, ["k"], "v")  // => {:k: v}
putNested({}, ["k", "k"], "v")  // => {:k: {:k: v}}
putNested({:k: {:j: 1}}, ["k", "k"], "v")  // => {:k: {:k: v, :j: 1}}
```

---

#### `union(s1, s2)`
```slug
fn slug.map#union(s1:map, s2:map):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s1` | map | — |
| `s2` | map | — |


#### Examples

```slug
union({:k1: 1}, {:k2: 2})  // => {:k1: 1, :k2: 2}
union({:k1: 1}, {})  // => {:k1: 1}
union({}, {:k2: 2})  // => {:k2: 2}
union({:k1: 1}, {:k1: 9})  // => {:k1: 1}
```
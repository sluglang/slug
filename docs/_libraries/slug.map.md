---
title: map (slug)
---

## slug.map

### Functions

#### `difference(s1, s2)`
```slug
fn slug.map#difference(@map s1, @map s2) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s1` | @map  | — |
| `s2` | @map  | — |


#### Examples

```slug
difference({:k1: 1}, {:k1: 2})  // => {}
difference({:k1: 1, :k2: 1}, {:k2: 2})  // => {:k1: 1}
```

---

#### `intersect(s1, s2)`
```slug
fn slug.map#intersect(@map s1, @map s2) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s1` | @map  | — |
| `s2` | @map  | — |


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
fn slug.map#merge(@map base, @map patch) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `base` | @map  | — |
| `patch` | @map  | — |


#### Examples

```slug
merge({:a: 1, :b: 2}, {:b: 3, :c: 4})  // => {:c: 4, :a: 1, :b: 3}
merge({:a: 1}, {})  // => {:a: 1}
merge({}, {:a: 1})  // => {:a: 1}
```

---

#### `patch(base, patchData)`
```slug
fn slug.map#patch(@map base, @map patchData) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `base` | @map  | — |
| `patchData` | @map  | — |


#### Examples

```slug
patch({:a: {:x: 1}}, {:a: {:y: 2}})  // => {:a: {:x: 1, :y: 2}}
patch({:a: 1}, {:a: {:b: 2}})  // => {:a: {:b: 2}}
patch({:a: {:b: 2}}, {:a: 1})  // => {:a: 1}
```

---

#### `putNested(keys, map, value)`
```slug
fn slug.map#putNested(@list keys, @map map, value) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `keys` | @list  | — |
| `map` | @map  | — |
| `value` |  | — |


#### Examples

```slug
putNested(["k"], {}, "v")  // => {:k: v}
putNested(["k", "k"], {}, "v")  // => {:k: {:k: v}}
putNested(["k", "k"], {:k: {:j: 1}}, "v")  // => {:k: {:k: v, :j: 1}}
```

---

#### `union(s1, s2)`
```slug
fn slug.map#union(@map s1, @map s2) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s1` | @map  | — |
| `s2` | @map  | — |


#### Examples

```slug
union({:k1: 1}, {:k2: 2})  // => {:k1: 1, :k2: 2}
union({:k1: 1}, {})  // => {:k1: 1}
union({}, {:k2: 2})  // => {:k2: 2}
union({:k1: 1}, {:k1: 9})  // => {:k1: 1}
```
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

### Functions

#### `difference(s1, s2)`
```slug
fn slug.map#difference(@map s1, @map s2) -> @map
```


returns `s1` with all keys that appear in `s2` removed.

| Parameter | Type | Default |
| --- | --- | --- |
| `s1` | @map  | — |
| `s2` | @map  | — |


#### Examples

```slug
difference({:k1: 1}, {:k1: 2})  // => {}
difference({:k2: 1, :k1: 1}, {:k2: 2})  // => {:k1: 1}
```

---

#### `intersect(s1, s2)`
```slug
fn slug.map#intersect(@map s1, @map s2) -> @map
```


returns a map containing only keys present in both maps.

Values are taken from `s1`. Keys that exist only in `s1` or only
in `s2` are excluded from the result.

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


merges two maps, with `patch` values overwriting `base` values on key conflicts.

Shallow — nested maps are replaced entirely, not recursively merged.
For deep merging of nested maps, use `patch` instead.

| Parameter | Type | Default |
| --- | --- | --- |
| `base` | @map  | — |
| `patch` | @map  | — |


#### Examples

```slug
merge({:a: 1, :b: 2}, {:b: 3, :c: 4})  // => {:a: 1, :b: 3, :c: 4}
merge({:a: 1}, {})  // => {:a: 1}
merge({}, {:a: 1})  // => {:a: 1}
```

---

#### `patch(base, patchData)`
```slug
fn slug.map#patch(@map base, @map patchData) -> @map
```


recursively merges two maps, deeply merging nested map values.

When both `base` and `patchData` have a map value at the same key,
those maps are merged recursively. Otherwise the `patchData` value
overwrites the `base` value. For a shallow merge, use `merge` instead.

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


sets a deeply nested key in a map, creating intermediate maps as needed.

Keys are converted to symbols. If an intermediate key already exists
and maps to a map, that map is updated. If it maps to a non-map value,
it is replaced with a new map containing the nested key.

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


returns a map containing all keys from both maps.

On key conflicts, `s1` values take precedence over `s2` values.

| Parameter | Type | Default |
| --- | --- | --- |
| `s1` | @map  | — |
| `s2` | @map  | — |


#### Examples

```slug
union({:k1: 1}, {:k2: 2})  // => {:k2: 2, :k1: 1}
union({:k1: 1}, {})  // => {:k1: 1}
union({}, {:k2: 2})  // => {:k2: 2}
union({:k1: 1}, {:k1: 9})  // => {:k1: 1}
```
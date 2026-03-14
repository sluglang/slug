---
title: math (slug)
---

## slug.math

### Functions

#### `ceil(n)`
```slug
fn slug.math#ceil(@num n) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | @num  | — |


#### Examples

```slug
ceil(1)  // => 1
ceil(1.2)  // => 2
ceil(-1.2)  // => -1
```

---

#### `clampZero(n)`
```slug
fn slug.math#clampZero(@num n) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | @num  | — |

---

#### `floor(n)`
```slug
fn slug.math#floor(@num n) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | @num  | — |


#### Examples

```slug
floor(1)  // => 1
floor(1.2)  // => 1
floor(-1.2)  // => -2
```

---

#### `max(nil)`
```slug
fn slug.math#max(nil) -> @num
```
nil


#### Examples

```slug
max(1, 2, 3)  // => 3
max(7, 9, 8)  // => 9
max(4, 5, 3)  // => 5
```

---

#### `mean(xs)`
```slug
fn slug.math#mean(xs) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` |  | — |

---

#### `min(a, b)`
```slug
fn slug.math#min(@num a, ...b) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` | @num  | — |
| `b` |  | — |


#### Examples

```slug
min(1, 2, 3)  // => 1
min(4, 3, 5)  // => 3
min(6, 5, 4)  // => 4
```

---

#### `percentileSorted(xs, p)`
```slug
fn slug.math#percentileSorted(xs, p) -> @num
```


percentileSorted: xs must be ascending

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` |  | — |
| `p` |  | — |

---

#### `rndRange(min, max)`
```slug
fn slug.math#rndRange(@num min, @num max) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `min` | @num  | — |
| `max` | @num  | — |

---

#### `sqrt(n)`
```slug
fn slug.math#sqrt(@num n) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | @num  | — |


#### Examples

```slug
sqrt(0)  // => 0
sqrt(4)  // => 2
sqrt(9)  // => 3
```

---

#### `stdev(xs, mean)`
```slug
fn slug.math#stdev(xs, mean) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` |  | — |
| `mean` |  | — |
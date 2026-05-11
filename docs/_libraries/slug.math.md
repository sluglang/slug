---
title: math (slug)
---

## slug.math

slug.math — numeric utilities

Mathematical functions for common numeric operations including
rounding, extrema, statistics, and random number generation.

### TOC

- [`ceil(n)`](#ceiln)
- [`clampZero(n)`](#clampzeron)
- [`floor(n)`](#floorn)
- [`max(nil)`](#maxnil)
- [`mean(xs)`](#meanxs)
- [`min(a, b)`](#mina-b)
- [`percentileSorted(xs, p)`](#percentilesortedxs-p)
- [`rndRange(min, max)`](#rndrangemin-max)
- [`sqrt(n)`](#sqrtn)
- [`stdev(xs, mean)`](#stdevxs-mean)

### Functions

#### `ceil(n)`
```slug
fn slug.math#ceil(n:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | num | — |


#### Examples

```slug
ceil(1)  // => 1
ceil(1.2)  // => 2
ceil(-1.2)  // => -1
```

---

#### `clampZero(n)`
```slug
fn slug.math#clampZero(n:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | num | — |

---

#### `floor(n)`
```slug
fn slug.math#floor(n:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | num | — |


#### Examples

```slug
floor(1)  // => 1
floor(1.2)  // => 1
floor(-1.2)  // => -2
```

---

#### `max(nil)`
```slug
fn slug.math#max(nil):num
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
fn slug.math#mean(xs):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` |  | — |

---

#### `min(a, b)`
```slug
fn slug.math#min(a:num, ...b):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` | num | — |
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
fn slug.math#percentileSorted(xs, p):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` |  | — |
| `p` |  | — |

---

#### `rndRange(min, max)`
```slug
fn slug.math#rndRange(min:num, max:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `min` | num | — |
| `max` | num | — |

**Effects:** `random`

---

#### `sqrt(n)`
```slug
fn slug.math#sqrt(n:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | num | — |


#### Examples

```slug
sqrt(0)  // => 0
sqrt(4)  // => 2
sqrt(9)  // => 3
```

---

#### `stdev(xs, mean)`
```slug
fn slug.math#stdev(xs, mean):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` |  | — |
| `mean` |  | — |
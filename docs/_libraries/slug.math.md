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
- [`max(a, b)`](#maxa-b)
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


returns the least integer greater than or equal to `n`.

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


returns `n` if it is >= 0, otherwise returns 0.

Useful for ensuring subtracted values never go negative.

| Parameter | Type | Default |
| --- | --- | --- |
| `n` | num | — |

---

#### `floor(n)`
```slug
fn slug.math#floor(n:num):num
```


returns the greatest integer less than or equal to `n`.

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

#### `max(a, b)`
```slug
fn slug.math#max(a:num, ...b):num
fn slug.math#max(a:num, b:num):num
```


returns the largest of two or more numbers.

Accepts a variadic list of additional values beyond the required first two.

| Parameter | Type | Default |
| --- | --- | --- |
| `a` | num | — |
| `b` |  | — |


#### Examples

```slug
max(1, 2, 3)  // => 3
max(7, 9, 8)  // => 9
max(4, 5, 3)  // => 5
```

---

#### `mean(xs)`
```slug
fn slug.math#mean(xs:list<num>):num
```


returns the arithmetic mean of a list of numbers.

Returns `0` for an empty list.

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` | list<num> | — |

---

#### `min(a, b)`
```slug
fn slug.math#min(a:num, ...b):num
```


returns the smallest of two or more numbers.

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
fn slug.math#percentileSorted(xs:list<num>, p:num):num
```


returns the `p`th percentile of a pre-sorted ascending list.

Uses linear interpolation between adjacent values. The list must be
sorted in ascending order before calling — use `slug.list#sort` or
`slug.benchmark`'s internal `sortAsc` if needed.

Returns `0` for an empty list and the single element for a one-element list.

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` | list<num> | — |
| `p` | num | — |

---

#### `rndRange(min, max)`
```slug
fn slug.math#rndRange(min:num, max:num):num
```


returns a random integer in the range `[min, max)` (exclusive upper bound).

@effects('random')

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


returns the square root of `n`.

Returns `NaN` for negative values.

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
fn slug.math#stdev(xs:list<num>, mean:num):num
```


returns the sample standard deviation of a list of numbers.

Uses Bessel's correction (divides by `n - 1`). Returns `0` for lists
with fewer than two elements. Requires the precomputed `mean` as a
second argument to avoid recomputing it internally.

| Parameter | Type | Default |
| --- | --- | --- |
| `xs` | list<num> | — |
| `mean` | num | — |
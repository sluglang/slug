---
title: time (slug)
---

## slug.time

### Functions

#### `clock()`
```slug
fn slug.time#clock() -> @num
```

---

#### `clockNanos()`
```slug
fn slug.time#clockNanos() -> @num
```

---

#### `delta(f)`
```slug
fn slug.time#delta(f) -> @fn
```


delta creates a function that measures time difference between calls
Parameters:
  f: function that returns a time value
Returns:
  function that returns time elapsed since first call to f

| Parameter | Type | Default |
| --- | --- | --- |
| `f` |  | — |

---

#### `fmtClock(millis, fmt)`
```slug
fn slug.time#fmtClock(@num millis, @str fmt) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `millis` | @num  | — |
| `fmt` | @str  | — |

---

#### `minsToMillis(mins)`
```slug
fn slug.time#minsToMillis(@num mins) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `mins` | @num  | — |


#### Examples

```slug
minsToMillis(1)  // => 60000
```

---

#### `secsToMillis(secs)`
```slug
fn slug.time#secsToMillis(@num secs) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `secs` | @num  | — |


#### Examples

```slug
secsToMillis(1)  // => 1000
```

---

#### `sleep(millis)`
```slug
fn slug.time#sleep(@num millis) -> nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `millis` | @num  | — |
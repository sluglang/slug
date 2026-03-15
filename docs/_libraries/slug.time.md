---
title: time (slug)
---

## slug.time

slug.time — time and duration utilities

Wall-clock access, sleep, formatting, and duration conversion helpers.
All time values are in milliseconds unless the function name specifies
otherwise (e.g. `clockNanos`).

### TOC

- [`clock()`](#clock)
- [`clockNanos()`](#clocknanos)
- [`delta(f)`](#deltaf)
- [`fmtClock(millis, fmt)`](#fmtclockmillis-fmt)
- [`minsToMillis(mins)`](#minstomillismins)
- [`secsToMillis(secs)`](#secstomillissecs)
- [`sleep(millis)`](#sleepmillis)

### Functions

#### `clock()`
```slug
fn slug.time#clock() -> @num
```


returns the current wall-clock time in milliseconds since the Unix epoch.

@effects('time')

**Effects:** `time`

---

#### `clockNanos()`
```slug
fn slug.time#clockNanos() -> @num
```


returns the current time as a nanosecond counter.

Intended for high-resolution timing. The epoch is implementation-defined
— use differences between two calls rather than absolute values.

@effects('time')

**Effects:** `time`

---

#### `delta(f)`
```slug
fn slug.time#delta(f) -> @fn
```


returns a function that measures elapsed time since it was created.

`f` should be a time source function (e.g. `clock` or `clockNanos`).
Each call to the returned function returns the elapsed time since `delta`
was first called.

```slug
val { delta, clock } = import("slug.time")

val elapsed = delta(clock)
// ... do some work ...
println(elapsed())  // => milliseconds elapsed
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` |  | — |

---

#### `fmtClock(millis, fmt)`
```slug
fn slug.time#fmtClock(@num millis, @str fmt) -> @str
```


formats a millisecond timestamp using a format string.

Format follows standard Go time layout conventions.

| Parameter | Type | Default |
| --- | --- | --- |
| `millis` | @num  | — |
| `fmt` | @str  | — |

---

#### `minsToMillis(mins)`
```slug
fn slug.time#minsToMillis(@num mins) -> @num
```


converts minutes to milliseconds.

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


converts seconds to milliseconds.

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


suspends execution for `millis` milliseconds.

@effects('time')

| Parameter | Type | Default |
| --- | --- | --- |
| `millis` | @num  | — |

**Effects:** `time`
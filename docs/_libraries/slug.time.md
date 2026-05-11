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
fn slug.time#clock():num
```

**Effects:** `time`

---

#### `clockNanos()`
```slug
fn slug.time#clockNanos():num
```

**Effects:** `time`

---

#### `delta(f)`
```slug
fn slug.time#delta(f):fn
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` |  | — |

---

#### `fmtClock(millis, fmt)`
```slug
fn slug.time#fmtClock(millis:num, fmt:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `millis` | num | — |
| `fmt` | str | — |

---

#### `minsToMillis(mins)`
```slug
fn slug.time#minsToMillis(mins:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `mins` | num | — |


#### Examples

```slug
minsToMillis(1)  // => 60000
```

---

#### `secsToMillis(secs)`
```slug
fn slug.time#secsToMillis(secs:num):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `secs` | num | — |


#### Examples

```slug
secsToMillis(1)  // => 1000
```

---

#### `sleep(millis)`
```slug
fn slug.time#sleep(millis:num):nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `millis` | num | — |

**Effects:** `time`
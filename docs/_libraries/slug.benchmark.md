---
title: benchmark (slug)
---

## slug.benchmark

slug.benchmark
Slug-native micro benchmarking (no opts maps)

### Constants

#### `UnitMs`

```slug
str slug.benchmark#UnitMs
```

#### `UnitNs`

```slug
str slug.benchmark#UnitNs
```

#### `UnitUs`

```slug
str slug.benchmark#UnitUs
```

### Functions

#### `compare(nil)`
```slug
fn slug.benchmark#compare(nil) -> @num
```


compare(work, benches, ...same defaults...)
benches: [ {name:"x", fun: fn(){...}}, ... ]
nil

---

#### `micro(name, workFn, warmupMs, sampleMs, samples, minIters, maxIters, subtractOverhead, unit)`
```slug
fn slug.benchmark#micro(@str name, @fn workFn, @num warmupMs = 100, @num sampleMs = 200, @num samples = 20, @num minIters = 1, @num maxIters = 10000000, @bool subtractOverhead = true, @str unit = UnitNs) -> @map
```


micro(name, workFn, warmupMs=100, sampleMs=200, samples=20, minIters=1, maxIters=10000000, subtractOverhead=true, unit="ns")

returns:
{ name, unit, itersPerSample, samples, stats:{...}, raw:{timesPerIter:[...] } }

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | @str  | — |
| `workFn` | @fn  | — |
| `warmupMs` | @num  | `100` |
| `sampleMs` | @num  | `200` |
| `samples` | @num  | `20` |
| `minIters` | @num  | `1` |
| `maxIters` | @num  | `10000000` |
| `subtractOverhead` | @bool  | `true` |
| `unit` | @str  | `UnitNs` |

---

#### `printCompareReport(report)`
```slug
fn slug.benchmark#printCompareReport(report) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `report` |  | — |

---

#### `printResult(res)`
```slug
fn slug.benchmark#printResult(res) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
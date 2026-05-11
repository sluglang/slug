---
title: benchmark (slug)
---

## slug.benchmark

slug.benchmark — micro-benchmarking for Slug functions

Provides tools to measure and compare the performance of Slug functions
with statistical rigour. Each benchmark runs a warmup phase, calibrates
iteration count to fill a target sample window, then collects multiple
samples to compute percentile statistics.

Overhead from the benchmarking harness itself is measured and subtracted
by default, giving a cleaner picture of the work under test.

## Quick start

```slug
val { micro, printResult, compare, printCompareReport } = import("slug.benchmark")

val result = micro("string concat", fn() { "hello" + " " + "world" })
result /> printResult
```

## Units

Results are reported in nanoseconds by default. Use `UnitUs` or `UnitMs`
for slower workloads.

### TOC

- [UnitMs](#unitms)
- [UnitNs](#unitns)
- [UnitUs](#unitus)
- [`compare(benches, warmupMs, sampleMs, samples, minIters, maxIters, subtractOverhead, unit)`](#comparebenches-warmupms-samplems-samples-miniters-maxiters-subtractoverhead-unit)
- [`micro(name, workFn, warmupMs, sampleMs, samples, minIters, maxIters, subtractOverhead, unit)`](#microname-workfn-warmupms-samplems-samples-miniters-maxiters-subtractoverhead-unit)
- [`printCompareReport(report)`](#printcomparereportreport)
- [`printResult(res)`](#printresultres)

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

#### `compare(benches, warmupMs, sampleMs, samples, minIters, maxIters, subtractOverhead, unit)`
```slug
fn slug.benchmark#compare(benches:list, warmupMs:num = 100, sampleMs:num = 200, samples:num = 20, minIters:num = 1, maxIters:num = 10000000, subtractOverhead:bool = true, unit:str = UnitNs):map
```


benchmarks a list of named functions and returns results sorted by p50.

| Parameter | Type | Default |
| --- | --- | --- |
| `benches` | list | — |
| `warmupMs` | num | `100` |
| `sampleMs` | num | `200` |
| `samples` | num | `20` |
| `minIters` | num | `1` |
| `maxIters` | num | `10000000` |
| `subtractOverhead` | bool | `true` |
| `unit` | str | `UnitNs` |

**Effects:** `time`

---

#### `micro(name, workFn, warmupMs, sampleMs, samples, minIters, maxIters, subtractOverhead, unit)`
```slug
fn slug.benchmark#micro(name:str, workFn:fn, warmupMs:num = 100, sampleMs:num = 200, samples:num = 20, minIters:num = 1, maxIters:num = 10000000, subtractOverhead:bool = true, unit:str = UnitNs):map
```


runs a single micro-benchmark and returns a result map with statistics.

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | str | — |
| `workFn` | fn | — |
| `warmupMs` | num | `100` |
| `sampleMs` | num | `200` |
| `samples` | num | `20` |
| `minIters` | num | `1` |
| `maxIters` | num | `10000000` |
| `subtractOverhead` | bool | `true` |
| `unit` | str | `UnitNs` |

**Effects:** `time`

---

#### `printCompareReport(report)`
```slug
fn slug.benchmark#printCompareReport(report):nil
```


prints a formatted comparison report to stdout.

| Parameter | Type | Default |
| --- | --- | --- |
| `report` |  | — |

---

#### `printResult(res)`
```slug
fn slug.benchmark#printResult(res):nil
```


prints a formatted summary of a micro result to stdout.

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
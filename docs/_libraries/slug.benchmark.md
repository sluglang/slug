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
- [`compare(nil)`](#comparenil)
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

#### `compare(nil)`
```slug
fn slug.benchmark#compare(nil):map
```


benchmarks a list of named functions and returns results sorted by p50.

Each entry in `benches` must be a map with `name` (@str) and `fun` (@fn) fields.
All benchmarks share the same timing parameters. Results include a `ratio`
field showing performance relative to the fastest (lowest p50) entry.

Pass the result to `printCompareReport` for formatted console output.

```slug
val report = compare([
  { name: "concat", fun: fn() { "a" + "b" } },
  { name: "fmt",    fun: fn() { fmt("{}{}", "a", "b") } },
])
report /> printCompareReport
```

@effects('time')
nil

**Effects:** `time`

---

#### `micro(name, workFn, warmupMs, sampleMs, samples, minIters, maxIters, subtractOverhead, unit)`
```slug
fn slug.benchmark#micro(name:str, workFn:fn, warmupMs:num = 100, sampleMs:num = 200, samples:num = 20, minIters:num = 1, maxIters:num = 10000000, subtractOverhead:bool = true, unit:str = UnitNs):map
```


runs a single micro-benchmark and returns a result map with statistics.

Warms up the JIT/runtime for `warmupMs`, then calibrates iteration count
so each sample takes approximately `sampleMs` to collect. Runs `samples`
measurements and computes percentile statistics on the results.

Returns a map with the shape:
```
{
  name, unit, itersPerSample, samples,
  stats: { min, p50, p90, p99, max, mean, stdev },
  raw:   { timesPerIter: [...] }
}
```

Pass the result to `printResult` for formatted console output.

@effects('time')

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

```
Benchmark report
  concat         : p50 38 ns  x1.0
  fmt            : p50 91 ns  x2.4
```

| Parameter | Type | Default |
| --- | --- | --- |
| `report` |  | — |

---

#### `printResult(res)`
```slug
fn slug.benchmark#printResult(res):nil
```


prints a formatted summary of a `micro` result to stdout.

```
string concat
  p50: 42 ns  p90: 51  p99: 78
  mean: 44  stdev: 8  min: 38  max: 91
  iters/sample: 10000  samples: 20
```

| Parameter | Type | Default |
| --- | --- | --- |
| `res` |  | — |
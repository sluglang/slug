# Module 6: Mini Project - Data Pipeline

Build a complete mini project in small steps.

## Goal

Given a list of numbers:

1. Keep only odd numbers.
2. Square each value.
3. Sum the result.
4. Return a structured report.

## Step 1: Transform and sum

```slug
val {*} = import("slug.std")

val run = fn(xs) {
  xs
    /> filter(fn(v) { v % 2 != 0 })
    /> map(fn(v) { v * v })
    /> reduce(0, fn(acc, v) { acc + v })
}

println(run([1, 2, 3, 4, 5]))
```

Expected output:

```text
35
```

## Step 2: Add input validation

```slug
val Error = struct {
  @str type = "Error",
  @str msg,
}

val runSafe = fn(xs) {
  if (len(xs) == 0) {
    throw Error { type: "InputError", msg: "expected at least one value" }
  }
  run(xs)
}
```

## Step 3: Return a report map

```slug
val runReport = fn(xs) {
  val odds = xs /> filter(fn(v) { v % 2 != 0 })
  val squares = odds /> map(fn(v) { v * v })
  val total = squares /> reduce(0, fn(acc, v) { acc + v })

  {
    inputCount: len(xs),
    processedCount: len(odds),
    total: total,
  }
}

println(runReport([1, 2, 3, 4, 5]))
```

Common mistakes:
- Mutating intermediate structures instead of recomputing immutable values.

### Challenge

Add `maxSquare` to the report.

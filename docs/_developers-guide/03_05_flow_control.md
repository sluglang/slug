# Module 5: Flow Control

This module covers branching, loops via recursion, and robust error handling.

## Lesson 5.1: Conditionals

### Mental model

`if` is an expression, so both branches should produce a value.

```slug
val max = fn(a, b) {
  if (a > b) { a } else { b }
}

println(max(3, 5))
```

Expected output:

```text
5
```

## Lesson 5.2: Looping with tail recursion and `recur`

`recur` must be in tail position.

```slug
val sumTo = fn(n, acc = 0) {
  if (n == 0) {
    acc
  } else {
    recur(n - 1, acc + n)
  }
}

println(sumTo(5))
```

Expected output:

```text
15
```

Common mistakes:
- Calling `recur(...)` and then doing more work after it.

## Lesson 5.3: Throwing errors

```slug
val Error = struct {
  type:str = "Error",
  msg:str,
  code = nil,
  data = nil,
  cause = nil,
}

val divide = fn(a, b) {
  if (b == 0) {
    throw Error { type: "ValidationError", msg: "divisor cannot be zero" }
  }
  a / b
}
```

## Lesson 5.4: `defer`, `defer onsuccess`, `defer onerror(err)`

### Mental model

- `defer`: always runs when scope exits.
- `defer onsuccess`: only on success path.
- `defer onerror(err)`: only on thrown error path.

```slug
val run = fn(x) {
  defer { println("always") }
  defer onsuccess { println("success") }
  defer onerror(err) { println("failed:", err.msg) }

  if (x < 0) {
    throw Error { type: "ValidationError", msg: "x must be >= 0" }
  }

  x * 2
}

println(run(2))
```

Common mistakes:
- Treating `onsuccess` or `onerror` as global keywords outside `defer`.

### Try it

Write `parsePositive(n)` that throws on `n <= 0` and logs failures with `defer onerror(err)`.

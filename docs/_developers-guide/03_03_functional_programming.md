# Module 3: Functional Programming

This module focuses on transforming data with small composable functions.

## Lesson 3.1: Map, filter, reduce

### Mental model

- `map`: transform each item
- `filter`: keep some items
- `reduce`: combine into one result

```slug
val {*} = import("slug.std")

val xs = [1, 2, 3, 4, 5]
val squares = xs /> map(fn(v) { v * v })
val evens = xs /> filter(fn(v) { v % 2 == 0 })
val sum = xs /> reduce(0, fn(acc, v) { acc + v })

println(squares)
println(evens)
println(sum)
```

Expected output:

```text
[1, 4, 9, 16, 25]
[2, 4]
15
```

## Lesson 3.2: Match expressions

```slug
val classify = fn(v) {
  match v {
    0 => "zero"
    1 => "one"
    _ => "other"
  }
}

println(classify(1))
println(classify(10))
```

Common mistakes:
- Forgetting `_` fallback case.

## Lesson 3.3: Destructuring patterns

```slug
val headOrZero = fn(xs) {
  match xs {
    [h, ..._] => h
    [] => 0
  }
}

println(headOrZero([10, 20]))
println(headOrZero([]))
```

Pattern tools:

- `_` wildcard
- `...rest` spread capture
- `^name` pin
- `{| ... |}` exact-map pattern

```slug
val expected = 42
match 42 {
  ^expected => println("matched")
  _ => println("nope")
}
```

## Lesson 3.4: Higher-order functions

```slug
val applyTwice = fn(f, v) { f(f(v)) }
val inc = fn(x) { x + 1 }
println(applyTwice(inc, 10))
```

### Try it

Write `times(n, f, x)` using `recur`.

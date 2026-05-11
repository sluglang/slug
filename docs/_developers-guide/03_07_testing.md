# Module 7: Testing

This module walks through practical test-writing patterns.

## Lesson 7.1: A basic behaviour test

```slug
val {*} = import("slug.test")

val add = fn(a, b) { a + b }

@test
val addWorks = fn() {
  assertEqual(add(2, 3), 5)
}
```

### Why this matters

A tiny, focused behaviour test is easier to debug than one large test.

## Lesson 7.2: Parameterized tests with `@testWith`

```slug
val {*} = import("slug.test")

@testWith(
  [3, 5], 8,
  [10, -5], 5,
  [0, 0], 0,
)
val addCases = fn(a, b) {
  a + b
}
```

### Mental model
- each input tuple is executed,
- return value is compared to expected value.

## Lesson 7.3: Testing errors

```slug
val Error = struct { type:str = "Error", msg:str }

val mustPositive = fn(n) {
  if (n <= 0) { throw Error { type: "ValidationError", msg: "n must be positive" } }
  n
}

@test
val mustPositiveThrows = fn() {
  fn() { mustPositive(0) } /> assertThrows()
}
```

## Lesson 7.4: Running tests

```shell
slug test path_to_source.slug
```

Common mistakes:
- Asserting too many behaviours in one test.
- Using unclear test names.

### Try it

Add one new `@testWith` case and one new error-path test.

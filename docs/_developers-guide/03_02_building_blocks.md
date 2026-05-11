# Module 2: Core Building Blocks

This module introduces the syntax you will use constantly.

## Lesson 2.1: Core values and types

### Mental model

Slug is expression-oriented: most constructs produce values.

Common value shapes:

- `nil`, `true`, `false`
- numbers (DEC64-inspired)
- strings and bytes (`0x"ff00"`)
- lists (`[1, 2, 3]`)
- maps (`{name: "Slug"}`)
- functions (`fn(...) { ... }`)
- task handles (from `spawn`)
- symbols (`:ok`, `:"Content-Type"`)

## Lesson 2.2: Comments

```slug
# line comment
// also a line comment
/* block comment */
/** doc comment */
```

Common mistakes:
- Forgetting doc comments are a distinct form used for docs/metadata.

## Lesson 2.3: Strings and interpolation

Build up in steps:

```slug
val name = "Slug"
val msg = "Hello {{name}}"
val raw = 'C:\temp\file.txt'
println(msg, raw)
```

Expected output:

```text
Hello Slug C:\temp\file.txt
```

## Lesson 2.4: Numeric and bytes literals

```slug
val users = 1_000_000
val hex = 0x10_ff
val bytes = 0x"414243"
println(users, hex, bytes)
```

Expected output (shape):

```text
1000000 4351 <bytes value>
```

## Lesson 2.5: `var` vs `val`

### Mental model

- `val`: bind once.
- `var`: reassignable binding.

```slug
var counter = 0
val label = "requests"
counter = counter + 1
println(counter, label)
```

Common mistakes:
- Reassigning a `val`.

## Lesson 2.6: Semicolons and newlines

```slug
val a = 1
val b = 2
println(a + b)
```

Line continuation example:

```slug
val sql =
    "select *"
    + " from users"
```

Common mistakes:
- Breaking lines where parser expects expression completion.

## Lesson 2.7: Trailing commas

```slug
val m = {
  user: "slug",
}

println(
  m,
)
```

Trailing commas are allowed in maps/lists/call args/tags.

## Lesson 2.8: Everyday builtins

```slug
val {*} = import("slug.std")
println(len([1, 2, 3]))
print("hello")
println(" world")
```

## Lesson 2.9: Modules and exports

```slug
// math.slug
@export
val add = fn(a, b) { a + b }

// app.slug
val math = import("math")
println(math.add(2, 3))
```

Expected output:

```text
5
```

## Lesson 2.10: Functions, defaults, and named args

```slug
val greet = fn(name, title = "Mx") { "Hello {{title}} {{name}}" }
println(greet("Slug"))
println(greet(name: "Slug", title: "Dr"))
```

## Lesson 2.11: Pipelines (`/>`)

### Mental model

`x /> f(y)` rewrites to `f(x, y)`.

```slug
val double = fn(n) { n * 2 }
println(10 /> double)
```

Pipeline into `match`:

```slug
10 /> match {
  10 => "ten"
  _ => "other"
} /> println()
```

Common mistakes:
- Assuming `/>` is method dispatch only; it is general call piping.

## Lesson 2.12: Tagged dispatch

```slug
fn add(a:num, b:num) { a + b }
fn add(a:str, b:str) { a + b }

println(add(1, 2))
println(add("a", "b"))
```

Common type names: `num`, `str`, `bool`, `list`, `map`, `bytes`, `fn`, `task`, `sym`, `chan`.

### Try it

Add a list overload for `add` that concatenates two lists, then print `add([1], [2])`.

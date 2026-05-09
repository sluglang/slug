# Module 1: Getting Started

In this module, you will run Slug code, understand module resolution, and inspect CLI inputs.

## Lesson 1.1: Your first program

### Mental model

A Slug file is a sequence of statements. The CLI loads one entry module and executes it.

Create `hello.slug`:

```slug
println("Hello, Slug!")
```

Run it:

```shell
slug hello.slug
```

Expected output:

```text
Hello, Slug!
```

Common mistakes:
- Running from a different folder than the file path you pass.

### Try it

Print two values with `println("hello", 123)` and verify both appear.

## Lesson 1.2: How Slug resolves imports

### Mental model

`import("x")` is resolved relative to the entry module first, then library paths.

If you run:

```shell
slug ./tests/bytes.slug
```

then `import("slug.std")` is searched in:

1. `./tests/slug/std.slug`
2. `$SLUG_HOME/lib/slug/std.slug`

If the CLI target is not a local file, Slug treats it as a module name and searches local then library paths.

Common mistakes:
- Assuming imports are always resolved from current working directory only.

## Lesson 1.3: Program arguments

### Mental model

Everything after the entry target is user input.

- `argv()` returns raw argument list.
- `argm()` returns parsed options + positional args.

```slug
println(argv())
println(argm())
```

Run:

```shell
slug script.slug --user knuckles input.txt
```

Typical shape:

```text
argv: ["--user", "knuckles", "input.txt"]
argm: { options: {user: "knuckles"}, positional: ["input.txt"] }
```

### Try it

Print just `argm().options` and `argm().positional` separately.

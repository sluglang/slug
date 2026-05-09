# Module 4: Data Structures

You will use lists, maps, symbols, and structs together.

## Lesson 4.1: Lists

```slug
val xs = [10, 20, 30, 40]
println(xs[1])
println(xs[-1])
println(xs[1:3])
println(xs[0:2])
println(xs[2:])
```

Common mistakes:
- Confusing slice bounds (`start:end`) with single index access.

## Lesson 4.2: Maps and symbol keys

### Mental model

Bare map keys are symbols.

```slug
val m = {name: "Slug", age: 2}
println(m[:name])
println(m.name)
```

## Lesson 4.3: Symbols

```slug
println(:ok)
println(:"Content-Type")
println(sym("user id"))
println(label(:ok))
```

## Lesson 4.4: Struct schemas and values

Step-by-step:

```slug
val User = struct {
  @str name,
  @num age,
  active = true,
}

val u1 = User { name: "Slug", age: 2 }
val u2 = u1 copy { age: 3 }

println(type(u1))
println(keys(u1))
println(u2.age)
```

## Lesson 4.5: Matching maps and structs

```slug
match u2 {
  User { name, age } => println(name, age)
  _ => println("unknown")
}

match m {
  {| :name: n, :age: a |} => println(n, a)
  _ => println("missing fields")
}
```

Common mistakes:
- Using non-exact map patterns when exact field set is required.

### Try it

Create `Account` with `id`, `email`, and default `active = true`, then copy it with a new email.

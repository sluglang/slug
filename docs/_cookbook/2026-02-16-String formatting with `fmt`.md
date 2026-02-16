---
title: String formatting with `fmt`
tags: [ slug.std.fmt, slug.std ]
---

### **Problem**

You want readable, consistent strings for logs, errors, tables, and user-facing output—without a pile of `+`
concatenation.

### **Idiom: Use `fmt(template, ...args)` with `{}` placeholders**

`fmt` builds a string from a template plus arguments.

- `{}` inserts the next argument (auto-indexed).
- `{0}`, `{1}`, … inserts a specific argument (positional).
- `{:<10s}`, `{:>8}`, `{:^10s}` control alignment and width.
- `{:.2f}` controls numeric precision.
- `{:,}` adds thousands separators.
- `{:.1%}` multiplies by 100 and adds `%`.
- `\{` and `\}` let you print literal braces.

### **Recipe**

#### 1) Start simple
```slug
var {*} = import("slug.std")

println(fmt("Hello {}", "Slug"))
```

#### 2) Use positional placeholders when order matters
```slug
println(fmt("{1} then {0}", "A", "B"))   // "B then A"
```

#### 3) Control numeric precision (and percent)
```slug
println(fmt("x={:.2f} y={:.2f}", 1.2, 3.4))  // "x=1.20 y=3.40"
println(fmt("success={:.1%}", 0.1234))       // "success=12.3%"
```

#### 4) Print integers or grouped numbers
```slug
println(fmt("id={:d}", 12.9))          // "id=12"
println(fmt("total={:,}", 1234567))    // "total=1,234,567"
```

#### 5) Make quick fixed-width columns (alignment + width)
```slug
println(fmt("|{:>8}|", 12.3))       // right-aligned
println(fmt("|{:<10s}|", "Slug"))   // left-aligned (string)
println(fmt("|{:^10s}|", "Slug"))   // centered (string)
```

#### 6) Escape braces when you mean literal `{` / `}`
```slug
println(fmt("value=\\{x\\}"))  // "value={x}"
```

### **Notes**

- Auto-indexed (`{}`) and positional (`{0}`) placeholders are both supported; prefer `{}` until you need reordering.
- Width/alignment is handy for debug output and small CLI tables—keep it simple and it stays readable tomorrow.

---
title: Variable swapping (without a temp variable)
tags: [ slug.std, slug.test ]
---

### **Problem**

You have two variables and want to swap their values cleanly—without introducing a temporary variable.

### **Idiom: Redeclare with list decomposition**

Slug lets you destructure lists, and (in the right scope) redeclare variables. Combine those to swap in one expression:

- Build a pair `[a, b]`
- Destructure it back into the opposite order `[b, a]`

### **Recipe**

#### 1) Swap two variables in-place

```slug
var {*} = import("slug.test")

var a = 1
var b = 2

// swap via list decomposition (no temp variable)
var [b, a] = [a, b]

a /> assertEqual(2)
b /> assertEqual(1)
```

#### 2) Works with any values (not just numbers)

```slug
var {*} = import("slug.test")

var left  = "L"
var right = "R"

var [right, left] = [left, right]

left  /> assertEqual("R")
right /> assertEqual("L")
```

### **Notes**

- This pattern relies on **(1) list decomposition** and **(2) redeclaration** in the current scope.

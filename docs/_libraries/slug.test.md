---
title: test (slug)
---

## slug.test

slug.test — assertion and test utilities

All assertion functions throw `AssertionError` on failure and return
the tested value on success, allowing assertions to be chained or
used inline in expressions.

## Quick start

```slug
val { assertEqual, assertThrows, runSafe } = import("slug.test")

assertEqual(1 + 1, 2)
assertEqual("hello" /> toUpper(), "HELLO")
assertThrows(fn() { throw Error{ type: "Oops", msg: "!" } }, "Oops")
```

## Short aliases

All assertion functions have short aliases for concise test code:

| Full name                | Alias   |
|--------------------------|---------|
| `assertEqual`            | `eq`    |
| `assertLessThan`         | `lt`    |
| `assertLessThanOrEqual`  | `lteq`  |
| `assertGreaterThan`      | `gt`    |
| `assertGreaterThanOrEqual` | `gteq` |
| `assertThrows`           | `throws`|
| `assertTrue`             | `ok`    |
| `assertFalse`            | `not`   |

## Error safety

Use `runSafe` to execute a function without propagating errors —
useful for testing error paths or running multiple assertions where
a single failure should not abort the rest.

### TOC

- [`assert(a, msg)`](#asserta-msg)
- [`assertEqual(value, expected, msg)`](#assertequalvalue-expected-msg)
- [`assertErrorType(f, expected, msg)`](#asserterrortypef-expected-msg)
- [`assertFalse(a, msg)`](#assertfalsea-msg)
- [`assertGreaterThan(value, expected, msg)`](#assertgreaterthanvalue-expected-msg)
- [`assertGreaterThanOrEqual(value, expected, msg)`](#assertgreaterthanorequalvalue-expected-msg)
- [`assertLessThan(value, expected, msg)`](#assertlessthanvalue-expected-msg)
- [`assertLessThanOrEqual(value, expected, msg)`](#assertlessthanorequalvalue-expected-msg)
- [`assertNil(value, msg)`](#assertnilvalue-msg)
- [`assertNotEqual(value, expected, msg)`](#assertnotequalvalue-expected-msg)
- [`assertNotNil(value, msg)`](#assertnotnilvalue-msg)
- [`assertThrows(f, expected, msg)`](#assertthrowsf-expected-msg)
- [`assertTrue(a, msg)`](#asserttruea-msg)
- [`eq(value, expected, msg)`](#eqvalue-expected-msg)
- [`fail(msg)`](#failmsg)
- [`gt(value, expected, msg)`](#gtvalue-expected-msg)
- [`gteq(value, expected, msg)`](#gteqvalue-expected-msg)
- [`isAssertError(v)`](#isasserterrorv)
- [`lt(value, expected, msg)`](#ltvalue-expected-msg)
- [`lteq(value, expected, msg)`](#lteqvalue-expected-msg)
- [`not(a, msg)`](#nota-msg)
- [`ok(a, msg)`](#oka-msg)
- [`runSafe(f)`](#runsafef)
- [`throws(f, expected, msg)`](#throwsf-expected-msg)

### Functions

#### `assert(a, msg)`
```slug
fn slug.test#assert(a, msg = nil):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertEqual(value, expected, msg)`
```slug
fn slug.test#assertEqual(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertErrorType(f, expected, msg)`
```slug
fn slug.test#assertErrorType(f:fn, expected, msg:str = nil):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` | fn | — |
| `expected` |  | — |
| `msg` | str | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertFalse(a, msg)`
```slug
fn slug.test#assertFalse(a, msg = nil):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertGreaterThan(value, expected, msg)`
```slug
fn slug.test#assertGreaterThan(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertGreaterThanOrEqual(value, expected, msg)`
```slug
fn slug.test#assertGreaterThanOrEqual(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertLessThan(value, expected, msg)`
```slug
fn slug.test#assertLessThan(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertLessThanOrEqual(value, expected, msg)`
```slug
fn slug.test#assertLessThanOrEqual(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertNil(value, msg)`
```slug
fn slug.test#assertNil(value, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertNotEqual(value, expected, msg)`
```slug
fn slug.test#assertNotEqual(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertNotNil(value, msg)`
```slug
fn slug.test#assertNotNil(value, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertThrows(f, expected, msg)`
```slug
fn slug.test#assertThrows(f:fn, expected, msg:str = nil):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` | fn | — |
| `expected` |  | — |
| `msg` | str | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `assertTrue(a, msg)`
```slug
fn slug.test#assertTrue(a, msg = nil):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `eq(value, expected, msg)`
```slug
fn slug.test#eq(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `fail(msg)`
```slug
fn slug.test#fail(msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `gt(value, expected, msg)`
```slug
fn slug.test#gt(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `gteq(value, expected, msg)`
```slug
fn slug.test#gteq(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `isAssertError(v)`
```slug
fn slug.test#isAssertError(v):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `lt(value, expected, msg)`
```slug
fn slug.test#lt(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `lteq(value, expected, msg)`
```slug
fn slug.test#lteq(value, expected, msg = nil):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `not(a, msg)`
```slug
fn slug.test#not(a, msg = nil):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `ok(a, msg)`
```slug
fn slug.test#ok(a, msg = nil):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `Error{type:AssertionError}`

---

#### `runSafe(f)`
```slug
fn slug.test#runSafe(f:fn):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` | fn | — |

---

#### `throws(f, expected, msg)`
```slug
fn slug.test#throws(f:fn, expected, msg:str = nil):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` | fn | — |
| `expected` |  | — |
| `msg` | str | `nil` |

**Throws:** `Error{type:AssertionError}`
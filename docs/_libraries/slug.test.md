---
title: test (slug)
---

## slug.test

### Functions

#### `assert(a, msg)`
```slug
fn slug.test#assert(a, msg = nil) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertEqual(value, expected, msg)`
```slug
fn slug.test#assertEqual(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertErrorType(f, expected, msg)`
```slug
fn slug.test#assertErrorType(@fn f, expected, @str msg = nil) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` | @fn  | — |
| `expected` |  | — |
| `msg` | @str  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertFalse(a, msg)`
```slug
fn slug.test#assertFalse(a, msg = nil) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertGreaterThan(value, expected, msg)`
```slug
fn slug.test#assertGreaterThan(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertGreaterThanOrEqual(value, expected, msg)`
```slug
fn slug.test#assertGreaterThanOrEqual(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertLessThan(value, expected, msg)`
```slug
fn slug.test#assertLessThan(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertLessThanOrEqual(value, expected, msg)`
```slug
fn slug.test#assertLessThanOrEqual(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertNil(value, msg)`
```slug
fn slug.test#assertNil(value, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertNotEqual(value, expected, msg)`
```slug
fn slug.test#assertNotEqual(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertNotNil(value, msg)`
```slug
fn slug.test#assertNotNil(value, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertThrows(f, expected, msg)`
```slug
fn slug.test#assertThrows(@fn f, expected, @str msg = nil) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` | @fn  | — |
| `expected` |  | — |
| `msg` | @str  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `assertTrue(a, msg)`
```slug
fn slug.test#assertTrue(a, msg = nil) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `eq(value, expected, msg)`
```slug
fn slug.test#eq(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `fail(msg)`
```slug
fn slug.test#fail(msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `gt(value, expected, msg)`
```slug
fn slug.test#gt(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `gteq(value, expected, msg)`
```slug
fn slug.test#gteq(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `isAssertError(v)`
```slug
fn slug.test#isAssertError(v) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

---

#### `lt(value, expected, msg)`
```slug
fn slug.test#lt(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `lteq(value, expected, msg)`
```slug
fn slug.test#lteq(value, expected, msg = nil) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `value` |  | — |
| `expected` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `not(a, msg)`
```slug
fn slug.test#not(a, msg = nil) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `ok(a, msg)`
```slug
fn slug.test#ok(a, msg = nil) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `a` |  | — |
| `msg` |  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`

---

#### `runSafe(f)`
```slug
fn slug.test#runSafe(@fn f) -> @map
```


runSafe executes f() and returns a map:
  {"ok": value} on normal return
  {"error": thrownValue, "trace": stacktrace(thrownValue)} on throw

Implementation uses `defer onerror` to intercept thrown payload.

| Parameter | Type | Default |
| --- | --- | --- |
| `f` | @fn  | — |

---

#### `throws(f, expected, msg)`
```slug
fn slug.test#throws(@fn f, expected, @str msg = nil) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `f` | @fn  | — |
| `expected` |  | — |
| `msg` | @str  | `nil` |

**Throws:** `@struct(Error{type:AssertionError})`
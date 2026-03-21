---
title: test
---

## test

Slug Test — unit test runner

Discovers and runs `@test` and `@testWith` tagged functions in the
given source files, reporting passes, failures, and errors.

Usage:
```bash
slug test [options] <test files...>
```

Test examples:
```slug
@test
val simpleTest = fn() {
  assert(1 + 1 == 2)
}

@testWith(
  [1, 2, 3], 3,
  [7, 9, 8], 9,
  [4, 5, 3], 5
)
val max = fn(a, ...b) { ... }
```

Options:
- `-h, --help`      show this screen
- `-v, --version`   show version
- `--verbose`       show passing tests as well as failures
---
title: sys (slug)
---

## slug.sys

### Functions

#### `env(str)`
```slug
fn slug.sys#env(@str str) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `exec(cmd, timeout)`
```slug
fn slug.sys#exec(@str cmd, @num timeout = 0) -> [@str, @str]
```

| Parameter | Type | Default |
| --- | --- | --- |
| `cmd` | @str  | — |
| `timeout` | @num  | `0` |

---

#### `exit(exitCode)`
```slug
fn slug.sys#exit(@num exitCode) -> nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `exitCode` | @num  | — |

---

#### `killProc(handle)`
```slug
fn slug.sys#killProc(@num handle) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |

---

#### `setEnv(str)`
```slug
fn slug.sys#setEnv(@str str) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `spawnProc(cmd)`
```slug
fn slug.sys#spawnProc(@list cmd) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `cmd` | @list  | — |

---

#### `waitProc(handle, timeout)`
```slug
fn slug.sys#waitProc(@num handle, @num timeout = 0) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |
| `timeout` | @num  | `0` |
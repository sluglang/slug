---
title: sys (slug)
---

## slug.sys

slug.sys — operating system interface

Process management, environment variables, and shell command execution.

## Process handles

`spawnProc`, `waitProc`, and `killProc` work with numeric process handles.
Spawn a process with `spawnProc`, wait for it to finish with `waitProc`,
and terminate it early with `killProc`.

For simple synchronous command execution, prefer `exec`.

### TOC

- [`env(str)`](#envstr)
- [`exec(cmd, timeout)`](#execcmd-timeout)
- [`exit(exitCode)`](#exitexitcode)
- [`killProc(handle)`](#killprochandle)
- [`setEnv(key, value)`](#setenvkey-value)
- [`spawnProc(cmd)`](#spawnproccmd)
- [`waitProc(handle, timeout)`](#waitprochandle-timeout)

### Functions

#### `env(str)`
```slug
fn slug.sys#env(str:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

**Effects:** `io`

---

#### `exec(cmd, timeout)`
```slug
fn slug.sys#exec(cmd:str, timeout:num = 0):[@str, @str]
```

| Parameter | Type | Default |
| --- | --- | --- |
| `cmd` | str | — |
| `timeout` | num | `0` |

**Effects:** `io`

---

#### `exit(exitCode)`
```slug
fn slug.sys#exit(exitCode:num):nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `exitCode` | num | — |

---

#### `killProc(handle)`
```slug
fn slug.sys#killProc(handle:num):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |

**Effects:** `io`

---

#### `setEnv(key, value)`
```slug
fn slug.sys#setEnv(key:str, value:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `key` | str | — |
| `value` | str | — |

**Effects:** `io`

---

#### `spawnProc(cmd)`
```slug
fn slug.sys#spawnProc(cmd:list):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `cmd` | list | — |

**Effects:** `io`

---

#### `waitProc(handle, timeout)`
```slug
fn slug.sys#waitProc(handle:num, timeout:num = 0):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |
| `timeout` | num | `0` |

**Effects:** `io`
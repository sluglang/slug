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
fn slug.sys#env(@str str) -> @str
```


reads an environment variable by name.

Returns `nil` if the variable is not set.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

**Effects:** `io`

---

#### `exec(cmd, timeout)`
```slug
fn slug.sys#exec(@str cmd, @num timeout = 0) -> [@str, @str]
```


runs a shell command synchronously and returns `[stdout, stderr]`.

Blocks until the command completes or `timeout` elapses. A `timeout`
of `0` means wait indefinitely. Both stdout and stderr are returned
as strings regardless of exit code.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `cmd` | @str  | — |
| `timeout` | @num  | `0` |

**Effects:** `io`

---

#### `exit(exitCode)`
```slug
fn slug.sys#exit(@num exitCode) -> nil
```


terminates the current process with the given exit code.

Deferred functions are not run. Exit code `0` indicates success;
non-zero indicates failure.

| Parameter | Type | Default |
| --- | --- | --- |
| `exitCode` | @num  | — |

---

#### `killProc(handle)`
```slug
fn slug.sys#killProc(@num handle) -> @bool
```


sends a termination signal to a spawned process.

Returns `true` if the signal was delivered successfully.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |

**Effects:** `io`

---

#### `setEnv(key, value)`
```slug
fn slug.sys#setEnv(@str key, @str value) -> @str
```


sets an environment variable.

The format is `"NAME", "value"`. Returns the value that was set.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `key` | @str  | — |
| `value` | @str  | — |

**Effects:** `io`

---

#### `spawnProc(cmd)`
```slug
fn slug.sys#spawnProc(@list cmd) -> @num
```


spawns a process from a command list and returns a process handle.

The command list format is `["executable", "arg1", "arg2", ...]`.
Use `waitProc` to block until the process finishes, or `killProc`
to terminate it.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `cmd` | @list  | — |

**Effects:** `io`

---

#### `waitProc(handle, timeout)`
```slug
fn slug.sys#waitProc(@num handle, @num timeout = 0) -> @num
```


waits for a spawned process to finish and returns its exit code.

Returns `-1` if `timeout` elapses before the process exits.
A `timeout` of `0` means wait indefinitely.

@effects('io')

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |
| `timeout` | @num  | `0` |

**Effects:** `io`
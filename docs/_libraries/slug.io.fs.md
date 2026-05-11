---
title: fs (slug.io)
---

## slug.io.fs

slug.io.fs — filesystem I/O

Read, write, and navigate the filesystem. File handles from `openFile`
must be closed with `closeFile` — use `defer closeFile(handle)` for safety.

## Simple file operations

```slug
val { readFile, writeFile, exists } = import("slug.io.fs")

val content = readFile("config.toml")
writeFile("output.txt", "hello\n")
```

## Streaming reads

```slug
val { openFile, readLines, closeFile, READ_MODE } = import("slug.io.fs")

val handle = openFile("data.csv", READ_MODE)
defer closeFile(handle)
val lines = readLines(handle)
```

## File modes

- `READ_MODE`   — open for reading (`"r"`)
- `WRITE_MODE`  — open for writing, truncates existing content (`"w"`)
- `APPEND_MODE` — open for appending (`"a"`)

@effects('fs')

### TOC

- [APPEND_MODE](#append_mode)
- [READ_MODE](#read_mode)
- [WRITE_MODE](#write_mode)
- [`appendFile(contents, path)`](#appendfilecontents-path)
- [`closeFile(handle)`](#closefilehandle)
- [`exists(path)`](#existspath)
- [`info(path)`](#infopath)
- [`isDir(path)`](#isdirpath)
- [`listFilesRecursive(path, filter, acc)`](#listfilesrecursivepath-filter-acc)
- [`ls(path)`](#lspath)
- [`mkDirs(path)`](#mkdirspath)
- [`openFile(path, mode)`](#openfilepath-mode)
- [`readFile(path)`](#readfilepath)
- [`readLine(handle)`](#readlinehandle)
- [`readLines(file, lines)`](#readlinesfile-lines)
- [`rm(path)`](#rmpath)
- [`write(handle, content)`](#writehandle-content)
- [`writeFile(contents, path)`](#writefilecontents-path)

### Constants

#### `APPEND_MODE`

```slug
str slug.io.fs#APPEND_MODE
```

#### `READ_MODE`

```slug
str slug.io.fs#READ_MODE
```

#### `WRITE_MODE`

```slug
str slug.io.fs#WRITE_MODE
```

### Functions

#### `appendFile(contents, path)`
```slug
fn slug.io.fs#appendFile(contents:str, path:str):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `contents` | str | — |
| `path` | str | — |

---

#### `closeFile(handle)`
```slug
fn slug.io.fs#closeFile(handle:num):nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |

---

#### `exists(path)`
```slug
fn slug.io.fs#exists(path:str):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `info(path)`
```slug
fn slug.io.fs#info(path:str):map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `isDir(path)`
```slug
fn slug.io.fs#isDir(path:str):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `listFilesRecursive(path, filter, acc)`
```slug
fn slug.io.fs#listFilesRecursive(path:list, filter:fn = fn((s)) {true}, acc = []):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | list | — |
| `filter` | fn | `fn((s)) {true}` |
| `acc` |  | `[]` |

---

#### `ls(path)`
```slug
fn slug.io.fs#ls(path:str):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `mkDirs(path)`
```slug
fn slug.io.fs#mkDirs(...path):bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` |  | — |

---

#### `openFile(path, mode)`
```slug
fn slug.io.fs#openFile(path:str, mode:str):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |
| `mode` | str | — |

---

#### `readFile(path)`
```slug
fn slug.io.fs#readFile(path:str):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `readLine(handle)`
```slug
fn slug.io.fs#readLine(handle:num):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |

---

#### `readLines(file, lines)`
```slug
fn slug.io.fs#readLines(file:num, lines = []):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `file` | num | — |
| `lines` |  | `[]` |

---

#### `rm(path)`
```slug
fn slug.io.fs#rm(path:str):nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `write(handle, content)`
```slug
fn slug.io.fs#write(handle:num, content:str):num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |
| `content` | str | — |

---

#### `writeFile(contents, path)`
```slug
fn slug.io.fs#writeFile(contents:str, path:str):nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `contents` | str | — |
| `path` | str | — |
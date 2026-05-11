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


appends `contents` to the end of a file, creating it if absent.

Returns the number of bytes written.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `contents` | str | — |
| `path` | str | — |

---

#### `closeFile(handle)`
```slug
fn slug.io.fs#closeFile(handle:num):nil
```


closes an open file handle.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |

---

#### `exists(path)`
```slug
fn slug.io.fs#exists(path:str):bool
```


returns true if `path` exists and is accessible.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `info(path)`
```slug
fn slug.io.fs#info(path:str):map
```


returns metadata about a file or directory as a map.

The map contains at minimum `name`, `size`, `isDir`, and `modTime` fields.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `isDir(path)`
```slug
fn slug.io.fs#isDir(path:str):bool
```


returns true if `path` is a directory.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `listFilesRecursive(path, filter, acc)`
```slug
fn slug.io.fs#listFilesRecursive(path:list, filter:fn = fn((s)) {true}, acc = []):list
```


recursively collects all files under the given path(s) matching `filter`.

`path` is a list of starting paths. Directories are expanded recursively.
`filter` receives each file path and returns true to include it.

```slug
listFilesRecursive(["src"], fn(s) { endsWith(s, ".slug") })
// => ["src/main.slug", "src/lib/util.slug", ...]
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


returns a list of filenames in a directory (non-recursive).

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `mkDirs(path)`
```slug
fn slug.io.fs#mkDirs(...path):bool
```


creates all directories in the given path(s), including intermediates.

Returns true on success.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `path` |  | — |

---

#### `openFile(path, mode)`
```slug
fn slug.io.fs#openFile(path:str, mode:str):num
```


opens a file in the specified mode and returns a file handle.

Always close the handle with `closeFile`. Use `defer closeFile(handle)`
to ensure it is closed even if an error occurs.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |
| `mode` | str | — |

---

#### `readFile(path)`
```slug
fn slug.io.fs#readFile(path:str):str
```


reads the entire contents of a file and returns it as a string.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `readLine(handle)`
```slug
fn slug.io.fs#readLine(handle:num):str
```


reads the next line from an open file handle.

Returns `nil` at end of file. The returned string includes the trailing
newline — use `str[0:-1]` to strip it.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |

---

#### `readLines(file, lines)`
```slug
fn slug.io.fs#readLines(file:num, lines = []):list
```


reads all remaining lines from an open file handle into a list.

Trailing newlines are stripped from each line. Returns an empty list
if the file is already at EOF.

| Parameter | Type | Default |
| --- | --- | --- |
| `file` | num | — |
| `lines` |  | `[]` |

---

#### `rm(path)`
```slug
fn slug.io.fs#rm(path:str):nil
```


removes a file or empty directory at `path`.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | str | — |

---

#### `write(handle, content)`
```slug
fn slug.io.fs#write(handle:num, content:str):num
```


writes `content` to an open file handle.

Returns the number of bytes written.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | num | — |
| `content` | str | — |

---

#### `writeFile(contents, path)`
```slug
fn slug.io.fs#writeFile(contents:str, path:str):nil
```


writes `contents` to a file, creating or truncating it.

@effects('fs')

| Parameter | Type | Default |
| --- | --- | --- |
| `contents` | str | — |
| `path` | str | — |
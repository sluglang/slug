---
title: fs (slug.io)
---

## slug.io.fs

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
fn slug.io.fs#appendFile(@str contents, @str path) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `contents` | @str  | — |
| `path` | @str  | — |

---

#### `closeFile(handle)`
```slug
fn slug.io.fs#closeFile(@num handle) -> nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |

---

#### `exists(path)`
```slug
fn slug.io.fs#exists(@str path) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | @str  | — |

---

#### `info(path)`
```slug
fn slug.io.fs#info(@str path) -> @map
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | @str  | — |

---

#### `isDir(path)`
```slug
fn slug.io.fs#isDir(@str path) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | @str  | — |

---

#### `listFilesRecursive(path, filter, acc)`
```slug
fn slug.io.fs#listFilesRecursive(@list path, @fn filter = fn((s)) {true}, acc = []) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | @list  | — |
| `filter` | @fn  | `fn((s)) {true}` |
| `acc` |  | `[]` |

---

#### `ls(path)`
```slug
fn slug.io.fs#ls(@str path) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | @str  | — |

---

#### `mkDirs(path)`
```slug
fn slug.io.fs#mkDirs(...path) -> @bool
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` |  | — |

---

#### `openFile(path, mode)`
```slug
fn slug.io.fs#openFile(@str path, @str mode) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | @str  | — |
| `mode` | @str  | — |

---

#### `readFile(path)`
```slug
fn slug.io.fs#readFile(@str path) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | @str  | — |

---

#### `readLine(handle)`
```slug
fn slug.io.fs#readLine(@num handle) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |

---

#### `readLines(file, lines)`
```slug
fn slug.io.fs#readLines(@num file, lines = []) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `file` | @num  | — |
| `lines` |  | `[]` |

---

#### `rm(path)`
```slug
fn slug.io.fs#rm(@str path) -> nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `path` | @str  | — |

---

#### `write(handle, content)`
```slug
fn slug.io.fs#write(@num handle, @str content) -> @num
```

| Parameter | Type | Default |
| --- | --- | --- |
| `handle` | @num  | — |
| `content` | @str  | — |

---

#### `writeFile(contents, path)`
```slug
fn slug.io.fs#writeFile(@str contents, @str path) -> nil
```

| Parameter | Type | Default |
| --- | --- | --- |
| `contents` | @str  | — |
| `path` | @str  | — |
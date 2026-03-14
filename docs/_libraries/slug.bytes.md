---
title: bytes (slug)
---

## slug.bytes

### Functions

#### `base64Decode(s)`
```slug
fn slug.bytes#base64Decode(@str s) -> @bytes
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s` | @str  | — |


#### Examples

```slug
base64Decode("")  // => 0x""
base64Decode("aGVsbG8gc2x1Zw==")  // => 0x"68656c6c6f20736c7567"
```

---

#### `base64Encode(b)`
```slug
fn slug.bytes#base64Encode(@bytes b) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | @bytes  | — |


#### Examples

```slug
base64Encode(0x"")  // => ""
base64Encode(0x"68656c6c6f20736c7567")  // => "aGVsbG8gc2x1Zw=="
```

---

#### `bytesToHexStr(b)`
```slug
fn slug.bytes#bytesToHexStr(@bytes b) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | @bytes  | — |


#### Examples

```slug
bytesToHexStr(0x"")  // => ""
bytesToHexStr(0x"68656c6c6f20736c7567")  // => "68656c6c6f20736c7567"
```

---

#### `bytesToNumbers(b, i, acc)`
```slug
fn slug.bytes#bytesToNumbers(@bytes b, i = 0, acc = []) -> @list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | @bytes  | — |
| `i` |  | `0` |
| `acc` |  | `[]` |


#### Examples

```slug
bytesToNumbers(0x"a8ff04")  // => [168, 255, 4]
```

---

#### `bytesToStr(b)`
```slug
fn slug.bytes#bytesToStr(@bytes b) -> @str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | @bytes  | — |


#### Examples

```slug
bytesToStr(0x"")  // => ""
bytesToStr(0x"68656c6c6f20736c7567")  // => "hello slug"
```

---

#### `hexStrToBytes(hex)`
```slug
fn slug.bytes#hexStrToBytes(@str hex) -> @bytes
```

| Parameter | Type | Default |
| --- | --- | --- |
| `hex` | @str  | — |


#### Examples

```slug
hexStrToBytes("")  // => 0x""
hexStrToBytes("68656c6c6f20736c7567")  // => 0x"68656c6c6f20736c7567"
hexStrToBytes("a8ff04")  // => 0x"a8ff04"
```

---

#### `repeat(b, count, acc)`
```slug
fn slug.bytes#repeat(@bytes b, @num count, @bytes acc = 0x"") -> @bytes
```

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | @bytes  | — |
| `count` | @num  | — |
| `acc` | @bytes  | `0x""` |


#### Examples

```slug
repeat(0x"ff", 3)  // => 0x"ffffff"
```

---

#### `strToBytes(s)`
```slug
fn slug.bytes#strToBytes(@str s) -> @bytes
```

| Parameter | Type | Default |
| --- | --- | --- |
| `s` | @str  | — |


#### Examples

```slug
strToBytes("")  // => 0x""
strToBytes("hello slug")  // => 0x"68656c6c6f20736c7567"
```
---
title: bytes (slug)
---

## slug.bytes

slug.bytes — byte buffer utilities

Conversions between `bytes`, `str`, hex strings, base64, and numeric lists.
All functions are pure and operate on immutable byte values.

Byte literals use the `0x"<hex>"` syntax, e.g. `0x"ff0a"`.

### TOC

- [`base64Decode(s)`](#base64decodes)
- [`base64Encode(b)`](#base64encodeb)
- [`bytesToHexStr(b)`](#bytestohexstrb)
- [`bytesToNumbers(b, i, acc)`](#bytestonumbersb-i-acc)
- [`bytesToStr(b)`](#bytestostrb)
- [`hexStrToBytes(hex)`](#hexstrtobyteshex)
- [`repeat(b, count, acc)`](#repeatb-count-acc)
- [`strToBytes(s)`](#strtobytess)

### Functions

#### `base64Decode(s)`
```slug
fn slug.bytes#base64Decode(s:str):bytes
```


decodes a standard base64 string to a byte buffer.

| Parameter | Type | Default |
| --- | --- | --- |
| `s` | str | — |


#### Examples

```slug
base64Decode("")  // => 0x""
base64Decode("aGVsbG8gc2x1Zw==")  // => 0x"68656c6c6f20736c7567"
```

---

#### `base64Encode(b)`
```slug
fn slug.bytes#base64Encode(b:bytes):str
```


encodes a byte buffer as a standard base64 string.

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | bytes | — |


#### Examples

```slug
base64Encode(0x"")  // => ""
base64Encode(0x"68656c6c6f20736c7567")  // => "aGVsbG8gc2x1Zw=="
```

---

#### `bytesToHexStr(b)`
```slug
fn slug.bytes#bytesToHexStr(b:bytes):str
```


encodes a byte buffer as a lowercase hex string.

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | bytes | — |


#### Examples

```slug
bytesToHexStr(0x"")  // => ""
bytesToHexStr(0x"68656c6c6f20736c7567")  // => "68656c6c6f20736c7567"
```

---

#### `bytesToNumbers(b, i, acc)`
```slug
fn slug.bytes#bytesToNumbers(b:bytes, i = 0, acc = []):list
```


converts a byte buffer to a list of numeric byte values (0–255).

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | bytes | — |
| `i` |  | `0` |
| `acc` |  | `[]` |


#### Examples

```slug
bytesToNumbers(0x"a8ff04")  // => [168, 255, 4]
```

---

#### `bytesToStr(b)`
```slug
fn slug.bytes#bytesToStr(b:bytes):str
```


converts a byte buffer to a UTF-8 string.

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | bytes | — |


#### Examples

```slug
bytesToStr(0x"")  // => ""
bytesToStr(0x"68656c6c6f20736c7567")  // => "hello slug"
```

---

#### `hexStrToBytes(hex)`
```slug
fn slug.bytes#hexStrToBytes(hex:str):bytes
```


decodes a lowercase hex string to a byte buffer.

| Parameter | Type | Default |
| --- | --- | --- |
| `hex` | str | — |


#### Examples

```slug
hexStrToBytes("")  // => 0x""
hexStrToBytes("68656c6c6f20736c7567")  // => 0x"68656c6c6f20736c7567"
hexStrToBytes("a8ff04")  // => 0x"a8ff04"
```

---

#### `repeat(b, count, acc)`
```slug
fn slug.bytes#repeat(b:bytes, count:num, acc:bytes = 0x""):bytes
```


repeats a byte buffer `count` times and returns the concatenated result.

| Parameter | Type | Default |
| --- | --- | --- |
| `b` | bytes | — |
| `count` | num | — |
| `acc` | bytes | `0x""` |


#### Examples

```slug
repeat(0x"ff", 3)  // => 0x"ffffff"
```

---

#### `strToBytes(s)`
```slug
fn slug.bytes#strToBytes(s:str):bytes
```


converts a UTF-8 string to a byte buffer.

| Parameter | Type | Default |
| --- | --- | --- |
| `s` | str | — |


#### Examples

```slug
strToBytes("")  // => 0x""
strToBytes("hello slug")  // => 0x"68656c6c6f20736c7567"
```
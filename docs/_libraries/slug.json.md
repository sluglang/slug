---
title: json (slug)
---

## slug.json

slug.json — JSON encoding and decoding

A pure Slug implementation of JSON serialisation and deserialisation
conforming to RFC 8259.

## Type mapping

| Slug type   | JSON type  | Notes                                      |
|-------------|------------|--------------------------------------------|
| `@str`      | string     |                                            |
| `@num`      | number     |                                            |
| `@bool`     | boolean    |                                            |
| `nil`       | null       |                                            |
| `@list`     | array      |                                            |
| `@map`      | object     | keys serialised as strings; sorted         |
| `@sym`      | string     | symbol label used as string value          |
| `@bytes`    | string     | encoded as `"b64:<base64>"` prefix string  |

## Key handling

Object keys are always serialised as strings. Both symbol keys (`:name`)
and string keys (`"name"`) are supported on encode. On decode, all
object keys are returned as string keys — use dot access (`m.name`) or
string bracket access (`m["name"]`) to read decoded values.

## Bytes encoding

`@bytes` values are encoded as `"b64:<base64>"` strings. On decode,
any string with this prefix is automatically decoded back to `@bytes`.

### Functions

#### `decode(jsonStr)`
```slug
fn slug.json#decode(@str jsonStr) -> ?
```


decodes a JSON string into a Slug value.

Object keys are always decoded as string keys. Use dot access (`m.name`)
or string bracket access (`m["name"]`) to read fields — never symbol
bracket access (`m[:name]`).

Strings with the `b64:` prefix are decoded to `@bytes` values.
Throws `JsonError` on malformed input.

| Parameter | Type | Default |
| --- | --- | --- |
| `jsonStr` | @str  | — |

**Throws:** `@struct(Error{type:JsonError})`


#### Examples

```slug
decode(""hello"")  // => "hello"
decode("42")  // => 42
decode("true")  // => true
decode("false")  // => false
decode("null")  // => nil
decode("[1,2,3]")  // => [1, 2, 3]
decode(""b64:Zm9v"")  // => 0x"666f6f"
decode("{"name":"Alice","age":30}")  // => {age: 30, name: Alice}
```

---

#### `encode(v)`
```slug
fn slug.json#encode(v) -> @str
```


encodes a Slug value as a compact JSON string.

Map keys are sorted alphabetically. Symbols are encoded as their label
string. Bytes are encoded as `"b64:<base64>"`. Structs are not supported
and will throw `JsonError`.

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |

**Throws:** `@struct(Error{type:JsonError})`


#### Examples

```slug
encode("hello")  // => ""hello""
encode(42)  // => "42"
encode(true)  // => "true"
encode(false)  // => "false"
encode(nil)  // => "null"
encode(0x"ff")  // => ""b64:/w==""
encode([1, 2, 3])  // => "[1,2,3]"
encode({:age: 30, :type: :fn})  // => "{"age":30,"type":"fn"}"
encode({:name: Alice, :age: 30})  // => "{"age":30,"name":"Alice"}"
encode({name: Alice, age: 30})  // => "{"age":30,"name":"Alice"}"
```

---

#### `pretty(v, indent)`
```slug
fn slug.json#pretty(v, @num indent = 2) -> @str
```


encodes a Slug value as a pretty-printed JSON string.

The `indent` parameter controls the number of spaces per indentation
level. Follows the same type mapping as `encode`.

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |
| `indent` | @num  | `2` |

**Throws:** `@struct(Error{type:JsonError})`


#### Examples

```slug
pretty("hello", 2)  // => ""hello""
pretty([1, 2, 3], 4)  // => "[
    1,
    2,
    3
]"
pretty(0x"ff", 2)  // => ""b64:/w==""
pretty({:type: :fn, :age: 30}, 2)  // => "{
  "age": 30,
  "type": "fn"
}"
pretty({:name: Alice, :age: 30}, 2)  // => "{
  "age": 30,
  "name": "Alice"
}"
pretty({name: Alice, age: 30}, 2)  // => "{
  "age": 30,
  "name": "Alice"
}"
```
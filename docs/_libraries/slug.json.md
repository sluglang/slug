---
title: json (slug)
---

## slug.json

### Functions

#### `decode(jsonStr)`
```slug
fn slug.json#decode(@str jsonStr) -> ?
```


decodes a JSON string; object keys are decoded as string keys; prefer dot access (m.name)
for ergonomic field access via dot lookup tolerance.

| Parameter | Type | Default |
| --- | --- | --- |
| `jsonStr` | @str  | — |


#### Examples

```slug
decode(""hello"")  // => "hello"
decode("42")  // => 42
decode("true")  // => true
decode("false")  // => false
decode("null")  // => nil
decode("[1,2,3]")  // => [1, 2, 3]
decode(""b64:Zm9v"")  // => 0x"666f6f"
decode("{"name":"Alice","age":30}")  // => {name: Alice, age: 30}
```

---

#### `encode(v)`
```slug
fn slug.json#encode(v) -> @str
```

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
pretty({:age: 30, :type: :fn}, 2)  // => "{
  "age": 30,
  "type": "fn"
}"
pretty({:name: Alice, :age: 30}, 2)  // => "{
  "age": 30,
  "name": "Alice"
}"
pretty({age: 30, name: Alice}, 2)  // => "{
  "age": 30,
  "name": "Alice"
}"
```
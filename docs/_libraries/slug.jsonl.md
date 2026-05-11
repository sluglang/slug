---
title: jsonl (slug)
---

## slug.jsonl

slug.jsonl — JSON Lines encoding and decoding

Encodes lists of values into JSONL (`.jsonl`) and decodes JSONL strings
into lists of values. Uses `slug.json` for per-line JSON handling.

### TOC

- [`decode(jsonl, skipEmpty)`](#decodejsonl-skipempty)
- [`encode(values, eol, trailingEol)`](#encodevalues-eol-trailingeol)

### Functions

#### `decode(jsonl, skipEmpty)`
```slug
fn slug.jsonl#decode(jsonl:str, skipEmpty:bool = true):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `jsonl` | str | — |
| `skipEmpty` | bool | `true` |


#### Examples

```slug
decode("{"a":1}
{"b":2}")  // => [{a: 1}, {b: 2}]
decode("{"a":1}
{"b":2}")  // => [{a: 1}, {b: 2}]
decode("{"a":1}

{"b":2}")  // => [{a: 1}, {b: 2}]
```

---

#### `encode(values, eol, trailingEol)`
```slug
fn slug.jsonl#encode(values:list, eol:str = "\n", trailingEol:bool = false):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `values` | list | — |
| `eol` | str | `"\n"` |
| `trailingEol` | bool | `false` |


#### Examples

```slug
encode([{name: a}, {name: b}])  // => "{"name":"a"}
{"name":"b"}"
encode([1, true, nil], "
", true)  // => "1
true
null
"
```
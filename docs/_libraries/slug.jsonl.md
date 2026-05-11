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


decodes JSON Lines text into a list of values.

By default blank lines are ignored. Set `skipEmpty` to `false` to enforce
strict parsing and treat empty lines as errors.

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


encodes a list of values as JSON Lines text.

Each value is encoded as one JSON value per line.

- `eol` controls the line separator (default `"\n"`).
- `trailingEol` appends one final `eol` when `true`.

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
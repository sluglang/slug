---
title: csv (slug)
---

## slug.csv

### Functions

#### `fromCsvString(csvStr, sep, quote)`
```slug
fn slug.csv#fromCsvString(@str csvStr, @str sep = ",", @str quote = "\"") -> @list
```


returned rows are lists, access columns by index: row[0], row[1]

| Parameter | Type | Default |
| --- | --- | --- |
| `csvStr` | @str  | — |
| `sep` | @str  | `","` |
| `quote` | @str  | `"\""` |


#### Examples

```slug
fromCsvString("a,b,c")  // => [["a", "b", "c"]]
fromCsvString("a,b,c
d,"e
 f",g")  // => [["a", "b", "c"], ["d", "e
 f", "g"]]
```

---

#### `toCsv(rows, sep, quote, eol, acc)`
```slug
fn slug.csv#toCsv(@list rows, @str sep = ",", @str quote = "\"", @str eol = "\r\n", @str acc = "") -> @str
```


converts a list of lists to a CSV string

| Parameter | Type | Default |
| --- | --- | --- |
| `rows` | @list  | — |
| `sep` | @str  | `","` |
| `quote` | @str  | `"\""` |
| `eol` | @str  | `"\r\n"` |
| `acc` | @str  | `""` |


#### Examples

```slug
toCsv([["a", "b"]])  // => "a,b
"
toCsv([["a", "b"], ["c", "d"]])  // => "a,b
c,d
"
toCsv([["a
", "qu"ote"]])  // => ""a
","qu""ote"
"
```
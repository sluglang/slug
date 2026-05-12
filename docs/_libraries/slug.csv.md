---
title: csv (slug)
---

## slug.csv

slug.csv — CSV parsing and serialisation

Parses CSV strings into lists of rows and serialises lists of rows back
to CSV strings. Handles quoted fields, embedded newlines, escaped quotes,
and configurable separators.

Rows are represented as `list` values. Columns are accessed by index:
`row[0]` for the first column, `row[1]` for the second, and so on.

## Example

```slug
val { fromCsvString, toCsv } = import("slug.csv")

val rows = fromCsvString("name,age\r\nAlice,30\r\nBob,25")
rows[0]  // => ["name", "age"]
rows[1]  // => ["Alice", "30"]

toCsv([["name", "age"], ["Alice", "30"]])
// => "name,age\r\nAlice,30\r\n"
```

## Quoting rules

Fields are automatically quoted during serialisation if they contain
the separator, the quote character, a newline, or a carriage return.
Quote characters within a field are escaped by doubling: `"` → `""`.

### TOC

- [`fromCsvString(csvStr, sep, quote)`](#fromcsvstringcsvstr-sep-quote)
- [`toCsv(rows, sep, quote, eol, acc)`](#tocsvrows-sep-quote-eol-acc)

### Functions

#### `fromCsvString(csvStr, sep, quote)`
```slug
fn slug.csv#fromCsvString(csvStr:str, sep:str = ",", quote:str = "\""):list<list<str>>
```


parses a CSV string into a list of rows, where each row is a list of strings.

Handles quoted fields (including embedded newlines and escaped quotes),
CRLF and LF line endings, and configurable separator and quote characters.
Access columns by index: `row[0]`, `row[1]`, etc.

| Parameter | Type | Default |
| --- | --- | --- |
| `csvStr` | str | — |
| `sep` | str | `","` |
| `quote` | str | `"\""` |


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
fn slug.csv#toCsv(rows:list<list<str>>, sep:str = ",", quote:str = "\"", eol:str = "\r\n", acc:str = ""):str
```


converts a list of rows (each a list of strings) to a CSV string.

Fields are quoted automatically when they contain the separator, quote
character, or newline characters. The default line ending is `\r\n`
(RFC 4180). Each row is terminated by `eol`.

| Parameter | Type | Default |
| --- | --- | --- |
| `rows` | list<list<str>> | — |
| `sep` | str | `","` |
| `quote` | str | `"\""` |
| `eol` | str | `"\r\n"` |
| `acc` | str | `""` |


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
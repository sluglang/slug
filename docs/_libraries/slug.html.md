---
title: html (slug)
---

## slug.html

slug.html — basic HTML utilities

Lightweight helpers for working with HTML fragments. Not a full HTML
parser — intended for simple extraction and unescaping tasks on
well-formed markup.

### TOC

- [`extractInnerText(html, tag, start)`](#extractinnertexthtml-tag-start)
- [`htmlUnescape(str)`](#htmlunescapestr)

### Functions

#### `extractInnerText(html, tag, start)`
```slug
fn slug.html#extractInnerText(html, tag, start = 0):str
```


extracts the inner content of the first matching tag in an HTML string.

Handles nested tags of the same type correctly — the inner content is
everything between the outermost opening and closing tag. Returns an
empty string if the tag is not found or `start` is negative.

The `start` parameter can be used to search from a specific character
offset, enabling successive extractions from the same document.

```slug
extractInnerText("<div><p>hello</p></div>", "p")
// => "hello"

extractInnerText("<a href='h'><a>hi</a></a>", "a")
// => "<a>hi</a>"
```

| Parameter | Type | Default |
| --- | --- | --- |
| `html` |  | — |
| `tag` |  | — |
| `start` |  | `0` |


#### Examples

```slug
extractInnerText("<a>hi</a>", "a", -1)  // => ""
extractInnerText("<a>hi</a>", "div")  // => ""
extractInnerText("<a>hi</a>", "a")  // => "hi"
extractInnerText("<a href='h'><a>hi</a></a>", "a")  // => "<a>hi</a>"
extractInnerText("<a><a href='h'>hi</a></a>", "a")  // => "<a href='h'>hi</a>"
```

---

#### `htmlUnescape(str)`
```slug
fn slug.html#htmlUnescape(str):str
```


decodes common HTML entities in a string.

Handles `&amp;`, `&lt;`, `&gt;`, `&quot;`, and `&#39;`.
Does not handle numeric entities beyond `&#39;` or named entities
beyond the five listed above.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` |  | — |


#### Examples

```slug
htmlUnescape("E &amp; S")  // => "E & S"
```
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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` |  | — |


#### Examples

```slug
htmlUnescape("E &amp; S")  // => "E & S"
```
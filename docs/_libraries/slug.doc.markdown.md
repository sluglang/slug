---
title: markdown (slug.doc)
---

## slug.doc.markdown

slug.doc.markdown — Markdown API documentation generator

Reflects over live Slug modules using slug.meta#describe and renders
a single-page markdown file with standard API documentation style.

Each module gets a level-2 section with its doc comment as a lead paragraph,
followed by subsections for Constants, Structs, and Functions. Overloaded
functions are grouped under a single heading with multiple signature blocks.

Output is designed for rendering on GitHub, static site generators, or
any standard markdown renderer.

### TOC

- [`pagePerModule(moduleNames, moduleToc)`](#pagepermodulemodulenames-moduletoc)
- [`singlePage(moduleNames, title, moduleToc)`](#singlepagemodulenames-title-moduletoc)

### Functions

#### `pagePerModule(moduleNames, moduleToc)`
```slug
fn slug.doc.markdown#pagePerModule(moduleNames:list, moduleToc:bool = false):list
```

| Parameter | Type | Default |
| --- | --- | --- |
| `moduleNames` | list | — |
| `moduleToc` | bool | `false` |

---

#### `singlePage(moduleNames, title, moduleToc)`
```slug
fn slug.doc.markdown#singlePage(moduleNames:list, title:str = "Slug API Reference", moduleToc:bool = false):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `moduleNames` | list | — |
| `title` | str | `"Slug API Reference"` |
| `moduleToc` | bool | `false` |
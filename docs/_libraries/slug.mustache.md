---
title: mustache (slug)
---

## slug.mustache

slug.mustache — Mustache template renderer

A native Slug implementation of the Mustache spec v1.4.3, including
the core suites and template inheritance.

## Basic usage

```slug
val { render, parse } = import("slug.mustache")

render("Hello, {{name}}!", { name: "World" })
// => "Hello, World!"

render("{{#items}}{{.}} {{/items}}", { items: ["a", "b", "c"] })
// => "a b c "
```

## Tag types

| Tag              | Meaning                                      |
|------------------|----------------------------------------------|
| `{{name}}`       | HTML-escaped variable                        |
| `{{{name}}}`     | Unescaped variable                           |
| `{{&name}}`      | Unescaped variable (alternate syntax)        |
| `{{#section}}`   | Section — renders if truthy/non-empty        |
| `{{^section}}`   | Inverted section — renders if falsey/empty   |
| `{{/section}}`   | Closes a section                             |
| `{{>partial}}`   | Partial — renders another template by name   |
| `{{<parent}}`    | Template inheritance — extend a parent       |
| `{{$block}}`     | Block — defines or overrides a named region  |
| `{{! comment }}` | Comment — ignored                            |
| `{{=<% %>=}}`    | Set delimiters                               |

## Partials

Pass a map of `name -> template_string` as the `partials` argument.
Partials inherit the current context and are resolved at render time.

## Template inheritance

Use `{{<parent}}...{{/parent}}` to extend a parent template.
Override named blocks with `{{$blockName}}...{{/blockName}}`.

## Performance

For templates rendered multiple times, parse once and pass the AST
to `render` or `renderCached` to avoid repeated parsing.
`renderCached` additionally caches parsed partials across calls.

### TOC

- [`parse(template)`](#parsetemplate)
- [`render(templateOrAst, data, partials)`](#rendertemplateorast-data-partials)
- [`renderCached(templateOrAst, data, partials, cache)`](#rendercachedtemplateorast-data-partials-cache)

### Functions

#### `parse(template)`
```slug
fn slug.mustache#parse(template:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `template` | str | — |

---

#### `render(templateOrAst, data, partials)`
```slug
fn slug.mustache#render(templateOrAst, data, partials = nil):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `templateOrAst` |  | — |
| `data` |  | — |
| `partials` |  | `nil` |

**Throws:** `Error{type:MustacheError}`

---

#### `renderCached(templateOrAst, data, partials, cache)`
```slug
fn slug.mustache#renderCached(templateOrAst, data, partials = nil, cache = {}):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `templateOrAst` |  | — |
| `data` |  | — |
| `partials` |  | `nil` |
| `cache` |  | `{}` |

**Throws:** `Error{type:MustacheError}`
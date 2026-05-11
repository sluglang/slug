---
title: views (slug.web)
---

## slug.web.views

slug.web.views — Mustache template rendering for web responses

Loads and renders Mustache templates from the filesystem, with support
for partials, layouts, HTMX fragment rendering, and template caching
in production mode.

## Directory structure

Templates are loaded from a root directory (default `"views"`):

```
views/
  pages/
    home.mustache          -- main page template
    home/
      content.mustache     -- page-specific partial
  layouts/
    main.mustache          -- shared layout
  partials/
    nav.mustache           -- shared partial
    nav/
      item.mustache        -- nested partial
```

## Partial name resolution

Inside a template, partials are referenced by logical name:

- `layouts.main` → `views/layouts/main.mustache`
- `partials.nav.item` → `views/partials/nav/item.mustache`
- `content` (in `pages/home`) → `views/pages/home/content.mustache`

Dots in partial names are converted to path separators.
If a partial cannot be found in the view-specific directory, it falls
back to `views/partials/`.

## HTMX fragment rendering

When `renderReply` receives a request with an `HX-Request` header, it
renders only the fragment specified by `Reply.fragment` (or the
`HX-Target` header if `fragment` is nil) rather than the full page.

## Dev vs production mode

In dev mode (`cfg("dev", true)`), templates are re-read from disk on
every request. In production mode, templates are cached after the first
load. Toggle with `cfg("dev", false)`.

### TOC

- [`new(root)`](#newroot)
- [`render(v, name, data)`](#renderv-name-data)
- [`renderReply(vx, req, reply)`](#renderreplyvx-req-reply)

### Functions

#### `new(root)`
```slug
fn slug.web.views#new(root:str = "views"):Views
```

| Parameter | Type | Default |
| --- | --- | --- |
| `root` | str | `"views"` |

---

#### `render(v, name, data)`
```slug
fn slug.web.views#render(v, name, data = {}):str
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |
| `name` |  | — |
| `data` |  | `{}` |

---

#### `renderReply(vx, req, reply)`
```slug
fn slug.web.views#renderReply(vx, req, reply):Response
```

| Parameter | Type | Default |
| --- | --- | --- |
| `vx` |  | — |
| `req` |  | — |
| `reply` |  | — |
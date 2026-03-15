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

### Functions

#### `new(root)`
```slug
fn slug.web.views#new(@str root = "views") -> @struct(Views)
```


creates a new `Views` instance with the given template root directory.

In dev mode templates are reloaded on every request. In production they
are cached after first load. Dev mode is enabled by default unless
`cfg("dev", false)` is set.

| Parameter | Type | Default |
| --- | --- | --- |
| `root` | @str  | `"views"` |

---

#### `render(v, name, data)`
```slug
fn slug.web.views#render(v, name, data = {}) -> @str
```


renders a page template by name and returns the HTML string.

`name` is the page name without the `.mustache` extension, resolved
under `views/pages/`. Partials are loaded relative to `views/pages/<name>/`
with fallback to `views/layouts/` and `views/partials/`.

```slug
val views = new()
views /> render("home", { title: "Hello", user: currentUser })
```

| Parameter | Type | Default |
| --- | --- | --- |
| `v` |  | — |
| `name` |  | — |
| `data` |  | `{}` |

---

#### `renderReply(vx, req, reply)`
```slug
fn slug.web.views#renderReply(vx, req, reply) -> @struct(Response)
```


renders a `Reply` into an HTML `Response`, with HTMX fragment support.

If the request has an `HX-Request` header, renders only the fragment
specified by `Reply.fragment` (or `HX-Target` if `fragment` is nil).
Otherwise renders the full page view.

The response status and extra headers from the `Reply` are applied to
the final `Response`.

| Parameter | Type | Default |
| --- | --- | --- |
| `vx` |  | — |
| `req` |  | — |
| `reply` |  | — |
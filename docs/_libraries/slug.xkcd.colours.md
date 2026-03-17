---
title: colours (slug.xkcd)
---

## slug.xkcd.colours

slug.xkcd.colours — xkcd colour survey colour index

A map of ~950 colour names to their hex values, sourced from the
[xkcd colour survey](https://blog.xkcd.com/2010/05/03/color-survey-results/).

Keys are camelCase symbol names (e.g. `:electricLime`, `:neonBlue`).
Values are hex colour strings (e.g. `"#a8ff04"`).

Used by `slug.term.colour` for `xkcd` and `bgXkcd` colour functions.

```slug
val { xkcd } = import("slug.term.colour")

"Hello!" /> xkcd(:electricLime) /> println
```

### Constants

#### `XkcdColourIndex`

```slug
map slug.xkcd.colours#XkcdColourIndex
```
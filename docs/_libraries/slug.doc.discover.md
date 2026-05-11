---
title: discover (slug.doc)
---

## slug.doc.discover

slug.doc.discover — Slug module discovery helper

Scans a source directory recursively for `.slug` files and converts each
file path into a dotted module name suitable for `import(...)`.

The scan is performed relative to the provided directory. Returned module
names are normalized by:
- stripping the base directory prefix
- converting `/` and `\` path separators to `.`
- removing any leading dot
- dropping the `.slug` suffix

Example:
```
discoverModules("lib")
// => ["slug.doc.discover", "slug.doc.manifest", "slug.doc.markdown", ...]
```

If `dir` is `nil`, an empty list is returned.

### Functions

#### `discoverModules(dir)`
```slug
fn slug.doc.discover#discoverModules(dir):list
```


Scans a directory recursively for .slug files and derives dotted module
names from their paths relative to the base directory.

e.g. base="lib", path="lib/slug/io/fs.slug" => "slug.io.fs"

| Parameter | Type | Default |
| --- | --- | --- |
| `dir` |  | — |
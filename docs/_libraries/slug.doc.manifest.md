---
title: manifest (slug.doc)
---

## slug.doc.manifest

slug.doc.manifest — MANIFEST.ai generator

Reflects over live Slug modules using slug.meta#describe and renders
a structured MANIFEST.ai file in the Option 3 block format.

Each exported symbol produces a signature line followed by an indented
block of metadata sub-lines. Only sub-lines with content are emitted —
absent tags produce no output.

Supported tags

On functions:
  @returns('@type')       return type annotation; rendered on the signature line
  @effects('fs io ...')   space-separated effect labels; rendered as "effects: ..."
  @throws('@type')        thrown error type; rendered under "errors:" block
  @throwsDesc('...')      cause description paired with @throws; appended to thrown line

On struct fields:
  @type('@type')          field type annotation; rendered on the field line
  @desc('...')            field description; rendered as ": ..." on the field line
                          field lines are omitted entirely when both tags are absent

Doc comments:
  The first paragraph of a /** doc comment is extracted and rendered as
  the desc: sub-line. Paragraphs are separated by blank lines. The first
  paragraph lines are joined with spaces into a single desc string.
  Subsequent paragraphs are ignored in MANIFEST output.

Output format example:
  fn slug.repo#loadQueries(@str base = DefaultBase) -> @map
    desc: "scans base dir for .sql files; returns nested map of query functions"
    effects: fs
    errors:
      thrown: struct(Error) "unexpected filesystem failure"

  struct slug.repo#SqlQuery{@str sql, @list args, @sym type}
    desc: "result of extractArgs; carries rewritten SQL and inferred statement type"
    field cause @str: "the original error that triggered this one"

### TOC

- [`manifest(moduleNames)`](#manifestmodulenames)

### Functions

#### `manifest(moduleNames)`
```slug
fn slug.doc.manifest#manifest(@list moduleNames) -> @str
```


generates a MANIFEST.ai file for the given list of module names.

modules are sorted alphabetically. within each module, symbols are
sorted alphabetically. the output contains two sections: @section modules
lists all module names, @section symbols lists all exported symbols in
the Option 3 block format with indented metadata sub-lines.

pass the full dotted module names as they would appear in an import statement.

| Parameter | Type | Default |
| --- | --- | --- |
| `moduleNames` | @list  | — |
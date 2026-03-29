---
title: Module-local logging
tags: [ slug.log, slug.std, logging ]
---

### **Problem**

You want each module to have its own logger with a stable source name.

Hardcoding the source is fragile:

```slug
val log = import("slug.log").logger("playground")
````

This breaks when modules are copied or renamed.

### **Idiom: Derive the logger source from the module**

Use `moduleName()` from `slug.std` when constructing the logger.

```slug
val {*} = import(
    "slug.std",
)

val log = import("slug.log").logger(moduleName())
```

The module provides its own identity at the call site.

This is the preferred, copy-paste-safe setup.

### **Examples**

#### Basic usage

```slug
val {*} = import("slug.std")

val log = import("slug.log").logger(moduleName())

log.info("starting worker")
```

#### In pipelines

```slug
val {*} = import("slug.std")

val log = import("slug.log").logger(moduleName())

process()
    /> log.cInfo("processed input")
    /> validate
    /> log.cDebug("validated result")
```

### **Why this works**

* The logger source is always correct for the calling module.
* No hidden context or runtime inspection is required.
* The pattern is portable across files and libraries.
* Logging remains explicit and composable.

### **Notes**

* `moduleName()` returns the fully-qualified module name (for example `"slug.io.http"`).
* `slug.log` does not infer the caller; the caller provides its identity explicitly.
* This pattern composes cleanly with other module-level setup (imports, config, etc.).

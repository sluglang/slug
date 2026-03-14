---
title: doc
---

## doc

doc — Slug documentation tool

Generates MANIFEST.ai or markdown documentation by reflecting over
Slug modules in a source directory.

Usage:
  slug doc [options] [--] [manifest|markdown]

Commands:
  manifest   generate a MANIFEST.ai file from source modules (default)
  markdown   generate human-readable markdown documentation

Options:
  -h --help      show this screen
  -v --version   show version
  --dir          source directory to scan for .slug modules (default: .)
  --out          output file path; omit to print to stdout
---
title: watch
---

## watch

A utility command that can watch a directory for file changes and when detected run/rerun a command

### Slug Watch

A file watcher utility that monitors directories for file changes and automatically
runs or reruns a specified command when changes are detected. Useful for development
workflows where you want to automatically rebuild, test, or restart your application
when source files are modified.

Usage:
```bash
slug watch [options] [--] <cmd>
```

Options
- `-h --help`:      Show this screen.
- `-v --version`:   Show version.
- `--exts`:         The file extensions to monitor, supports multiple values, [default: [".slug", ".toml"]].
- `--roots`:        The root directories to monitor, supports multiple values, [default: ["."]].
- `--clear`:        Clear the terminal on restart [default: false].
- `--show-time`:    Show the execution time of the command [default: true].
- `--poll-ms`:      How often to check for file changes in milliseconds [default: 200].
- `--debounce-ms`:  Debounce time for file checks in milliseconds [default: 150].
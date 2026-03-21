---
title: watch
---

## watch

Slug Watch — file watcher and command runner

Monitors directories for file changes and automatically reruns a
command when changes are detected. Useful for development workflows
where you want to automatically rebuild, test, or restart your
application when source files are modified.

Usage:
```bash
slug watch [options] <cmd> [args...]
```

Options:
- `-h, --help`        show this screen
- `-v, --version`     show version
- `--exts`            file extensions to monitor, may be passed multiple times (default: .slug .toml)
- `--roots`           root directories to monitor, may be passed multiple times (default: .)
- `--clear`           clear the terminal on restart (default: false)
- `--showTime`        show execution time of the command (default: true)
- `--pollMs`          how often to check for file changes in ms (default: 200)
- `--debounceMs`      debounce window for file change events in ms (default: 150)
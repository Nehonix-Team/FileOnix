<p align="center">
  <img src="assets/FileOnix.png" width="600" alt="FileOnix Logo">
</p>

# Nehonix FileOnix v2

**High-Performance Native Watcher System**

FileOnix v2 is the complete ground-up rewrite of QuickDev — a professional, native file watcher and development server engineered for extreme performance. Built with a Go core and a premium terminal UI, it serves as the native hot-reload engine for modern TypeScript and JavaScript applications.

## Core Philosophy

FileOnix v2 bridges the gap between high-speed native monitoring and developer experience. It prioritizes the **Bun** runtime for near-instant execution while maintaining full compatibility with `tsx`, `ts-node`, and standard Node.js. Every visual detail of the CLI — from the gradient banner to the timestamped event log — is designed to feel premium.

## What's New in v2

- **Total rewrite** — Clean architecture, no legacy code
- **Premium CLI** — Chrome/silver + electric blue theme matching the FileOnix logo
- **Smarter detection** — MD5 file hashing eliminates false restarts from touch/save-without-change
- **Batch mode** — Groups rapid file changes into single restarts
- **Auto runtime detection** — Finds bun → tsx → ts-node → node automatically
- **Graceful shutdown** — SIGTERM with configurable timeout before SIGKILL
- **Zero dependencies** — Pure Go stdlib, no vendor bloat

## Installation

### Via npm (recommended)
```bash
npm install -g fileonix
```

### Via xfpm
```bash
xfpm install fileonix
```

### Build from source
```bash
git clone https://github.com/Nehonix-Team/fileonix
cd fileonix
go build -o bin/fileonix .
```

## Quick Start

```bash
# Zero config — watches src/, auto-detects runtime
fileonix -script src/index.ts

# Full control
fileonix -script server.ts -watch src,config -runner bun -clear -batch

# Generate config file
fileonix init
```

## Configuration

`fileonix.config.json` or `.fileonixrc.json` in your project root:

```json
{
  "script":           "src/index.ts",
  "watch":            ["src", "internal"],
  "ignore":           ["node_modules", "dist", ".git"],
  "typescriptRunner": "bun",
  "clearScreen":      true,
  "debounceMs":       100,
  "batchMode":        false,
  "useFileHash":      true,
  "maxRestarts":      -1,
  "gracefulMs":       3000
}
```

## CLI Options

| Flag | Description |
|------|-------------|
| `-script <file>` | Entry point to watch and execute |
| `-watch <dirs>` | Comma-separated directories to watch |
| `-runner <n>` | Runtime: `bun` \| `tsx` \| `ts-node` \| `node` |
| `-ignore <dirs>` | Additional directories to ignore |
| `-ext <exts>` | Extensions to watch (default: `.ts,.js`) |
| `-delay <ms>` | Debounce delay in ms (default: `100`) |
| `-batch` | Enable batch mode |
| `-clear` | Clear screen on restart |
| `--no-hash` | Disable hash-based change detection |
| `init` | Generate a config file |

## Environment Variables

FileOnix injects these into your process:

| Variable | Value |
|----------|-------|
| `FILEONIX` | `1` |
| `FILEONIX_RESTART` | Restart count (integer) |

## Architecture

```
fileonix/
├── main.go                    # Entry point, arg routing
└── internal/
    ├── config/config.go       # Config loading (file + CLI args)
    ├── ui/ui.go               # Premium terminal UI
    └── watcher/watcher.go     # Core FS engine + process manager
```

## License

MIT — Nehonix Team

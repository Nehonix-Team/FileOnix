<p align="center">
  <img src="assets/FileOnix.rmbg.png" width="600" alt="FileOnix Logo">
</p>

# Nehonix FileOnix

**High-Performance Native Watcher System**

Nehonix FileOnix (formerly QuickDev) is a professional, native file watcher and development server. Engineered for extreme performance, it serves as the native hot-reload engine for modern TypeScript and JavaScript applications, offering significant speed advantages over traditional Node-based watchers.

## Core Philosophy

FileOnix is designed to bridge the gap between high-speed native monitoring and developer productivity. It prioritizes the **Bun** runtime for near-instant execution while maintaining full compatibility with `tsx`, `ts-node`, and standard Node.js.

## Key Features

- **Native Performance** — Core monitoring engine written in Go for low-latency filesystem events and minimal CPU overhead.
- **Bun-First Execution** — Automatically detects and utilizes the Bun runtime for the fastest possible hot-reload cycles.
- **Smart Process Management** — Graceful shutdowns, restart limits, and environment variable preservation.
- **Sophisticated Monitoring** — File hashing for precise change detection, parallel processing, and debounced events.
- **Zero-Emoji Professional CLI** — Clean, symbol-based terminal output designed for enterprise environments.

## Installation

### Recommended: Via xfpm (XyPriss Fast Package Manager)

Install FileOnix using [xfpm](https://github.com/Nehonix-Team/XFMP):

```bash
xfpm install fileonix
```

### Manual Installation

FileOnix ships as a pre-compiled native binary. You can also install it globally via npm:

```bash
npm install -g fileonix
```

## Configuration

FileOnix is designed to be "zero-config" but offers deep customization via `fileonix.config.json` or `.fileonixrc.json` in your project root.

```json
{
  "script": "src/index.ts",
  "watch": ["src", "internal"],
  "ignore": ["node_modules", "dist"],
  "typescriptRunner": "bun",
  "clearScreen": true
}
```

## Usage

### Basic Command

```bash
fileonix -script src/index.ts
```

### Advanced Command

```bash
fileonix -script server.ts --watch="src,config" --ext=".ts,.js" --batch=true
```

## Community & Resources

- **XyPriss Framework**: [Documentation](https://xypriss.nehonix.com)
- **Package Manager**: [XFPM Repository](https://github.com/Nehonix-Team/XFMP)
- **Support**: [Nehonix Team](https://github.com/Nehonix-Team)

## License

Licensed under the MIT License.

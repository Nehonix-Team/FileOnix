#!/usr/bin/env node

/**
 * ╔══════════════════════════════════════════════════════════════╗
 * ║                    FileOnix  ·  Installer                   ║
 * ║              Nehonix Team  —  github.com/Nehonix-Team        ║
 * ╚══════════════════════════════════════════════════════════════╝
 *
 * Post-install script — downloads the correct platform binary
 * from GitHub Releases and makes it ready to run.
 */

"use strict"; 

const fs = require("fs");
const path = require("path");
const os = require("os");
const https = require("https");

// ─── Resolve package version safely ──────────────────────────────────────────
let VERSION = "unknown";
try {
  VERSION = require("../package.json").version;
} catch (_) {
  // package.json not reachable from scripts/ — silently ignore
}

const GITHUB_REPO = "Nehonix-Team/FileOnix";
const MAX_REDIRECTS = 10;

// ─── Terminal styling ─────────────────────────────────────────────────────────
const isTTY = process.stdout.isTTY;

// Force colors even if npm masks TTY
const C = {
  reset: "\x1b[0m",
  bold: "\x1b[1m",
  dim: "\x1b[2m",

  // FileOnix palette  — cyan / blue / gray
  navy: "\x1b[34m", // Blue
  skyBlue: "\x1b[36m", // Cyan
  brightBlue: "\x1b[36m", // Cyan
  steel: "\x1b[90m", // Gray
  white: "\x1b[37m", // White

  // Status colours
  green: "\x1b[32m", // Green
  yellow: "\x1b[33m", // Yellow
  red: "\x1b[31m", // Red
  orange: "\x1b[33m", // Yellow
};

// ─── Print helpers ────────────────────────────────────────────────────────────

function banner() {
  const line = `${C.brightBlue}${C.bold}`;
  const dim = `${C.dim}${C.steel}`;

  console.log();
  console.log(
    `${line}  ███████╗██╗██╗     ███████╗ ██████╗ ███╗   ██╗██╗██╗  ██╗${C.reset}`,
  );
  console.log(
    `${line}  ██╔════╝██║██║     ██╔════╝██╔═══██╗████╗  ██║██║╚██╗██╔╝${C.reset}`,
  );
  console.log(
    `${line}  █████╗  ██║██║     █████╗  ██║   ██║██╔██╗ ██║██║ ╚███╔╝ ${C.reset}`,
  );
  console.log(
    `${line}  ██╔══╝  ██║██║     ██╔══╝  ██║   ██║██║╚██╗██║██║ ██╔██╗ ${C.reset}`,
  );
  console.log(
    `${line}  ██║     ██║███████╗███████╗╚██████╔╝██║ ╚████║██║██╔╝ ██╗${C.reset}`,
  );
  console.log(
    `${line}  ╚═╝     ╚═╝╚══════╝╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝╚═╝  ╚═╝${C.reset}`,
  );
  console.log();
  console.log(
    `${dim}  ══════════════════════════════════════════════════════════${C.reset}`,
  );
  console.log(
    `  ${C.steel}${C.bold}  Nehonix FileOnix${C.reset}${C.dim}${C.steel}  ·  v${VERSION}  ·  Binary Installer${C.reset}`,
  );
  console.log(
    `${dim}  ══════════════════════════════════════════════════════════${C.reset}`,
  );
  console.log();
}

function log(level, msg) {
  const icons = {
    info: `${C.brightBlue}${C.bold}  ◆${C.reset}`,
    success: `${C.green}${C.bold}  ✔${C.reset}`,
    warn: `${C.yellow}${C.bold}  ⚠${C.reset}`,
    error: `${C.red}${C.bold}  ✖${C.reset}`,
    step: `${C.skyBlue}${C.bold}  ▸${C.reset}`,
    muted: `${C.dim}  ·${C.reset}`,
  };
  console.log(`${icons[level] || icons.info}  ${msg}${C.reset}`);
}

function section(title) {
  console.log();
  console.log(
    `  ${C.steel}${C.dim}─────────────────────────────────────────${C.reset}`,
  );
  console.log(`  ${C.brightBlue}${C.bold}${title}${C.reset}`);
  console.log(
    `  ${C.steel}${C.dim}─────────────────────────────────────────${C.reset}`,
  );
}

// ─── Spinner (TTY only) ───────────────────────────────────────────────────────

function createSpinner(label) {
  if (!isTTY || noColor) {
    process.stdout.write(`  ${C.brightBlue}↓${C.reset}  ${label} …\n`);
    return { update: () => {}, stop: () => {} };
  }

  const frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
  let i = 0;
  const iv = setInterval(() => {
    const frame = `${C.brightBlue}${C.bold}${frames[i++ % frames.length]}${C.reset}`;
    process.stdout.write(`\r  ${frame}  ${C.steel}${label}${C.reset}    `);
  }, 80);

  return {
    update(msg) {
      label = msg;
    },
    stop(success, finalMsg) {
      clearInterval(iv);
      const icon = success
        ? `${C.green}${C.bold}✔${C.reset}`
        : `${C.red}${C.bold}✖${C.reset}`;
      process.stdout.write(`\r  ${icon}  ${finalMsg || label}\n`);
    },
  };
}

// ─── Progress bar ─────────────────────────────────────────────────────────────

function renderProgressBar(received, total) {
  if (!isTTY || noColor) return;
  const width = 36;
  const pct = total > 0 ? Math.min(received / total, 1) : 0;
  const filled = Math.round(width * pct);
  const empty = width - filled;
  const bar = `${C.skyBlue}${"█".repeat(filled)}${C.dim}${"░".repeat(empty)}${C.reset}`;
  const mb = (n) => (n / 1_048_576).toFixed(1) + " MB";
  const pctStr = `${Math.round(pct * 100)}%`.padStart(4);
  const sizeStr =
    total > 0 ? `${mb(received)} / ${mb(total)}` : `${mb(received)}`;
  process.stdout.write(
    `\r  ${bar}  ${C.steel}${pctStr}  ${C.dim}${sizeStr}${C.reset}   `,
  );
}

// ─── Platform detection ───────────────────────────────────────────────────────

function getPlatformInfo() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = {
    win32: "windows",
    darwin: "darwin",
    linux: "linux",
  };
  const archMap = {
    x64: "amd64",
    arm64: "arm64",
  };

  const platformName = platformMap[platform];
  const archName = archMap[arch];

  if (!platformName) {
    throw new Error(
      `Unsupported platform: ${C.orange}${platform}${C.reset}\n` +
        `  Supported platforms: ${Object.keys(platformMap).join(", ")}`,
    );
  }
  if (!archName) {
    throw new Error(
      `Unsupported architecture: ${C.orange}${arch}${C.reset}\n` +
        `  Supported architectures: ${Object.keys(archMap).join(", ")}`,
    );
  }

  const extension = platform === "win32" ? ".exe" : "";
  const binaryName = `fileonix-${platformName}-${archName}${extension}`;
  const downloadUrl = `https://github.com/${GITHUB_REPO}/releases/latest/download/${binaryName}`;

  return { platform, arch, platformName, archName, binaryName, downloadUrl };
}

// ─── Download with redirects + progress ──────────────────────────────────────

function downloadFile(url, destination, spinner, redirectCount = 0) {
  return new Promise((resolve, reject) => {
    if (redirectCount > MAX_REDIRECTS) {
      return reject(new Error("Too many redirects — aborting download."));
    }

    const file = fs.createWriteStream(destination);
    let received = 0;
    let total = 0;

    const cleanup = (err) => {
      file.destroy();
      try {
        if (fs.existsSync(destination)) fs.unlinkSync(destination);
      } catch (_) {}
      reject(err);
    };

    https
      .get(
        url,
        { headers: { "User-Agent": `fileonix-installer/${VERSION}` } },
        (res) => {
          // Redirect
          if (
            res.statusCode === 301 ||
            res.statusCode === 302 ||
            res.statusCode === 307
          ) {
            file.close();
            try {
              fs.unlinkSync(destination);
            } catch (_) {}
            const loc = res.headers.location;
            if (!loc)
              return reject(new Error("Redirect with no Location header."));
            spinner.update(`Redirecting … (${redirectCount + 1})`);
            return downloadFile(loc, destination, spinner, redirectCount + 1)
              .then(resolve)
              .catch(reject);
          }

          if (res.statusCode !== 200) {
            return cleanup(
              new Error(`HTTP ${res.statusCode}: ${res.statusMessage}`),
            );
          }

          total = parseInt(res.headers["content-length"] || "0", 10);
          res.pipe(file);

          res.on("data", (chunk) => {
            received += chunk.length;
            renderProgressBar(received, total);
          });

          file.on("finish", () => {
            file.close();
            if (isTTY && !noColor) process.stdout.write("\n");
            resolve();
          });

          file.on("error", cleanup);
          res.on("error", cleanup);
        },
      )
      .on("error", cleanup);
  });
}

// ─── Verify binary is non-empty and executable ────────────────────────────────

function verifyBinary(binaryPath) {
  const stat = fs.statSync(binaryPath);
  if (stat.size < 1024) {
    throw new Error(
      `Binary looks invalid (${stat.size} bytes). ` +
        `The release asset may be missing or the download was interrupted.`,
    );
  }
}

// ─── Main ─────────────────────────────────────────────────────────────────────

async function main() {
  banner();

  // ── 1. Detect platform ───────────────────────────────────────────────────
  section("Platform Detection");

  let platformInfo;
  try {
    platformInfo = getPlatformInfo();
  } catch (err) {
    log("error", err.message);
    process.exit(1);
  }

  const { platform, arch, platformName, archName, binaryName, downloadUrl } =
    platformInfo;

  log("info", `OS       ${C.steel}${platformName} (${platform})${C.reset}`);
  log("info", `Arch     ${C.steel}${archName} (${arch})${C.reset}`);
  log("info", `Binary   ${C.brightBlue}${binaryName}${C.reset}`);
  log("info", `Source   ${C.dim}${downloadUrl}${C.reset}`);

  // ── 2. Prepare bin/ directory ────────────────────────────────────────────
  section("Setup");

  const binDir = path.join(__dirname, "..", "bin");
  const binaryPath = path.join(binDir, binaryName);

  try {
    fs.mkdirSync(binDir, { recursive: true });
    log("success", `bin/ directory ready`);
  } catch (err) {
    log("error", `Cannot create bin/ directory: ${err.message}`);
    process.exit(1);
  }

  // ── 3. Download (skip if already cached) ────────────────────────────────
  section("Download");

  if (fs.existsSync(binaryPath)) {
    try {
      verifyBinary(binaryPath);
      log("success", `Binary already cached — skipping download`);
      log("muted", `${C.dim}${binaryPath}${C.reset}`);
    } catch (_) {
      log("warn", "Cached binary appears corrupt — re-downloading…");
      fs.unlinkSync(binaryPath);
    }
  }

  if (!fs.existsSync(binaryPath)) {
    const spinner = createSpinner(`Downloading ${binaryName}`);
    try {
      await downloadFile(downloadUrl, binaryPath, spinner);
      verifyBinary(binaryPath);
      spinner.stop(true, `Downloaded ${C.brightBlue}${binaryName}${C.reset}`);
    } catch (err) {
      spinner.stop(false, `Download failed`);
      console.log();
      log("error", `${err.message}`);
      console.log();
      log("muted", `Possible causes:`);
      log(
        "muted",
        `  ${C.steel}1. Release not yet published on GitHub${C.reset}`,
      );
      log(
        "muted",
        `  ${C.steel}2. Platform / architecture not supported${C.reset}`,
      );
      log("muted", `  ${C.steel}3. Network connectivity issue${C.reset}`);
      console.log();
      log(
        "step",
        `Check releases: ${C.brightBlue}https://github.com/${GITHUB_REPO}/releases${C.reset}`,
      );
      console.log();
      process.exit(1);
    }
  }

  // ── 4. Make executable (Unix) ────────────────────────────────────────────
  if (platform !== "win32") {
    try {
      fs.chmodSync(binaryPath, 0o755);
      log("success", "Executable permissions set (chmod 755)");
    } catch (err) {
      log("warn", `Could not chmod binary: ${err.message}`);
    }
  }

  // ── 5. Done ───────────────────────────────────────────────────────────────
  console.log();
  console.log(
    `  ${C.steel}${C.dim}════════════════════════════════════════════${C.reset}`,
  );
  console.log(
    `  ${C.green}${C.bold}  🎉  FileOnix installed successfully!${C.reset}`,
  );
  console.log(
    `  ${C.steel}${C.dim}════════════════════════════════════════════${C.reset}`,
  );
  console.log();
  log("success", `Binary  ${C.brightBlue}${binaryName}${C.reset}`);
  log(
    "success",
    `Ready   ${C.green}fileonix${C.reset} command available globally`,
  );
  console.log();
  log("step", `Quick start:`);
  log("muted", `  ${C.brightBlue}fileonix -script your-script.js${C.reset}`);
  log("muted", `  ${C.brightBlue}fileonix --help${C.reset}`);
  console.log();
  log(
    "muted",
    `Docs & releases: ${C.dim}https://github.com/${GITHUB_REPO}${C.reset}`,
  );
  console.log();
  process.exit(0);
}

main().catch((err) => {
  console.error(
    `\n${C.red}${C.bold}  ✖  Unexpected error:${C.reset}  ${err.message}\n`,
  );
  process.exit(1);
});

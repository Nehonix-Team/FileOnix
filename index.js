#!/usr/bin/env node

/**
 * Nehonix FileOnix - Professional-grade file watcher and development server
 *
 * This is the main entry point for the npm package.
 * It detects the platform and executes the appropriate binary.
 */

const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");
const os = require("os");

/**
 * Get the platform-specific binary name
 */
function getBinaryName() {
  const platform = os.platform();
  const arch = os.arch();

  let platformName;
  let archName;
  let extension = "";

  // Map Node.js platform names to our binary names
  switch (platform) {
    case "win32":
      platformName = "windows";
      extension = ".exe";
      break;
    case "darwin":
      platformName = "darwin";
      break;
    case "linux":
      platformName = "linux";
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }

  // Map Node.js arch names to our binary names
  switch (arch) {
    case "x64":
      archName = "amd64";
      break;
    case "arm64":
      archName = "arm64";
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }

  return `fileonix-${platformName}-${archName}${extension}`;
}

/**
 * Get the path to the binary
 */
function getBinaryPath() {
  const binaryName = getBinaryName();
  const binaryPath = path.join(__dirname, "bin", binaryName);

  if (!fs.existsSync(binaryPath)) {
    throw new Error(`Binary not found: ${binaryPath}`);
  }

  return binaryPath;
}

/**
 * Execute the binary with the provided arguments
 */
function main() {
  try {
    const binaryPath = getBinaryPath();
    const args = process.argv.slice(2);

    // Spawn the binary process
    const child = spawn(binaryPath, args, {
      stdio: "inherit",
      windowsHide: false,
    });

    // Handle process exit
    child.on("exit", (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
      } else {
        process.exit(code);
      }
    });

    // Handle errors
    child.on("error", (err) => {
      console.error("Failed to start fileonix:", err.message);
      process.exit(1);
    });
  } catch (error) {
    console.error("Error:", error.message);
    console.error("\nIf this error persists, please report it at:");
    console.error("https://github.com/nehonix/fileonix/issues");
    process.exit(1);
  }
}

// Only run if this file is executed directly
if (require.main === module) {
  main();
}

module.exports = { getBinaryName, getBinaryPath, main };

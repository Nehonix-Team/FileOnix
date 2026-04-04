#!/usr/bin/env node

/**
 * Post-install script for Nehonix QuickDev
 *
 * This script runs after npm install and downloads the appropriate binary
 * from GitHub releases.
 */

const fs = require("fs");
const path = require("path");
const os = require("os");
const https = require("https");
const { execSync } = require("child_process");

// Configuration
const GITHUB_REPO = "Nehonix-Team/FileOnix";
const VERSION = require("../package.json").version;

/**
 * Get the platform-specific binary name and download URL
 */
function getBinaryInfo() {
  const platform = os.platform();
  const arch = os.arch();

  let platformName;
  let archName;
  let extension = "";

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
      console.error(`Unsupported platform: ${platform}`);
      process.exit(1);
  }

  switch (arch) {
    case "x64":
      archName = "amd64";
      break;
    case "arm64":
      archName = "arm64";
      break;
    default:
      console.error(`Unsupported architecture: ${arch}`);
      process.exit(1);
  }

  const binaryName = `fileonix-${platformName}-${archName}${extension}`;
  const downloadUrl = `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${binaryName}`;

  return { binaryName, downloadUrl };
}

/**
 * Download a file from URL
 */
function downloadFile(url, destination) {
  return new Promise((resolve, reject) => {
    console.log(`Downloading: ${url}`);

    const file = fs.createWriteStream(destination);

    https
      .get(url, (response) => {
        // Handle redirects
        if (response.statusCode === 302 || response.statusCode === 301) {
          file.close();
          fs.unlinkSync(destination);
          return downloadFile(response.headers.location, destination)
            .then(resolve)
            .catch(reject);
        }

        if (response.statusCode !== 200) {
          file.close();
          fs.unlinkSync(destination);
          return reject(
            new Error(
              `Download failed: ${response.statusCode} ${response.statusMessage}`,
            ),
          );
        }

        response.pipe(file);

        file.on("finish", () => {
          file.close();
          resolve();
        });

        file.on("error", (err) => {
          file.close();
          fs.unlinkSync(destination);
          reject(err);
        });
      })
      .on("error", (err) => {
        file.close();
        fs.unlinkSync(destination);
        reject(err);
      });
  });
}

/**
 * Main installation logic
 */
async function main() {
  console.log("Setting up Nehonix FileOnix...");

  try {
    const { binaryName, downloadUrl } = getBinaryInfo();
    const binDir = path.join(__dirname, "..", "bin");
    const binaryPath = path.join(binDir, binaryName);

    // Create bin directory if it doesn't exist
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }

    // Check if binary already exists
    if (fs.existsSync(binaryPath)) {
      console.log(`✓ Binary already exists: ${binaryName}`);
    } else {
      console.log(`Downloading binary for your platform: ${binaryName}`);

      try {
        await downloadFile(downloadUrl, binaryPath);
        console.log(`✓ Downloaded: ${binaryName}`);
      } catch (error) {
        console.error(`❌ Failed to download binary: ${error.message}`);
        console.error(`URL: ${downloadUrl}`);
        console.error("");
        console.error("This might happen if:");
        console.error("1. The release is not yet available on GitHub");
        console.error("2. Your platform is not supported");
        console.error("3. Network connectivity issues");
        console.error("");
        console.error(
          "Please check: https://github.com/Nehonix-Team/FileOnix/releases",
        );
        process.exit(1);
      }
    }

    // Make the binary executable on Unix systems
    if (os.platform() !== "win32") {
      try {
        fs.chmodSync(binaryPath, 0o755);
        console.log(`✓ Made binary executable`);
      } catch (error) {
        console.warn(
          "Warning: Could not make binary executable:",
          error.message,
        );
      }
    }

    console.log("");
    console.log(`🎉 FileOnix installed successfully!`);
    console.log(`✓ Binary: ${binaryName}`);
    console.log(`✓ You can now use 'fileonix' command globally`);
    console.log("");
    console.log("Get started:");
    console.log("  fileonix -script your-script.js");
    console.log("");
    console.log("For more information:");
    console.log("  fileonix --help");
    console.log("  https://github.com/Nehonix-Team/FileOnix");
  } catch (error) {
    console.error("❌ Installation failed:", error.message);
    process.exit(1);
  }
}

main().catch(console.error);

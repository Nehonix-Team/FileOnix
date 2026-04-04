#!/usr/bin/env node

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const os = require("os");
const { version } = require("../package.json");

// Ensure bin directory exists
const binDir = path.join(__dirname, "..", "bin");
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir);
}

// Platform configurations
const platforms = [
  { GOOS: "windows", GOARCH: "amd64", suffix: ".exe" },
  { GOOS: "windows", GOARCH: "arm64", suffix: ".exe" },
  { GOOS: "linux", GOARCH: "amd64", suffix: "" },
  { GOOS: "linux", GOARCH: "arm64", suffix: "" },
  { GOOS: "darwin", GOARCH: "amd64", suffix: "" },
  { GOOS: "darwin", GOARCH: "arm64", suffix: "" },
];

// Determine if we're running on Windows
const isWindows = os.platform() === "win32";

// Function to create tar archive
function createTarArchive(outputName, outputPath, tarPath) {
  if (isWindows) {
    // On Windows, we'll use 7zip if available, otherwise fall back to tar
    try {
      // Try using 7zip
      execSync(`7z a -ttar "${tarPath}.tmp" "${outputName}"`, { cwd: binDir });
      execSync(`7z a -tgzip "${tarPath}" "${tarPath}.tmp"`, { cwd: binDir });
      // Clean up temporary file
      fs.unlinkSync(path.join(binDir, `${tarPath}.tmp`));
    } catch (error) {
      // Fallback to tar if available
      try {
        execSync(`tar -czf "${tarPath}" -C "${binDir}" "${outputName}"`);
      } catch (tarError) {
        console.warn(
          "Warning: Could not create tar.gz archive. Please install 7zip or tar.",
        );
        // Copy the binary as is
        fs.copyFileSync(outputPath, tarPath);
      }
    }
  } else {
    // On Unix systems, use tar directly
    execSync(
      `chmod +x "${outputPath}" && tar -czf "${tarPath}" -C "${binDir}" "${outputName}"`,
    );
  }
}

// Build for each platform
platforms.forEach((platform) => {
  const { GOOS, GOARCH, suffix } = platform;
  const outputName = `fileonix-${GOOS}-${GOARCH}${suffix}`;
  const outputPath = path.join(binDir, outputName);

  console.log(`Building for ${GOOS} ${GOARCH}...`);

  try {
    // Build the binary
    execSync(`go build -o "${outputPath}" ./internal`, {
      env: {
        ...process.env,
        GOOS,
        GOARCH,
        CGO_ENABLED: "0",
      },
      stdio: "inherit",
    });

    // Create tar.gz archive
    const tarName = `fileonix-${GOOS}-${GOARCH}.tar.gz`;
    const tarPath = path.join(binDir, tarName);

    // Set file permissions (Windows doesn't need chmod)
    if (!isWindows) {
      fs.chmodSync(outputPath, 0o755);
    }

    // Create archive
    createTarArchive(outputName, outputPath, tarPath);

    console.log(`✓ Built ${outputName}`);
  } catch (error) {
    console.error(`✗ Failed to build for ${GOOS} ${GOARCH}:`, error.message);
    process.exit(1);
  }
});

console.log("\nAll builds completed successfully!");

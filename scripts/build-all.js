#!/usr/bin/env node

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const os = require("os");

const rootDir = path.join(__dirname, "..");
const buildShPath = path.join(rootDir, "build.sh");

// On Unix systems, delegate to hardened build.sh
if (os.platform() !== "win32" && fs.existsSync(buildShPath)) {
  try {
    const args = process.argv.slice(2).join(" ");
    execSync(`bash "${buildShPath}" ${args}`, {
      cwd: rootDir,
      stdio: "inherit",
    });
    process.exit(0);
  } catch (err) {
    process.exit(err.status || 1);
  }
}

// Fallback for environments without bash
const binDir = path.join(rootDir, "bin");
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

const platforms = [
  { GOOS: "windows", GOARCH: "amd64", suffix: ".exe" },
  { GOOS: "windows", GOARCH: "arm64", suffix: ".exe" },
  { GOOS: "linux", GOARCH: "amd64", suffix: "" },
  { GOOS: "linux", GOARCH: "arm64", suffix: "" },
  { GOOS: "darwin", GOARCH: "amd64", suffix: "" },
  { GOOS: "darwin", GOARCH: "arm64", suffix: "" },
];

const isWindows = os.platform() === "win32";

function stripBuildinfo(targetPath) {
  try {
    const script = `
import sys, re
target = sys.argv[1]
try:
    with open(target, 'r+b') as f:
        data = bytearray(f.read())
        changed = False
        for m in re.finditer(b'\\\\xff Go buildinf:', data):
            data[m.start():m.start()+32] = b'\\x00' * 32
            changed = True
        for m in re.finditer(b'path\\\\t', data):
            end = data.find(b'\\x00', m.start())
            if end != -1:
                data[m.start():end] = b'\\x00' * (end - m.start())
                changed = True
        if changed:
            f.seek(0)
            f.write(data)
            f.truncate()
except Exception:
    pass
`;
    execSync(`python3 -c "${script.replace(/\n/g, " ")}" "${targetPath}"`, {
      stdio: "ignore",
    });
  } catch (_) {}
}

function createTarArchive(outputName, outputPath, tarPath) {
  if (isWindows) {
    try {
      execSync(`7z a -ttar "${tarPath}.tmp" "${outputName}"`, { cwd: binDir });
      execSync(`7z a -tgzip "${tarPath}" "${tarPath}.tmp"`, { cwd: binDir });
      fs.unlinkSync(path.join(binDir, `${tarPath}.tmp`));
    } catch (error) {
      try {
        execSync(`tar -czf "${tarPath}" -C "${binDir}" "${outputName}"`);
      } catch (tarError) {
        fs.copyFileSync(outputPath, tarPath);
      }
    }
  } else {
    execSync(
      `chmod +x "${outputPath}" && tar -czf "${tarPath}" -C "${binDir}" "${outputName}"`
    );
  }
}

platforms.forEach((platform) => {
  const { GOOS, GOARCH, suffix } = platform;
  const outputName = `fileonix-${GOOS}-${GOARCH}${suffix}`;
  const outputPath = path.join(binDir, outputName);

  console.log(`Building for ${GOOS} ${GOARCH}...`);

  try {
    execSync(
      `go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o "${outputPath}" ./internal`,
      {
        cwd: rootDir,
        env: {
          ...process.env,
          GOOS,
          GOARCH,
          CGO_ENABLED: "0",
        },
        stdio: "inherit",
      }
    );

    stripBuildinfo(outputPath);

    const tarName = `fileonix-${GOOS}-${GOARCH}.tar.gz`;
    const tarPath = path.join(binDir, tarName);

    if (!isWindows) {
      fs.chmodSync(outputPath, 0o755);
    }

    createTarArchive(outputName, outputPath, tarPath);
    console.log(`✓ Built ${outputName}`);
  } catch (error) {
    console.error(`✗ Failed to build for ${GOOS} ${GOARCH}:`, error.message);
    process.exit(1);
  }
});

console.log("\nAll builds completed successfully!");

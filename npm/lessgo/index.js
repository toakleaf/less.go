"use strict";

const os = require("os");
const path = require("path");
const { execFileSync, spawn } = require("child_process");
const fs = require("fs");

/**
 * Platform to package name mapping
 */
const PLATFORM_PACKAGES = {
  "darwin-arm64": "@lessgo/darwin-arm64",
  "darwin-x64": "@lessgo/darwin-x64",
  "linux-x64": "@lessgo/linux-x64",
  "linux-arm64": "@lessgo/linux-arm64",
  "win32-x64": "@lessgo/win32-x64",
  "win32-arm64": "@lessgo/win32-arm64",
};

/**
 * Get the platform key for the current system
 * @returns {string} Platform key (e.g., "darwin-arm64")
 */
function getPlatformKey() {
  const platform = os.platform();
  const arch = os.arch();
  return `${platform}-${arch}`;
}

/**
 * Get the package name for the current platform
 * @returns {string|null} Package name or null if unsupported
 */
function getPlatformPackage() {
  const key = getPlatformKey();
  return PLATFORM_PACKAGES[key] || null;
}

/**
 * Try to resolve the binary path from an installed platform package
 * @param {string} packageName - The platform-specific package name
 * @returns {string|null} Path to binary or null if not found
 */
function resolveBinaryFromPackage(packageName) {
  try {
    // Try to resolve the package
    const packagePath = require.resolve(`${packageName}/package.json`);
    const packageDir = path.dirname(packagePath);
    const packageJson = require(packagePath);

    // The binary name should be in the package's bin field
    const binName = os.platform() === "win32" ? "lessc-go.exe" : "lessc-go";
    const binPath = path.join(packageDir, "bin", binName);

    if (fs.existsSync(binPath)) {
      return binPath;
    }

    // Fallback: check if there's a bin field in package.json
    if (packageJson.bin) {
      const binValue =
        typeof packageJson.bin === "string"
          ? packageJson.bin
          : packageJson.bin["lessc-go"];
      if (binValue) {
        const resolvedBin = path.join(packageDir, binValue);
        if (fs.existsSync(resolvedBin)) {
          return resolvedBin;
        }
      }
    }
  } catch (e) {
    // Package not installed
  }
  return null;
}

/**
 * Get the path to the lessc-go binary for the current platform
 * @returns {string} Path to the binary
 * @throws {Error} If binary cannot be found
 */
function getBinaryPath() {
  const platformKey = getPlatformKey();
  const packageName = getPlatformPackage();

  if (!packageName) {
    throw new Error(
      `Unsupported platform: ${platformKey}. ` +
        `less.go currently supports: ${Object.keys(PLATFORM_PACKAGES).join(", ")}`
    );
  }

  const binaryPath = resolveBinaryFromPackage(packageName);

  if (!binaryPath) {
    throw new Error(
      `Could not find the lessc-go binary. The platform-specific package ` +
        `"${packageName}" may not be installed. Try reinstalling lessgo`
    );
  }

  return binaryPath;
}

/**
 * Compile a LESS file or string
 * @param {string} input - File path or LESS content
 * @param {Object} options - Compilation options
 * @param {boolean} [options.string=false] - If true, treat input as LESS content instead of file path
 * @param {string[]} [options.paths] - Include paths for @import
 * @param {boolean} [options.compress=false] - Minify output
 * @param {boolean} [options.sourceMap=false] - Generate source map
 * @param {string} [options.sourceMapFilename] - Source map filename
 * @param {Object} [options.globalVars] - Global variables
 * @param {Object} [options.modifyVars] - Modify variables
 * @returns {Promise<{css: string, map?: string}>} Compilation result
 */
async function compile(input, options = {}) {
  const binaryPath = getBinaryPath();
  const args = [];

  // Build arguments from options
  if (options.string) {
    args.push("--stdin");
  }

  if (options.paths && options.paths.length > 0) {
    args.push(`--include-path=${options.paths.join(os.platform() === "win32" ? ";" : ":")}`);
  }

  if (options.compress) {
    args.push("--compress");
  }

  if (options.sourceMap) {
    args.push("--source-map");
    if (options.sourceMapFilename) {
      args.push(`--source-map-filename=${options.sourceMapFilename}`);
    }
  }

  if (options.globalVars) {
    for (const [key, value] of Object.entries(options.globalVars)) {
      args.push(`--global-var=${key}=${value}`);
    }
  }

  if (options.modifyVars) {
    for (const [key, value] of Object.entries(options.modifyVars)) {
      args.push(`--modify-var=${key}=${value}`);
    }
  }

  // Add input file if not using stdin
  if (!options.string) {
    args.push(input);
  }

  return new Promise((resolve, reject) => {
    const proc = spawn(binaryPath, args, {
      stdio: ["pipe", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";

    proc.stdout.on("data", (data) => {
      stdout += data.toString();
    });

    proc.stderr.on("data", (data) => {
      stderr += data.toString();
    });

    proc.on("error", (err) => {
      reject(new Error(`Failed to run lessc-go: ${err.message}`));
    });

    proc.on("close", (code) => {
      if (code !== 0) {
        reject(new Error(stderr || `lessc-go exited with code ${code}`));
      } else {
        resolve({ css: stdout });
      }
    });

    // If using stdin, write the input
    if (options.string) {
      proc.stdin.write(input);
      proc.stdin.end();
    }
  });
}

/**
 * Synchronously compile a LESS file or string
 * @param {string} input - File path or LESS content
 * @param {Object} options - Compilation options (same as compile)
 * @returns {{css: string, map?: string}} Compilation result
 */
function compileSync(input, options = {}) {
  const binaryPath = getBinaryPath();
  const args = [];

  // Build arguments from options
  if (options.string) {
    args.push("--stdin");
  }

  if (options.paths && options.paths.length > 0) {
    args.push(`--include-path=${options.paths.join(os.platform() === "win32" ? ";" : ":")}`);
  }

  if (options.compress) {
    args.push("--compress");
  }

  if (options.sourceMap) {
    args.push("--source-map");
    if (options.sourceMapFilename) {
      args.push(`--source-map-filename=${options.sourceMapFilename}`);
    }
  }

  if (options.globalVars) {
    for (const [key, value] of Object.entries(options.globalVars)) {
      args.push(`--global-var=${key}=${value}`);
    }
  }

  if (options.modifyVars) {
    for (const [key, value] of Object.entries(options.modifyVars)) {
      args.push(`--modify-var=${key}=${value}`);
    }
  }

  // Add input file if not using stdin
  if (!options.string) {
    args.push(input);
  }

  try {
    const result = execFileSync(binaryPath, args, {
      encoding: "utf8",
      input: options.string ? input : undefined,
      maxBuffer: 50 * 1024 * 1024, // 50MB buffer
    });

    return { css: result };
  } catch (err) {
    if (err.stderr) {
      throw new Error(err.stderr);
    }
    throw err;
  }
}

/**
 * Run the compiler with raw arguments (like CLI)
 * @param {string[]} args - Command line arguments
 * @returns {Promise<{code: number, stdout: string, stderr: string}>}
 */
async function run(args = []) {
  const binaryPath = getBinaryPath();

  return new Promise((resolve) => {
    const proc = spawn(binaryPath, args, {
      stdio: ["inherit", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";

    proc.stdout.on("data", (data) => {
      stdout += data.toString();
    });

    proc.stderr.on("data", (data) => {
      stderr += data.toString();
    });

    proc.on("error", (err) => {
      resolve({ code: 1, stdout: "", stderr: err.message });
    });

    proc.on("close", (code) => {
      resolve({ code: code || 0, stdout, stderr });
    });
  });
}

/**
 * Verify the binary is properly installed (used by postinstall)
 */
function verifyInstallation() {
  const platformKey = getPlatformKey();
  const packageName = getPlatformPackage();

  console.log(`lessgo: Detected platform ${platformKey}`);

  if (!packageName) {
    console.warn(
      `lessgo: Warning - Unsupported platform "${platformKey}". ` +
        `Binary execution will not work on this platform.`
    );
    return false;
  }

  try {
    const binaryPath = getBinaryPath();
    console.log(`lessgo: Found binary at ${binaryPath}`);

    // Try to run --version to verify it works
    const result = execFileSync(binaryPath, ["--version"], {
      encoding: "utf8",
      timeout: 5000,
    });
    console.log(`lessgo: ${result.trim()}`);
    return true;
  } catch (err) {
    console.warn(
      `lessgo: Warning - Could not verify binary installation. ` +
        `The package "${packageName}" may not be installed or the binary may not be executable.`
    );
    return false;
  }
}

// Handle --postinstall flag
if (process.argv.includes("--postinstall")) {
  verifyInstallation();
}

module.exports = {
  getBinaryPath,
  getPlatformKey,
  getPlatformPackage,
  compile,
  compileSync,
  run,
  verifyInstallation,
  PLATFORM_PACKAGES,
};

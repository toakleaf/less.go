/**
 * Platform to package name mapping
 */
export const PLATFORM_PACKAGES: {
  "darwin-arm64": string;
  "darwin-x64": string;
  "linux-x64": string;
  "linux-arm64": string;
  "win32-x64": string;
  "win32-arm64": string;
};

/**
 * Get the platform key for the current system
 * @returns Platform key (e.g., "darwin-arm64")
 */
export function getPlatformKey(): string;

/**
 * Get the package name for the current platform
 * @returns Package name or null if unsupported
 */
export function getPlatformPackage(): string | null;

/**
 * Get the path to the lessc-go binary for the current platform
 * @returns Path to the binary
 * @throws Error if binary cannot be found
 */
export function getBinaryPath(): string;

/**
 * Options for LESS compilation
 */
export interface CompileOptions {
  /**
   * If true, treat input as LESS content instead of file path
   */
  string?: boolean;

  /**
   * Include paths for @import
   */
  paths?: string[];

  /**
   * Minify output
   */
  compress?: boolean;

  /**
   * Generate source map
   */
  sourceMap?: boolean;

  /**
   * Source map filename
   */
  sourceMapFilename?: string;

  /**
   * Global variables to inject
   */
  globalVars?: Record<string, string>;

  /**
   * Variables to modify
   */
  modifyVars?: Record<string, string>;
}

/**
 * Result of LESS compilation
 */
export interface CompileResult {
  /**
   * The compiled CSS output
   */
  css: string;

  /**
   * The source map (if sourceMap option was enabled)
   */
  map?: string;
}

/**
 * Result of running the compiler
 */
export interface RunResult {
  /**
   * Exit code
   */
  code: number;

  /**
   * Standard output
   */
  stdout: string;

  /**
   * Standard error
   */
  stderr: string;
}

/**
 * Compile a LESS file or string
 * @param input - File path or LESS content
 * @param options - Compilation options
 * @returns Promise resolving to compilation result
 */
export function compile(
  input: string,
  options?: CompileOptions
): Promise<CompileResult>;

/**
 * Synchronously compile a LESS file or string
 * @param input - File path or LESS content
 * @param options - Compilation options
 * @returns Compilation result
 */
export function compileSync(
  input: string,
  options?: CompileOptions
): CompileResult;

/**
 * Run the compiler with raw arguments (like CLI)
 * @param args - Command line arguments
 * @returns Promise resolving to run result
 */
export function run(args?: string[]): Promise<RunResult>;

/**
 * Verify the binary is properly installed (used by postinstall)
 * @returns true if installation is valid, false otherwise
 */
export function verifyInstallation(): boolean;

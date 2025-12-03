import { compile, type CompileOptions, type PluginSpec } from 'lessgo';
import fs from 'node:fs';
import path from 'node:path';
import type { Plugin, ResolvedConfig, ViteDevServer } from 'vite';

/**
 * Options for the lessgo Vite plugin
 */
export interface LessgoPluginOptions {
  /**
   * Minify CSS output
   * @default false
   */
  compress?: boolean;

  /**
   * Additional include paths for @import resolution
   * These paths will be searched when resolving @import statements
   */
  paths?: string[];

  /**
   * Global variables to inject into LESS compilation
   * Variables are available in all compiled LESS files
   * @example { theme: '"dark"', primaryColor: '#007bff' }
   */
  globalVars?: Record<string, string>;

  /**
   * Variables to modify (override) in LESS compilation
   * These take precedence over variables defined in LESS files
   * @example { primaryColor: '#ff0000' }
   */
  modifyVars?: Record<string, string>;

  /**
   * LESS plugins to load
   * Can be plugin names (with or without 'less-plugin-' prefix),
   * paths to plugin files, or plugin specification objects
   * @example ['clean-css']
   * @example [{ name: 'clean-css', options: 'advanced' }]
   * @example ['./my-plugin.js']
   */
  plugins?: (PluginSpec | string)[];

  /**
   * Generate source maps
   * @default true in development, false in production
   */
  sourceMap?: boolean;

  /**
   * File patterns to include
   * @default /\.less$/
   */
  include?: RegExp | string | string[];

  /**
   * File patterns to exclude
   */
  exclude?: RegExp | string | string[];
}

// Internal constants
const VIRTUAL_PREFIX = '\0lessgo-compiled:';
const VIRTUAL_SUFFIX = '.css';
const LESS_FILE_REGEX = /\.less$/;
const VIRTUAL_ID_REGEX = /^\0lessgo-compiled:/;

/**
 * Normalize a pattern to a RegExp
 */
function toRegExp(pattern: RegExp | string | string[] | undefined): RegExp | null {
  if (!pattern) return null;
  if (pattern instanceof RegExp) return pattern;
  if (typeof pattern === 'string') {
    return new RegExp(pattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));
  }
  if (Array.isArray(pattern)) {
    const escaped = pattern.map((p) => p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));
    return new RegExp(`(${escaped.join('|')})`);
  }
  return null;
}

/**
 * Check if a path matches include/exclude patterns
 */
function shouldProcess(
  filePath: string,
  include: RegExp | null,
  exclude: RegExp | null
): boolean {
  // Must match include pattern (or no include pattern specified)
  if (include && !include.test(filePath)) {
    return false;
  }
  // Must not match exclude pattern
  if (exclude && exclude.test(filePath)) {
    return false;
  }
  // Default: match .less files
  return LESS_FILE_REGEX.test(filePath);
}

/**
 * Format a compile error for Vite's error overlay
 */
function formatError(error: Error, filePath: string): Error {
  const formattedError = new Error(error.message);
  formattedError.name = 'LessgoCompileError';

  // Parse line/column from lessgo error messages if available
  // Format: "ParseError: ... in /path/to/file.less on line X, column Y"
  const lineMatch = error.message.match(/on line (\d+)/i);
  const columnMatch = error.message.match(/column (\d+)/i);

  // Attach source location for Vite's error overlay
  (formattedError as Error & { loc?: { file: string; line: number; column: number } }).loc = {
    file: filePath,
    line: lineMatch ? parseInt(lineMatch[1], 10) : 1,
    column: columnMatch ? parseInt(columnMatch[1], 10) : 0,
  };

  return formattedError;
}

/**
 * Vite plugin for using less.go (lessc-go) as the LESS preprocessor
 *
 * This plugin intercepts .less imports and compiles them using the
 * lessgo Node.js API, providing fast LESS compilation with the Go-based compiler.
 *
 * @param options - Plugin configuration options
 * @returns Vite plugin object
 *
 * @example
 * ```ts
 * // vite.config.ts
 * import { defineConfig } from 'vite';
 * import lessgo from '@lessgo/plugin-vite';
 *
 * export default defineConfig({
 *   plugins: [
 *     lessgo({
 *       compress: true,
 *       globalVars: {
 *         primaryColor: '#007bff',
 *       },
 *     }),
 *   ],
 * });
 * ```
 */
export default function lessgoPlugin(options: LessgoPluginOptions = {}): Plugin {
  // Map virtual IDs back to original .less file paths
  const virtualToLess = new Map<string, string>();

  // Track dependencies for HMR
  const fileDependencies = new Map<string, Set<string>>();

  // Resolved patterns
  const includePattern = toRegExp(options.include);
  const excludePattern = toRegExp(options.exclude);

  // Vite config reference
  let config: ResolvedConfig;
  let server: ViteDevServer | undefined;

  return {
    name: 'vite-plugin-lessgo',
    enforce: 'pre', // Run before Vite's built-in CSS handling

    configResolved(resolvedConfig) {
      config = resolvedConfig;
    },

    configureServer(devServer) {
      server = devServer;
    },

    resolveId(source, importer) {
      // Only process .less files
      if (!LESS_FILE_REGEX.test(source)) {
        return null;
      }

      // Check include/exclude patterns
      if (!shouldProcess(source, includePattern, excludePattern)) {
        return null;
      }

      // Resolve the actual file path
      let resolvedPath: string;
      if (path.isAbsolute(source)) {
        resolvedPath = source;
      } else if (importer) {
        // Handle importer that might be a virtual module
        const importerPath = importer.startsWith(VIRTUAL_PREFIX)
          ? virtualToLess.get(importer) || importer
          : importer;
        resolvedPath = path.resolve(path.dirname(importerPath), source);
      } else {
        resolvedPath = path.resolve(source);
      }

      // Normalize the path
      resolvedPath = path.normalize(resolvedPath);

      // Create a virtual module ID ending in .css
      const virtualId = VIRTUAL_PREFIX + resolvedPath + VIRTUAL_SUFFIX;
      virtualToLess.set(virtualId, resolvedPath);

      return virtualId;
    },

    async load(id) {
      // Only process our virtual modules
      if (!id.startsWith(VIRTUAL_PREFIX)) {
        return null;
      }

      const lessPath = virtualToLess.get(id);
      if (!lessPath) {
        return null;
      }

      if (!fs.existsSync(lessPath)) {
        throw new Error(`[lessgo] LESS file not found: ${lessPath}`);
      }

      try {
        // Build include paths - always include the file's directory
        const includePaths = [path.dirname(lessPath)];
        if (options.paths) {
          includePaths.push(...options.paths);
        }

        // Add node_modules paths for @import resolution
        let currentDir = path.dirname(lessPath);
        while (currentDir !== path.dirname(currentDir)) {
          const nodeModules = path.join(currentDir, 'node_modules');
          if (fs.existsSync(nodeModules)) {
            includePaths.push(nodeModules);
          }
          currentDir = path.dirname(currentDir);
        }

        // Determine source map setting
        const generateSourceMap =
          options.sourceMap !== undefined
            ? options.sourceMap
            : config.command === 'serve';

        // Build compile options
        const compileOptions: CompileOptions = {
          paths: includePaths,
          compress: options.compress,
          globalVars: options.globalVars,
          modifyVars: options.modifyVars,
          plugins: options.plugins,
          sourceMap: generateSourceMap,
        };

        // Compile using lessgo Node.js API
        const result = await compile(lessPath, compileOptions);

        // Track this file for HMR
        if (server) {
          this.addWatchFile(lessPath);
        }

        // Return compiled CSS for Vite to process
        return {
          code: result.css,
          map: result.map ? JSON.parse(result.map) : null,
        };
      } catch (error) {
        throw formatError(error as Error, lessPath);
      }
    },

    // Handle HMR updates
    handleHotUpdate({ file, server }) {
      if (!LESS_FILE_REGEX.test(file)) {
        return;
      }

      // Find all virtual modules that depend on this file
      const affectedModules: ReturnType<typeof server.moduleGraph.getModuleById>[] = [];

      // Check if this is a main file we've compiled
      for (const [virtualId, lessPath] of virtualToLess) {
        if (lessPath === file) {
          const mod = server.moduleGraph.getModuleById(virtualId);
          if (mod) {
            affectedModules.push(mod);
          }
        }
      }

      // Check if this is a dependency of any compiled file
      for (const [mainFile, deps] of fileDependencies) {
        if (deps.has(file)) {
          const virtualId = VIRTUAL_PREFIX + mainFile + VIRTUAL_SUFFIX;
          const mod = server.moduleGraph.getModuleById(virtualId);
          if (mod) {
            affectedModules.push(mod);
          }
        }
      }

      if (affectedModules.length > 0) {
        return affectedModules.filter(Boolean) as NonNullable<
          ReturnType<typeof server.moduleGraph.getModuleById>
        >[];
      }
    },
  };
}

// Named export for ESM
export { lessgoPlugin };

// Re-export types from lessgo for convenience
export type { PluginSpec, CompileOptions } from 'lessgo';

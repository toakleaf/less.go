import { PluginSpec } from 'lessgo';
export { CompileOptions, PluginSpec } from 'lessgo';
import { Plugin } from 'vite';

/**
 * Options for the lessgo Vite plugin
 */
interface LessgoPluginOptions {
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
declare function lessgoPlugin(options?: LessgoPluginOptions): Plugin;

export { type LessgoPluginOptions, lessgoPlugin as default, lessgoPlugin };

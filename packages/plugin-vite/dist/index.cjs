'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

var lessgo = require('lessgo');
var fs = require('fs');
var path = require('path');

function _interopDefault (e) { return e && e.__esModule ? e : { default: e }; }

var fs__default = /*#__PURE__*/_interopDefault(fs);
var path__default = /*#__PURE__*/_interopDefault(path);

// src/index.ts
var VIRTUAL_PREFIX = "\0lessgo-compiled:";
var VIRTUAL_SUFFIX = ".css";
var LESS_FILE_REGEX = /\.less$/;
function toRegExp(pattern) {
  if (!pattern) return null;
  if (pattern instanceof RegExp) return pattern;
  if (typeof pattern === "string") {
    return new RegExp(pattern.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"));
  }
  if (Array.isArray(pattern)) {
    const escaped = pattern.map((p) => p.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"));
    return new RegExp(`(${escaped.join("|")})`);
  }
  return null;
}
function shouldProcess(filePath, include, exclude) {
  if (include && !include.test(filePath)) {
    return false;
  }
  if (exclude && exclude.test(filePath)) {
    return false;
  }
  return LESS_FILE_REGEX.test(filePath);
}
function formatError(error, filePath) {
  const formattedError = new Error(error.message);
  formattedError.name = "LessgoCompileError";
  const lineMatch = error.message.match(/on line (\d+)/i);
  const columnMatch = error.message.match(/column (\d+)/i);
  formattedError.loc = {
    file: filePath,
    line: lineMatch ? parseInt(lineMatch[1], 10) : 1,
    column: columnMatch ? parseInt(columnMatch[1], 10) : 0
  };
  return formattedError;
}
function lessgoPlugin(options = {}) {
  const virtualToLess = /* @__PURE__ */ new Map();
  const fileDependencies = /* @__PURE__ */ new Map();
  const includePattern = toRegExp(options.include);
  const excludePattern = toRegExp(options.exclude);
  let config;
  let server;
  return {
    name: "vite-plugin-lessgo",
    enforce: "pre",
    // Run before Vite's built-in CSS handling
    configResolved(resolvedConfig) {
      config = resolvedConfig;
    },
    configureServer(devServer) {
      server = devServer;
    },
    resolveId(source, importer) {
      if (!LESS_FILE_REGEX.test(source)) {
        return null;
      }
      if (!shouldProcess(source, includePattern, excludePattern)) {
        return null;
      }
      let resolvedPath;
      if (path__default.default.isAbsolute(source)) {
        resolvedPath = source;
      } else if (importer) {
        const importerPath = importer.startsWith(VIRTUAL_PREFIX) ? virtualToLess.get(importer) || importer : importer;
        resolvedPath = path__default.default.resolve(path__default.default.dirname(importerPath), source);
      } else {
        resolvedPath = path__default.default.resolve(source);
      }
      resolvedPath = path__default.default.normalize(resolvedPath);
      const virtualId = VIRTUAL_PREFIX + resolvedPath + VIRTUAL_SUFFIX;
      virtualToLess.set(virtualId, resolvedPath);
      return virtualId;
    },
    async load(id) {
      if (!id.startsWith(VIRTUAL_PREFIX)) {
        return null;
      }
      const lessPath = virtualToLess.get(id);
      if (!lessPath) {
        return null;
      }
      if (!fs__default.default.existsSync(lessPath)) {
        throw new Error(`[lessgo] LESS file not found: ${lessPath}`);
      }
      try {
        const includePaths = [path__default.default.dirname(lessPath)];
        if (options.paths) {
          includePaths.push(...options.paths);
        }
        let currentDir = path__default.default.dirname(lessPath);
        while (currentDir !== path__default.default.dirname(currentDir)) {
          const nodeModules = path__default.default.join(currentDir, "node_modules");
          if (fs__default.default.existsSync(nodeModules)) {
            includePaths.push(nodeModules);
          }
          currentDir = path__default.default.dirname(currentDir);
        }
        const generateSourceMap = options.sourceMap !== void 0 ? options.sourceMap : config.command === "serve";
        const compileOptions = {
          paths: includePaths,
          compress: options.compress,
          globalVars: options.globalVars,
          modifyVars: options.modifyVars,
          plugins: options.plugins,
          sourceMap: generateSourceMap
        };
        const result = await lessgo.compile(lessPath, compileOptions);
        if (server) {
          this.addWatchFile(lessPath);
        }
        return {
          code: result.css,
          map: result.map ? JSON.parse(result.map) : null
        };
      } catch (error) {
        throw formatError(error, lessPath);
      }
    },
    // Handle HMR updates
    handleHotUpdate({ file, server: server2 }) {
      if (!LESS_FILE_REGEX.test(file)) {
        return;
      }
      const affectedModules = [];
      for (const [virtualId, lessPath] of virtualToLess) {
        if (lessPath === file) {
          const mod = server2.moduleGraph.getModuleById(virtualId);
          if (mod) {
            affectedModules.push(mod);
          }
        }
      }
      for (const [mainFile, deps] of fileDependencies) {
        if (deps.has(file)) {
          const virtualId = VIRTUAL_PREFIX + mainFile + VIRTUAL_SUFFIX;
          const mod = server2.moduleGraph.getModuleById(virtualId);
          if (mod) {
            affectedModules.push(mod);
          }
        }
      }
      if (affectedModules.length > 0) {
        return affectedModules.filter(Boolean);
      }
    }
  };
}

exports.default = lessgoPlugin;
exports.lessgoPlugin = lessgoPlugin;
//# sourceMappingURL=index.cjs.map
//# sourceMappingURL=index.cjs.map
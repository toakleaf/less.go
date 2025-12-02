import { defineConfig } from 'vitest/config';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const lessPath = path.resolve(__dirname, '../../reference/less.js/packages/less/src/less');

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: ['**/*.test.js'],
    root: __dirname,
    
  },
  resolve: {
    alias: {
      '@less': lessPath,
    },
  },
  esbuild: {
    target: 'node18',
  },
});

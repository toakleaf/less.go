import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: ['**/*.test.js'],
    root: __dirname,
  },
  resolve: {
    alias: {
      '@less': path.resolve(__dirname, '../../reference/less.js/packages/less/src/less')
    }
  }
});

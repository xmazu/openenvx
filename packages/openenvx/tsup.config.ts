import { copy } from 'esbuild-plugin-copy';
import { defineConfig } from 'tsup';

export default defineConfig({
  entry: ['src/index.ts'],
  format: ['esm'],
  dts: true,
  splitting: false,
  sourcemap: true,
  clean: true,
  esbuildPlugins: [
    copy({
      resolveFrom: 'cwd',
      assets: {
        from: ['./src/templates/**/*'],
        to: ['./dist/templates'],
      },
      globbyOptions: {
        dot: true,
      },
    }),
  ],
  esbuildOptions(options) {
    options.banner = {
      js: '#!/usr/bin/env node',
    };
  },
});

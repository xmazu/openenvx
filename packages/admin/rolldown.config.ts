import { defineConfig } from 'rolldown';

const externalDeps = [
  'react',
  'react-dom',
  'next',
  'drizzle-orm',
  '@tanstack/react-table',
  '@swc/helpers',
  'lucide-react',
  'radix-ui',
  '@radix-ui',
  'use-sync-external-store',
  'cmdk',
  'react-day-picker',
  'tailwind-merge',
  'class-variance-authority',
  'clsx',
  'postgres',
];

const external = (id: string) => {
  if (id.includes('node_modules')) {
    return true;
  }
  return externalDeps.some((dep) => id === dep || id.startsWith(`${dep}/`));
};

export default defineConfig([
  {
    input: 'src/server/index.ts',
    output: {
      file: 'dist/server.js',
      format: 'esm',
      sourcemap: true,
    },
    external,
    resolve: {
      conditionNames: ['import', 'module', 'default'],
    },
  },
  {
    input: 'src/client/index.ts',
    output: {
      file: 'dist/client.js',
      format: 'esm',
      sourcemap: true,
    },
    external,
    resolve: {
      conditionNames: ['import', 'module', 'default'],
    },
  },
]);

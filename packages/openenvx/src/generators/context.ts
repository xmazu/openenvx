import path from 'node:path';
import type {
  GenerateContext,
  PackageManager,
  ProjectConfig,
} from '../lib/types';

export function createContext(
  config: ProjectConfig,
  packageManager: PackageManager
): GenerateContext {
  return {
    config,
    targetDir: path.resolve(process.cwd(), config.name),
    state: {
      features: [],
      generated: [],
    },
    packageManager,
  };
}

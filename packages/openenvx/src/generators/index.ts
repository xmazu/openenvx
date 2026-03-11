import type { LogEntry, ProjectConfig } from '../lib/types';
import { detectPackageManager } from '../lib/utils';
import { createContext } from './context';
import { generateBase } from './steps/base';
import { installDependencies } from './steps/dependencies';
import { createProjectDirectory } from './steps/directory';
import { setupEnvironment } from './steps/environment';
import { generateFeatures } from './steps/features';
import { initGit } from './steps/git';
import { initShadcn } from './steps/shadcn';
import { addWorkspaceDependencies } from './steps/workspace';

export type { ProjectConfig } from '../lib/types';

export async function* generateProject(
  config: ProjectConfig
): AsyncGenerator<LogEntry, void, unknown> {
  const packageManager = await detectPackageManager();
  yield { message: `Using package manager: ${packageManager}`, level: 'info' };

  const ctx = createContext(config, packageManager);

  yield* createProjectDirectory(ctx);
  yield* generateBase(ctx);
  yield* generateFeatures(ctx);
  yield* setupEnvironment(ctx);
  yield* addWorkspaceDependencies(ctx);
  yield* installDependencies(ctx);
  yield* initShadcn(ctx);
  yield* initGit(ctx);
}

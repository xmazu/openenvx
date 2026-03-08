import { execSync } from 'node:child_process';
import path from 'node:path';
import { execa } from 'execa';
import fs from 'fs-extra';
import {
  addPackageDependency,
  appendEnvVariables,
  generateBaseTemplate,
  generateFeature,
} from '../lib/templates';

export interface ProjectConfig {
  database: string;
  features: {
    stripe: boolean;
    storage: boolean;
    email: boolean;
  };
  name: string;
  projectName: string;
}

export interface State {
  features: string[];
  generated: string[];
}

export type LogLevel = 'info' | 'success' | 'warning' | 'error' | 'spinner';
export interface LogEntry {
  level: LogLevel;
  message: string;
}

export type PackageManager = 'bun' | 'pnpm';

interface GenerateContext {
  config: ProjectConfig;
  hasOexctl: boolean;
  packageManager: PackageManager;
  state: State;
  targetDir: string;
}

async function detectPackageManager(): Promise<PackageManager> {
  try {
    await execa('bun', ['--version'], { stdio: 'ignore' });
    return 'bun';
  } catch {
    try {
      await execa('pnpm', ['--version'], { stdio: 'ignore' });
      return 'pnpm';
    } catch {
      throw new Error(
        'No package manager found. Please install Bun (https://bun.sh) or pnpm (https://pnpm.io).'
      );
    }
  }
}

function checkOexctlInstalled(): boolean {
  try {
    execSync('which oexctl', { stdio: 'ignore' });
    return true;
  } catch {
    return false;
  }
}

function createContext(
  config: ProjectConfig,
  packageManager: PackageManager,
  hasOexctl: boolean
): GenerateContext {
  return {
    config,
    hasOexctl,
    targetDir: path.resolve(process.cwd(), config.name),
    state: {
      features: [],
      generated: [],
    },
    packageManager,
  };
}

async function* createProjectDirectory(
  ctx: GenerateContext
): AsyncGenerator<LogEntry> {
  if (await fs.pathExists(ctx.targetDir)) {
    throw new Error(`Directory "${ctx.config.name}" already exists`);
  }

  yield { message: 'Creating project directory...', level: 'spinner' };
  await fs.ensureDir(ctx.targetDir);
  yield { message: 'Project directory created', level: 'success' };
}

async function* generateBase(ctx: GenerateContext): AsyncGenerator<LogEntry> {
  yield { message: 'Generating base template...', level: 'spinner' };
  await generateBaseTemplate(
    ctx.targetDir,
    ctx.config,
    ctx.packageManager,
    ctx.hasOexctl
  );
  ctx.state.generated.push('base');
  yield { message: 'Base template generated', level: 'success' };
}

async function* generateFeatures(
  ctx: GenerateContext
): AsyncGenerator<LogEntry> {
  if (ctx.config.features.stripe) {
    yield { message: 'Generating Stripe feature...', level: 'spinner' };
    await generateFeature(ctx.targetDir, 'stripe', ctx.config);
    ctx.state.features.push('stripe');
    ctx.state.generated.push('stripe');
    yield { message: 'Stripe feature generated', level: 'success' };
  }

  if (ctx.config.features.storage) {
    yield { message: 'Generating Storage feature...', level: 'spinner' };
    await generateFeature(ctx.targetDir, 'storage', ctx.config);
    ctx.state.features.push('storage');
    ctx.state.generated.push('storage');
    yield { message: 'Storage feature generated', level: 'success' };
  }

  if (ctx.config.features.email) {
    yield { message: 'Generating Email feature...', level: 'spinner' };
    await generateFeature(ctx.targetDir, 'email', ctx.config);
    ctx.state.features.push('email');
    ctx.state.generated.push('email');
    yield { message: 'Email feature generated', level: 'success' };
  }
}

async function* setupEnvironment(
  ctx: GenerateContext
): AsyncGenerator<LogEntry> {
  yield { message: 'Appending environment variables...', level: 'spinner' };
  await appendEnvVariables(ctx.targetDir, ctx.config, ctx.hasOexctl);

  await fs.ensureDir(path.join(ctx.targetDir, '.openenvx'));
  await fs.writeJson(
    path.join(ctx.targetDir, '.openenvx', 'state.json'),
    ctx.state,
    { spaces: 2 }
  );
  yield { message: 'Environment configured', level: 'success' };
}

async function* addWorkspaceDependencies(
  ctx: GenerateContext
): AsyncGenerator<LogEntry> {
  if (ctx.config.features.stripe) {
    yield {
      message: 'Adding Stripe workspace dependency...',
      level: 'spinner',
    };
    await addPackageDependency(
      path.join(ctx.targetDir, 'apps', 'dashboard', 'package.json'),
      `@${ctx.config.projectName}/stripe`,
      'workspace:*'
    );
  }

  if (ctx.config.features.storage) {
    yield {
      message: 'Adding Storage workspace dependency...',
      level: 'spinner',
    };
    await addPackageDependency(
      path.join(ctx.targetDir, 'apps', 'dashboard', 'package.json'),
      `@${ctx.config.projectName}/storage`,
      'workspace:*'
    );
  }

  if (ctx.config.features.email) {
    yield { message: 'Adding Email workspace dependency...', level: 'spinner' };
    await addPackageDependency(
      path.join(ctx.targetDir, 'apps', 'dashboard', 'package.json'),
      `@${ctx.config.projectName}/email`,
      'workspace:*'
    );
  }

  if (
    ctx.config.features.stripe ||
    ctx.config.features.storage ||
    ctx.config.features.email
  ) {
    yield { message: 'Workspace dependencies added', level: 'success' };
  }
}

async function* installDependencies(
  ctx: GenerateContext
): AsyncGenerator<LogEntry> {
  yield {
    message: `Installing dependencies with ${ctx.packageManager}...`,
    level: 'spinner',
  };

  const installCmd = ctx.packageManager === 'bun' ? 'bun' : 'pnpm';
  const installArgs = ctx.packageManager === 'bun' ? ['install'] : ['install'];

  await execa(installCmd, installArgs, {
    cwd: ctx.targetDir,
    stdout: 'inherit',
    stderr: 'inherit',
  });

  yield { message: 'Dependencies installed', level: 'success' };
}

async function* initGit(ctx: GenerateContext): AsyncGenerator<LogEntry> {
  yield { message: 'Initializing Git repository...', level: 'spinner' };
  await execa('git', ['init'], { cwd: ctx.targetDir });
  await execa('git', ['add', '.'], { cwd: ctx.targetDir });
  await execa(
    'git',
    ['commit', '-m', 'Initial commit from create-openenvx-app'],
    {
      cwd: ctx.targetDir,
    }
  );
  yield { message: 'Git repository initialized', level: 'success' };
}

const SHADCN_COMPONENTS = [
  'alert',
  'avatar',
  'badge',
  'button',
  'card',
  'dropdown-menu',
  'form',
  'input',
  'label',
  'separator',
  'sheet',
  'sidebar',
  'skeleton',
  'tooltip',
] as const;

async function* initShadcn(ctx: GenerateContext): AsyncGenerator<LogEntry> {
  yield { message: 'Adding shadcn/ui components...', level: 'spinner' };

  const uiPackageDir = path.join(ctx.targetDir, 'packages', 'ui');

  const runCmd = ctx.packageManager === 'bun' ? 'bunx' : 'pnpm';
  const runArgs =
    ctx.packageManager === 'bun'
      ? ['shadcn@latest', 'add', '-y', '--overwrite', ...SHADCN_COMPONENTS]
      : [
          'exec',
          'shadcn@latest',
          'add',
          '-y',
          '--overwrite',
          ...SHADCN_COMPONENTS,
        ];

  await execa(runCmd, runArgs, {
    cwd: uiPackageDir,
    stdout: 'inherit',
    stderr: 'inherit',
  });

  yield { message: 'shadcn/ui components added', level: 'success' };
}

export async function* generateProject(
  config: ProjectConfig
): AsyncGenerator<LogEntry, void, unknown> {
  const packageManager = await detectPackageManager();
  yield { message: `Using package manager: ${packageManager}`, level: 'info' };

  const hasOexctl = checkOexctlInstalled();
  if (hasOexctl) {
    yield {
      message: 'oexctl detected - configuring proxy URLs',
      level: 'info',
    };
  } else {
    yield {
      message: 'oexctl not detected - using fallback ports',
      level: 'info',
    };
    yield {
      message: 'Install oexctl for better URLs: openenvx install',
      level: 'info',
    };
  }

  const ctx = createContext(config, packageManager, hasOexctl);

  yield* createProjectDirectory(ctx);
  yield* generateBase(ctx);
  yield* generateFeatures(ctx);
  yield* setupEnvironment(ctx);
  yield* addWorkspaceDependencies(ctx);
  yield* installDependencies(ctx);
  yield* initShadcn(ctx);
  // yield* installDependencies(ctx);
  yield* initGit(ctx);
}

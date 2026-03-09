import path from 'node:path';
import fs from 'fs-extra';
import { globby } from 'globby';
import Handlebars from 'handlebars';
import { dependencyCatalog } from './constants';
import type { PackageManager, ProjectConfig } from './types';
import { getTemplatesDir } from './utils';

function generateRootPackageJson(
  projectName: string,
  packageManager: 'bun' | 'pnpm'
): string {
  const basePackageJson = {
    name: projectName,
    version: '0.0.1',
    private: true,
    scripts: {
      build: 'turbo build',
      dev: 'turbo dev',
      lint: 'turbo lint',
      'db:generate': 'turbo db:generate',
      'db:migrate': 'turbo db:migrate',
      'db:studio': 'turbo db:studio',
    },
    devDependencies: {
      turbo: '^2.8.13',
      typescript: 'catalog:',
    },
    packageManager: packageManager === 'bun' ? 'bun@1.3.2' : 'pnpm@10.30.3',
  };

  const bunSpecific =
    packageManager === 'bun'
      ? { workspaces: ['apps/*', 'packages/*'], catalog: dependencyCatalog }
      : {};

  return JSON.stringify({ ...basePackageJson, ...bunSpecific }, null, 2);
}

function generatePnpmWorkspaceYaml(): string {
  return `packages:
  - "apps/*"
  - "packages/*"

catalog:
${Object.entries(dependencyCatalog)
  .map(([name, version]) => `  ${name}: ${version}`)
  .join('\n')}
`;
}

export async function generateBaseTemplate(
  targetDir: string,
  config: ProjectConfig,
  packageManager: PackageManager
): Promise<void> {
  const templatesDir = getTemplatesDir('base');

  Handlebars.registerHelper('scopedName', (name: string) => {
    return `@${config.projectName}/${name}`;
  });

  const templateFiles = await globby('**/*.hbs', {
    cwd: templatesDir,
    dot: true,
  });

  for (const templateFile of templateFiles) {
    if (templateFile === 'package.json.hbs') {
      continue;
    }

    const templatePath = path.join(templatesDir, templateFile);
    const templateContent = await fs.readFile(templatePath, 'utf-8');

    const template = Handlebars.compile(templateContent);
    const rendered = template(config);

    const targetFile = templateFile.replace('.hbs', '');
    const targetFilePath = path.join(targetDir, targetFile);

    await fs.ensureDir(path.dirname(targetFilePath));
    await fs.writeFile(targetFilePath, rendered);
  }

  const rootPackageJson = generateRootPackageJson(
    config.projectName,
    packageManager
  );
  await fs.writeFile(path.join(targetDir, 'package.json'), rootPackageJson);

  if (packageManager === 'pnpm') {
    await fs.writeFile(
      path.join(targetDir, 'pnpm-workspace.yaml'),
      generatePnpmWorkspaceYaml()
    );
  }

  const nonTemplateFiles = await globby(['**/*', '!**/*.hbs'], {
    cwd: templatesDir,
    dot: true,
    onlyFiles: true,
  });

  for (const file of nonTemplateFiles) {
    if (file === 'components.json') {
      continue;
    }

    const sourcePath = path.join(templatesDir, file);
    const targetFilePath = path.join(targetDir, file);

    await fs.ensureDir(path.dirname(targetFilePath));
    await fs.copy(sourcePath, targetFilePath);
  }

  // Make shell scripts executable
  const scriptsDir = path.join(targetDir, 'scripts');
  if (await fs.pathExists(scriptsDir)) {
    const scriptFiles = await fs.readdir(scriptsDir);
    for (const scriptFile of scriptFiles) {
      if (scriptFile.endsWith('.sh')) {
        await fs.chmod(path.join(scriptsDir, scriptFile), 0o755);
      }
    }
  }

  // Generate services configuration for oexctl
  const servicesTemplatePath = getTemplatesDir(
    path.join('base', '.openenvx', 'services.yaml.hbs')
  );
  if (await fs.pathExists(servicesTemplatePath)) {
    const servicesContent = await fs.readFile(servicesTemplatePath, 'utf-8');
    const servicesTemplate = Handlebars.compile(servicesContent);
    const renderedServices = servicesTemplate(config);
    const servicesTargetPath = path.join(
      targetDir,
      '.openenvx',
      'services.yaml'
    );
    await fs.ensureDir(path.dirname(servicesTargetPath));
    await fs.writeFile(servicesTargetPath, renderedServices);
  }
}

export async function generateFeature(
  targetDir: string,
  feature: 'stripe' | 'storage' | 'email',
  config: ProjectConfig
): Promise<void> {
  const templatesDir = getTemplatesDir(path.join('features', feature));

  if (!(await fs.pathExists(templatesDir))) {
    console.warn(`Templates for feature ${feature} not found`);
    return;
  }

  Handlebars.registerHelper('scopedName', (name: string) => {
    return `@${config.projectName}/${name}`;
  });

  const templateFiles = await globby('**/*.hbs', {
    cwd: templatesDir,
    dot: true,
  });

  for (const templateFile of templateFiles) {
    const templatePath = path.join(templatesDir, templateFile);
    const templateContent = await fs.readFile(templatePath, 'utf-8');

    const template = Handlebars.compile(templateContent);
    const rendered = template(config);

    const targetFile = templateFile.replace('.hbs', '');
    const targetFilePath = path.join(targetDir, targetFile);

    await fs.ensureDir(path.dirname(targetFilePath));
    await fs.writeFile(targetFilePath, rendered);
  }

  const nonTemplateFiles = await globby(['**/*', '!**/*.hbs'], {
    cwd: templatesDir,
    dot: true,
    onlyFiles: true,
  });

  for (const file of nonTemplateFiles) {
    const sourcePath = path.join(templatesDir, file);
    const targetFilePath = path.join(targetDir, file);

    await fs.ensureDir(path.dirname(targetFilePath));
    await fs.copy(sourcePath, targetFilePath);
  }
}

export async function appendEnvVariables(
  targetDir: string,
  config: ProjectConfig,
  hasOexctl: boolean
): Promise<void> {
  const envTemplatePath = getTemplatesDir(path.join('features', 'env.hbs'));

  if (!(await fs.pathExists(envTemplatePath))) {
    return;
  }

  const templateContent = await fs.readFile(envTemplatePath, 'utf-8');
  const template = Handlebars.compile(templateContent);
  const templateContext = {
    ...config,
    hasOexctl,
  };
  const rendered = template(templateContext);

  // Append to both apps' .env files (create if missing, e.g. when base template
  // doesn't include .env or structure changes)
  const apps = ['web', 'dashboard'];
  for (const app of apps) {
    const envPath = path.join(targetDir, 'apps', app, '.env');
    await fs.ensureDir(path.dirname(envPath));
    if (await fs.pathExists(envPath)) {
      const existingContent = await fs.readFile(envPath, 'utf-8');
      await fs.writeFile(envPath, `${existingContent}\n${rendered}`);
    } else {
      await fs.writeFile(envPath, rendered);
    }
  }
}

export async function addPackageDependency(
  packageJsonPath: string,
  depName: string,
  version: string,
  isDev = false
): Promise<void> {
  const pkg = await fs.readJson(packageJsonPath);
  const depType = isDev ? 'devDependencies' : 'dependencies';

  if (!pkg[depType]) {
    pkg[depType] = {};
  }

  pkg[depType][depName] = version;
  await fs.writeJson(packageJsonPath, pkg, { spaces: 2 });
}

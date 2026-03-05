import {
  cancel,
  group,
  intro,
  log,
  multiselect,
  note,
  outro,
  text,
} from '@clack/prompts';
import { Command } from 'commander';
import color from 'picocolors';
import {
  generateProject,
  type ProjectConfig,
} from './generators/project-generator';

const PROJECT_NAME_REGEX = /^[a-z0-9-_]+$/i;

const program = new Command();

program
  .name('openenvx')
  .description('OpenEnvx CLI - Create and manage OpenEnvx SaaS apps')
  .version('0.0.1');

program
  .command('init')
  .description('Initialize a new OpenEnvx project')
  .argument('[project-directory]', 'Directory to create the project in')
  .action(async (projectDirectory) => {
    intro(color.bgCyan(color.black(' create-openenvx-app ')));

    const groupResult = await group(
      {
        name: () =>
          text({
            message: 'What is your project named?',
            placeholder: projectDirectory || 'my-app',
            initialValue: projectDirectory,
            validate: (value: string | undefined) => {
              if (!value) {
                return 'Project name is required';
              }
              if (!PROJECT_NAME_REGEX.test(value)) {
                return 'Only letters, numbers, hyphens, and underscores allowed';
              }
            },
          }),
        features: () =>
          multiselect({
            message: 'Select features to include:',
            options: [
              { value: 'stripe', label: 'Stripe Payments' },
              { value: 'storage', label: 'S3 File Storage' },
              { value: 'email', label: 'Email (Resend)' },
            ],
          }),
      },
      {
        onCancel: () => {
          cancel('Operation cancelled.');
          process.exit(0);
        },
      }
    );

    const config: ProjectConfig = {
      name: groupResult.name,
      projectName: groupResult.name,
      features: {
        stripe: groupResult.features.includes('stripe'),
        storage: groupResult.features.includes('storage'),
        email: groupResult.features.includes('email'),
      },
      database: 'postgresql',
    };

    log.step('Creating your project...');

    try {
      for await (const entry of generateProject(config)) {
        if (entry.level === 'spinner') {
          log.step(entry.message);
        } else if (entry.level === 'success') {
          log.success(entry.message);
        } else if (entry.level === 'error') {
          log.error(entry.message);
        } else {
          log.message(entry.message);
        }
      }

      const nextSteps = `cd ${config.name}
cp .env.example .env.local
bun dev`;

      note(nextSteps, 'Next steps');
      outro(color.green('Project created successfully!'));
    } catch (err) {
      log.error(err instanceof Error ? err.message : 'Unknown error occurred');
      process.exit(1);
    }
  });

program.parse();

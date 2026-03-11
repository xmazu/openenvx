import { generateBaseTemplate } from '../../lib/templates';
import type { GenerateContext, LogEntry } from '../../lib/types';

export async function* generateBase(
  ctx: GenerateContext
): AsyncGenerator<LogEntry> {
  yield { message: 'Generating base template...', level: 'spinner' };
  await generateBaseTemplate(ctx.targetDir, ctx.config, ctx.packageManager);
  ctx.state.generated.push('base');
  yield { message: 'Base template generated', level: 'success' };

  if (
    ctx.config.database === 'postgres' ||
    ctx.config.database === 'postgresql'
  ) {
    yield { message: 'Configuring services...', level: 'spinner' };
    ctx.state.features.push('postgres');
    yield {
      message: 'docker-compose.yml configured',
      level: 'success',
    };
  }
}

import { appendEnvVariables } from '../../lib/templates';
import type { GenerateContext, LogEntry } from '../../lib/types';

export async function* setupEnvironment(
  ctx: GenerateContext
): AsyncGenerator<LogEntry> {
  yield { message: 'Appending environment variables...', level: 'spinner' };
  await appendEnvVariables(ctx.targetDir, ctx.config);

  yield { message: 'Environment configured', level: 'success' };
}

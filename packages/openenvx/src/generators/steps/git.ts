import { execa } from 'execa';
import type { GenerateContext, LogEntry } from '../../lib/types';

export async function* initGit(ctx: GenerateContext): AsyncGenerator<LogEntry> {
  yield { message: 'Initializing Git repository...', level: 'spinner' };
  await execa('git', ['init'], { cwd: ctx.targetDir });
  await execa('git', ['add', '.'], { cwd: ctx.targetDir });
  await execa('git', ['commit', '-m', 'Initial commit'], {
    cwd: ctx.targetDir,
  });
  yield { message: 'Git repository initialized', level: 'success' };
}

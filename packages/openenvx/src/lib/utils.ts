import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { execa } from 'execa';
import fs from 'fs-extra';
import type { PackageManager } from './types';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export function getTemplatesDir(subPath: string): string {
  const builtPath = path.join(__dirname, 'template', subPath);
  if (fs.existsSync(builtPath)) {
    return builtPath;
  }
  const devPath = path.join(__dirname, '../..', 'template', subPath);
  if (fs.existsSync(devPath)) {
    return devPath;
  }
  return devPath;
}

export async function detectPackageManager(): Promise<PackageManager> {
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

export function filterShadcnTooltipMessage(output: string): string {
  const lines = output.split('\n');
  const result: string[] = [];

  let skipping = false;
  let inFenceBlock = false;

  for (const line of lines) {
    if (
      !skipping &&
      line.includes(
        'The `tooltip` component has been added. Remember to wrap your app with the `TooltipProvider` component.'
      )
    ) {
      skipping = true;
      inFenceBlock = false;
      continue;
    }

    if (skipping) {
      const trimmed = line.trimStart();

      if (trimmed.startsWith('```')) {
        if (!inFenceBlock) {
          inFenceBlock = true;
          continue;
        }

        skipping = false;
        inFenceBlock = false;
        continue;
      }

      continue;
    }

    result.push(line);
  }

  return result.join('\n').trimEnd();
}

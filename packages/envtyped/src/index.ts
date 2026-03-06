import { z } from 'zod';

export type EnvSchema = Record<string, z.ZodType>;

export type InferEnv<T extends EnvSchema> = {
  [K in keyof T]: z.output<T[K]>;
};

export interface CreateEnvOptions<T extends EnvSchema> {
  env?: Record<string, string | undefined>;
  onValidationError?: (
    issues: Array<{ path: string; message: string }>
  ) => never;
  schema: T;
}

export class EnvValidationError extends Error {
  readonly issues: Array<{ path: string; message: string }>;

  constructor(
    message: string,
    issues: Array<{ path: string; message: string }>
  ) {
    super(message);
    this.name = 'EnvValidationError';
    this.issues = issues;
  }
}

export function createEnv<T extends EnvSchema>(
  schema: T,
  env: Record<string, string | undefined> = process.env
): InferEnv<T> {
  const parsedEnv: Record<string, unknown> = {};
  const issues: Array<{ path: string; message: string }> = [];

  for (const [key, validator] of Object.entries(schema)) {
    const rawValue = env[key];

    const result = validator.safeParse(rawValue);

    if (result.success) {
      parsedEnv[key] = result.data;
    } else {
      const errorMessage = result.error.issues
        .map((err) => err.message)
        .join(', ');
      issues.push({
        path: key,
        message: errorMessage,
      });
    }
  }

  if (issues.length > 0) {
    const formattedIssues = issues
      .map((issue) => `  - ${issue.path}: ${issue.message}`)
      .join('\n');
    throw new EnvValidationError(
      `Environment variable validation failed:\n${formattedIssues}`,
      issues
    );
  }

  return parsedEnv as InferEnv<T>;
}

export function createEnvWithOptions<T extends EnvSchema>(
  options: CreateEnvOptions<T>
): InferEnv<T> {
  try {
    return createEnv(options.schema, options.env);
  } catch (error) {
    if (error instanceof EnvValidationError && options.onValidationError) {
      options.onValidationError(error.issues);
    }
    throw error;
  }
}

export function booleanString() {
  return z
    .union([z.literal('true'), z.literal('false')])
    .transform((val) => val === 'true');
}

export function numberString() {
  return z
    .string()
    .refine((val) => !Number.isNaN(Number(val)), {
      message: 'Expected a number',
    })
    .transform((val) => Number(val));
}

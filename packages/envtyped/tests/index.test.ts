import { describe, expect, it } from 'vitest';
import { z } from 'zod';
import {
  booleanString,
  createEnv,
  createEnvWithOptions,
  EnvValidationError,
  numberString,
} from '../src/index';

describe('createEnv', () => {
  it('should validate and parse environment variables', () => {
    const env = createEnv(
      {
        PORT: z.string().transform((val) => Number(val)),
        NODE_ENV: z.enum(['development', 'production', 'test']),
        API_URL: z.string().url(),
      },
      {
        PORT: '3000',
        NODE_ENV: 'development',
        API_URL: 'https://api.example.com',
      }
    );

    expect(env.PORT).toBe(3000);
    expect(env.NODE_ENV).toBe('development');
    expect(env.API_URL).toBe('https://api.example.com');
  });

  it('should apply default values', () => {
    const env = createEnv(
      {
        PORT: z.string().default('3000'),
        NODE_ENV: z.enum(['development', 'production']).default('development'),
      },
      {}
    );

    expect(env.PORT).toBe('3000');
    expect(env.NODE_ENV).toBe('development');
  });

  it('should throw EnvValidationError for missing required variables', () => {
    expect(() =>
      createEnv(
        {
          DATABASE_URL: z.string().min(1),
        },
        {}
      )
    ).toThrow(EnvValidationError);
  });

  it('should throw EnvValidationError for invalid values', () => {
    expect(() =>
      createEnv(
        {
          PORT: z.number(),
        },
        {
          PORT: 'not-a-number',
        }
      )
    ).toThrow(EnvValidationError);
  });

  it('should provide detailed error messages', () => {
    try {
      createEnv(
        {
          PORT: z.number(),
          DATABASE_URL: z.string().url(),
        },
        {
          PORT: 'invalid',
          DATABASE_URL: 'not-a-url',
        }
      );
    } catch (error) {
      expect(error).toBeInstanceOf(EnvValidationError);
      expect((error as EnvValidationError).issues).toHaveLength(2);
      expect((error as EnvValidationError).issues[0].path).toBe('PORT');
      expect((error as EnvValidationError).issues[1].path).toBe('DATABASE_URL');
    }
  });

  it('should support transforms', () => {
    const env = createEnv(
      {
        FEATURE_FLAGS: z.string().transform((val) => val.split(',')),
      },
      {
        FEATURE_FLAGS: 'flag1,flag2,flag3',
      }
    );

    expect(env.FEATURE_FLAGS).toEqual(['flag1', 'flag2', 'flag3']);
  });

  it('should support refinements', () => {
    const env = createEnv(
      {
        PORT: z
          .string()
          .transform((val) => Number(val))
          .refine((val) => val > 0 && val < 65_536, {
            message: 'PORT must be between 1 and 65535',
          }),
      },
      {
        PORT: '8080',
      }
    );

    expect(env.PORT).toBe(8080);
  });

  it('should throw for out-of-range values', () => {
    expect(() =>
      createEnv(
        {
          PORT: z
            .string()
            .transform((val) => Number(val))
            .refine((val) => val > 0 && val < 65_536, {
              message: 'PORT must be between 1 and 65535',
            }),
        },
        {
          PORT: '99999',
        }
      )
    ).toThrow(EnvValidationError);
  });

  it('should support optional fields with undefined', () => {
    const env = createEnv(
      {
        OPTIONAL_VAR: z.string().optional(),
        REQUIRED_VAR: z.string(),
      },
      {
        REQUIRED_VAR: 'value',
      }
    );

    expect(env.OPTIONAL_VAR).toBeUndefined();
    expect(env.REQUIRED_VAR).toBe('value');
  });

  it('should support optional fields with default', () => {
    const env = createEnv(
      {
        OPTIONAL_VAR: z.string().optional().default('default-value'),
        REQUIRED_VAR: z.string(),
      },
      {
        REQUIRED_VAR: 'value',
      }
    );

    expect(env.OPTIONAL_VAR).toBe('default-value');
    expect(env.REQUIRED_VAR).toBe('value');
  });
});

describe('createEnvWithOptions', () => {
  it('should work with options object', () => {
    const env = createEnvWithOptions({
      schema: {
        PORT: z.string().default('3000'),
      },
      env: {},
    });

    expect(env.PORT).toBe('3000');
  });

  it('should call custom error handler', () => {
    const customErrorHandler = () => {
      throw new Error('Custom error');
    };

    expect(() =>
      createEnvWithOptions({
        schema: {
          REQUIRED: z.string(),
        },
        env: {},
        onValidationError: customErrorHandler,
      })
    ).toThrow('Custom error');
  });
});

describe('booleanString', () => {
  it('should parse "true" as boolean true', () => {
    const env = createEnv(
      {
        DEBUG: booleanString(),
      },
      {
        DEBUG: 'true',
      }
    );

    expect(env.DEBUG).toBe(true);
  });

  it('should parse "false" as boolean false', () => {
    const env = createEnv(
      {
        DEBUG: booleanString(),
      },
      {
        DEBUG: 'false',
      }
    );

    expect(env.DEBUG).toBe(false);
  });

  it('should throw for invalid boolean strings', () => {
    expect(() =>
      createEnv(
        {
          DEBUG: booleanString(),
        },
        {
          DEBUG: 'yes',
        }
      )
    ).toThrow(EnvValidationError);
  });

  it('should support default values', () => {
    const env = createEnv(
      {
        DEBUG: z
          .union([z.literal('true'), z.literal('false')])
          .default('false')
          .transform((val) => val === 'true'),
      },
      {}
    );

    expect(env.DEBUG).toBe(false);
  });
});

describe('numberString', () => {
  it('should parse valid numbers', () => {
    const env = createEnv(
      {
        PORT: numberString(),
        TIMEOUT: numberString(),
      },
      {
        PORT: '3000',
        TIMEOUT: '5000',
      }
    );

    expect(env.PORT).toBe(3000);
    expect(env.TIMEOUT).toBe(5000);
  });

  it('should parse negative numbers', () => {
    const env = createEnv(
      {
        OFFSET: numberString(),
      },
      {
        OFFSET: '-100',
      }
    );

    expect(env.OFFSET).toBe(-100);
  });

  it('should parse decimal numbers', () => {
    const env = createEnv(
      {
        RATE: numberString(),
      },
      {
        RATE: '3.14159',
      }
    );

    expect(env.RATE).toBeCloseTo(3.14, 2);
  });

  it('should throw for invalid numbers', () => {
    expect(() =>
      createEnv(
        {
          PORT: numberString(),
        },
        {
          PORT: 'not-a-number',
        }
      )
    ).toThrow(EnvValidationError);
  });

  it('should support default values', () => {
    const env = createEnv(
      {
        PORT: z
          .string()
          .default('3000')
          .transform((val) => Number(val)),
      },
      {}
    );

    expect(env.PORT).toBe(3000);
  });
});

describe('EnvValidationError', () => {
  it('should have correct name and message', () => {
    const issues = [{ path: 'VAR', message: 'Required' }];
    const error = new EnvValidationError('Validation failed', issues);

    expect(error.name).toBe('EnvValidationError');
    expect(error.message).toBe('Validation failed');
    expect(error.issues).toEqual(issues);
  });

  it('should be instanceof Error', () => {
    const error = new EnvValidationError('test', []);
    expect(error).toBeInstanceOf(Error);
  });
});

# envtyped

A typed environment variable parser using [Zod](https://zod.dev/) with support for defaults, transforms, and validation.

## Installation

```bash
bun add envtyped zod
# or
npm install envtyped zod
# or
yarn add envtyped zod
```

## Usage

### Basic Usage

```typescript
import { createEnv, z } from "envtyped";

const env = createEnv({
  PORT: z.string().default("3000"),
  NODE_ENV: z.enum(["development", "production", "test"]),
  DATABASE_URL: z.string().url(),
});

// env is fully typed
env.PORT; // string
env.NODE_ENV; // 'development' | 'production' | 'test'
env.DATABASE_URL; // string
```

### With Custom Environment

```typescript
const env = createEnv(
  {
    API_KEY: z.string().min(1),
    DEBUG: z.boolean().default(false),
  },
  {
    API_KEY: "secret-key",
    DEBUG: "true",
  },
);
```

### Transforms

```typescript
const env = createEnv({
  PORT: z.string().transform((val) => Number(val)),
  FEATURE_FLAGS: z.string().transform((val) => val.split(",")),
});

// env.PORT is number
// env.FEATURE_FLAGS is string[]
```

### Refinements

```typescript
const env = createEnv({
  PORT: z
    .string()
    .transform((val) => Number(val))
    .refine((val) => val > 0 && val < 65536, {
      message: "PORT must be between 1 and 65535",
    }),
});
```

## Helper Functions

### `booleanString()`

Parses string `'true'` or `'false'` to actual boolean:

```typescript
import { booleanString } from "envtyped";

const env = createEnv({
  DEBUG: booleanString().default("false"),
  ENABLE_CACHE: booleanString(),
});

// env.DEBUG is boolean (defaults to false)
// env.ENABLE_CACHE is boolean
```

### `numberString()`

Parses string numbers to actual numbers:

```typescript
import { numberString } from "envtyped";

const env = createEnv({
  PORT: numberString().default("3000"),
  TIMEOUT: numberString(),
});

// env.PORT is number (defaults to 3000)
// env.TIMEOUT is number
```

## Error Handling

When validation fails, an `EnvValidationError` is thrown with detailed information:

```typescript
import { createEnv, EnvValidationError, z } from "envtyped";

try {
  const env = createEnv({
    PORT: z.number(),
    DATABASE_URL: z.string().url(),
  });
} catch (error) {
  if (error instanceof EnvValidationError) {
    console.error("Validation failed:", error.message);
    console.error("Issues:", error.issues);
    // [
    //   { path: 'PORT', message: 'Expected number, received nan' },
    //   { path: 'DATABASE_URL', message: 'Invalid url' }
    // ]
  }
}
```

## TypeScript

`envtyped` provides full type inference:

```typescript
import { createEnv, z } from "envtyped";

const env = createEnv({
  PORT: z.number().default(3000),
  NODE_ENV: z.enum(["development", "production"]),
});

// TypeScript knows the exact types
const port: number = env.PORT;
const nodeEnv: "development" | "production" = env.NODE_ENV;
```

## API

### `createEnv(schema, env?)`

Creates a validated environment object.

- `schema`: Record of Zod schemas keyed by environment variable name
- `env`: Optional environment object (defaults to `process.env`)
- Returns: Typed object with parsed values
- Throws: `EnvValidationError` if validation fails

### `createEnvWithOptions(options)`

Creates a validated environment object with options.

```typescript
createEnvWithOptions({
  schema: {
    /* zod schemas */
  },
  env: {
    /* environment variables */
  },
  onValidationError: (issues) => {
    /* custom error handler */
  },
});
```

### `EnvValidationError`

Error class thrown when validation fails.

Properties:

- `message`: Human-readable error message
- `issues`: Array of `{ path: string; message: string }` objects

## License

MIT

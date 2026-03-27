import { z } from 'zod';
import type { FieldConfig } from './resource-types';

const COLOR_REGEX = /^#[0-9A-Fa-f]{6}$/;

export function generateZodSchema(
  fields: FieldConfig[],
  customSchema?: z.ZodObject<Record<string, z.ZodTypeAny>>
): z.ZodObject<Record<string, z.ZodTypeAny>> {
  const shape: Record<string, z.ZodTypeAny> = {};

  for (const field of fields) {
    if (field.hidden) {
      continue;
    }

    let schema: z.ZodTypeAny = buildFieldSchema(field);

    if (!field.required) {
      schema = schema.optional().nullable();
    }

    shape[field.name] = schema;
  }

  const baseSchema = z.object(shape);

  if (customSchema) {
    return baseSchema.merge(customSchema) as z.ZodObject<
      Record<string, z.ZodTypeAny>
    >;
  }

  return baseSchema;
}

function buildBaseStringSchema(field: FieldConfig): z.ZodString {
  let schema = z.string();

  if ('minLength' in field && typeof field.minLength === 'number') {
    schema = schema.min(field.minLength);
  }

  if ('maxLength' in field && typeof field.maxLength === 'number') {
    schema = schema.max(field.maxLength);
  }

  if ('pattern' in field && typeof field.pattern === 'string') {
    schema = schema.regex(new RegExp(field.pattern));
  }

  return schema;
}

function buildNumberSchema(field: FieldConfig): z.ZodNumber {
  let schema = z.number();

  if ('min' in field && typeof field.min === 'number') {
    schema = schema.min(field.min);
  }

  if ('max' in field && typeof field.max === 'number') {
    schema = schema.max(field.max);
  }

  return schema;
}

function buildFieldSchema(field: FieldConfig): z.ZodTypeAny {
  if (field.validation) {
    return field.validation;
  }

  switch (field.type) {
    case 'text':
    case 'textarea':
    case 'slug':
    case 'phone':
      return buildBaseStringSchema(field);

    case 'email':
      return z.string().email();

    case 'password': {
      let schema = z.string();
      if ('minLength' in field && typeof field.minLength === 'number') {
        schema = schema.min(field.minLength);
      }
      return schema;
    }

    case 'number':
    case 'integer':
      return z.coerce.number().pipe(buildNumberSchema(field));

    case 'boolean':
      return z.boolean();

    case 'date':
    case 'datetime':
      return z.coerce.date();

    case 'url':
      return z.string().url();

    case 'color':
      return z.string().regex(COLOR_REGEX);

    case 'select': {
      const selectField = field as {
        options: Array<{ value: string } | string>;
      };
      const options = selectField.options || [];
      const values = options.map((opt) =>
        typeof opt === 'object' ? opt.value : opt
      );
      if (values.length === 0) {
        return z.string();
      }
      return z.enum(values as [string, ...string[]]);
    }

    case 'multiselect': {
      const multiField = field as {
        options: Array<{ value: string } | string>;
      };
      const opts = multiField.options || [];
      const vals = opts.map((opt) =>
        typeof opt === 'object' ? opt.value : opt
      );
      if (vals.length === 0) {
        return z.array(z.string());
      }
      return z.array(z.enum(vals as [string, ...string[]]));
    }

    case 'json':
      return z.unknown();

    case 'reference':
      return z.union([z.string(), z.number()]);

    case 'array': {
      const arrayField = field as {
        of: { type: string; fields?: FieldConfig[] };
        minItems?: number;
        maxItems?: number;
      };
      let schema: z.ZodTypeAny;
      if (arrayField.of.type === 'object' && arrayField.of.fields) {
        schema = z.array(generateZodSchema(arrayField.of.fields));
      } else {
        schema = z.array(z.string());
      }
      if (arrayField.minItems !== undefined) {
        schema = (schema as z.ZodArray<z.ZodTypeAny>).min(arrayField.minItems);
      }
      if (arrayField.maxItems !== undefined) {
        schema = (schema as z.ZodArray<z.ZodTypeAny>).max(arrayField.maxItems);
      }
      return schema;
    }

    default:
      return z.unknown();
  }
}

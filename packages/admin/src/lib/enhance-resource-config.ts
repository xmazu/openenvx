import type { FieldConfig } from '@/lib/resource-types';
import type {
  IntrospectedColumn,
  IntrospectedTable,
  ResourceConfig,
  ResourceItem,
} from '@/types/resources';

function validateResourceConfig(
  resourceName: string,
  config: ResourceConfig,
  tableSchema: IntrospectedTable
): void {
  if (!config.fields) {
    return;
  }

  const definedFieldNames = new Set(config.fields.map((f) => f.name));

  for (const column of tableSchema.columns) {
    if (column.isPrimaryKey) {
      continue;
    }

    if (column.defaultValue !== null) {
      continue;
    }

    if (!(column.isNullable || definedFieldNames.has(column.name))) {
      console.warn(
        `[${resourceName}] Column "${column.name}" is NOT NULL in the database but is not defined in your resource config. ` +
          'This may cause validation errors when creating records.'
      );
    }
  }

  for (const field of config.fields) {
    const column = tableSchema.columns.find((c) => c.name === field.name);
    if (!column) {
      continue;
    }

    if (field.required === true && column.isNullable) {
      console.warn(
        `[${resourceName}] Field "${field.name}" is marked as required in your config, ` +
          'but the database column allows NULL values. Consider setting required: false ' +
          'or making the column NOT NULL in your schema.'
      );
    }

    if (
      field.required !== true &&
      !column.isNullable &&
      column.defaultValue === null &&
      !column.isPrimaryKey
    ) {
      console.warn(
        `[${resourceName}] Field "${field.name}" is NOT NULL in the database with no default value, ` +
          'but is not marked as required in your config. Consider setting required: true ' +
          'to prevent validation errors.'
      );
    }
  }
}

export function enhanceResourceConfigWithIntrospection(
  config: ResourceConfig | undefined,
  tableSchema: IntrospectedTable | undefined,
  resourceName?: string
): ResourceConfig | undefined {
  if (!config) {
    return undefined;
  }

  if (!(tableSchema && config.fields) || config.fields.length === 0) {
    return config;
  }

  if (resourceName) {
    validateResourceConfig(resourceName, config, tableSchema);
  }

  const enhancedFields: FieldConfig[] = config.fields.map((field) => {
    const column = tableSchema.columns.find((c) => c.name === field.name);
    if (!column) {
      return field;
    }

    return {
      ...field,
      _introspection: {
        dataType: column.dataType,
        isNullable: column.isNullable,
        isPrimaryKey: column.isPrimaryKey,
        defaultValue: column.defaultValue,
      },
    } as FieldConfig & { _introspection?: IntrospectedColumn };
  });

  return {
    ...config,
    fields: enhancedFields,
  };
}

export function enhanceResourcesWithIntrospection(
  resources: ResourceItem[],
  introspectionTables: IntrospectedTable[]
): ResourceItem[] {
  const tableMap = new Map(introspectionTables.map((t) => [t.name, t]));

  return resources.map((resource) => {
    const tableSchema = tableMap.get(resource.name);
    if (!(tableSchema && resource.config)) {
      return resource;
    }

    return {
      ...resource,
      config: enhanceResourceConfigWithIntrospection(
        resource.config,
        tableSchema,
        resource.name
      ),
    };
  });
}

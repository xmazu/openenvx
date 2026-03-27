import { detectCircularDependencies } from './circular-deps';
import type {
  FieldConfig,
  ForeignKeyInfo,
  IntrospectedColumn,
  ResourceConfig,
  ResourceDefinition,
} from './resource-types';
import { autoGenerateField } from './resource-types';

export interface DefineResourceOptions extends Omit<ResourceConfig, 'fields'> {
  fields?: FieldConfig[];
  tableName?: string;
}

export class ResourceRegistry {
  private readonly resources = new Map<string, ResourceDefinition>();

  register(resource: ResourceDefinition): void {
    this.resources.set(resource.tableName, resource);
  }

  get(tableName: string): ResourceDefinition | undefined {
    return this.resources.get(tableName);
  }

  has(tableName: string): boolean {
    return this.resources.has(tableName);
  }

  getAll(): ResourceDefinition[] {
    return Array.from(this.resources.values());
  }

  getTableNames(): string[] {
    return Array.from(this.resources.keys());
  }

  clear(): void {
    this.resources.clear();
  }
}

export const resourceRegistry = new ResourceRegistry();

export function defineResource(
  tableName: string,
  options: DefineResourceOptions = {}
): ResourceDefinition {
  const fields = options.fields || [];

  detectCircularDependencies(fields);

  const resource: ResourceDefinition = {
    tableName,
    config: {
      ...options,
      fields,
    },
  };
  resourceRegistry.register(resource);
  return resource;
}

export function mergeFields(
  manualFields: FieldConfig[] | undefined,
  autoFields: FieldConfig[]
): FieldConfig[] {
  if (!manualFields || manualFields.length === 0) {
    return autoFields;
  }

  const manualFieldMap = new Map(manualFields.map((f) => [f.name, f]));
  const merged: FieldConfig[] = [...manualFields];

  for (const autoField of autoFields) {
    if (!manualFieldMap.has(autoField.name)) {
      merged.push(autoField);
    }
  }

  return merged;
}

export function createMergedConfig(
  manualConfig: ResourceConfig | undefined,
  autoConfig: ResourceConfig
): ResourceConfig {
  if (!manualConfig) {
    return autoConfig;
  }

  return {
    label: manualConfig.label ?? autoConfig.label,
    icon: manualConfig.icon ?? autoConfig.icon,
    description: manualConfig.description ?? autoConfig.description,
    fields: mergeFields(manualConfig.fields, autoConfig.fields || []),
    list: { ...autoConfig.list, ...manualConfig.list },
    form: { ...autoConfig.form, ...manualConfig.form },
    hooks: { ...autoConfig.hooks, ...manualConfig.hooks },
    access: { ...autoConfig.access, ...manualConfig.access },
    canCreate: manualConfig.canCreate ?? autoConfig.canCreate ?? true,
    canEdit: manualConfig.canEdit ?? autoConfig.canEdit ?? true,
    canDelete: manualConfig.canDelete ?? autoConfig.canDelete ?? true,
    canShow: manualConfig.canShow ?? autoConfig.canShow ?? true,
    exclude: manualConfig.exclude ?? autoConfig.exclude ?? false,
  };
}

export function buildConfigFromColumns(
  columns: IntrospectedColumn[],
  foreignKeys?: ForeignKeyInfo[]
): ResourceConfig {
  const fields = columns.map((col) => autoGenerateField(col, foreignKeys));

  return {
    label: formatLabel(columns[0]?.name || 'Resource'),
    fields,
    canCreate: true,
    canEdit: true,
    canDelete: true,
    canShow: true,
    list: {
      columns: fields
        .filter((f) => !['id', 'created_at', 'updated_at'].includes(f.name))
        .slice(0, 5)
        .map((f) => f.name),
      searchable: fields
        .filter((f) => ['text', 'textarea', 'email'].includes(f.type))
        .slice(0, 3)
        .map((f) => f.name),
    },
    form: {
      layout: 'vertical',
      columns: 1,
    },
  };
}

const LABEL_REGEX_UNDERSCORE = /_/g;
const LABEL_REGEX_CAMEL = /([A-Z])/g;
const LABEL_REGEX_LEADING_SPACE = /^\s+/;
const LABEL_REGEX_EXTRA_SPACES = /\s+/g;

function formatLabel(name: string): string {
  return name
    .replace(LABEL_REGEX_UNDERSCORE, ' ')
    .replace(LABEL_REGEX_CAMEL, ' $1')
    .replace(LABEL_REGEX_LEADING_SPACE, '')
    .replace(LABEL_REGEX_EXTRA_SPACES, ' ')
    .toLowerCase()
    .replace(/\b\w/g, (l) => l.toUpperCase());
}

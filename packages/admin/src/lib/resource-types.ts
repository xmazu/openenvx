import type { ZodType } from 'zod';

export type FieldType =
  | 'text'
  | 'textarea'
  | 'number'
  | 'integer'
  | 'boolean'
  | 'date'
  | 'datetime'
  | 'email'
  | 'password'
  | 'select'
  | 'multiselect'
  | 'json'
  | 'reference'
  | 'url'
  | 'color'
  | 'phone'
  | 'slug'
  | 'rich-text'
  | 'array'
  | 'custom';

export interface SelectOption {
  label: string;
  value: string | number;
}

export interface ReferenceConfig {
  displayField?: string;
  filter?: (query: unknown) => unknown;
  table: string;
  valueField?: string;
}

export interface ComputedConfig {
  deps: string[];
  displayOnly?: boolean;
  fn: (data: Record<string, unknown>) => unknown;
}

export interface BaseFieldConfig {
  className?: string;
  computed?: ComputedConfig | ((data: Record<string, unknown>) => unknown);
  condition?: (data: Record<string, unknown>) => boolean;
  defaultValue?: unknown;
  description?: string;
  hidden?: boolean;
  label?: string;
  name: string;
  placeholder?: string;
  readOnly?: boolean;
  required?: boolean;
  type: FieldType;
  validate?: (value: unknown) => boolean | string;
  validation?: ZodType<unknown>;
  width?: string;
}

export interface TextFieldConfig extends BaseFieldConfig {
  maxLength?: number;
  minLength?: number;
  pattern?: string;
  type: 'text';
}

export interface TextareaFieldConfig extends BaseFieldConfig {
  maxLength?: number;
  minLength?: number;
  rows?: number;
  type: 'textarea';
}

export interface NumberFieldConfig extends BaseFieldConfig {
  max?: number;
  min?: number;
  step?: number;
  type: 'number' | 'integer';
}

export interface BooleanFieldConfig extends BaseFieldConfig {
  type: 'boolean';
}

export interface DateFieldConfig extends BaseFieldConfig {
  max?: string;
  min?: string;
  type: 'date' | 'datetime';
}

export interface EmailFieldConfig extends BaseFieldConfig {
  type: 'email';
}

export interface PasswordFieldConfig extends BaseFieldConfig {
  type: 'password';
}

export interface SelectFieldConfig extends BaseFieldConfig {
  options: SelectOption[] | string[];
  type: 'select' | 'multiselect';
}

export interface JsonFieldConfig extends BaseFieldConfig {
  type: 'json';
}

export interface ReferenceFieldConfig extends BaseFieldConfig {
  displayColumn?: string;
  multiple?: boolean;
  reference: ReferenceConfig;
  type: 'reference';
}

export interface UrlFieldConfig extends BaseFieldConfig {
  type: 'url';
}

export interface ColorFieldConfig extends BaseFieldConfig {
  type: 'color';
}

export interface PhoneFieldConfig extends BaseFieldConfig {
  type: 'phone';
}

export interface SlugFieldConfig extends BaseFieldConfig {
  source?: string;
  type: 'slug';
}

export interface RichTextFieldConfig extends BaseFieldConfig {
  type: 'rich-text';
}

export interface ArrayFieldConfig extends BaseFieldConfig {
  maxItems?: number;
  minItems?: number;
  of: { fields?: FieldConfig[]; type: 'object' | 'text' };
  type: 'array';
}

export interface CustomFieldConfig extends BaseFieldConfig {
  component: React.ComponentType<FieldComponentProps>;
  type: 'custom';
}

export interface FieldComponentProps {
  disabled?: boolean;
  error?: string;
  field: FieldConfig;
  onBlur?: () => void;
  onChange: (value: unknown) => void;
  value: unknown;
}

export type FieldConfig =
  | TextFieldConfig
  | TextareaFieldConfig
  | NumberFieldConfig
  | BooleanFieldConfig
  | DateFieldConfig
  | EmailFieldConfig
  | PasswordFieldConfig
  | SelectFieldConfig
  | JsonFieldConfig
  | ReferenceFieldConfig
  | UrlFieldConfig
  | ColorFieldConfig
  | PhoneFieldConfig
  | SlugFieldConfig
  | RichTextFieldConfig
  | ArrayFieldConfig
  | CustomFieldConfig;

export interface ListFilter {
  field: string;
  options?: string[];
  type: 'select' | 'date-range' | 'text';
}

export interface BulkActionConfig {
  confirm?: {
    destructive?: boolean;
    message: string;
    title: string;
  };
  dialog?: {
    fields: Array<{
      label: string;
      name: string;
      options?: string[];
      type: 'text' | 'select';
    }>;
    title: string;
  };
  icon?: string;
  key: string;
  label: string;
}

export interface ListViewConfig {
  actions?: {
    clone?: boolean;
    delete?: boolean;
    edit?: boolean;
    show?: boolean;
  };
  bulkActions?: BulkActionConfig[];
  columns?: string[];
  defaultSort?: { direction: 'asc' | 'desc'; field: string };
  filters?: ListFilter[];
  pageSizeOptions?: number[];
  perPage?: number;
  searchable?: string[];
}

export type FormLayoutItem =
  | {
      columns?: number;
      fields: string[];
      label?: string;
      title?: string;
      type: 'tab';
    }
  | {
      columns?: number;
      fields: string[];
      label?: string;
      title?: string;
      type: 'collapsible';
    }
  | {
      columns?: number;
      fields: string[];
      label?: string;
      title?: string;
      type: 'group';
    };

export interface FormViewConfig {
  columns?: number;
  groups?: FormLayoutItem[];
  layout?: 'vertical' | 'horizontal' | 'grid';
  showDescriptions?: boolean;
}

export type AccessFunction = (
  user: unknown,
  record?: Record<string, unknown>
) => boolean;

export interface AccessConfig {
  create?: AccessFunction;
  delete?: AccessFunction;
  read?: AccessFunction;
  update?: AccessFunction;
}

export interface HookContext {
  db?: unknown;
  params: Record<string, string>;
  request: Request;
  response: Response;
  user?: unknown;
}

export interface ListParams {
  filters?: Array<{ field: string; operator: string; value: unknown }>;
  pagination?: { current?: number; pageSize?: number };
  sorters?: Array<{ direction: 'asc' | 'desc'; field: string }>;
}

export interface ResourceHooks {
  afterCreate?: (
    data: Record<string, unknown>,
    context: HookContext
  ) => Promise<void>;
  afterDelete?: (id: string | number, context: HookContext) => Promise<void>;
  afterList?: (data: unknown[], context: HookContext) => Promise<unknown[]>;
  afterUpdate?: (
    data: Record<string, unknown>,
    id: string | number,
    context: HookContext
  ) => Promise<void>;
  beforeCreate?: (
    data: Record<string, unknown>,
    context: HookContext
  ) => Promise<Record<string, unknown>>;
  beforeDelete?: (
    id: string | number,
    context: HookContext
  ) => Promise<boolean>;
  beforeList?: (
    params: ListParams,
    context: HookContext
  ) => Promise<ListParams>;
  beforeUpdate?: (
    data: Record<string, unknown>,
    id: string | number,
    context: HookContext
  ) => Promise<Record<string, unknown>>;
}

export interface ResourceConfig {
  access?: AccessConfig;
  canCreate?: boolean;
  canDelete?: boolean;
  canEdit?: boolean;
  canShow?: boolean;
  description?: string;
  exclude?: boolean;
  fields: FieldConfig[];
  form?: FormViewConfig;
  hooks?: ResourceHooks;
  icon?: string;
  identifier?: string;
  label?: string;
  list?: ListViewConfig;
}

export interface ResourceDefinition {
  config: ResourceConfig;
  tableName: string;
}

export type DefinedResource = ResourceDefinition;

export interface IntrospectedColumn {
  dataType: string;
  defaultValue: string | null;
  isNullable: boolean;
  isPrimaryKey: boolean;
  name: string;
}

export interface IntrospectedTable {
  columns: IntrospectedColumn[];
  foreignKeys: ForeignKeyInfo[];
  name: string;
  primaryKey: string | null;
}

export interface ForeignKeyInfo {
  column: string;
  foreignColumn: string;
  foreignTable: string;
}

export const POSTGRES_TYPE_MAP: Record<string, FieldType> = {
  'character varying': 'text',
  varchar: 'text',
  character: 'text',
  char: 'text',
  text: 'textarea',
  integer: 'integer',
  bigint: 'integer',
  smallint: 'integer',
  numeric: 'number',
  decimal: 'number',
  real: 'number',
  'double precision': 'number',
  boolean: 'boolean',
  date: 'date',
  'timestamp without time zone': 'datetime',
  'timestamp with time zone': 'datetime',
  'time without time zone': 'text',
  'time with time zone': 'text',
  json: 'json',
  jsonb: 'json',
  uuid: 'text',
  ARRAY: 'json',
};

const EMAIL_PATTERN = /^email$/i;
const PASSWORD_PATTERN = /^password$/i;
const PHONE_PATTERN = /^phone|tel$/i;
const URL_PATTERN = /^url|link|website$/i;
const COLOR_PATTERN = /^color$/i;
const SLUG_PATTERN = /^slug$/i;
const RICH_TEXT_PATTERN = /^content|body|description$/i;
const BOOLEAN_PATTERN = /^is_|has_|can_/i;
const DATETIME_PATTERN = /_at$|_on$|date$|time$/i;

export function inferFieldTypeFromName(columnName: string): FieldType | null {
  const patterns: [RegExp, FieldType][] = [
    [EMAIL_PATTERN, 'email'],
    [PASSWORD_PATTERN, 'password'],
    [PHONE_PATTERN, 'phone'],
    [URL_PATTERN, 'url'],
    [COLOR_PATTERN, 'color'],
    [SLUG_PATTERN, 'slug'],
    [RICH_TEXT_PATTERN, 'rich-text'],
    [BOOLEAN_PATTERN, 'boolean'],
    [DATETIME_PATTERN, 'datetime'],
  ];

  for (const [pattern, type] of patterns) {
    if (pattern.test(columnName)) {
      return type;
    }
  }

  return null;
}

export function autoGenerateField(
  column: IntrospectedColumn,
  foreignKeys?: ForeignKeyInfo[]
): FieldConfig {
  const fk = foreignKeys?.find((fk) => fk.column === column.name);
  if (fk) {
    return {
      name: column.name,
      type: 'reference',
      required: !column.isNullable && column.defaultValue === null,
      readOnly: column.isPrimaryKey,
      reference: {
        table: fk.foreignTable,
        displayField: 'name',
      },
    } as FieldConfig;
  }

  const nameBasedType = inferFieldTypeFromName(column.name);
  const pgType = column.dataType.toLowerCase();
  const type = nameBasedType || POSTGRES_TYPE_MAP[pgType] || 'text';

  const baseConfig = {
    name: column.name,
    type,
    required: !column.isNullable && column.defaultValue === null,
    readOnly: column.isPrimaryKey,
  };

  switch (type) {
    case 'select':
      return { ...baseConfig, type: 'select', options: [] };
    case 'textarea':
      return { ...baseConfig, type: 'textarea', rows: 3 };
    case 'number':
    case 'integer':
      return { ...baseConfig, type };
    default:
      return baseConfig as FieldConfig;
  }
}

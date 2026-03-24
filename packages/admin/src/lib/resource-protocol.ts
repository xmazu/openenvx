/**
 * Declarative CRUD Protocol for @openenvx/admin
 *
 * This module provides a type-safe way to define admin resources
 * with automatic form and table generation from database introspection.
 */

// ============================================================================
// Field Types
// ============================================================================

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
  | 'rich-text';

// ============================================================================
// Base Field Configuration
// ============================================================================

export interface BaseFieldConfig {
  /** Custom CSS class for the field container */
  className?: string;
  /** Default value */
  defaultValue?: unknown;
  /** Helper text shown below the field */
  description?: string;
  /** Whether to hide this field from forms */
  hidden?: boolean;
  /** Display label (defaults to formatted column name) */
  label?: string;
  /** Database column name */
  name: string;
  /** Placeholder text */
  placeholder?: string;
  /** Whether the field is read-only */
  readOnly?: boolean;
  /** Whether the field is required */
  required?: boolean;
  /** Field type - determines the input component */
  type: FieldType;
  /** Validation function */
  validate?: (value: unknown) => boolean | string;
  /** Field width (e.g., '100%', '50%', '200px') */
  width?: string;
}

// ============================================================================
// Specific Field Configurations
// ============================================================================

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

export interface SelectOption {
  label: string;
  value: string | number;
}

export interface SelectFieldConfig extends BaseFieldConfig {
  options: SelectOption[] | string[];
  type: 'select' | 'multiselect';
}

export interface JsonFieldConfig extends BaseFieldConfig {
  type: 'json';
}

export interface ReferenceFieldConfig extends BaseFieldConfig {
  /** The column to display (defaults to 'name', 'title', 'email', or 'id') */
  displayColumn?: string;
  /** Whether to allow multiple selections */
  multiple?: boolean;
  /** The referenced table name */
  referenceTable: string;
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
  /** Field to generate slug from (e.g., 'title') */
  source?: string;
  type: 'slug';
}

export interface RichTextFieldConfig extends BaseFieldConfig {
  type: 'rich-text';
}

// ============================================================================
// Union Type for All Field Configs
// ============================================================================

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
  | RichTextFieldConfig;

// ============================================================================
// List View Configuration
// ============================================================================

export interface ListViewConfig {
  /** Custom actions for each row */
  actions?: {
    show?: boolean;
    edit?: boolean;
    delete?: boolean;
    clone?: boolean;
  };
  /** Whether to enable bulk actions */
  bulkActions?: boolean;
  /** Columns to display (defaults to first 5 fields or all if less) */
  columns?: string[];
  /** Default sort column (prefix with - for descending) */
  defaultSort?: string;
  /** Whether to enable filters */
  filters?: boolean;
  /** Available page size options */
  pageSizeOptions?: number[];
  /** Default page size */
  perPage?: number;
  /** Fields that can be searched */
  searchable?: string[];
}

// ============================================================================
// Form View Configuration
// ============================================================================

export type FormLayout = 'vertical' | 'horizontal' | 'grid';

export interface FormFieldGroup {
  /** Column span for grid layout */
  columns?: number;
  /** Fields in this group */
  fields: string[];
  /** Group title */
  title?: string;
}

export interface FormViewConfig {
  /** Number of columns for grid layout */
  columns?: number;
  /** Field groups for organizing the form */
  groups?: FormFieldGroup[];
  /** Form layout style */
  layout?: FormLayout;
  /** Whether to show field descriptions */
  showDescriptions?: boolean;
}

// ============================================================================
// Resource Configuration
// ============================================================================

export interface ResourceConfig {
  /** Whether to enable the create operation */
  canCreate?: boolean;
  /** Whether to enable the delete operation */
  canDelete?: boolean;
  /** Whether to enable the edit operation */
  canEdit?: boolean;
  /** Whether to enable the show/view operation */
  canShow?: boolean;
  /** Description shown in the resource list */
  description?: string;
  /** Exclude this resource from the admin panel */
  exclude?: boolean;
  /** Field definitions - if provided, only these fields are shown */
  fields?: FieldConfig[];
  /** Form view configuration (used for both create and edit) */
  form?: FormViewConfig;
  /** Icon name from Lucide icons */
  icon?: string;
  /** Custom identifier for the resource (defaults to table name) */
  identifier?: string;
  /** Display label (defaults to formatted table name) */
  label?: string;
  /** List view configuration */
  list?: ListViewConfig;
}

// ============================================================================
// Resource Registry
// ============================================================================

export interface ResourceDefinition {
  /** Resource configuration */
  config: ResourceConfig;
  /** Table name in the database */
  tableName: string;
}

export type ResourceRegistry = Map<string, ResourceDefinition>;

// ============================================================================
// Database Introspection Types
// ============================================================================

export interface IntrospectedColumn {
  dataType: string;
  defaultValue: string | null;
  isNullable: boolean;
  isPrimaryKey: boolean;
  name: string;
}

export interface IntrospectedTable {
  columns: IntrospectedColumn[];
  name: string;
  primaryKey: string | null;
}

// ============================================================================
// Auto-Generated Field Mapping
// ============================================================================

/**
 * Maps PostgreSQL data types to field types
 */
export const POSTGRES_TYPE_MAP: Record<string, FieldType> = {
  // Text types
  'character varying': 'text',
  varchar: 'text',
  character: 'text',
  char: 'text',
  text: 'textarea',

  // Numeric types
  integer: 'integer',
  bigint: 'integer',
  smallint: 'integer',
  numeric: 'number',
  decimal: 'number',
  real: 'number',
  'double precision': 'number',

  // Boolean
  boolean: 'boolean',

  // Date/Time
  date: 'date',
  'timestamp without time zone': 'datetime',
  'timestamp with time zone': 'datetime',
  'time without time zone': 'text',
  'time with time zone': 'text',

  // JSON
  json: 'json',
  jsonb: 'json',

  // UUID
  uuid: 'text',

  // Arrays (simplified to JSON for now)
  ARRAY: 'json',
};

// Regex patterns for field type inference (defined at top level for performance)
const EMAIL_PATTERN = /^email$/i;
const PASSWORD_PATTERN = /^password$/i;
const PHONE_PATTERN = /^phone|tel$/i;
const URL_PATTERN = /^url|link|website$/i;
const COLOR_PATTERN = /^color$/i;
const SLUG_PATTERN = /^slug$/i;
const RICH_TEXT_PATTERN = /^content|body|description$/i;
const BOOLEAN_PATTERN = /^is_|has_|can_/i;
const DATETIME_PATTERN = /_at$|_on$|date$|time$/i;

/**
 * Infer field type from column name patterns
 */
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

/**
 * Auto-generate field configuration from introspected column
 */
export function autoGenerateField(column: IntrospectedColumn): FieldConfig {
  // First, try to infer from column name
  const nameBasedType = inferFieldTypeFromName(column.name);

  // Then, map from PostgreSQL type
  const pgType = column.dataType.toLowerCase();
  const type = nameBasedType || POSTGRES_TYPE_MAP[pgType] || 'text';

  const baseConfig = {
    name: column.name,
    type,
    required: !column.isNullable && column.defaultValue === null,
    readOnly: column.isPrimaryKey,
  };

  // Add type-specific configurations
  switch (type) {
    case 'select':
      return {
        ...baseConfig,
        type: 'select',
        options: [],
      };
    case 'textarea':
      return {
        ...baseConfig,
        type: 'textarea',
        rows: 3,
      };
    case 'number':
    case 'integer':
      return {
        ...baseConfig,
        type,
      };
    default:
      return baseConfig as FieldConfig;
  }
}

// ============================================================================
// Utility Types
// ============================================================================

export type InferFieldValue<T extends FieldConfig> = T extends {
  type: 'boolean';
}
  ? boolean
  : T extends { type: 'number' | 'integer' }
    ? number
    : T extends { type: 'json' }
      ? unknown
      : T extends { type: 'multiselect' }
        ? (string | number)[]
        : string;

export type ResourceData<T extends FieldConfig[]> = {
  [K in T[number] as K['name']]: InferFieldValue<K>;
};

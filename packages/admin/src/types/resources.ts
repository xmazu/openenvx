import type { FieldConfig, ResourceHooks } from '@/lib/resource-types';

export interface NestedResourceConfig {
  fields?: FieldConfig[];
  icon?: string;
  label: string;
  parentField?: string;
}

export interface ResourceConfig {
  description?: string;
  displayField?: string;
  fields?: FieldConfig[];
  form?: {
    layout?: 'vertical' | 'horizontal' | 'grid';
    columns?: number;
  };
  hooks?: ResourceHooks;
  icon?: string;
  label: string;
  list?: {
    columns?: string[];
    searchable?: string[];
    perPage?: number;
  };
  nested?: Record<string, NestedResourceConfig>;
}

export type ResourcesConfig = Record<string, ResourceConfig>;

export interface ResourceItem {
  config?: ResourceConfig;
  create?: string;
  displayField?: string;
  edit?: string;
  icon?: string;
  label: string;
  list: string;
  meta?: {
    label?: string;
    icon?: string;
    [key: string]: unknown;
  };
  name: string;
  nested?: Record<string, NestedResourceItem>;
  show?: string;
}

export interface NestedResourceItem {
  create?: string;
  edit?: string;
  icon?: string;
  label: string;
  list: string;
  meta?: {
    label?: string;
    icon?: string;
    [key: string]: unknown;
  };
  name: string;
  parentField?: string;
  show?: string;
}

export interface TreeMenuItem {
  children?: TreeMenuItem[];
  icon?: string;
  key: string;
  label: string;
  meta?: {
    label?: string;
    icon?: string;
    [key: string]: unknown;
  };
  name: string;
  route?: string;
}

export interface BreadcrumbItem {
  href?: string;
  icon?: string;
  label: string;
}

export interface IntrospectedColumn {
  dataType: string;
  defaultValue: string | null;
  isNullable: boolean;
  isPrimaryKey: boolean;
  name: string;
}

export interface ForeignKeyInfo {
  column: string;
  foreignColumn: string;
  foreignTable: string;
}

export interface IntrospectedTable {
  columns: IntrospectedColumn[];
  foreignKeys: ForeignKeyInfo[];
  name: string;
  primaryKey: string | null;
}

export interface IntrospectionData {
  tables: IntrospectedTable[];
  version: string;
}

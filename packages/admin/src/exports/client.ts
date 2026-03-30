'use client';

export * from '../components/admin-provider';
export * from '../context';
export { createAdminAuthClient } from '../lib/create-admin-auth-client';
export * from '../pages/dynamic-admin-page';
export type { AuthClient, AuthSession, AuthUser } from '../types';
export type {
  BreadcrumbItem,
  ForeignKeyInfo,
  IntrospectedColumn,
  IntrospectedTable,
  IntrospectionData,
  NestedResourceConfig,
  NestedResourceItem,
  ResourceConfig,
  ResourceItem,
  ResourcesConfig,
  TreeMenuItem,
} from '../types/resources';

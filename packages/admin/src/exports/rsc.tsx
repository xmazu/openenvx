import type React from 'react';
import { AdminProvider } from '../components/admin-provider';
import { enhanceResourcesWithIntrospection } from '../lib/enhance-resource-config';
import { fetchAllSchemas } from '../server/introspection';
import type { AuthClient } from '../types';
import type {
  IntrospectedTable,
  IntrospectionData,
  ResourceItem,
} from '../types/resources';
import type { Admin } from './server';

export interface AdminServerProviderProps {
  admin: Admin;
  authClient?: AuthClient;
  children: React.ReactNode;
}

function validateResource(
  resource: ResourceItem,
  tableNames: Set<string>,
  tables: IntrospectedTable[]
): void {
  if (!tableNames.has(resource.name)) {
    console.warn(
      `[Admin] Resource "${resource.name}" is defined but table "${resource.name}" was not found in the database introspection.`
    );
  }

  if (!resource.nested) {
    return;
  }

  for (const [nestedName, nestedConfig] of Object.entries(resource.nested)) {
    const nestedTable = tables.find((t) => t.name === nestedName);
    if (!nestedTable) {
      console.warn(
        `[Admin] Nested resource "${nestedName}" for "${resource.name}" is defined but table "${nestedName}" was not found in the database introspection.`
      );
      continue;
    }

    if (nestedConfig.parentField) {
      const hasField = nestedTable.columns.some(
        (c) => c.name === nestedConfig.parentField
      );
      if (!hasField) {
        console.warn(
          `[Admin] Nested resource "${nestedName}" specifies parentField "${nestedConfig.parentField}" but column not found in table "${nestedName}".`
        );
      }
    }
  }
}

function validateResources(
  resources: ResourceItem[],
  introspection: IntrospectionData
): void {
  const tableNames = new Set(introspection.tables.map((t) => t.name));
  for (const resource of resources) {
    validateResource(resource, tableNames, introspection.tables);
  }
}

export async function AdminServerProvider({
  admin,
  authClient,
  children,
}: AdminServerProviderProps) {
  const resources = admin.resources;
  let introspection: IntrospectionData | undefined;
  let enhancedResources = resources;

  if (!Array.isArray(resources)) {
    console.error(
      '[AdminServerProvider] Expected resources to be an array, but got:',
      typeof resources,
      resources
    );
    return (
      <AdminProvider authClient={authClient} resources={[]}>
        {children}
      </AdminProvider>
    );
  }

  try {
    const tables = await fetchAllSchemas();
    introspection = {
      tables,
      version: '1.0.0',
    };

    validateResources(resources, introspection);
    enhancedResources = enhanceResourcesWithIntrospection(resources, tables);
  } catch (error) {
    console.warn('Failed to fetch database introspection:', error);
  }

  return (
    <AdminProvider authClient={authClient} resources={enhancedResources}>
      {children}
    </AdminProvider>
  );
}

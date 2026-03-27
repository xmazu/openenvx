import { useEffect, useState } from 'react';
import type { FieldConfig, ResourceConfig } from '@/lib/resource-types';
import { autoGenerateField } from '@/lib/resource-types';
import type { TableSchema } from '@/server/introspection';

const LABEL_REGEX_UNDERSCORE = /_/g;
const LABEL_REGEX_CAMEL = /([A-Z])/g;
const LABEL_REGEX_LEADING_SPACE = /^\s+/;
const LABEL_REGEX_EXTRA_SPACES = /\s+/g;

export interface UseResourceConfigResult {
  config: ResourceConfig | null;
  error: string | null;
  loading: boolean;
}

export function useResourceConfig(tableName: string): UseResourceConfigResult {
  const [result, setResult] = useState<UseResourceConfigResult>({
    config: null,
    error: null,
    loading: true,
  });

  useEffect(() => {
    async function fetchConfig() {
      try {
        const response = await fetch(
          `/api/admin/resources/${tableName}/config`
        );

        if (!response.ok) {
          throw new Error(
            `Failed to fetch resource config: ${response.statusText}`
          );
        }

        const data = await response.json();
        setResult({
          config: data.config,
          error: null,
          loading: false,
        });
      } catch (err) {
        setResult({
          config: null,
          error: err instanceof Error ? err.message : 'Unknown error',
          loading: false,
        });
      }
    }

    if (tableName) {
      fetchConfig();
    }
  }, [tableName]);

  return result;
}

export function buildConfigFromSchema(schema: TableSchema): ResourceConfig {
  const fields: FieldConfig[] = schema.columns.map((col) =>
    autoGenerateField(col, schema.foreignKeys)
  );

  return {
    label: formatLabel(schema.name),
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

function formatLabel(name: string): string {
  return name
    .replace(LABEL_REGEX_UNDERSCORE, ' ')
    .replace(LABEL_REGEX_CAMEL, ' $1')
    .replace(LABEL_REGEX_LEADING_SPACE, '')
    .replace(LABEL_REGEX_EXTRA_SPACES, ' ')
    .toLowerCase()
    .replace(/\b\w/g, (l) => l.toUpperCase());
}

import { unstable_cache } from 'next/cache';
import { resourceRegistry } from '@/lib/define-resource';
import type { IResourceItem } from '@/types';
import { fetchTables } from './introspection';

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

export const getResources = unstable_cache(
  async (): Promise<IResourceItem[]> => {
    const tableNames = await fetchTables();
    const manualConfigs = resourceRegistry.getAll();

    return tableNames
      .filter((name) => {
        const manual = manualConfigs.find((m) => m.tableName === name);
        return !manual?.config.exclude;
      })
      .map((name) => {
        const manual = manualConfigs.find((m) => m.tableName === name)?.config;

        return {
          name,
          identifier: manual?.identifier ?? name,
          label: manual?.label ?? formatLabel(name),
          list: `/${name}`,
          create: `/${name}/create`,
          edit: `/${name}/edit`,
          show: `/${name}/show`,
          meta: {
            icon: manual?.icon,
            label: manual?.label,
          },
        };
      });
  },
  ['admin-resources'],
  { revalidate: 60, tags: ['admin-schema'] }
);

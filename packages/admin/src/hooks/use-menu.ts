'use client';

import { useContext, useMemo } from 'react';
import type { IResourceItem, TreeMenuItem } from '@/types';
import { ResourcesContext } from './use-resources';

export interface UseMenuResult {
  menuItems: TreeMenuItem[];
  selectedKey?: string;
}

export function useMenu(): UseMenuResult {
  const context = useContext(ResourcesContext);

  return useMemo(() => {
    const menuItems = buildMenuItems(context?.resources ?? []);

    return {
      menuItems,
      selectedKey: context?.selectedKey,
    };
  }, [context]);
}

function buildMenuItems(resources: IResourceItem[]): TreeMenuItem[] {
  const items: TreeMenuItem[] = [];

  for (const resource of resources) {
    // Skip resources without a list route
    if (!resource.list) {
      continue;
    }

    const item: TreeMenuItem = {
      name: resource.name,
      key: resource.identifier ?? resource.name,
      route: resource.list,
      label: resource.meta?.label ?? resource.label ?? resource.name,
      icon: resource.meta?.icon ?? resource.icon,
      meta: resource.meta,
    };

    items.push(item);
  }

  return items;
}

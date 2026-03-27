'use client';

import React, { createContext, useContext, useMemo } from 'react';
import type { IResourceItem, TreeMenuItem } from '@/types';

interface ResourcesContextValue {
  menuItems: TreeMenuItem[];
  resources: IResourceItem[];
  selectedKey?: string;
}

export const ResourcesContext = createContext<ResourcesContextValue>({
  menuItems: [],
  resources: [],
});

export function useResources(): ResourcesContextValue {
  return useContext(ResourcesContext);
}

interface ResourcesProviderProps {
  children: React.ReactNode;
  resources: IResourceItem[];
}

export function ResourcesProvider({
  children,
  resources,
}: ResourcesProviderProps) {
  const value = useMemo(() => {
    const menuItems = buildMenuItems(resources);

    return {
      resources,
      menuItems,
      selectedKey: undefined,
    };
  }, [resources]);

  return React.createElement(ResourcesContext.Provider, { value }, children);
}

function buildMenuItems(resources: IResourceItem[]): TreeMenuItem[] {
  return resources.map((resource) => ({
    name: resource.name,
    key: resource.identifier ?? resource.name,
    route: resource.list,
    label: resource.meta?.label ?? resource.label ?? resource.name,
    icon: resource.meta?.icon ?? resource.icon,
    meta: resource.meta,
  }));
}

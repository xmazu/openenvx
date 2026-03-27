'use client';

import { AdminContextProvider } from '@/hooks';
import { ResourcesProvider } from '@/hooks/use-resources';
import type { IResourceItem } from '@/types';
import { Layout } from '@/ui/layout/layout';

export interface AdminProviderProps {
  children: React.ReactNode;
  resources: IResourceItem[];
}

export const AdminProvider = ({ children, resources }: AdminProviderProps) => {
  return (
    <AdminContextProvider>
      <ResourcesProvider resources={resources}>
        <Layout>{children}</Layout>
      </ResourcesProvider>
    </AdminContextProvider>
  );
};

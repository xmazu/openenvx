'use client';

import { AuthProvider } from '@/context/auth-context';
import { AdminContextProvider } from '@/hooks';
import { ResourcesProvider } from '@/hooks/use-resources';
import type { AuthClient, IResourceItem } from '@/types';
import { Layout } from '@/ui/layout/layout';

export interface AdminProviderProps {
  authClient: AuthClient;
  children: React.ReactNode;
  resources: IResourceItem[];
}

export const AdminProvider = ({
  children,
  resources,
  authClient,
}: AdminProviderProps) => {
  return (
    <AuthProvider authClient={authClient} skipSessionFetch>
      <AdminContextProvider>
        <ResourcesProvider resources={resources}>
          <Layout>{children}</Layout>
        </ResourcesProvider>
      </AdminContextProvider>
    </AuthProvider>
  );
};

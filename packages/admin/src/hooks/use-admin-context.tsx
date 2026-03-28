'use client';

import { createContext, useContext, useMemo } from 'react';

import type { DataProvider } from '@/server/data-provider';
import { createPostgRESTDataProvider } from '@/server/data-provider';

const ADMIN_API_URL = '/api/admin';

interface AdminContextValue {
  dataProvider: DataProvider;
}

const AdminContext = createContext<AdminContextValue | null>(null);

export function useAdminContext() {
  const context = useContext(AdminContext);
  if (!context) {
    throw new Error(
      'useAdminContext must be used within an AdminContextProvider'
    );
  }
  return context;
}

export function useDataProvider() {
  const { dataProvider } = useAdminContext();
  return dataProvider;
}

interface AdminContextProviderProps {
  children: React.ReactNode;
  dataProvider?: DataProvider;
}

export function AdminContextProvider({
  children,
  dataProvider: customDataProvider,
}: AdminContextProviderProps) {
  const dataProvider = useMemo(
    () =>
      customDataProvider ??
      createPostgRESTDataProvider({
        apiUrl: ADMIN_API_URL,
      }),
    [customDataProvider]
  );

  return (
    <AdminContext.Provider value={{ dataProvider }}>
      {children}
    </AdminContext.Provider>
  );
}

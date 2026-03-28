'use client';

import { useAuth } from '@/context/auth-context';
import type { AuthProvider } from '@/types';

export function useActiveAuthProvider(): AuthProvider | undefined {
  const { user, signOut } = useAuth();

  return {
    getIdentity: async <TData = unknown>() => user as TData,
    logout: async () => {
      await signOut();
      return { success: true };
    },
  };
}

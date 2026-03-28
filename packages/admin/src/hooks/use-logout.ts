'use client';

import { useCallback, useState } from 'react';
import { useAuth } from '@/context/auth-context';

export interface UseLogoutResult {
  isPending: boolean;
  mutate: () => Promise<void>;
}

export function useLogout(): UseLogoutResult {
  const { signOut } = useAuth();
  const [isPending, setIsPending] = useState(false);

  const mutate = useCallback(async () => {
    setIsPending(true);
    try {
      await signOut();
    } finally {
      setIsPending(false);
    }
  }, [signOut]);

  return {
    mutate,
    isPending,
  };
}

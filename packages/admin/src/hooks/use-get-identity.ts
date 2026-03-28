'use client';

import { useAuthUser } from '@/context/auth-context';

export interface UseGetIdentityResult<TData> {
  data?: TData;
  error?: Error;
  isError: boolean;
  isLoading: boolean;
}

export function useGetIdentity<TData = unknown>(): UseGetIdentityResult<TData> {
  const user = useAuthUser();

  return {
    data: user as TData | undefined,
    isLoading: false,
    isError: false,
    error: undefined,
  };
}

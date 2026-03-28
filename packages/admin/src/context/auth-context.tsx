'use client';

import { createContext, useCallback, useContext, useState } from 'react';
import type { AuthClient, AuthSession, AuthUser } from '@/types';

export interface AuthContextValue {
  isAuthenticated: boolean;
  isLoading: boolean;
  session: AuthSession | null;
  signOut: () => Promise<void>;
  user: AuthUser | null;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export interface AuthProviderProps {
  authClient: AuthClient;
  children: React.ReactNode;
  skipSessionFetch?: boolean;
}

export function AuthProvider({
  children,
  authClient,
  skipSessionFetch = false,
}: AuthProviderProps) {
  const [session, setSession] = useState<AuthSession | null>(null);
  const [isLoading, _setIsLoading] = useState(!skipSessionFetch);

  // const fetchSession = useCallback(async () => {
  //   try {
  //     const sessionData = await authClient.getSession();
  //     setSession(sessionData);
  //   } catch {
  //     setSession(null);
  //   } finally {
  //     setIsLoading(false);
  //   }
  // }, [authClient]);

  // useEffect(() => {
  //   if (skipSessionFetch) {
  //     setIsLoading(false);
  //     return;
  //   }

  //   // fetchSession();

  //   if (authClient.onSessionChange) {
  //     const unsubscribe = authClient.onSessionChange((newSession) => {
  //       setSession(newSession);
  //     });
  //     return unsubscribe;
  //   }
  // }, [ skipSessionFetch]);

  const signOut = useCallback(async () => {
    await authClient.signOut();
    setSession(null);
  }, [authClient]);

  const value: AuthContextValue = {
    session,
    user: session?.user ?? null,
    isLoading,
    isAuthenticated: !!session,
    signOut,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export function useAuthSession(): AuthSession | null {
  return useAuth().session;
}

export function useAuthUser(): AuthUser | null {
  return useAuth().user;
}

export function useIsAuthenticated(): boolean {
  return useAuth().isAuthenticated;
}

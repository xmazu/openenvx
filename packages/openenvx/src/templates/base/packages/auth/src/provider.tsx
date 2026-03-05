'use client';

import { createContext, type ReactNode, useContext } from 'react';

const AuthContext = createContext(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  return <AuthContext.Provider value={null}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}

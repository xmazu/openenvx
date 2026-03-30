import type React from 'react';

export type BaseKey = string | number;

export interface ResourceMeta {
  group?: boolean;
  icon?: string;
  label?: string;
  [key: string]: unknown;
}

export interface TreeMenuItem {
  children?: TreeMenuItem[];
  icon?: string;
  key?: string;
  label?: string;
  meta?: ResourceMeta;
  name: string;
  route?: string;
}

export interface BreadcrumbItem {
  href?: string;
  icon?: string;
  label: string;
}

export interface AutoSaveIndicatorProps {
  elements?: {
    idle?: React.ReactNode;
    loading?: React.ReactNode;
    success?: React.ReactNode;
    error?: React.ReactNode;
    pending?: React.ReactNode;
  };
  status: 'idle' | 'loading' | 'success' | 'error' | 'pending';
}

export interface AuthProvider {
  check?: (params: unknown) => Promise<unknown>;
  getIdentity?: <TData = unknown>() => Promise<TData>;
  getPermissions?: (params: unknown) => Promise<unknown>;
  login?: (params: unknown) => Promise<unknown>;
  logout?: (params: unknown) => Promise<unknown>;
  onError?: (error: unknown) => Promise<unknown>;
}

export interface AuthUser {
  email?: string;
  id: string;
  image?: string;
  name?: string;
  role?: string | string[];
  [key: string]: unknown;
}

export interface AuthSession {
  token?: string;
  user: AuthUser;
}

export interface AuthClient {
  getSession: () => Promise<AuthSession | null>;
  onSessionChange?: (
    callback: (session: AuthSession | null) => void
  ) => () => void;
  signOut: () => Promise<void>;
}

export interface AdminOptions {
  authProvider?: AuthProvider;
  title?: {
    text?: string;
    icon?: React.ReactNode;
  };
}

export interface UserFriendlyNameOptions {
  plural?: boolean;
  singular?: boolean;
}

// Allow string literals for backward compatibility
export type UserFriendlyNameOption =
  | 'plural'
  | 'singular'
  | UserFriendlyNameOptions;

export type TranslateFunction = (
  key: string,
  defaultMessage?: string,
  options?: Record<string, unknown>
) => string;

export interface LinkComponentProps {
  children: React.ReactNode;
  className?: string;
  href: string;
  onClick?: (e: React.MouseEvent) => void;
  replace?: boolean;
}

export interface ButtonHookResult {
  disabled?: boolean;
  hidden?: boolean;
  LinkComponent: React.ComponentType<LinkComponentProps>;
  label?: string;
  to?: string;
}

export interface RefreshButtonHookResult {
  label?: string;
  loading: boolean;
  onClick: () => void;
}

import type { AuthClient, AuthSession } from '../types';

interface BetterAuthUser {
  email: string;
  id: string;
  image?: string;
  name?: string;
  role?: string | string[];
}

interface BetterAuthSession {
  session: {
    token: string;
  };
  user: BetterAuthUser;
}

interface BetterAuthClient {
  getSession: () => Promise<{ data: BetterAuthSession | null }>;
  signOut: () => Promise<void>;
}

/**
 * Creates an AuthClient from a Better Auth client instance.
 *
 * @param betterAuthClient - The Better Auth client from `createAuthClient()`
 * @returns AuthClient compatible with @openenvx/admin
 *
 * @example
 * ```typescript
 * import { createAdminAuthClient } from '@openenvx/admin/client';
 * import { createAuthClient } from 'better-auth/react';
 *
 * const betterAuth = createAuthClient({ baseURL: '...' });
 * export const authClient = createAdminAuthClient(betterAuth);
 * ```
 */
export function createAdminAuthClient(
  betterAuthClient: BetterAuthClient
): AuthClient {
  const mapSession = (data: BetterAuthSession | null): AuthSession | null => {
    if (!data) {
      return null;
    }
    return {
      user: {
        id: data.user.id,
        email: data.user.email,
        name: data.user.name,
        image: data.user.image,
        role: data.user.role,
      },
      token: data.session.token,
    };
  };

  return {
    getSession: async () => {
      const { data } = await betterAuthClient.getSession();
      return mapSession(data);
    },

    signOut: () => betterAuthClient.signOut(),

    onSessionChange: (callback) => {
      const interval = setInterval(async () => {
        const { data } = await betterAuthClient.getSession();
        callback(mapSession(data));
      }, 5000);

      return () => clearInterval(interval);
    },
  };
}

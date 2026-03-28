'use client';

import type { AuthClient, AuthSession } from '@openenvx/admin/client';
import { createAuthClient } from 'better-auth/react';

const betterAuthClient = createAuthClient({
  baseURL: process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000',
});

export const authClient: AuthClient = {
  getSession: async () => {
    const { data } = await betterAuthClient.getSession();
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
    } as AuthSession;
  },

  signOut: () => betterAuthClient.signOut(),

  onSessionChange: (callback) => {
    const interval = setInterval(async () => {
      const { data } = await betterAuthClient.getSession();
      if (!data) {
        callback(null);
        return;
      }

      callback({
        user: {
          id: data.user.id,
          email: data.user.email,
          name: data.user.name,
          image: data.user.image,
          role: data.user.role,
        },
        token: data.session.token,
      } as AuthSession);
    }, 5000);

    return () => clearInterval(interval);
  },
};

export { betterAuthClient };

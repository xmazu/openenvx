import { createAdminAuthClient } from '@openenvx/admin/client';
import { createAuthClient } from 'better-auth/react';

const betterAuthClient = createAuthClient({
  baseURL: process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000',
});

export const authClient = createAdminAuthClient(betterAuthClient);
export { betterAuthClient };

import {
  createAdmin,
  createBetterAuthTokenExtractor,
} from '@openenvx/admin/server';
import type { NextRequest } from 'next/server';

const extractToken = createBetterAuthTokenExtractor({
  betterAuthSecret: process.env.BETTER_AUTH_SECRET!,
  jwtSecret: process.env.ADMIN_JWT_SECRET!,
  dbRole: process.env.ADMIN_DB_ROLE || 'admin_service',
  requiredRole: process.env.ADMIN_REQUIRED_ROLE || 'super_admin',
  tokenExpirySeconds: Number(process.env.ADMIN_TOKEN_EXPIRY) || 300,
});

const admin = createAdmin({
  postgrestUrl: process.env.POSTGREST_URL || 'http://localhost:3001',
  getToken: (req: NextRequest) => {
    // Auth middleware stores token in header to avoid double extraction
    const tokenFromHeader = req.headers.get('x-admin-token');
    if (tokenFromHeader) {
      return tokenFromHeader;
    }
    // Fallback to extraction
    return extractToken(req);
  },
  auth: {
    getToken: extractToken,
    tokenHeader: 'x-admin-token',
  },
  resources: {},
});

export const { GET, POST, PUT, PATCH, DELETE } = admin.handler;

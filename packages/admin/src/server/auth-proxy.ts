import { jwtVerify, SignJWT } from 'jose';
import type { NextRequest } from 'next/server';

export interface BetterAuthUser {
  id: string;
  email: string;
  name?: string;
  role?: string | string[];
  image?: string;
  [key: string]: unknown;
}

export interface BetterAuthTokenExtractorConfig {
  betterAuthSecret: string;
  jwtSecret: string;
  dbRole?: string;
  requiredRole?: string;
  tokenExpirySeconds?: number;
}

const DEFAULT_TOKEN_EXPIRY = 300;
const SESSION_COOKIE_PATTERN = /better-auth.session=([^;]+)/;

function normalizeRoles(role: string | string[] | undefined): string[] {
  if (!role) {
    return [];
  }
  if (Array.isArray(role)) {
    return role;
  }
  return role.split(',').map((r) => r.trim());
}

function hasRole(userRoles: string[], requiredRole: string): boolean {
  return userRoles.includes(requiredRole);
}

export function createBetterAuthTokenExtractor(config: BetterAuthTokenExtractorConfig) {
  const {
    betterAuthSecret,
    jwtSecret,
    dbRole = 'admin_service',
    requiredRole = 'super_admin',
    tokenExpirySeconds = DEFAULT_TOKEN_EXPIRY,
  } = config;

  const betterAuthKey = new TextEncoder().encode(betterAuthSecret);
  const adminKey = new TextEncoder().encode(jwtSecret);

  return async function getToken(request: NextRequest): Promise<string | null> {
    const cookie = request.headers.get('cookie') || '';
    const sessionMatch = cookie.match(SESSION_COOKIE_PATTERN);
    
    if (!sessionMatch) {
      return null;
    }

    try {
      const token = decodeURIComponent(sessionMatch[1]);
      const { payload } = await jwtVerify(token, betterAuthKey, {
        algorithms: ['HS256'],
      });

      if (!payload.user) {
        return null;
      }

      const user = payload.user as BetterAuthUser;
      const userRoles = normalizeRoles(user.role);

      if (!hasRole(userRoles, requiredRole)) {
        return null;
      }

      const adminToken = await new SignJWT({
        role: dbRole,
        sub: user.id,
        email: user.email,
        name: user.name,
        originalRole: user.role || requiredRole,
      })
        .setProtectedHeader({ alg: 'HS256' })
        .setIssuedAt()
        .setExpirationTime(`${tokenExpirySeconds}s`)
        .sign(adminKey);

      return adminToken;
    } catch {
      return null;
    }
  };
}

import { SignJWT } from 'jose';
import type { NextRequest } from 'next/server';

export interface BetterAuthUser {
  email: string;
  id: string;
  image?: string;
  name?: string;
  role?: string | string[];
  [key: string]: unknown;
}

export interface BetterAuthSession {
  session: {
    id: string;
    createdAt: string;
    updatedAt: string;
    userId: string;
    expiresAt: string;
    token: string;
    ipAddress?: string;
    userAgent?: string;
  };
  user: BetterAuthUser;
}

export interface BetterAuthTokenExtractorConfig {
  /** URL to Better Auth API (e.g., 'http://localhost:3000') */
  betterAuthUrl: string;
  cookieName?: string;
  dbRole?: string;
  jwtSecret: string;
  requiredRole?: string;
  tokenExpirySeconds?: number;
}

const DEFAULT_TOKEN_EXPIRY = 300;

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

/**
 * Creates a token extractor that validates Better Auth session via API call
 * and generates a PostgREST JWT token with the user's role.
 *
 * Better Auth uses cookie-based sessions (not JWT), so we validate by calling
 * the Better Auth API /api/auth/session endpoint server-side.
 */
export function createBetterAuthTokenExtractor(
  config: BetterAuthTokenExtractorConfig
) {
  const {
    betterAuthUrl,
    cookieName = 'better-auth.session_token',
    jwtSecret,
    dbRole = 'admin_service',
    requiredRole = 'super_admin',
    tokenExpirySeconds = DEFAULT_TOKEN_EXPIRY,
  } = config;

  const adminKey = new TextEncoder().encode(jwtSecret);
  const cookiePattern = new RegExp(`${cookieName}=([^;]+)`);

  return async function getToken(request: NextRequest): Promise<string | null> {
    const cookie = request.headers.get('cookie') || '';
    const sessionMatch = cookie.match(cookiePattern);

    if (!sessionMatch) {
      console.error(`[Admin Auth] No ${cookieName} cookie found`);
      return null;
    }

    try {
      console.error('[Admin Auth] Validating session with Better Auth API...');

      // Call Better Auth API to validate session server-side
      // Better Auth uses cookie-based sessions, not JWT
      const sessionResponse = await fetch(
        `${betterAuthUrl}/api/auth/get-session`,
        {
          headers: {
            cookie: request.headers.get('cookie') || '',
          },
        }
      );

      if (!sessionResponse.ok) {
        console.error(
          '[Admin Auth] Session validation failed:',
          sessionResponse.status
        );
        return null;
      }

      const sessionData = (await sessionResponse.json()) as BetterAuthSession;

      if (!sessionData.user) {
        console.error('[Admin Auth] No user in session response');
        return null;
      }

      const user = sessionData.user;
      console.error(
        '[Admin Auth] User authenticated:',
        user.email,
        'Roles:',
        user.role
      );

      const userRoles = normalizeRoles(user.role);

      if (!hasRole(userRoles, requiredRole)) {
        console.error(
          `[Admin Auth] User missing required role: ${requiredRole}`
        );
        return null;
      }

      console.error('[Admin Auth] Creating PostgREST token with role:', dbRole);

      // Create PostgREST JWT token (this is our own JWT for PostgREST, not Better Auth)
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

      console.error('[Admin Auth] PostgREST token created successfully');
      return adminToken;
    } catch (error) {
      console.error('[Admin Auth] Token extraction failed:', error);
      return null;
    }
  };
}

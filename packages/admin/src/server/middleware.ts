import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';

/**
 * Configuration for the admin authentication middleware
 */
export interface AuthMiddlewareConfig {
  /** Function to check if user is authenticated. Return null/false if not authenticated. */
  isAuthenticated: (request: NextRequest) => Promise<boolean> | boolean;
  /** Path to redirect unauthenticated users to (default: '/auth/login') */
  loginPath?: string;

  /** Optional callback when auth check fails. Can be used for logging. */
  onAuthFailure?: (request: NextRequest) => void;

  /** Routes that should be accessible without authentication */
  publicRoutes?: string[];
}

/**
 * Default public routes that should always be accessible
 * Note: API routes are NOT handled by middleware - they should protect themselves
 */
const DEFAULT_PUBLIC_ROUTES = [
  '/_next',
  '/api', // Skip all API routes - they handle their own auth
  '/static',
  '/favicon.ico',
  '/login',
  '/auth',
];

/**
 * Creates a Next.js middleware that enforces authentication for admin routes.
 *
 * @param config - Configuration for auth checking and redirects
 * @returns Next.js middleware function
 *
 * @example
 * ```typescript
 * // middleware.ts
 * import { createAuthMiddleware } from '@openenvx/admin/server';
 * import { createBetterAuthChecker } from './auth';
 *
 * export const middleware = createAuthMiddleware({
 *   loginPath: '/auth/login',
 *   isAuthenticated: createBetterAuthChecker({ secret: process.env.BETTER_AUTH_SECRET! }),
 * });
 *
 * export const config = {
 *   matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
 * };
 * ```
 */
export function createAuthMiddleware(config: AuthMiddlewareConfig) {
  const {
    loginPath = '/auth/login',
    publicRoutes = [],
    isAuthenticated,
    onAuthFailure,
  } = config;

  const allPublicRoutes = [...DEFAULT_PUBLIC_ROUTES, ...publicRoutes];

  return async function middleware(request: NextRequest) {
    const { pathname } = request.nextUrl;

    // Skip auth check for public routes
    if (allPublicRoutes.some((route) => pathname.startsWith(route))) {
      return NextResponse.next();
    }

    // Check if user is authenticated
    const authenticated = await isAuthenticated(request);

    if (!authenticated) {
      onAuthFailure?.(request);

      // Redirect to login page, preserving the original URL
      const loginUrl = new URL(loginPath, request.url);
      loginUrl.searchParams.set('redirect', pathname);
      return NextResponse.redirect(loginUrl);
    }

    return NextResponse.next();
  };
}

/**
 * Helper to check if a Better Auth session cookie exists and is valid.
 * This is a lightweight check suitable for middleware (doesn't verify JWT signature).
 *
 * @param cookieName - Name of the session cookie (default: 'better-auth.session')
 * @returns Middleware-compatible auth checker
 *
 * @example
 * ```typescript
 * createAuthMiddleware({
 *   isAuthenticated: createBetterAuthChecker('better-auth.session'),
 * });
 * ```
 */
export function createBetterAuthChecker(cookieName = 'better-auth.session') {
  return function isAuthenticated(request: NextRequest): boolean {
    const sessionCookie = request.cookies.get(cookieName);
    return !!sessionCookie?.value;
  };
}

/**
 * Creates an auth checker that verifies the session with the full JWT validation.
 * Use this when you need role-based access control in middleware.
 *
 * Note: This uses jose for JWT verification and is async.
 *
 * @param config - Configuration for JWT verification
 * @returns Async auth checker function
 */
export interface JWTAuthCheckerConfig {
  cookieName?: string;
  requiredRole?: string;
  secret: string;
}

export function createJWTAuthChecker(config: JWTAuthCheckerConfig) {
  const { secret, cookieName = 'better-auth.session', requiredRole } = config;

  // Dynamic import to avoid loading jose in middleware unnecessarily
  return async function isAuthenticated(
    request: NextRequest
  ): Promise<boolean> {
    const sessionCookie = request.cookies.get(cookieName);
    if (!sessionCookie?.value) {
      return false;
    }

    try {
      const { jwtVerify } = await import('jose');
      const secretKey = new TextEncoder().encode(secret);
      const token = decodeURIComponent(sessionCookie.value);

      const { payload } = await jwtVerify(token, secretKey, {
        algorithms: ['HS256'],
      });

      // If a specific role is required, check for it
      if (requiredRole) {
        const user = payload.user as { role?: string | string[] } | undefined;
        const roles = user?.role;
        const userRoles = Array.isArray(roles)
          ? roles
          : roles?.split(',').map((r) => r.trim()) || [];
        return userRoles.includes(requiredRole);
      }

      return true;
    } catch {
      return false;
    }
  };
}

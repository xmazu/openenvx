import type { NextRequest, NextResponse } from 'next/server';

export interface AuthMiddlewareConfig {
  betterAuthUrl: string;
  cookieName?: string;
  loginPath?: string;
  onAuthFailure?: (request: NextRequest) => void;
  publicRoutes?: string[];
  requiredRole?: string;
}

const DEFAULT_PUBLIC_ROUTES = [
  '/_next',
  '/api',
  '/static',
  '/favicon.ico',
  '/login',
  '/auth',
];

interface BetterAuthSession {
  session: {
    id: string;
    userId: string;
    expiresAt: string;
    token: string;
  } | null;
  user: {
    id: string;
    email: string;
    name?: string;
    image?: string;
    role?: string | string[];
  } | null;
}

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
 * Creates a Next.js middleware that enforces authentication for admin routes.
 * Validates session by calling Better Auth API (/api/auth/session).
 * Automatically skips /api/admin/* routes (those are validated by the proxy).
 */
export function createAuthMiddleware(config: AuthMiddlewareConfig) {
  const {
    betterAuthUrl,
    cookieName = 'better-auth.session_token',
    loginPath = '/auth/login',
    publicRoutes = [],
    requiredRole,
    onAuthFailure,
  } = config;

  const allPublicRoutes = [...DEFAULT_PUBLIC_ROUTES, ...publicRoutes];

  return async function middleware(
    request: NextRequest
  ): Promise<NextResponse | Response> {
    const { pathname } = request.nextUrl;
    const { NextResponse } = await import('next/server');

    if (allPublicRoutes.some((route) => pathname.startsWith(route))) {
      return NextResponse.next();
    }

    const cookie = request.headers.get('cookie') || '';
    const cookiePattern = new RegExp(`${cookieName}=([^;]+)`);
    const sessionMatch = cookie.match(cookiePattern);

    if (!sessionMatch) {
      onAuthFailure?.(request);
      const loginUrl = new URL(loginPath, request.url);
      loginUrl.searchParams.set('redirect', pathname);
      return NextResponse.redirect(loginUrl);
    }

    try {
      const sessionResponse = await fetch(
        `${betterAuthUrl}/api/auth/get-session`,
        {
          headers: { cookie },
        }
      );

      if (!sessionResponse.ok) {
        onAuthFailure?.(request);
        const loginUrl = new URL(loginPath, request.url);
        loginUrl.searchParams.set('redirect', pathname);
        return NextResponse.redirect(loginUrl);
      }

      const sessionData = (await sessionResponse.json()) as BetterAuthSession;

      if (!(sessionData.session && sessionData.user)) {
        onAuthFailure?.(request);
        const loginUrl = new URL(loginPath, request.url);
        loginUrl.searchParams.set('redirect', pathname);
        return NextResponse.redirect(loginUrl);
      }

      if (requiredRole) {
        const userRoles = normalizeRoles(sessionData.user.role);
        if (!hasRole(userRoles, requiredRole)) {
          onAuthFailure?.(request);
          return NextResponse.redirect(new URL('/unauthorized', request.url));
        }
      }

      return NextResponse.next();
    } catch {
      onAuthFailure?.(request);
      const loginUrl = new URL(loginPath, request.url);
      loginUrl.searchParams.set('redirect', pathname);
      return NextResponse.redirect(loginUrl);
    }
  };
}

/**
 * Creates a lightweight auth checker that only verifies cookie presence.
 * Use with composeMiddleware when you need custom auth logic.
 */
export function createBetterAuthChecker(
  cookieName = 'better-auth.session_token'
) {
  return function isAuthenticated(request: NextRequest): boolean {
    const sessionCookie = request.cookies.get(cookieName);
    return !!sessionCookie?.value;
  };
}

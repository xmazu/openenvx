import { NextRequest, NextResponse } from 'next/server';

/**
 * Configuration for the API auth wrapper
 */
export interface WithAuthConfig {
  /**
   * Function to check if request is authenticated.
   * Should return the token if authenticated, null/undefined if not.
   * The token will be passed to the request headers for downstream use.
   */
  getToken: (request: NextRequest) => Promise<string | null> | string | null;

  /** Optional custom error response when authentication fails */
  onAuthFailure?: (request: NextRequest) => NextResponse;

  /** Header name to store the extracted token (default: 'x-admin-token') */
  tokenHeader?: string;
}

type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

type RouteHandler = (
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) => Promise<NextResponse> | NextResponse;

type AuthRouteHandler = (
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) => Promise<NextResponse>;

const DEFAULT_AUTH_FAILURE = (): NextResponse => {
  return NextResponse.json(
    { error: 'Unauthorized', message: 'Authentication required' },
    { status: 401 }
  );
};

/**
 * Wraps API route handlers to enforce authentication.
 * Extracts the token via getToken and stores it in a header for downstream use.
 *
 * @param handlers - Object with HTTP method handlers (GET, POST, etc.)
 * @param config - Auth configuration with token extractor
 * @returns Wrapped handlers that check auth before executing
 *
 * @example
 * ```typescript
 * // app/api/admin/[...path]/route.ts
 * import { createAdmin, withAuth, createBetterAuthTokenExtractor } from '@openenvx/admin/server';
 *
 * const getToken = createBetterAuthTokenExtractor({
 *   betterAuthSecret: process.env.BETTER_AUTH_SECRET!,
 *   jwtSecret: process.env.ADMIN_JWT_SECRET!,
 * });
 *
 * const admin = createAdmin({
 *   postgrestUrl: process.env.POSTGREST_URL!,
 *   getToken: (req) => req.headers.get('x-admin-token'), // Read from header
 * });
 *
 * export const { GET, POST, PUT, PATCH, DELETE } = withAuth(admin.handler, {
 *   getToken,
 *   tokenHeader: 'x-admin-token',
 * });
 * ```
 */
export function withAuth(
  handlers: Record<HTTPMethod, RouteHandler>,
  config: WithAuthConfig
): Record<HTTPMethod, AuthRouteHandler> {
  const { getToken, onAuthFailure, tokenHeader = 'x-admin-token' } = config;
  const handleAuthFailure = onAuthFailure ?? DEFAULT_AUTH_FAILURE;

  const wrapHandler = (handler: RouteHandler): AuthRouteHandler => {
    return async function authHandler(
      request: NextRequest,
      context: { params: Promise<{ path: string[] }> }
    ): Promise<NextResponse> {
      const token = await getToken(request);

      if (!token) {
        return handleAuthFailure(request);
      }

      // Clone request with token in header to avoid re-extraction
      const headers = new Headers(request.headers);
      headers.set(tokenHeader, token);

      const requestWithToken = new NextRequest(request.url, {
        method: request.method,
        headers,
        body: request.body,
      });

      return handler(requestWithToken, context);
    };
  };

  return {
    GET: wrapHandler(handlers.GET),
    POST: wrapHandler(handlers.POST),
    PUT: wrapHandler(handlers.PUT),
    PATCH: wrapHandler(handlers.PATCH),
    DELETE: wrapHandler(handlers.DELETE),
  };
}

/**
 * Creates a simple auth wrapper for a single handler function.
 *
 * @param handler - The route handler to wrap
 * @param config - Auth configuration
 * @returns Wrapped handler with auth check
 */
export function withAuthHandler(
  handler: (request: NextRequest) => Promise<NextResponse> | NextResponse,
  config: Pick<WithAuthConfig, 'getToken' | 'onAuthFailure'>
): AuthRouteHandler {
  const { getToken, onAuthFailure } = config;
  const handleAuthFailure = onAuthFailure ?? DEFAULT_AUTH_FAILURE;

  return async function authHandler(
    request: NextRequest,
    _context: { params: Promise<{ path: string[] }> }
  ): Promise<NextResponse> {
    const token = await getToken(request);

    if (!token) {
      return handleAuthFailure(request);
    }

    return handler(request);
  };
}

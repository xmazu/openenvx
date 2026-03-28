import {
  createAuthMiddleware,
  createBetterAuthChecker,
} from '@openenvx/admin/server';

export const middleware = createAuthMiddleware({
  loginPath: process.env.AUTH_LOGIN_PATH || '/auth/login',
  isAuthenticated: createBetterAuthChecker('better-auth.session'),
  publicRoutes: [
    '/login',
    '/auth/login',
    '/auth/register',
    '/auth/forgot-password',
  ],
});

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};

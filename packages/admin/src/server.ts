export { type AdminConfig, createAdmin } from './server/admin';
export {
  type BetterAuthTokenExtractorConfig,
  type BetterAuthUser,
  createBetterAuthTokenExtractor,
} from './server/auth-proxy';
export type {
  ColumnInfo,
  ForeignKeyInfo,
  TableSchema,
} from './server/introspection';
export {
  type AuthMiddlewareConfig,
  createAuthMiddleware,
  createBetterAuthChecker,
  createJWTAuthChecker,
  type JWTAuthCheckerConfig,
} from './server/middleware';
export { getResources } from './server/resources';
export {
  createPostgRESTProxy,
  type PostgRESTProxyConfig,
} from './server/router';
export {
  type WithAuthConfig,
  withAuth,
  withAuthHandler,
} from './server/with-auth';

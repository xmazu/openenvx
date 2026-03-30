export { defineResources } from '../lib/define-resources';
export {
  type Admin,
  type AdminAuthConfig,
  type AdminConfig,
  createAdmin,
} from '../server/admin';
export {
  type BetterAuthTokenExtractorConfig,
  type BetterAuthUser,
  createBetterAuthTokenExtractor,
} from '../server/auth-proxy';
export type {
  ColumnInfo,
  ForeignKeyInfo,
  TableSchema,
} from '../server/introspection';
export {
  type AuthMiddlewareConfig,
  createAuthMiddleware,
  createBetterAuthChecker,
} from '../server/middleware';
export {
  composeMiddleware,
  createConditionalMiddleware,
  createPathExcludingMiddleware,
  type Middleware,
  type MiddlewareFunction,
  type MiddlewareNextFunction,
} from '../server/middleware-compose';
export {
  createPostgRESTProxy,
  type PostgRESTProxyConfig,
} from '../server/router';
export type {
  ResourceConfig,
  ResourceItem,
  ResourcesConfig,
} from '../types/resources';

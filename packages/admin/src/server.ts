export { type AdminConfig, createAdmin } from './server/admin';
export type {
  ColumnInfo,
  ForeignKeyInfo,
  TableSchema,
} from './server/introspection';
export { getResources } from './server/resources';
export {
  createPostgRESTProxy,
  type PostgRESTProxyConfig,
} from './server/router';

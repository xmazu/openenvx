import { type NextRequest, NextResponse } from 'next/server';
import { createMergedConfig, resourceRegistry } from '@/lib/define-resource';
import { autoGenerateField } from '@/lib/resource-types';
import {
  fetchColumns,
  fetchReferenceData,
  fetchTableSchema,
  fetchTables,
  type TableSchema,
} from './introspection';

const LABEL_REGEX_UNDERSCORE = /_/g;
const LABEL_REGEX_CAMEL = /([A-Z])/g;
const LABEL_REGEX_LEADING_SPACE = /^\s+/;
const LABEL_REGEX_EXTRA_SPACES = /\s+/g;

export interface PostgRESTProxyConfig {
  getToken?: (request: NextRequest) => Promise<string | null> | string | null;
  postgrestUrl: string;
  transformRequest?: (
    request: NextRequest
  ) => Promise<NextRequest> | NextRequest;
}

export interface RouteContext {
  params: Promise<{ path: string[] }>;
}

function formatLabel(name: string): string {
  return name
    .replace(LABEL_REGEX_UNDERSCORE, ' ')
    .replace(LABEL_REGEX_CAMEL, ' $1')
    .replace(LABEL_REGEX_LEADING_SPACE, '')
    .replace(LABEL_REGEX_EXTRA_SPACES, ' ')
    .toLowerCase()
    .replace(/\b\w/g, (l) => l.toUpperCase());
}

function buildConfigFromSchema(schema: TableSchema) {
  const fields = schema.columns.map((col) =>
    autoGenerateField(col, schema.foreignKeys)
  );

  return {
    label: formatLabel(schema.name),
    fields,
    canCreate: true,
    canEdit: true,
    canDelete: true,
    canShow: true,
    list: {
      columns: fields
        .filter((f) => !['id', 'created_at', 'updated_at'].includes(f.name))
        .slice(0, 5)
        .map((f) => f.name),
      searchable: fields
        .filter((f) => ['text', 'textarea', 'email'].includes(f.type))
        .slice(0, 3)
        .map((f) => f.name),
    },
    form: {
      layout: 'vertical' as const,
      columns: 1,
    },
  };
}

export function createPostgRESTProxy(config: PostgRESTProxyConfig) {
  const { postgrestUrl, getToken, transformRequest } = config;

  async function handleResourceConfig(
    _request: NextRequest,
    path: string[]
  ): Promise<NextResponse | null> {
    if (path[0] !== 'resources' || path[2] !== 'config') {
      return null;
    }

    const tableName = path[1];

    try {
      const schema = await fetchTableSchema(tableName);
      const autoConfig = buildConfigFromSchema(schema);

      const manualConfig = resourceRegistry.get(tableName)?.config;
      const mergedConfig = createMergedConfig(manualConfig, autoConfig);

      return NextResponse.json({ config: mergedConfig });
    } catch (error) {
      return NextResponse.json(
        { error: 'Failed to build resource config', message: String(error) },
        { status: 500 }
      );
    }
  }

  async function handleIntrospection(
    request: NextRequest,
    path: string[]
  ): Promise<NextResponse | null> {
    if (path[0] !== 'introspection') {
      return null;
    }

    const subPath = path.slice(1);

    try {
      if (subPath[0] === 'tables' && request.method === 'GET') {
        const tables = await fetchTables();
        return NextResponse.json(tables.map((name) => ({ table_name: name })));
      }

      if (subPath[0] === 'columns' && subPath[1] && request.method === 'GET') {
        const tableName = subPath[1];
        const columns = await fetchColumns(tableName);
        return NextResponse.json(
          columns.map((col) => ({
            column_name: col.name,
            data_type: col.dataType,
            is_nullable: col.isNullable ? 'YES' : 'NO',
            column_default: col.defaultValue,
          }))
        );
      }

      if (subPath[0] === 'schema' && subPath[1] && request.method === 'GET') {
        const tableName = subPath[1];
        const schema = await fetchTableSchema(tableName);
        return NextResponse.json(schema);
      }

      return NextResponse.json(
        { error: 'Invalid introspection endpoint' },
        { status: 404 }
      );
    } catch (error) {
      return NextResponse.json(
        { error: 'Introspection error', message: String(error) },
        { status: 500 }
      );
    }
  }

  async function handleRelationships(
    request: NextRequest,
    path: string[]
  ): Promise<NextResponse | null> {
    if (path[0] !== 'relationships') {
      return null;
    }

    const tableName = path[1];
    if (!tableName) {
      return NextResponse.json(
        { error: 'Table name required' },
        { status: 400 }
      );
    }

    try {
      const url = new URL(request.url);
      const search = url.searchParams.get('search') || undefined;
      const limit = Number(url.searchParams.get('limit')) || 50;

      const data = await fetchReferenceData(tableName, search, limit);
      return NextResponse.json(data);
    } catch (error) {
      return NextResponse.json(
        { error: 'Failed to fetch reference data', message: String(error) },
        { status: 500 }
      );
    }
  }

  async function proxyRequest(
    request: NextRequest,
    context: RouteContext
  ): Promise<NextResponse> {
    const params = await context.params;
    const path = params.path || [];

    const resourceConfigResponse = await handleResourceConfig(request, path);
    if (resourceConfigResponse) {
      return resourceConfigResponse;
    }

    const introspectionResponse = await handleIntrospection(request, path);
    if (introspectionResponse) {
      return introspectionResponse;
    }

    const relationshipsResponse = await handleRelationships(request, path);
    if (relationshipsResponse) {
      return relationshipsResponse;
    }

    const pathStr = path.join('/');
    const url = new URL(request.url);
    const targetUrl = new URL(pathStr + url.search, postgrestUrl);

    const headers = new Headers(request.headers);
    headers.delete('host');
    headers.set('host', new URL(postgrestUrl).host);

    const token = await getToken?.(request);

    if (!token) {
      return NextResponse.json(
        { error: 'Unauthorized', message: 'Authentication required' },
        { status: 401 }
      );
    }
    headers.set('authorization', `Bearer ${token}`);

    if (transformRequest) {
      await transformRequest(request);
    }

    try {
      console.error('[PostgREST Proxy] Forwarding request:', {
        method: request.method,
        url: targetUrl.toString(),
        hasToken: !!token,
      });

      const response = await fetch(targetUrl.toString(), {
        method: request.method,
        headers,
        body: ['GET', 'HEAD'].includes(request.method)
          ? null
          : await request.text(),
      });

      console.error('[PostgREST Proxy] Response:', {
        status: response.status,
        statusText: response.statusText,
      });

      if (!response.ok) {
        const errorBody = await response.text();
        console.error('[PostgREST Proxy] Error response body:', errorBody);
      }

      const responseHeaders = new Headers(response.headers);
      responseHeaders.delete('content-encoding');

      return new NextResponse(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders,
      });
    } catch (error) {
      console.error('[PostgREST Proxy] Fetch error:', error);
      return NextResponse.json(
        { error: 'Proxy error', message: String(error) },
        { status: 502 }
      );
    }
  }

  return {
    GET: proxyRequest,
    POST: proxyRequest,
    PUT: proxyRequest,
    PATCH: proxyRequest,
    DELETE: proxyRequest,
  };
}

import { type NextRequest, NextResponse } from 'next/server';
import type { ResourceItem } from '@/types/resources';
import { fetchReferenceData } from './introspection';

export interface PostgRESTProxyConfig {
  getToken?: (request: NextRequest) => Promise<string | null> | string | null;
  postgrestUrl: string;
  resources?: ResourceItem[];
  transformRequest?: (
    request: NextRequest
  ) => Promise<NextRequest> | NextRequest;
}

export interface RouteContext {
  params: Promise<{ path: string[] }>;
}

export function createPostgRESTProxy(config: PostgRESTProxyConfig) {
  const { postgrestUrl, getToken, resources = [], transformRequest } = config;

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
      const sourceResource = url.searchParams.get('sourceResource');
      const sourceField = url.searchParams.get('sourceField');

      let displayField = 'name';

      if (sourceResource && sourceField) {
        const sourceRes = resources.find((r) => r.name === sourceResource);
        if (sourceRes?.config?.fields) {
          const field = sourceRes.config.fields.find(
            (f) => f.name === sourceField && f.type === 'reference'
          );
          if (field && 'reference' in field && field.reference?.displayField) {
            displayField = field.reference.displayField;
          }
        }
      }

      const data = await fetchReferenceData(
        tableName,
        search,
        limit,
        displayField
      );
      return NextResponse.json({ data, displayField });
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
      const response = await fetch(targetUrl.toString(), {
        method: request.method,
        headers,
        body: ['GET', 'HEAD'].includes(request.method)
          ? null
          : await request.text(),
      });

      const responseHeaders = new Headers(response.headers);
      responseHeaders.delete('content-encoding');

      return new NextResponse(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders,
      });
    } catch (error) {
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

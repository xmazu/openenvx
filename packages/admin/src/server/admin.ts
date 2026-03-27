import { NextRequest } from 'next/server';
import type {
  HookContext,
  ListParams,
  ResourceHooks,
} from '@/lib/resource-types';
import { createPostgRESTProxy, type PostgRESTProxyConfig } from './router';

export interface ResourceConfig {
  hooks?: ResourceHooks;
}

export interface AdminConfig extends PostgRESTProxyConfig {
  resources?: Record<string, ResourceConfig>;
}

interface HookRegistry {
  get: (resourceName: string) => ResourceHooks | undefined;
  has: (resourceName: string) => boolean;
}

function createHookRegistry(
  resources: Record<string, ResourceConfig>
): HookRegistry {
  const registry = new Map<string, ResourceHooks>();

  for (const [name, config] of Object.entries(resources)) {
    if (config.hooks) {
      registry.set(name, config.hooks);
    }
  }

  return {
    get: (name: string) => registry.get(name),
    has: (name: string) => registry.has(name),
  };
}

export function createAdmin(config: AdminConfig) {
  const { resources = {}, ...proxyConfig } = config;
  const hookRegistry = createHookRegistry(resources);

  async function executeBeforeCreate(
    resourceName: string,
    data: Record<string, unknown>,
    context: HookContext
  ): Promise<Record<string, unknown>> {
    const hooks = hookRegistry.get(resourceName);
    if (!hooks?.beforeCreate) {
      return data;
    }
    return (await hooks.beforeCreate(data, context)) ?? data;
  }

  async function executeBeforeUpdate(
    resourceName: string,
    data: Record<string, unknown>,
    id: string | number,
    context: HookContext
  ): Promise<Record<string, unknown>> {
    const hooks = hookRegistry.get(resourceName);
    if (!hooks?.beforeUpdate) {
      return data;
    }
    return (await hooks.beforeUpdate(data, id, context)) ?? data;
  }

  async function executeBeforeDelete(
    resourceName: string,
    id: string | number,
    context: HookContext
  ): Promise<boolean> {
    const hooks = hookRegistry.get(resourceName);
    if (!hooks?.beforeDelete) {
      return true;
    }
    return await hooks.beforeDelete(id, context);
  }

  async function executeBeforeList(
    resourceName: string,
    params: ListParams,
    context: HookContext
  ): Promise<ListParams> {
    const hooks = hookRegistry.get(resourceName);
    if (!hooks?.beforeList) {
      return params;
    }
    return (await hooks.beforeList(params, context)) ?? params;
  }

  async function executeAfterList(
    resourceName: string,
    data: unknown[],
    context: HookContext
  ): Promise<unknown[]> {
    const hooks = hookRegistry.get(resourceName);
    if (!hooks?.afterList) {
      return data;
    }
    return (await hooks.afterList(data, context)) ?? data;
  }

  async function executeAfterCreate(
    resourceName: string,
    data: Record<string, unknown>,
    context: HookContext
  ): Promise<void> {
    const hooks = hookRegistry.get(resourceName);
    if (hooks?.afterCreate) {
      await hooks.afterCreate(data, context);
    }
  }

  async function executeAfterUpdate(
    resourceName: string,
    data: Record<string, unknown>,
    id: string | number,
    context: HookContext
  ): Promise<void> {
    const hooks = hookRegistry.get(resourceName);
    if (hooks?.afterUpdate) {
      await hooks.afterUpdate(data, id, context);
    }
  }

  async function executeAfterDelete(
    resourceName: string,
    id: string | number,
    context: HookContext
  ): Promise<void> {
    const hooks = hookRegistry.get(resourceName);
    if (hooks?.afterDelete) {
      await hooks.afterDelete(id, context);
    }
  }

  const proxy = createPostgRESTProxy(proxyConfig);

  const handler = {
    GET: async (
      request: NextRequest,
      context: { params: Promise<{ path: string[] }> }
    ) => {
      const params = await context.params;
      const path = params.path || [];
      const resourceName = path[0];

      if (resourceName && hookRegistry.has(resourceName) && !path[1]) {
        const url = new URL(request.url);
        const listParams: ListParams = {
          pagination: {
            current: Number(url.searchParams.get('page')) || 1,
            pageSize: Number(url.searchParams.get('pageSize')) || 10,
          },
        };

        const hookContext = createHookContext(request);
        const modifiedParams = await executeBeforeList(
          resourceName,
          listParams,
          hookContext
        );

        const modifiedUrl = new URL(request.url);
        if (modifiedParams.pagination) {
          modifiedUrl.searchParams.set(
            'page',
            String(modifiedParams.pagination.current || 1)
          );
          modifiedUrl.searchParams.set(
            'pageSize',
            String(modifiedParams.pagination.pageSize || 10)
          );
        }

        const modifiedRequest = new NextRequest(modifiedUrl.toString(), {
          method: request.method,
          headers: request.headers,
        });

        const result = await proxy.GET(modifiedRequest, context);
        const body = await result.json();

        if (body.data) {
          body.data = await executeAfterList(
            resourceName,
            body.data,
            hookContext
          );
        }

        return new Response(JSON.stringify(body), {
          status: result.status,
          headers: result.headers,
        });
      }

      return proxy.GET(request, context);
    },

    POST: async (
      request: NextRequest,
      context: { params: Promise<{ path: string[] }> }
    ) => {
      const params = await context.params;
      const path = params.path || [];
      const resourceName = path[0];

      if (resourceName && hookRegistry.has(resourceName)) {
        const hookContext = createHookContext(request);
        const body = await request.json();
        const modifiedData = await executeBeforeCreate(
          resourceName,
          body,
          hookContext
        );

        const newRequest = new NextRequest(request.url, {
          method: 'POST',
          headers: request.headers,
          body: JSON.stringify(modifiedData),
        });

        const result = await proxy.POST(newRequest, context);
        const responseBody = await result.json();

        if (responseBody.data) {
          await executeAfterCreate(
            resourceName,
            responseBody.data,
            hookContext
          );
        }

        return new Response(JSON.stringify(responseBody), {
          status: result.status,
          headers: result.headers,
        });
      }

      return proxy.POST(request, context);
    },

    PATCH: async (
      request: NextRequest,
      context: { params: Promise<{ path: string[] }> }
    ) => {
      const params = await context.params;
      const path = params.path || [];
      const resourceName = path[0];
      const id = path[1];

      if (resourceName && id && hookRegistry.has(resourceName)) {
        const hookContext = createHookContext(request);
        const body = await request.json();
        const modifiedData = await executeBeforeUpdate(
          resourceName,
          body,
          id,
          hookContext
        );

        const newRequest = new NextRequest(request.url, {
          method: 'PATCH',
          headers: request.headers,
          body: JSON.stringify(modifiedData),
        });

        const result = await proxy.PATCH(newRequest, context);
        const responseBody = await result.json();

        if (responseBody.data) {
          await executeAfterUpdate(
            resourceName,
            responseBody.data,
            id,
            hookContext
          );
        }

        return new Response(JSON.stringify(responseBody), {
          status: result.status,
          headers: result.headers,
        });
      }

      return proxy.PATCH(request, context);
    },

    DELETE: async (
      request: NextRequest,
      context: { params: Promise<{ path: string[] }> }
    ) => {
      const params = await context.params;
      const path = params.path || [];
      const resourceName = path[0];
      const id = path[1];

      if (resourceName && id && hookRegistry.has(resourceName)) {
        const hookContext = createHookContext(request);
        const shouldDelete = await executeBeforeDelete(
          resourceName,
          id,
          hookContext
        );

        if (!shouldDelete) {
          return new Response(
            JSON.stringify({ error: 'Delete cancelled by hook' }),
            {
              status: 400,
              headers: { 'Content-Type': 'application/json' },
            }
          );
        }

        const result = await proxy.DELETE(request, context);
        await executeAfterDelete(resourceName, id, hookContext);

        return result;
      }

      return proxy.DELETE(request, context);
    },

    PUT: proxy.PUT,
  };

  return {
    handler,
    registry: hookRegistry,
  };
}

function createHookContext(request: NextRequest): HookContext {
  return {
    params: {},
    request,
    response: new Response(),
  };
}

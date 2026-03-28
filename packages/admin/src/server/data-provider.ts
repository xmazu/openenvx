export interface BaseRecord {
  id?: string | number;
  [key: string]: unknown;
}

export interface Pagination {
  current?: number;
  currentPage?: number;
  mode?: 'client' | 'server';
  pageSize?: number;
}

export interface Sorter {
  field: string;
  order: 'asc' | 'desc';
}

export interface CrudFilter {
  field: string;
  operator: string;
  value: unknown;
}

export interface GetListParams {
  filters?: CrudFilter[];
  meta?: Record<string, unknown>;
  pagination?: Pagination;
  resource: string;
  sorters?: Sorter[];
}

export interface GetListResponse<TData extends BaseRecord = BaseRecord> {
  data: TData[];
  total: number;
}

export interface GetOneParams {
  id: string | number;
  meta?: Record<string, unknown>;
  resource: string;
}

export interface GetOneResponse<TData extends BaseRecord = BaseRecord> {
  data: TData;
}

export interface CreateParams<TVariables = Record<string, unknown>> {
  meta?: Record<string, unknown>;
  resource: string;
  variables: TVariables;
}

export interface CreateResponse<TData extends BaseRecord = BaseRecord> {
  data: TData;
}

export interface UpdateParams<TVariables = Record<string, unknown>> {
  id: string | number;
  meta?: Record<string, unknown>;
  resource: string;
  variables: TVariables;
}

export interface UpdateResponse<TData extends BaseRecord = BaseRecord> {
  data: TData;
}

export interface DeleteOneParams<TVariables = Record<string, unknown>> {
  id: string | number;
  meta?: Record<string, unknown>;
  resource: string;
  variables?: TVariables;
}

export interface DeleteOneResponse<TData extends BaseRecord = BaseRecord> {
  data: TData;
}

export interface DataProvider {
  create: <
    TData extends BaseRecord = BaseRecord,
    TVariables = Record<string, unknown>,
  >(
    params: CreateParams<TVariables>
  ) => Promise<CreateResponse<TData>>;
  custom?: <
    TData extends BaseRecord = BaseRecord,
    TQuery = unknown,
    TPayload = unknown,
  >(params: {
    url: string;
    method: string;
    query?: TQuery;
    payload?: TPayload;
    headers?: Record<string, string>;
  }) => Promise<{ data: TData }>;
  deleteOne: <
    TData extends BaseRecord = BaseRecord,
    TVariables = Record<string, unknown>,
  >(
    params: DeleteOneParams<TVariables>
  ) => Promise<DeleteOneResponse<TData>>;
  getApiUrl: () => string;
  getList: <TData extends BaseRecord = BaseRecord>(
    params: GetListParams
  ) => Promise<GetListResponse<TData>>;
  getOne: <TData extends BaseRecord = BaseRecord>(
    params: GetOneParams
  ) => Promise<GetOneResponse<TData>>;
  update: <
    TData extends BaseRecord = BaseRecord,
    TVariables = Record<string, unknown>,
  >(
    params: UpdateParams<TVariables>
  ) => Promise<UpdateResponse<TData>>;
}

export interface DataProviderConfig {
  /** Proxy API base path (e.g., '/api/admin') */
  apiUrl: string;
  /** Default headers for all requests */
  headers?: Record<string, string>;
}

export function createPostgRESTDataProvider(
  config: DataProviderConfig
): DataProvider {
  const { apiUrl: proxyUrl, headers: defaultHeaders = {} } = config;

  const getHeaders = (): HeadersInit => {
    return {
      'Content-Type': 'application/json',
      Accept: 'application/json',
      ...defaultHeaders,
    };
  };

  const buildQueryParams = (params: GetListParams): string => {
    const queryParts: string[] = [];

    // Pagination
    if (params.pagination) {
      const pageSize = params.pagination.pageSize ?? 10;
      const currentPage =
        (params.pagination as { current?: number; currentPage?: number })
          .current ??
        (params.pagination as { current?: number; currentPage?: number })
          .currentPage ??
        1;
      const offset = (currentPage - 1) * pageSize;
      queryParts.push(`limit=${pageSize}&offset=${offset}`);
    }

    // Sorting
    if (params.sorters && params.sorters.length > 0) {
      const orderBy = params.sorters
        .map((s) => `${s.field}.${s.order === 'desc' ? 'desc' : 'asc'}`)
        .join(',');
      queryParts.push(`order=${orderBy}`);
    }

    // Filters
    if (params.filters && params.filters.length > 0) {
      for (const filter of params.filters) {
        if ('field' in filter && filter.operator === 'eq') {
          queryParts.push(`${filter.field}=eq.${filter.value}`);
        }
      }
    }

    return queryParts.length > 0 ? `?${queryParts.join('&')}` : '';
  };

  const dataProvider: DataProvider = {
    getList: async <TData extends BaseRecord = BaseRecord>({
      resource,
      pagination,
      sorters,
      filters,
    }: GetListParams): Promise<GetListResponse<TData>> => {
      const queryParams = buildQueryParams({
        resource,
        pagination,
        sorters,
        filters,
      });
      const response = await fetch(`${proxyUrl}/${resource}${queryParams}`, {
        method: 'GET',
        headers: getHeaders(),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch list: ${response.statusText}`);
      }

      // Get total count from Content-Range header
      const contentRange = response.headers.get('Content-Range');
      const total = contentRange
        ? Number.parseInt(contentRange.split('/')[1] || '0', 10)
        : 0;

      const data = await response.json();

      return {
        data: data as TData[],
        total,
      };
    },

    getOne: async <TData extends BaseRecord = BaseRecord>({
      resource,
      id,
    }: GetOneParams): Promise<GetOneResponse<TData>> => {
      const response = await fetch(
        `${proxyUrl}/${resource}?id=eq.${id}&limit=1`,
        {
          method: 'GET',
          headers: getHeaders(),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to fetch record: ${response.statusText}`);
      }

      const data = await response.json();

      if (!data || data.length === 0) {
        throw new Error(`Record not found: ${id}`);
      }

      return {
        data: data[0] as TData,
      };
    },

    create: async <
      TData extends BaseRecord = BaseRecord,
      TVariables = Record<string, unknown>,
    >({
      resource,
      variables,
    }: CreateParams<TVariables>): Promise<CreateResponse<TData>> => {
      const response = await fetch(`${proxyUrl}/${resource}`, {
        method: 'POST',
        headers: {
          ...getHeaders(),
          Prefer: 'return=representation',
        },
        body: JSON.stringify(variables),
      });

      if (!response.ok) {
        throw new Error(`Failed to create record: ${response.statusText}`);
      }

      const data = await response.json();

      return {
        data: (Array.isArray(data) ? data[0] : data) as TData,
      };
    },

    update: async <
      TData extends BaseRecord = BaseRecord,
      TVariables = Record<string, unknown>,
    >({
      resource,
      id,
      variables,
    }: UpdateParams<TVariables>): Promise<UpdateResponse<TData>> => {
      const response = await fetch(`${proxyUrl}/${resource}?id=eq.${id}`, {
        method: 'PATCH',
        headers: {
          ...getHeaders(),
          Prefer: 'return=representation',
        },
        body: JSON.stringify(variables),
      });

      if (!response.ok) {
        throw new Error(`Failed to update record: ${response.statusText}`);
      }

      const data = await response.json();

      return {
        data: (Array.isArray(data) ? data[0] : data) as TData,
      };
    },

    deleteOne: async <
      TData extends BaseRecord = BaseRecord,
      TVariables = Record<string, unknown>,
    >({
      resource,
      id,
    }: DeleteOneParams<TVariables>): Promise<DeleteOneResponse<TData>> => {
      const response = await fetch(`${proxyUrl}/${resource}?id=eq.${id}`, {
        method: 'DELETE',
        headers: {
          ...getHeaders(),
          Prefer: 'return=representation',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to delete record: ${response.statusText}`);
      }

      return {
        data: { id } as TData,
      };
    },

    getApiUrl: () => proxyUrl,

    custom: async <
      TData extends BaseRecord = BaseRecord,
      TQuery = unknown,
      TPayload = unknown,
    >({
      url,
      method,
      query,
      payload,
    }: {
      url: string;
      method: string;
      query?: TQuery;
      payload?: TPayload;
    }): Promise<{ data: TData }> => {
      const queryString = query
        ? `?${new URLSearchParams(query as Record<string, string>).toString()}`
        : '';

      const response = await fetch(`${proxyUrl}${url}${queryString}`, {
        method,
        headers: getHeaders(),
        body: payload ? JSON.stringify(payload) : undefined,
      });

      if (!response.ok) {
        throw new Error(`Custom request failed: ${response.statusText}`);
      }

      const data = await response.json();

      return {
        data: data as TData,
      };
    },
  };

  return dataProvider;
}

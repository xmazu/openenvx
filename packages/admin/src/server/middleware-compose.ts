import type { NextRequest, NextResponse } from 'next/server';

export type MiddlewareFunction = (
  request: NextRequest
) => Promise<NextResponse | Response> | NextResponse | Response;

export type MiddlewareNextFunction = () =>
  | Promise<NextResponse | Response>
  | NextResponse
  | Response;

export type Middleware = (
  request: NextRequest,
  next: MiddlewareNextFunction
) => Promise<NextResponse | Response> | NextResponse | Response;

export function composeMiddleware(
  middlewares: Middleware[]
): MiddlewareFunction {
  return async function composedMiddleware(
    request: NextRequest
  ): Promise<NextResponse | Response> {
    let index = 0;

    async function next(): Promise<NextResponse | Response> {
      if (index >= middlewares.length) {
        const { NextResponse } = await import('next/server');
        return NextResponse.next();
      }

      const middleware = middlewares[index++];
      return await middleware(request, next);
    }

    return await next();
  };
}

export function createConditionalMiddleware(
  matcher: (pathname: string) => boolean,
  middleware: Middleware
): Middleware {
  return async (request, next) => {
    const pathname = request.nextUrl.pathname;

    if (!matcher(pathname)) {
      return await next();
    }

    return await middleware(request, next);
  };
}

export function createPathExcludingMiddleware(
  excludePaths: string[],
  middleware: Middleware
): Middleware {
  return createConditionalMiddleware(
    (pathname) => !excludePaths.some((path) => pathname.startsWith(path)),
    middleware
  );
}

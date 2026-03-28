# @openenvx/admin

Admin panel for PostgREST with Better Auth integration.

## Important: Better Auth Uses Cookie-Based Sessions

**Better Auth does NOT use JWT tokens.** It uses traditional cookie-based sessions.

This package handles authentication correctly by:
- **Middleware**: Checks cookie + validates session via Better Auth API (`/api/auth/session`)
- **API Route**: Validates session server-side + creates PostgREST token
- **PostgREST**: Receives JWT token for row-level security

## Install

```bash
npm install @openenvx/admin
```

## Setup

### 1. Database Setup

Run the included SQL script in your PostgreSQL database to create the necessary roles:

```bash
# Copy and run the SQL script in your database (e.g., Neon SQL Editor)
cat node_modules/@openenvx/admin/postgrest-setup.sql
```

**Key points:**
- Creates `authenticator` role - the role PostgREST uses to connect
- Creates `anon` and `admin_service` roles for JWT authentication
- Grants necessary permissions for role switching

**Important:** Update the `DATABASE_URL` in your PostgREST config to use the `authenticator` role:
```
postgresql://authenticator:your-password@host/database
```

### 2. API Route

```typescript
// app/api/admin/[...path]/route.ts
import { createAdmin, createBetterAuthTokenExtractor } from "@openenvx/admin/server";

const extractToken = createBetterAuthTokenExtractor({
  betterAuthUrl: process.env.NEXT_PUBLIC_APP_URL!,
  jwtSecret: process.env.ADMIN_JWT_SECRET!, // For PostgREST token only
  requiredRole: "admin", // Optional: check user role
});

const admin = createAdmin({
  postgrestUrl: process.env.POSTGREST_URL!,
  auth: {
    getToken: extractToken,
  },
});

export const { GET, POST, PUT, PATCH, DELETE } = admin.handler;
```

### 3. Middleware (Simple)

```typescript
// middleware.ts
import { createAuthMiddleware } from "@openenvx/admin/server";

export const middleware = createAuthMiddleware({
  betterAuthUrl: process.env.NEXT_PUBLIC_APP_URL!,
  loginPath: "/auth/login",
  requiredRole: "admin", // Optional
  publicRoutes: ["/login", "/auth"],
});

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
```

### 4. Middleware (Composition)

```typescript
// middleware.ts
import {
  composeMiddleware,
  createAuthMiddleware,
  createPathExcludingMiddleware,
} from "@openenvx/admin/server";

const authMiddleware = createAuthMiddleware({
  betterAuthUrl: process.env.NEXT_PUBLIC_APP_URL!,
  loginPath: "/auth/login",
  requiredRole: "admin",
});

const loggingMiddleware = async (request, next) => {
  console.log(`[Middleware] ${request.method} ${request.url}`);
  return next();
};

export const middleware = composeMiddleware([
  loggingMiddleware,
  createPathExcludingMiddleware(['/api/internal'], authMiddleware),
]);

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
```

### 5. Layout

```typescript
// app/layout.tsx
import { AdminProvider } from "@openenvx/admin/components";
import { authClient } from "@/lib/auth";

const resources = [
  { name: "users", list: "/users", create: "/users/create" },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <AdminProvider resources={resources} authClient={authClient}>
      {children}
    </AdminProvider>
  );
}
```

## Auth Client

### With Better Auth

```typescript
// lib/auth.ts
import { createAdminAuthClient } from "@openenvx/admin/client";
import { createAuthClient } from "better-auth/react";

const betterAuth = createAuthClient({ baseURL: process.env.NEXT_PUBLIC_APP_URL });
export const authClient = createAdminAuthClient(betterAuth);
```

### Custom Auth

```typescript
// lib/auth.ts
import type { AuthClient } from "@openenvx/admin/client";

export const authClient: AuthClient = {
  getSession: async () => {
    const res = await fetch("/api/auth/session");
    if (!res.ok) return null;
    return res.json();
  },
  signOut: async () => {
    await fetch("/api/auth/logout", { method: "POST" });
  },
};
```

## Environment Variables

```env
POSTGREST_URL=http://localhost:3001
NEXT_PUBLIC_APP_URL=http://localhost:3000
ADMIN_JWT_SECRET=your-secret-min-32-chars-for-postgrest-only
```

**Note**: `ADMIN_JWT_SECRET` is used only for creating PostgREST JWT tokens, NOT for Better Auth validation.

## How It Works

### Authentication Flow

1. **User logs in** via Better Auth → session cookie is set (`better-auth.session_token`)
2. **Middleware** on every page request:
   - Checks if session cookie exists
   - Calls Better Auth API `/api/auth/session` to validate session
   - Checks user roles if required
   - Redirects to login if not authenticated
3. **API routes** (`/api/admin/*`) are **skipped** by middleware (they have their own auth in the proxy)
4. **Proxy** validates session again and creates PostgREST JWT token

### Why Skip /api/admin in Middleware?

API routes handle their own authentication via `createBetterAuthTokenExtractor`. The middleware:
- Protects **page routes** (the admin UI)
- Skips **API routes** `/api/*` (they protect themselves)

This avoids double-validation and allows the proxy to handle PostgREST token creation.

## Middleware Composition

Compose multiple middlewares like in TanStack Start:

```typescript
import {
  composeMiddleware,
  createConditionalMiddleware,
  createPathExcludingMiddleware,
} from "@openenvx/admin/server";

// Compose multiple middlewares
const middleware = composeMiddleware([
  loggingMiddleware,
  rateLimitMiddleware,
  authMiddleware,
]);

// Conditionally run middleware
const conditionalAuth = createConditionalMiddleware(
  (pathname) => pathname.startsWith('/admin'),
  authMiddleware
);

// Exclude paths from middleware
const adminOnly = createPathExcludingMiddleware(
  ['/api/webhooks', '/api/public'],
  authMiddleware
);
```

## API

### Server

#### Admin Setup
- `createAdmin(config)` - Creates admin API handler with PostgREST proxy
- `createBetterAuthTokenExtractor(config)` - Validates session and creates PostgREST token

#### Middleware
- `createAuthMiddleware(config)` - Creates auth middleware (cookie check + getSession)
- `createBetterAuthChecker(cookieName?)` - Lightweight cookie check only

#### Middleware Composition
- `composeMiddleware(middlewares[])` - Compose multiple middlewares
- `createConditionalMiddleware(matcher, middleware)` - Conditional middleware
- `createPathExcludingMiddleware(paths[], middleware)` - Exclude paths from middleware

### Client

- `AdminProvider` - Root provider component
- `createAdminAuthClient(betterAuthClient)` - Wraps Better Auth client
- `useAuth()` - Access auth context
- `useAuthUser()` - Get current user

## License

MIT

# @openenvx/admin

A standalone, auth-agnostic admin panel package for building data management interfaces with PostgREST.

## Features

- **Auth-Agnostic**: Works with any authentication system (Better Auth, NextAuth, Clerk, custom JWT, etc.)
- **PostgREST Integration**: Direct integration with PostgREST for database operations
- **Automatic Introspection**: Discovers your database schema automatically
- **Type-Safe**: Full TypeScript support
- **Customizable**: Extensive theming and component customization options
- **Server & Client**: Provides both server-side utilities and React components

## Installation

```bash
npm install @openenvx/admin
# or
yarn add @openenvx/admin
# or
bun add @openenvx/admin
```

## Quick Start

### 1. Server Setup (API Routes)

Create your admin API endpoint:

```typescript
// app/api/admin/[...path]/route.ts
import {
  createAdmin,
  createBetterAuthTokenExtractor,
  withAuth,
} from "@openenvx/admin/server";
import type { NextRequest } from "next/server";

const extractToken = createBetterAuthTokenExtractor({
  betterAuthSecret: process.env.BETTER_AUTH_SECRET!,
  jwtSecret: process.env.ADMIN_JWT_SECRET!,
  dbRole: "admin_service",
  requiredRole: "super_admin",
});

const admin = createAdmin({
  postgrestUrl: process.env.POSTGREST_URL!,
  getToken: (req: NextRequest) => {
    // Read token from header (set by withAuth) or extract fresh
    return req.headers.get("x-admin-token") || extractToken(req);
  },
});

export const { GET, POST, PUT, PATCH, DELETE } = withAuth(admin.handler, {
  getToken: extractToken,
});
```

### 2. Client Setup (Layout)

Wrap your app with the AdminProvider:

```typescript
// app/layout.tsx
import { AdminProvider } from "@openenvx/admin/components";
import { authClient } from "@yourproject/auth"; // Your auth implementation

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <AdminProvider resources={resources} authClient={authClient}>
      {children}
    </AdminProvider>
  );
}
```

### 3. Middleware Setup

Add authentication middleware:

```typescript
// middleware.ts
import { createAuthMiddleware, createBetterAuthChecker } from "@openenvx/admin/server";

export const middleware = createAuthMiddleware({
  loginPath: "/auth/login",
  isAuthenticated: createBetterAuthChecker("better-auth.session"),
  publicRoutes: ["/login", "/auth"],
});

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
```

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Browser   │────▶│  Middleware  │────▶│  Login Page │ (if not auth)
└─────────────┘     └──────────────┘     └─────────────┘
       │
       ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Admin API  │────▶│  withAuth()  │────▶│   401 Error │ (if no token)
│  (/api/...) │     └──────────────┘     └─────────────┘
└─────────────┘            │
       │                   │ (token valid)
       ▼                   ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  PostgREST  │◀────│  Admin JWT   │────▶│  PostgreSQL │
└─────────────┘     └──────────────┘     └─────────────┘
```

## Auth Integration

The admin package is completely auth-agnostic. You provide an `AuthClient` implementation:

```typescript
interface AuthClient {
  getSession: () => Promise<AuthSession | null>;
  signOut: () => Promise<void>;
  onSessionChange?: (callback: (session: AuthSession | null) => void) => () => void;
}

interface AuthSession {
  user: {
    id: string;
    email?: string;
    name?: string;
    role?: string | string[];
    image?: string;
  };
  token?: string;
}
```

### Example: Better Auth Integration

```typescript
// packages/auth/client.ts
import type { AuthClient, AuthSession } from "@openenvx/admin/client";
import { createAuthClient } from "better-auth/react";

const betterAuth = createAuthClient({ baseURL: "..." });

export const authClient: AuthClient = {
  getSession: async () => {
    const { data } = await betterAuth.getSession();
    if (!data) return null;
    return {
      user: {
        id: data.user.id,
        email: data.user.email,
        name: data.user.name,
        role: data.user.role,
      },
      token: data.session.token,
    } as AuthSession;
  },
  signOut: () => betterAuth.signOut(),
};
```

### Example: Custom JWT Integration

```typescript
// lib/auth-client.ts
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

## API Reference

### Server Exports (`@openenvx/admin/server`)

#### `createAdmin(config)`
Creates the admin API handler.

```typescript
const admin = createAdmin({
  postgrestUrl: string;
  getToken?: (request: NextRequest) => string | null | Promise<string | null>;
  resources?: Record<string, ResourceConfig>;
});
```

#### `withAuth(handlers, config)`
Wraps API handlers to enforce authentication.

```typescript
const protectedHandlers = withAuth(handlers, {
  getToken: (request) => string | null | Promise<string | null>;
  onAuthFailure?: (request) => NextResponse;
  tokenHeader?: string; // default: "x-admin-token"
});
```

#### `createAuthMiddleware(config)`
Creates Next.js middleware for page-level auth.

```typescript
const middleware = createAuthMiddleware({
  isAuthenticated: (request) => boolean | Promise<boolean>;
  loginPath?: string;        // default: "/auth/login"
  publicRoutes?: string[];   // additional public routes
  onAuthFailure?: (request) => void;
});
```

#### `createBetterAuthTokenExtractor(config)`
Extracts and validates Better Auth sessions.

```typescript
const getToken = createBetterAuthTokenExtractor({
  betterAuthSecret: string;
  jwtSecret: string;
  dbRole?: string;           // default: "admin_service"
  requiredRole?: string;     // default: "super_admin"
  tokenExpirySeconds?: number; // default: 300
});
```

### Client Exports (`@openenvx/admin/client`)

#### Components
- `AdminProvider` - Root provider with auth and resources

#### Hooks
- `useAuth()` - Access auth context
- `useAuthUser()` - Get current user
- `useLogout()` - Sign out functionality
- `useGetIdentity()` - Get user identity (compatible with Refine)

#### Types
- `AuthClient` - Interface for auth implementations
- `AuthSession` - Session type
- `AuthUser` - User type

## Configuration

### Environment Variables

```env
# Required
POSTGREST_URL=http://localhost:3001

# For Better Auth integration
BETTER_AUTH_SECRET=your-secret
ADMIN_JWT_SECRET=your-secret-min-32-chars

# Optional
ADMIN_DB_ROLE=admin_service
ADMIN_REQUIRED_ROLE=super_admin
ADMIN_TOKEN_EXPIRY=300
AUTH_LOGIN_PATH=/auth/login
```

### PostgREST Configuration

```env
PGRST_JWT_SECRET=your-secret-min-32-characters-long
PGRST_DB_ANON_ROLE=anon
```

## Security

- **Token Extraction**: Happens once at API boundary, passed via header
- **JWT Validation**: Cryptographic signature verification (HS256)
- **Role Checking**: Server-side role validation before token creation
- **Short-lived Tokens**: Admin JWTs expire quickly (5 min default)
- **No Token Storage**: Tokens are ephemeral, created per-request

## License

MIT

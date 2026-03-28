# Admin Panel

This admin panel is built with `@openenvx/admin` and uses your shared authentication package.

## Architecture

```
Browser → Middleware (cookie check) → Login Page (if not authenticated)
   ↓
Admin API → Auth Check (JWT validation) → 401 Error (if invalid)
   ↓ (valid)
PostgREST ← Admin JWT ← PostgreSQL
```

**Key Points:**
- Middleware only handles page redirects (not API routes)
- API routes validate JWT and return 401 if invalid
- All database access goes through PostgREST with authenticated JWT
- Auth is provided by your shared `@{{projectName}}/auth` package

## Quick Start

1. **Environment Setup**

Copy `.env.hbs` to `.env.local` and configure:

```env
# Login redirect path
AUTH_LOGIN_PATH=/auth/login

# Better Auth (from dashboard app)
BETTER_AUTH_SECRET=your-secret

# PostgREST
POSTGREST_URL=http://localhost:3001
ADMIN_JWT_SECRET=your-secret-min-32-chars

# Optional
ADMIN_DB_ROLE=admin_service
ADMIN_REQUIRED_ROLE=super_admin
ADMIN_TOKEN_EXPIRY=300
```

2. **Auth Package**

This app imports auth from `@{{projectName}}/auth`. The auth client is already configured:

```typescript
// packages/auth/src/client.ts
import { createAdminAuthClient } from "@openenvx/admin/client";
import { createAuthClient } from "better-auth/react";

const betterAuth = createAuthClient({ baseURL: "..." });
export const authClient = createAdminAuthClient(betterAuth);
```

3. **Middleware**

Already configured in `middleware.ts`:

```typescript
import { createAuthMiddleware, createBetterAuthChecker } from "@openenvx/admin/server";

export const middleware = createAuthMiddleware({
  loginPath: process.env.AUTH_LOGIN_PATH || "/auth/login",
  isAuthenticated: createBetterAuthChecker("better-auth.session"),
  publicRoutes: ["/login", "/auth/login", "/auth/register"],
});
```

4. **API Routes**

Already configured in `app/api/admin/[...path]/route.ts`:

```typescript
import { createAdmin, createBetterAuthTokenExtractor } from "@openenvx/admin/server";

const extractToken = createBetterAuthTokenExtractor({
  betterAuthSecret: process.env.BETTER_AUTH_SECRET!,
  jwtSecret: process.env.ADMIN_JWT_SECRET!,
});

const admin = createAdmin({
  postgrestUrl: process.env.POSTGREST_URL!,
  getToken: (req) => req.headers.get("x-admin-token"),
  auth: {
    getToken: extractToken,
    tokenHeader: "x-admin-token",
  },
});

export const { GET, POST, PUT, PATCH, DELETE } = admin.handler;
```

5. **Layout**

Already configured in `app/layout.tsx`:

```typescript
import { AdminProvider } from "@openenvx/admin/components";
import { authClient } from "@{{projectName}}/auth";

export default function RootLayout({ children }) {
  return (
    <AdminProvider resources={resources} authClient={authClient}>
      {children}
    </AdminProvider>
  );
}
```

## PostgREST Setup

### Environment Variables

```env
PGRST_JWT_SECRET=your-secret-min-32-characters-long
PGRST_DB_ANON_ROLE=anon
```

**Note:** `PGRST_JWT_SECRET` must match `ADMIN_JWT_SECRET` from this app.

### PostgreSQL Roles

```sql
-- Admin role (full permissions)
CREATE ROLE admin_service WITH LOGIN;
GRANT USAGE ON SCHEMA public TO admin_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO admin_service;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO admin_service;
ALTER DEFAULT PRIVILEGES IN SCHEMA public 
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO admin_service;

-- Anonymous role (minimal permissions)
CREATE ROLE anon WITH LOGIN;
```

## User Roles

Users need the configured role (default: `super_admin`) in Better Auth:

```typescript
// In dashboard app
await authClient.updateUser({
  role: "super_admin"
});
```

## Customization

### Adding Resources

Edit `app/layout.tsx`:

```typescript
const resources = [
  {
    name: "users",
    list: "/users",
    create: "/users/create",
    edit: "/users/edit",
    show: "/users/show",
  },
  // ...
];
```

### Changing Auth Configuration

Edit `packages/auth/src/client.ts` to customize how auth works.

### Middleware Options

Edit `middleware.ts`:

```typescript
export const middleware = createAuthMiddleware({
  loginPath: "/custom-login",
  isAuthenticated: async (request) => {
    // Custom auth check
    return true;
  },
  publicRoutes: ["/public-page"],
});
```

## Troubleshooting

### 401 Unauthorized (API)
- Missing or invalid Better Auth cookie
- `BETTER_AUTH_SECRET` mismatch between apps
- User doesn't have required role

### Redirect Loop
- Check `AUTH_LOGIN_PATH` is accessible (add to `publicRoutes`)
- Ensure login page doesn't require auth

### PostgREST 401
- `PGRST_JWT_SECRET` doesn't match `ADMIN_JWT_SECRET`
- JWT expired (check `ADMIN_TOKEN_EXPIRY`)
- Database role doesn't exist

### Missing Auth Client
- Ensure `@{{projectName}}/auth` exports `authClient`
- Check package is installed: `bun install`

## Security

- Admin JWTs are short-lived (5 minutes default)
- Created fresh for each request
- Cryptographically signed with HS256
- PostgREST validates signature before executing queries
- Middleware only redirects, API returns 401

## Further Reading

- [@openenvx/admin documentation](https://github.com/yourorg/openenvx/tree/main/packages/admin)
- [PostgREST JWT Guide](https://postgrest.org/en/stable/tutorials/tut1.html)
- [Better Auth Documentation](https://better-auth.com)

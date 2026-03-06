# OpenEnvX - Agent Guidelines

## Project Overview

OpenEnvX is a local-first development runtime for micro-SaaS builders focused on secure environment variable management with encryption. It's a multi-language monorepo combining TypeScript/JavaScript (for web and developer experience) with Go (for performance-critical CLI operations and security).

**Key Philosophy:** Secure by default, local-first, developer-friendly tooling for managing environment secrets.

## Tech Stack

### JavaScript/TypeScript

- **Package Manager:** Bun (v1.3.0)
- **Monorepo:** Turborepo (v2.8.13)
- **Build:** tsup
- **Linting/Formatting:** Biome (v2.4.0) with Ultracite config
- **Versioning:** Changesets
- **Frontend:** Next.js 16, React 19, Tailwind CSS v4
- **UI Components:** shadcn/ui (New York style)
- **Validation:** Zod v4.3.0
- **CLI Framework:** Commander.js + @clack/prompts
- **Testing:** Vitest

### Go

- **Go Version:** 1.24.0
- **CLI Framework:** Cobra
- **TUI:** Charmbracelet (huh, lipgloss, bubbletea)
- **Encryption:** age (X25519), AES-256-GCM
- **MCP Support:** Model Context Protocol Go SDK

## Project Structure

```
/apps/
  └── landing/          # Next.js marketing site
/packages/
  ├── openenvx/         # Project generator CLI (create-openenvx-app)
  └── envtyped/         # Typed env validation library (@openenvx/envtyped)
/envx/                  # Go CLI for secure env management
/runtime/               # oexctl - Go control plane CLI for local dev proxy
```

## oexctl - Control Plane CLI

The `oexctl` CLI provides local development proxying with automatic TLS and subdomain routing.

```bash
# Install (via openenvx init or curl)
openenvx init  # Installs oexctl binary

# Run an app with proxy
oexctl proxy run myapp -- npm run dev

# Access at https://myapp.localhost:1355
```

**Features:**
- Automatic TLS certificate generation
- Subdomain-based routing (*.localhost:1355)
- No DNS configuration needed
- Route management for multiple apps

## Conventions

### File Naming

**TypeScript/JavaScript:**

- Files: `kebab-case.ts` (e.g., `project-generator.ts`)
- Components: `kebab-case.tsx` (e.g., `hero-section.tsx`)
- Templates: `.hbs` extension for Handlebars
- Tests: `.test.ts` suffix

**Go:**

- Files: `snake_case.go` (e.g., `workspace.go`)
- Test files: `_test.go` suffix
- Packages: lowercase single word

### Code Style

**TypeScript:**

- Use `node:` prefix for built-ins (`node:path`, `node:fs`)
- Prefer `import type` for type-only imports
- Named exports preferred over default exports
- Use Zod for runtime validation
- Custom error classes for domain errors (e.g., `EnvValidationError`)

**Go:**

- Explicit error returns, no panic in production code

## Do's

- **Use Biome for linting/formatting** - Don't bypass the configured linter
- **Write tests for new features** - Vitest for JS, table-driven tests for Go
- **Use Changesets for versioning** - Always create a changeset for published packages
- **Use `node:` prefix** for Node.js built-in modules
- **Validate with Zod** at runtime boundaries
- **Handle errors explicitly** in Go (no panics in production paths)
- **Use async generators** for CLI progress reporting when appropriate
- **Follow existing naming conventions** in each language
- **Use TypeScript strict mode** - leverage the type system
- **Prefer explicit types** over `any` or implicit types

## Don'ts

- **Don't commit secrets** - The entire point is secure secret management
- **Don't use default exports** unnecessarily
- **Don't panic in Go** except in `main()` or `init()`
- **Don't mix sync/async** without clear intent
- **Don't modify `package-lock.json`** - this project uses Bun
- **Don't create circular dependencies** between packages

## Package Publishing

Packages are published to npm under the `@openenvx` scope:

- `@openenvx/envtyped` - Typed environment validation
- `openenvx` - Project generator CLI

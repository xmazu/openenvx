# OpenEnvX - Agent Guidelines

## Project Overview

OpenEnvX is a local-first development runtime for micro-SaaS builders. It consists of:
- **CLI Tool** (`packages/openenvx`) - Project generator for scaffolding SaaS applications
- **Landing Site** (`apps/landing`) - Marketing website
- **Admin Package** (`packages/admin`) - Reusable admin panel components

## Architecture

```
/apps/
  └── landing/          # Next.js 16 marketing site
/packages/
  ├── openenvx/         # CLI tool for generating SaaS projects
  └── admin/            # Admin panel package (PostgREST + shadcn/ui)
```

## Tech Stack

### Package Management & Build
- **Package Manager:** Bun 1.3.0
- **Monorepo:** Turborepo
- **Build:** tsup / rolldown (package-specific)

### Frontend
- **Framework:** Next.js 16, React 19
- **Styling:** Tailwind CSS v4
- **UI Components:** shadcn/ui, Radix UI primitives
- **Animation:** Motion (Framer Motion successor)
- **Icons:** Lucide React, Simple Icons

### CLI Framework
- **Commander.js** - CLI command structure
- **@clack/prompts** - Interactive prompts
- **Handlebars** - Template rendering
- **execa** - Process execution

### Development Tools
- **Linting/Formatting:** Biome with Ultracite config
- **Type Checking:** TypeScript 5.9 (strict mode)
- **Testing:** Vitest
- **Validation:** Zod v4

## Conventions

### File Naming
- Files: `kebab-case.ts` (e.g., `project-generator.ts`)
- Components: `kebab-case.tsx` (e.g., `hero-section.tsx`)
- Templates: `.hbs` extension for Handlebars
- Tests: `.test.ts` suffix

### Code Style
- Use `node:` prefix for built-ins (`node:path`, `node:fs`)
- Prefer `import type` for type-only imports
- Named exports preferred over default exports
- Use Zod for runtime validation
- Custom error classes for domain errors

## Do's

- **Use Biome for linting/formatting** - Don't bypass the configured linter
- **Write tests for new features** - Vitest for JS/TS
- **Use Changesets for versioning** - Always create a changeset for published packages
- **Use `node:` prefix** for Node.js built-in modules
- **Validate with Zod** at runtime boundaries
- **Follow existing naming conventions** in each package
- **Use TypeScript strict mode** - leverage the type system
- **Prefer explicit types** over `any` or implicit types

## Don'ts

- **Don't commit secrets** - The entire point is secure secret management
- **Don't use default exports** unnecessarily
- **Don't mix sync/async** without clear intent
- **Don't modify `package-lock.json`** - this project uses Bun
- **Don't create circular dependencies** between packages

## Package Publishing

Packages are published to npm:

- `openenvx` - Project generator CLI
- `@openenvx/admin` - Admin panel components

## Working in this Repo

- ALWAYS USE PARALLEL TOOLS WHEN APPLICABLE
- The default branch is `main`
- Prefer automation: execute actions without confirmation unless blocked by missing info or safety/irreversibility
- You may be running in a git worktree - all changes must be made in your current working directory
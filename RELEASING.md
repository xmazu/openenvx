# Releasing

This project uses [Changesets](https://github.com/changesets/changesets) to manage versions and releases.

## Developer Workflow

### Making Changes

When you make changes that should be released:

1. **Create a changeset** before or with your PR:

   ```bash
   bun changeset
   ```

   This will:
   - Ask which packages are affected
   - Ask what type of change it is (patch/minor/major)
   - Ask for a description of the change
   - Create a `.changeset/*.md` file

2. **Commit the changeset** file along with your changes:

   ```bash
   git add .
   git commit -m "feat: new feature"
   git push
   ```

3. **Create your PR** - the changeset bot will comment on your PR

### Automated Release Process

Once your PR is merged to `main`:

1. **GitHub Action runs** and detects changesets
2. **"Version Packages" PR** is created automatically with:
   - Bumped versions in package.json files
   - Updated CHANGELOG.md files
   - Removed changeset files (they become part of changelog)
3. **Maintainer merges** the Version Packages PR
4. **Packages are published** to npm automatically
5. **GitHub Release** is created with changelog

## Version Types

For 0.x versions (pre-1.0):

- **patch**: Bug fixes (0.1.0 → 0.1.1)
- **minor**: New features OR breaking changes (0.1.0 → 0.2.0)
- **major**: Goes to 1.0.0 - DON'T USE until ready!

**Important**: While in 0.x, use `minor` for breaking changes to stay in 0.x range. Only use `major` when you're ready to release v1.0.0.

## Manual Commands (if needed)

```bash
# Create a changeset
bun changeset

# Version packages locally (usually not needed)
bun version-packages

# Publish packages (CI does this automatically)
bun publish-packages
```

## Skipping Releases

If you have changes that shouldn't be released (e.g., docs, tests), just don't add a changeset. The changes will be included in the next release that has a changeset.

## Emergency Releases

If you need to release immediately without waiting for the Version Packages PR:

1. Merge the Version Packages PR
2. The publish happens automatically on merge

Or manually (requires npm access):

```bash
bun run build
bun publish-packages
```

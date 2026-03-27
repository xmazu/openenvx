import type { FieldConfig } from './resource-types';

interface ComputedConfig {
  deps: string[];
  fn: (data: Record<string, unknown>) => unknown;
}

export function detectCircularDependencies(fields: FieldConfig[]): void {
  const graph = new Map<string, Set<string>>();

  for (const field of fields) {
    const computed = getComputedConfig(field);
    if (computed?.deps) {
      graph.set(field.name, new Set(computed.deps));
    }
  }

  const visited = new Set<string>();
  const path = new Set<string>();

  function hasCycle(node: string, currentPath: Set<string>): boolean {
    if (currentPath.has(node)) {
      return true;
    }
    if (visited.has(node)) {
      return false;
    }

    visited.add(node);
    currentPath.add(node);

    const deps = graph.get(node);
    if (deps) {
      for (const dep of deps) {
        if (hasCycle(dep, currentPath)) {
          return true;
        }
      }
    }

    currentPath.delete(node);
    return false;
  }

  for (const field of fields) {
    const computed = getComputedConfig(field);
    if (computed?.deps) {
      visited.clear();
      path.clear();
      if (hasCycle(field.name, path)) {
        const cyclePath = findCyclePath(field.name, graph);
        throw new Error(
          `Circular dependency detected in field "${field.name}". ` +
            `Cycle: ${cyclePath.join(' -> ')} -> ${field.name}`
        );
      }
    }
  }
}

function getComputedConfig(field: FieldConfig): ComputedConfig | null {
  if (!field.computed) {
    return null;
  }
  if (typeof field.computed === 'function') {
    return null;
  }
  return field.computed as ComputedConfig;
}

function findCyclePath(
  startNode: string,
  graph: Map<string, Set<string>>
): string[] {
  const path: string[] = [];
  const visited = new Set<string>();

  function dfs(node: string): boolean {
    if (node === startNode && path.length > 0) {
      return true;
    }
    if (visited.has(node)) {
      return false;
    }

    visited.add(node);
    path.push(node);

    const deps = graph.get(node);
    if (deps) {
      for (const dep of deps) {
        if (dfs(dep)) {
          return true;
        }
      }
    }

    path.pop();
    return false;
  }

  dfs(startNode);
  return path;
}

export type ViewMode = 'dashboard' | 'list' | 'create' | 'edit' | 'show';

export interface ParsedPath {
  recordId: string | null;
  resourceName: string | null;
  viewMode: ViewMode;
}

export function parsePath(path: string[]): ParsedPath {
  if (path.length === 0) {
    return { resourceName: null, viewMode: 'dashboard', recordId: null };
  }

  const resourceName = path[0];

  if (path.length === 1) {
    return { resourceName, viewMode: 'list', recordId: null };
  }

  const action = path[1];

  if (action === 'create') {
    return { resourceName, viewMode: 'create', recordId: null };
  }

  if (action === 'edit' && path[2]) {
    return { resourceName, viewMode: 'edit', recordId: path[2] };
  }

  if ((action === 'show' || action === 'view') && path[2]) {
    return { resourceName, viewMode: 'show', recordId: path[2] };
  }

  if (path[1] && !['create', 'edit', 'show', 'view'].includes(path[1])) {
    return { resourceName, viewMode: 'show', recordId: path[1] };
  }

  return { resourceName, viewMode: 'list', recordId: null };
}

export function formatCellValue(value: unknown): string {
  if (value === null || value === undefined) {
    return '-';
  }
  if (typeof value === 'boolean') {
    return value ? 'Yes' : 'No';
  }
  if (typeof value === 'object') {
    if (value instanceof Date) {
      return value.toLocaleDateString();
    }
    return JSON.stringify(value).slice(0, 50);
  }
  return String(value).slice(0, 50);
}

'use client';

import { useCallback, useMemo, useState } from 'react';
import type { BulkAction } from '@/components/bulk-operations';
import { BulkOperations } from '@/components/bulk-operations';
import { useList, useResourceConfig } from '@/hooks';
import type { BulkActionConfig } from '@/lib/resource-types';
import { DeleteButton } from '@/ui/buttons/delete';
import { EditButton } from '@/ui/buttons/edit';
import { ShowButton } from '@/ui/buttons/show';
import { Checkbox } from '@/ui/shadcn/checkbox';
import { Skeleton } from '@/ui/shadcn/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/ui/shadcn/table';
import { ListView, ListViewHeader } from '@/ui/views/list-view';
import { formatCellValue } from './admin-utils';

interface ListPageViewProps {
  resourceName: string;
}

export function ListPageView({ resourceName }: ListPageViewProps) {
  const { config, loading: configLoading } = useResourceConfig(resourceName);
  const listResult = useList({
    resource: resourceName,
    pagination: { pageSize: 25 },
  });
  const [selectedIds, setSelectedIds] = useState<(string | number)[]>([]);

  const records = (listResult.result?.data || []) as Record<string, unknown>[];
  const isLoading = listResult.query?.isPending || configLoading;
  const hasBulkActions = (config?.list?.bulkActions?.length ?? 0) > 0;

  const displayColumns = useMemo(() => {
    if (!config?.list?.columns || config.list.columns.length === 0) {
      return ['id'];
    }
    return config.list.columns;
  }, [config]);

  const allRecordIds = useMemo(() => {
    return records.map((record) => String(record.id ?? '')).filter(Boolean);
  }, [records]);

  const isAllSelected =
    allRecordIds.length > 0 &&
    allRecordIds.every((id) => selectedIds.includes(id));

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(allRecordIds);
    } else {
      setSelectedIds([]);
    }
  };

  const handleSelectRow = (recordId: string | number, checked: boolean) => {
    if (checked) {
      setSelectedIds((prev: (string | number)[]) => [...prev, recordId]);
    } else {
      setSelectedIds((prev: (string | number)[]) =>
        prev.filter((id: string | number) => id !== recordId)
      );
    }
  };

  const handleClearSelection = () => {
    setSelectedIds([]);
  };

  const handleBulkAction = useCallback(
    async (
      actionKey: string,
      ids: (string | number)[],
      formData?: Record<string, unknown>
    ) => {
      console.log('Bulk action:', actionKey, ids, formData);
      await listResult.query.refetch();
      setSelectedIds([]);
    },
    [listResult.query]
  );

  const bulkActions: BulkAction[] = useMemo(() => {
    const configs =
      (config?.list?.bulkActions as BulkActionConfig[] | undefined) || [];
    return configs.map((config: BulkActionConfig) => ({
      key: config.key,
      label: config.label,
      icon: undefined,
      confirm: config.confirm,
      dialog: config.dialog
        ? {
            title: config.dialog.title,
            fields: config.dialog.fields,
            handler: async (
              ids: (string | number)[],
              formData: Record<string, unknown>
            ) => {
              await handleBulkAction(config.key, ids, formData);
            },
          }
        : undefined,
      handler: async (ids: (string | number)[]) => {
        await handleBulkAction(config.key, ids);
      },
    }));
  }, [config?.list?.bulkActions, handleBulkAction]);

  function renderTableBody() {
    if (isLoading) {
      return (
        <TableRow>
          {hasBulkActions && (
            <TableCell>
              <Skeleton className="h-4 w-4" />
            </TableCell>
          )}
          {displayColumns.map((column) => (
            <TableCell key={column}>
              <Skeleton className="h-4 w-full" />
            </TableCell>
          ))}
          <TableCell>
            <Skeleton className="h-8 w-20" />
          </TableCell>
        </TableRow>
      );
    }

    if (records.length === 0) {
      return (
        <TableRow>
          <TableCell
            className="h-24 text-center"
            colSpan={displayColumns.length + (hasBulkActions ? 2 : 1)}
          >
            No records found.
          </TableCell>
        </TableRow>
      );
    }

    return records.map((record: Record<string, unknown>) => {
      const recordId = String(record.id ?? Math.random());
      const isSelected = selectedIds.includes(recordId);
      return (
        <TableRow
          data-state={isSelected ? 'selected' : undefined}
          key={recordId}
        >
          {hasBulkActions && (
            <TableCell>
              <Checkbox
                checked={isSelected}
                onCheckedChange={(checked: boolean) =>
                  handleSelectRow(recordId, checked)
                }
              />
            </TableCell>
          )}
          {displayColumns.map((column) => (
            <TableCell key={column}>
              {formatCellValue(record[column])}
            </TableCell>
          ))}
          <TableCell>
            <div className="flex items-center gap-2">
              <ShowButton recordItemId={recordId} resource={resourceName} />
              <EditButton recordItemId={recordId} resource={resourceName} />
              <DeleteButton recordItemId={recordId} resource={resourceName} />
            </div>
          </TableCell>
        </TableRow>
      );
    });
  }

  return (
    <ListView>
      <ListViewHeader canCreate={config?.canCreate ?? true} />
      {hasBulkActions && selectedIds.length > 0 && (
        <BulkOperations
          actions={bulkActions}
          onClearSelection={handleClearSelection}
          selectedIds={selectedIds}
        />
      )}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {hasBulkActions && (
                <TableHead className="w-[40px]">
                  <Checkbox
                    aria-label="Select all"
                    checked={isAllSelected}
                    onCheckedChange={handleSelectAll}
                  />
                </TableHead>
              )}
              {displayColumns.map((column) => (
                <TableHead key={column}>{column}</TableHead>
              ))}
              <TableHead className="w-[150px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>{renderTableBody()}</TableBody>
        </Table>
      </div>
    </ListView>
  );
}

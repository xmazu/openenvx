'use client';

import { useMemo } from 'react';
import { useResourceConfig, useShow } from '@/hooks';
import { cn } from '@/lib/utils';
import { EditButton } from '@/ui/buttons/edit';
import { ListButton } from '@/ui/buttons/list';
import { Skeleton } from '@/ui/shadcn/skeleton';
import { ShowView, ShowViewHeader } from '@/ui/views/show-view';
import { formatCellValue } from './admin-utils';

interface ShowPageViewProps {
  recordId: string;
  resourceName: string;
}

export function ShowPageView({ resourceName, recordId }: ShowPageViewProps) {
  const { config, loading: configLoading } = useResourceConfig(resourceName);
  const { query } = useShow({
    resource: resourceName,
    id: recordId,
  });

  const record = query?.data?.data as Record<string, unknown> | undefined;

  const displayFields = useMemo(() => {
    if (!config?.fields) {
      return [];
    }
    return config.fields.filter((f) => !f.hidden);
  }, [config]);

  return (
    <ShowView>
      <ShowViewHeader />
      <div className={cn('rounded-md border p-6')}>
        {configLoading || query.isLoading || !record ? (
          <div className="space-y-4">
            <Skeleton className="h-6 w-full" />
            <Skeleton className="h-6 w-full" />
            <Skeleton className="h-6 w-full" />
          </div>
        ) : (
          <div className="space-y-4">
            {displayFields.map((field) => (
              <div
                className="flex justify-between border-b py-2"
                key={field.name}
              >
                <span className="font-medium text-muted-foreground">
                  {field.label || field.name}
                </span>
                <span>{formatCellValue(record[field.name])}</span>
              </div>
            ))}
            <div className="flex gap-2 pt-4">
              <EditButton recordItemId={recordId} resource={resourceName}>
                Edit
              </EditButton>
              <ListButton resource={resourceName} variant="outline">
                Back to List
              </ListButton>
            </div>
          </div>
        )}
      </div>
    </ShowView>
  );
}

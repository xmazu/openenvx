'use client';

import { useEffect, useState } from 'react';
import { AutoForm } from '@/components/auto-form';
import { useDataProvider, useResourceConfig } from '@/hooks';
import { Skeleton } from '@/ui/shadcn/skeleton';
import { EditView, EditViewHeader } from '@/ui/views/edit-view';

interface EditPageViewProps {
  recordId: string;
  resourceName: string;
}

export function EditPageView({ resourceName, recordId }: EditPageViewProps) {
  const { config, loading: configLoading } = useResourceConfig(resourceName);
  const dataProvider = useDataProvider();
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [record, setRecord] = useState<Record<string, unknown> | undefined>(
    undefined
  );
  const [isRecordLoading, setIsRecordLoading] = useState(true);

  useEffect(() => {
    const fetchRecord = async () => {
      setIsRecordLoading(true);
      setFormError(null);

      try {
        const response = await dataProvider.getOne({
          resource: resourceName,
          id: recordId,
        });
        setRecord(response.data as Record<string, unknown>);
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : 'Failed to fetch record';
        setFormError(errorMessage);
      } finally {
        setIsRecordLoading(false);
      }
    };

    if (recordId) {
      fetchRecord();
    }
  }, [dataProvider, resourceName, recordId]);

  const handleSubmit = async (data: Record<string, unknown>) => {
    setFormLoading(true);
    setFormError(null);

    try {
      await dataProvider.update({
        resource: resourceName,
        id: recordId,
        variables: data,
      });
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to update record';
      setFormError(errorMessage);
      throw err;
    } finally {
      setFormLoading(false);
    }
  };

  if (configLoading || !config || isRecordLoading || !record) {
    return (
      <EditView>
        <EditViewHeader />
        <div className="rounded-md border p-6">
          <div className="space-y-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        </div>
      </EditView>
    );
  }

  return (
    <EditView>
      <EditViewHeader />
      <div className="rounded-md border p-6">
        <AutoForm
          error={formError}
          initialData={record}
          isLoading={formLoading}
          mode="edit"
          onSubmit={handleSubmit}
          resourceConfig={config}
        />
      </div>
    </EditView>
  );
}

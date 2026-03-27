'use client';

import { useState } from 'react';
import { AutoForm } from '@/components/auto-form';
import { useDataProvider, useResourceConfig } from '@/hooks';
import { Skeleton } from '@/ui/shadcn/skeleton';
import { CreateView, CreateViewHeader } from '@/ui/views/create-view';

interface CreatePageViewProps {
  resourceName: string;
}

export function CreatePageView({ resourceName }: CreatePageViewProps) {
  const { config, loading: configLoading } = useResourceConfig(resourceName);
  const dataProvider = useDataProvider();
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const handleSubmit = async (data: Record<string, unknown>) => {
    setFormLoading(true);
    setFormError(null);

    try {
      await dataProvider.create({
        resource: resourceName,
        variables: data,
      });
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to create record';
      setFormError(errorMessage);
      throw err;
    } finally {
      setFormLoading(false);
    }
  };

  if (configLoading || !config) {
    return (
      <CreateView>
        <CreateViewHeader />
        <div className="rounded-md border p-6">
          <div className="space-y-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        </div>
      </CreateView>
    );
  }

  return (
    <CreateView>
      <CreateViewHeader />
      <div className="rounded-md border p-6">
        <AutoForm
          error={formError}
          isLoading={formLoading}
          mode="create"
          onSubmit={handleSubmit}
          resourceConfig={config}
        />
      </div>
    </CreateView>
  );
}

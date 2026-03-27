'use client';

import { Database, LayoutDashboard, Rows3, Table2 } from 'lucide-react';
import { useResources } from '@/hooks';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/ui/shadcn/card';
import { Skeleton } from '@/ui/shadcn/skeleton';

interface DashboardCardProps {
  description?: string;
  icon: React.ReactNode;
  loading?: boolean;
  title: string;
  value: string | number;
}

function DashboardCard({
  title,
  value,
  description,
  icon,
  loading,
}: DashboardCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="font-medium text-sm">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-8 w-20" />
        ) : (
          <>
            <div className="font-bold text-2xl">{value}</div>
            {description && (
              <p className="text-muted-foreground text-xs">{description}</p>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}

interface AdminDashboardProps {
  className?: string;
}

export function AdminDashboard({ className }: AdminDashboardProps) {
  const { resources } = useResources();

  function renderResourcesContent() {
    if (resources.length === 0) {
      return (
        <p className="text-muted-foreground text-sm">
          No resources configured.
        </p>
      );
    }

    return (
      <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
        {resources.map((resource) => (
          <a
            className="flex items-center gap-2 rounded-md border p-3 transition-colors hover:bg-muted"
            href={resource.list}
            key={resource.name}
          >
            <Table2 className="h-4 w-4 text-muted-foreground" />
            <span className="font-medium">
              {resource.label || resource.name}
            </span>
          </a>
        ))}
      </div>
    );
  }

  return (
    <div className={cn('flex flex-col gap-6', className)}>
      <div className="flex items-center gap-2">
        <LayoutDashboard className="h-6 w-6" />
        <h1 className="font-bold text-2xl">Dashboard</h1>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <DashboardCard
          description="Active resources in admin panel"
          icon={<Database className="h-4 w-4 text-muted-foreground" />}
          title="Resources"
          value={resources.length}
        />
        <DashboardCard
          description="Resources with list view"
          icon={<Table2 className="h-4 w-4 text-muted-foreground" />}
          title="Active Resources"
          value={resources.filter((r) => r.list).length}
        />
        <DashboardCard
          description="All systems operational"
          icon={<Rows3 className="h-4 w-4 text-muted-foreground" />}
          title="System Status"
          value="Active"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Available Resources</CardTitle>
        </CardHeader>
        <CardContent>{renderResourcesContent()}</CardContent>
      </Card>
    </div>
  );
}

export default AdminDashboard;

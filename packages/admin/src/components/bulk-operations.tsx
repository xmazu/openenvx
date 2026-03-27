'use client';

import { Loader2 } from 'lucide-react';
import { useCallback, useState } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/ui/shadcn/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/ui/shadcn/dialog';
import { Label } from '@/ui/shadcn/label';

function renderActionIcon(
  isLoading: boolean,
  Icon?: React.ComponentType<{ className?: string }>
) {
  if (isLoading) {
    return <Loader2 className="mr-1 h-4 w-4 animate-spin" />;
  }
  if (Icon) {
    return <Icon className="mr-1 h-4 w-4" />;
  }
  return null;
}

export interface BulkAction {
  confirm?: {
    destructive?: boolean;
    message: string;
    title: string;
  };
  dialog?: {
    fields: Array<{
      label: string;
      name: string;
      options?: string[];
      type: 'text' | 'select';
    }>;
    handler: (
      ids: (string | number)[],
      formData: Record<string, unknown>
    ) => Promise<void>;
    title: string;
  };
  handler?: (ids: (string | number)[]) => Promise<void>;
  icon?: React.ComponentType<{ className?: string }>;
  key: string;
  label: string;
}

export interface BulkOperationsProps {
  actions: BulkAction[];
  onClearSelection: () => void;
  selectedIds: (string | number)[];
}

export function BulkOperations({
  actions,
  onClearSelection,
  selectedIds,
}: BulkOperationsProps) {
  const [loading, setLoading] = useState<string | null>(null);
  const [confirmAction, setConfirmAction] = useState<BulkAction | null>(null);
  const [dialogAction, setDialogAction] = useState<BulkAction | null>(null);
  const [dialogFormData, setDialogFormData] = useState<Record<string, unknown>>(
    {}
  );

  const handleAction = useCallback(
    async (action: BulkAction) => {
      if (action.confirm) {
        setConfirmAction(action);
        return;
      }

      if (action.dialog) {
        setDialogAction(action);
        setDialogFormData({});
        return;
      }

      if (action.handler) {
        setLoading(action.key);
        try {
          await action.handler(selectedIds);
          onClearSelection();
        } catch (error) {
          console.error('Bulk action failed:', error);
        } finally {
          setLoading(null);
        }
      }
    },
    [selectedIds, onClearSelection]
  );

  const handleConfirm = async () => {
    if (!confirmAction) {
      return;
    }

    setLoading(confirmAction.key);
    try {
      if (confirmAction.handler) {
        await confirmAction.handler(selectedIds);
      }
      onClearSelection();
      setConfirmAction(null);
    } catch (error) {
      console.error('Bulk action failed:', error);
    } finally {
      setLoading(null);
    }
  };

  const handleDialogSubmit = async () => {
    if (!dialogAction?.dialog) {
      return;
    }

    setLoading(dialogAction.key);
    try {
      await dialogAction.dialog.handler(selectedIds, dialogFormData);
      onClearSelection();
      setDialogAction(null);
    } catch (error) {
      console.error('Bulk action failed:', error);
    } finally {
      setLoading(null);
    }
  };

  return (
    <>
      <div className="flex items-center gap-2 rounded-lg bg-muted px-4 py-2">
        <span className="font-medium text-sm">
          {selectedIds.length} selected
        </span>
        <div className="h-4 w-px bg-border" />
        {actions.map((action) => {
          const Icon = action.icon;
          return (
            <Button
              className={cn(
                action.confirm?.destructive &&
                  'text-destructive hover:text-destructive'
              )}
              disabled={loading === action.key}
              key={action.key}
              onClick={() => handleAction(action)}
              size="sm"
              variant="ghost"
            >
              {renderActionIcon(loading === action.key, Icon)}
              {action.label}
            </Button>
          );
        })}
        <Button
          className="ml-auto"
          onClick={onClearSelection}
          size="sm"
          variant="ghost"
        >
          Clear
        </Button>
      </div>

      {/* Confirmation Dialog */}
      <Dialog
        onOpenChange={() => setConfirmAction(null)}
        open={!!confirmAction}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{confirmAction?.confirm?.title}</DialogTitle>
            <DialogDescription>
              {confirmAction?.confirm?.message}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button onClick={() => setConfirmAction(null)} variant="outline">
              Cancel
            </Button>
            <Button
              disabled={!!loading}
              onClick={handleConfirm}
              variant={
                confirmAction?.confirm?.destructive ? 'destructive' : 'default'
              }
            >
              {loading ? (
                <Loader2 className="mr-1 h-4 w-4 animate-spin" />
              ) : (
                'Confirm'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog onOpenChange={() => setDialogAction(null)} open={!!dialogAction}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{dialogAction?.dialog?.title}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {dialogAction?.dialog?.fields.map((field) => (
              <div className="space-y-2" key={field.name}>
                <Label>{field.label}</Label>
                {field.type === 'select' ? (
                  <select
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    onChange={(e) =>
                      setDialogFormData((prev) => ({
                        ...prev,
                        [field.name]: e.target.value,
                      }))
                    }
                    value={(dialogFormData[field.name] as string) || ''}
                  >
                    <option value="">Select...</option>
                    {field.options?.map((opt) => (
                      <option key={opt} value={opt}>
                        {opt}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    onChange={(e) =>
                      setDialogFormData((prev) => ({
                        ...prev,
                        [field.name]: e.target.value,
                      }))
                    }
                    type="text"
                    value={(dialogFormData[field.name] as string) || ''}
                  />
                )}
              </div>
            ))}
          </div>
          <DialogFooter>
            <Button onClick={() => setDialogAction(null)} variant="outline">
              Cancel
            </Button>
            <Button disabled={!!loading} onClick={handleDialogSubmit}>
              {loading ? (
                <Loader2 className="mr-1 h-4 w-4 animate-spin" />
              ) : (
                'Apply'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

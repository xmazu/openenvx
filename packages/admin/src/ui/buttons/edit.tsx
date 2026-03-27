'use client';

import { Pencil } from 'lucide-react';
import Link from 'next/link';
import React from 'react';
import { Button } from '@/ui/shadcn/button';

type EditButtonProps = {
  resource?: string;
  recordItemId?: string | number;
} & React.ComponentProps<typeof Button>;

export const EditButton = React.forwardRef<
  React.ComponentRef<typeof Button>,
  EditButtonProps
>(({ resource, recordItemId, children, onClick, ...rest }, ref) => {
  if (!(resource && recordItemId)) {
    return null;
  }

  return (
    <Button {...rest} asChild ref={ref}>
      <Link
        href={`/${resource}/edit/${recordItemId}`}
        onClick={
          onClick as unknown as React.MouseEventHandler<HTMLAnchorElement>
        }
      >
        {children ?? (
          <div className="flex items-center gap-2 font-semibold">
            <Pencil className="h-4 w-4" />
            <span>Edit</span>
          </div>
        )}
      </Link>
    </Button>
  );
});

EditButton.displayName = 'EditButton';

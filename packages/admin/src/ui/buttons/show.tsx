'use client';

import { Eye } from 'lucide-react';
import Link from 'next/link';
import React from 'react';
import { Button } from '@/ui/shadcn/button';

type ShowButtonProps = {
  resource: string;
  recordItemId: string;
} & React.ComponentProps<typeof Button>;

export const ShowButton = React.forwardRef<
  React.ComponentRef<typeof Button>,
  ShowButtonProps
>(({ resource, recordItemId, children, onClick, ...rest }, ref) => {
  return (
    <Button {...rest} asChild ref={ref}>
      <Link
        href={`/${resource}/show/${recordItemId}`}
        onClick={
          onClick as unknown as React.MouseEventHandler<HTMLAnchorElement>
        }
      >
        {children ?? (
          <div className="flex items-center gap-2 font-semibold">
            <Eye className="h-4 w-4" />
            <span>Show</span>
          </div>
        )}
      </Link>
    </Button>
  );
});

ShowButton.displayName = 'ShowButton';

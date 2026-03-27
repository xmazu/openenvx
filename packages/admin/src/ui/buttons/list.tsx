'use client';

import { List } from 'lucide-react';
import Link from 'next/link';
import React from 'react';
import { Button } from '@/ui/shadcn/button';

type ListButtonProps = {
  resource: string;
} & React.ComponentProps<typeof Button>;

export const ListButton = React.forwardRef<
  React.ComponentRef<typeof Button>,
  ListButtonProps
>(({ resource, children, onClick, ...rest }, ref) => {
  return (
    <Button {...rest} asChild ref={ref}>
      <Link
        href={`/${resource}`}
        onClick={
          onClick as unknown as React.MouseEventHandler<HTMLAnchorElement>
        }
      >
        {children ?? (
          <div className="flex items-center gap-2 font-semibold">
            <List className="h-4 w-4" />
            <span>List</span>
          </div>
        )}
      </Link>
    </Button>
  );
});

ListButton.displayName = 'ListButton';

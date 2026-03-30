'use client';

import * as Icons from 'lucide-react';
import { ChevronRight, ListIcon } from 'lucide-react';
import Link from 'next/link';
import type React from 'react';
import { useAdminOptions, useMenu } from '@/hooks';
import { cn } from '@/lib/utils';
import type { TreeMenuItem } from '@/types';
import { Button } from '@/ui/shadcn/button';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/ui/shadcn/collapsible';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/ui/shadcn/dropdown-menu';
import {
  Sidebar as ShadcnSidebar,
  SidebarContent as ShadcnSidebarContent,
  SidebarHeader as ShadcnSidebarHeader,
  SidebarRail as ShadcnSidebarRail,
  SidebarTrigger as ShadcnSidebarTrigger,
  useSidebar as useShadcnSidebar,
} from '@/ui/shadcn/sidebar';

export function Sidebar() {
  const { open } = useShadcnSidebar();
  const { menuItems, selectedKey } = useMenu();

  return (
    <ShadcnSidebar className={cn('border-none')} collapsible="icon">
      <ShadcnSidebarRail />
      <SidebarHeader />
      <ShadcnSidebarContent
        className={cn(
          'transition-discrete',
          'duration-200',
          'flex',
          'flex-col',
          'gap-2',
          'pt-2',
          'pb-2',
          'border-r',
          'border-border',
          {
            'px-3': open,
            'px-1': !open,
          }
        )}
      >
        {menuItems.map((item: TreeMenuItem) => (
          <SidebarItem
            item={item}
            key={item.key || item.name}
            selectedKey={selectedKey}
          />
        ))}
      </ShadcnSidebarContent>
    </ShadcnSidebar>
  );
}

interface MenuItemProps {
  item: TreeMenuItem;
  selectedKey?: string;
}

function SidebarItem({ item, selectedKey }: MenuItemProps) {
  const { open } = useShadcnSidebar();

  if (item.meta?.group) {
    return <SidebarItemGroup item={item} selectedKey={selectedKey} />;
  }

  if (item.children && item.children.length > 0) {
    if (open) {
      return <SidebarItemCollapsible item={item} selectedKey={selectedKey} />;
    }
    return <SidebarItemDropdown item={item} selectedKey={selectedKey} />;
  }

  return <SidebarItemLink item={item} selectedKey={selectedKey} />;
}

function SidebarItemGroup({ item, selectedKey }: MenuItemProps) {
  const { children } = item;
  const { open } = useShadcnSidebar();

  return (
    <div className={cn('border-t', 'border-sidebar-border', 'pt-4')}>
      <span
        className={cn(
          'ml-3',
          'block',
          'text-xs',
          'font-semibold',
          'uppercase',
          'text-muted-foreground',
          'transition-all',
          'duration-200',
          {
            'h-8': open,
            'h-0': !open,
            'opacity-0': !open,
            'opacity-100': open,
            'pointer-events-none': !open,
            'pointer-events-auto': open,
          }
        )}
      >
        {getDisplayName(item)}
      </span>
      {children && children.length > 0 && (
        <div className={cn('flex', 'flex-col')}>
          {children.map((child: TreeMenuItem) => (
            <SidebarItem
              item={child}
              key={child.key || child.name}
              selectedKey={selectedKey}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function SidebarItemCollapsible({ item, selectedKey }: MenuItemProps) {
  const { name, children } = item;

  const chevronIcon = (
    <ChevronRight
      className={cn(
        'h-4',
        'w-4',
        'shrink-0',
        'text-muted-foreground',
        'transition-transform',
        'duration-200',
        'group-data-[state=open]:rotate-90'
      )}
    />
  );

  return (
    <Collapsible className={cn('w-full', 'group')} key={`collapsible-${name}`}>
      <CollapsibleTrigger asChild>
        <SidebarButton item={item} rightIcon={chevronIcon} />
      </CollapsibleTrigger>
      <CollapsibleContent className={cn('ml-6', 'flex', 'flex-col', 'gap-2')}>
        {children?.map((child: TreeMenuItem) => (
          <SidebarItem
            item={child}
            key={child.key || child.name}
            selectedKey={selectedKey}
          />
        ))}
      </CollapsibleContent>
    </Collapsible>
  );
}

function SidebarItemDropdown({ item, selectedKey }: MenuItemProps) {
  const { children } = item;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <SidebarButton item={item} />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" side="right">
        {children?.map((child: TreeMenuItem) => {
          const { key: childKey } = child;
          const isSelected = childKey === selectedKey;

          return (
            <DropdownMenuItem asChild key={childKey || child.name}>
              <Link
                className={cn('flex w-full items-center gap-2', {
                  'bg-accent text-accent-foreground': isSelected,
                })}
                href={child.route || ''}
              >
                <ItemIcon
                  icon={child.meta?.icon ?? child.icon}
                  isSelected={isSelected}
                />
                <span>{getDisplayName(child)}</span>
              </Link>
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function SidebarItemLink({ item, selectedKey }: MenuItemProps) {
  const isSelected = item.key === selectedKey;

  return <SidebarButton asLink={true} isSelected={isSelected} item={item} />;
}

function SidebarHeader() {
  const { title } = useAdminOptions();
  const { open, isMobile } = useShadcnSidebar();

  return (
    <ShadcnSidebarHeader
      className={cn(
        'p-0',
        'h-16',
        'border-b',
        'border-border',
        'flex-row',
        'items-center',
        'justify-between',
        'overflow-hidden'
      )}
    >
      <div
        className={cn(
          'whitespace-nowrap',
          'flex',
          'flex-row',
          'h-full',
          'items-center',
          'justify-start',
          'gap-2',
          'transition-discrete',
          'duration-200',
          {
            'pl-3': !open,
            'pl-5': open,
          }
        )}
      >
        <div>{title.icon}</div>
        <h2
          className={cn(
            'text-sm',
            'font-bold',
            'transition-opacity',
            'duration-200',
            {
              'opacity-0': !open,
              'opacity-100': open,
            }
          )}
        >
          {title.text}
        </h2>
      </div>

      <ShadcnSidebarTrigger
        className={cn('text-muted-foreground', 'mr-1.5', {
          'opacity-0': !open,
          'opacity-100': open || isMobile,
          'pointer-events-auto': open || isMobile,
          'pointer-events-none': !(open || isMobile),
        })}
      />
    </ShadcnSidebarHeader>
  );
}

function getDisplayName(item: TreeMenuItem) {
  return item.meta?.label ?? item.label ?? item.name;
}

interface IconProps {
  icon: string | undefined;
  isSelected?: boolean;
}

function ItemIcon({ icon, isSelected }: IconProps) {
  const IconComponent = icon
    ? (
        Icons as unknown as Record<
          string,
          React.ComponentType<{ className?: string }>
        >
      )[icon]
    : null;

  return (
    <div
      className={cn('w-4', {
        'text-muted-foreground': !isSelected,
        'text-sidebar-primary-foreground': isSelected,
      })}
    >
      {IconComponent ? (
        <IconComponent className="h-4 w-4" />
      ) : (
        <ListIcon className="h-4 w-4" />
      )}
    </div>
  );
}

type SidebarButtonProps = React.ComponentProps<typeof Button> & {
  item: TreeMenuItem;
  isSelected?: boolean;
  rightIcon?: React.ReactNode;
  asLink?: boolean;
  onClick?: () => void;
};

function SidebarButton({
  item,
  isSelected = false,
  rightIcon,
  asLink = false,
  className,
  onClick,
  ...props
}: SidebarButtonProps) {
  const buttonContent = (
    <>
      <ItemIcon icon={item.meta?.icon ?? item.icon} isSelected={isSelected} />
      <span
        className={cn('tracking-[-0.00875rem]', {
          'flex-1': rightIcon,
          'text-left': rightIcon,
          'line-clamp-1': !rightIcon,
          truncate: !rightIcon,
          'font-normal': !isSelected,
          'font-semibold': isSelected,
          'text-sidebar-primary-foreground': isSelected,
          'text-foreground': !isSelected,
        })}
      >
        {getDisplayName(item)}
      </span>
      {rightIcon}
    </>
  );

  return (
    <Button
      asChild={!!(asLink && item.route)}
      className={cn(
        '!px-3 flex w-full items-center justify-start gap-2 py-2 text-sm',
        {
          'bg-sidebar-primary': isSelected,
          'hover:!bg-sidebar-primary/90': isSelected,
          'text-sidebar-primary-foreground': isSelected,
          'hover:text-sidebar-primary-foreground': isSelected,
        },
        className
      )}
      onClick={onClick}
      size="lg"
      variant="ghost"
      {...props}
    >
      {asLink && item.route ? (
        <Link
          className={cn('flex w-full items-center gap-2')}
          href={item.route}
        >
          {buttonContent}
        </Link>
      ) : (
        buttonContent
      )}
    </Button>
  );
}

Sidebar.displayName = 'Sidebar';

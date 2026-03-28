import type { ReactNode } from 'react';

export interface ListRowSlotProps {
  actionsSlot: ReactNode;
  index: number;
  onSelect: () => void;
  record: Record<string, unknown>;
  selected: boolean;
}

export interface ListCellSlotProps {
  column: string;
  record: Record<string, unknown>;
  value: unknown;
}

export interface ListHeaderSlotProps {
  column: string;
  onSort: () => void;
  sortable: boolean;
  sortDirection?: 'asc' | 'desc';
}

export interface ListFilterSlotProps {
  field: string;
  onChange: (value: unknown) => void;
  value: unknown;
}

export interface ListActionsSlotProps {
  record: Record<string, unknown>;
}

// TODO:
export interface ListViewSlots {
  actions?: (
    props: ListActionsSlotProps & { defaultActions: ReactNode }
  ) => ReactNode;
  cell?: (props: ListCellSlotProps & { defaultCell: ReactNode }) => ReactNode;
  empty?: () => ReactNode;
  filter?: (
    props: ListFilterSlotProps & { defaultFilter: ReactNode }
  ) => ReactNode;
  header?: (
    props: ListHeaderSlotProps & { defaultHeader: ReactNode }
  ) => ReactNode;
  loading?: () => ReactNode;
  row?: (props: ListRowSlotProps & { defaultRow: ReactNode }) => ReactNode;
}

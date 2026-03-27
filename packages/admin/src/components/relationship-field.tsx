'use client';

import { ChevronDown, X } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import type { FieldConfig } from '@/lib/resource-types';
import { cn } from '@/lib/utils';
import { Button } from '@/ui/shadcn/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/ui/shadcn/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/ui/shadcn/popover';

type RelationshipFieldConfig = FieldConfig & {
  displayColumn?: string;
  multiple?: boolean;
  reference: {
    displayField?: string;
    filter?: (query: unknown) => unknown;
    table: string;
  };
};

interface RelationshipFieldProps {
  field: RelationshipFieldConfig;
}

interface RelationshipOption {
  label: string;
  value: string | number;
}

function getDisplayText(
  field: RelationshipFieldConfig,
  selectedSingle: RelationshipOption | null,
  selectedMultiple: RelationshipOption[]
): string {
  if (field.multiple) {
    if (selectedMultiple.length > 0) {
      return `${selectedMultiple.length} selected`;
    }
    return `Select ${field.label || field.name}...`;
  }

  if (selectedSingle) {
    return selectedSingle.label;
  }
  return `Select ${field.label || field.name}...`;
}

export function RelationshipField({ field }: RelationshipFieldProps) {
  const form = useFormContext();
  const [open, setOpen] = useState(false);
  const [options, setOptions] = useState<RelationshipOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const abortRef = useRef<AbortController | null>(null);

  const value = form.watch(field.name);
  const displayField =
    field.displayColumn || field.reference.displayField || 'name';

  const loadOptions = useCallback(
    async (search?: string) => {
      if (abortRef.current) {
        abortRef.current.abort();
      }
      abortRef.current = new AbortController();

      setLoading(true);
      try {
        const params = new URLSearchParams();
        if (search) {
          params.append('search', search);
        }

        const response = await fetch(
          `/api/admin/relationships/${field.reference.table}?${params}`,
          { signal: abortRef.current.signal }
        );

        if (!response.ok) {
          throw new Error('Failed to load options');
        }

        const data = await response.json();
        setOptions(
          data.map((item: Record<string, unknown>) => ({
            value: item.id as string | number,
            label: (item[displayField] as string) || String(item.id),
          }))
        );
      } catch (error) {
        if (error instanceof Error && error.name !== 'AbortError') {
          console.error('Failed to load relationship options:', error);
        }
      } finally {
        setLoading(false);
      }
    },
    [field.reference.table, displayField]
  );

  useEffect(() => {
    if (open && options.length === 0) {
      loadOptions();
    }
  }, [open, options.length, loadOptions]);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchQuery) {
        loadOptions(searchQuery);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery, loadOptions]);

  const handleSelect = (selectedValue: string | number) => {
    if (field.multiple) {
      const currentValue = (value as (string | number)[]) || [];
      if (currentValue.includes(selectedValue)) {
        form.setValue(
          field.name,
          currentValue.filter((v) => v !== selectedValue)
        );
      } else {
        form.setValue(field.name, [...currentValue, selectedValue]);
      }
    } else {
      form.setValue(field.name, selectedValue);
      setOpen(false);
    }
  };

  const handleRemove = (removeValue: string | number) => {
    if (field.multiple) {
      const currentValue = (value as (string | number)[]) || [];
      form.setValue(
        field.name,
        currentValue.filter((v) => v !== removeValue)
      );
    } else {
      form.setValue(field.name, null);
    }
  };

  const selectedSingle: RelationshipOption | null = field.multiple
    ? null
    : (options.find((opt) => opt.value === value) ?? null);

  const selectedMultiple: RelationshipOption[] = field.multiple
    ? options.filter((opt) =>
        (value as (string | number)[])?.includes(opt.value)
      )
    : [];

  return (
    <div className="space-y-2">
      <Popover onOpenChange={setOpen} open={open}>
        <PopoverTrigger asChild>
          <Button
            aria-expanded={open}
            className="w-full justify-between"
            disabled={form.formState.isSubmitting}
            role="combobox"
            variant="outline"
          >
            <span className="truncate">
              {getDisplayText(field, selectedSingle, selectedMultiple)}
            </span>
            <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[300px] p-0">
          <Command>
            <CommandInput
              onValueChange={setSearchQuery}
              placeholder={`Search ${field.reference.table}...`}
              value={searchQuery}
            />
            <CommandList>
              <CommandEmpty>
                {loading ? 'Loading...' : 'No results found.'}
              </CommandEmpty>
              <CommandGroup>
                {options.map((option) => {
                  const isSelected = field.multiple
                    ? (value as (string | number)[])?.includes(option.value)
                    : value === option.value;

                  return (
                    <CommandItem
                      key={option.value}
                      onSelect={() => handleSelect(option.value)}
                      value={String(option.value)}
                    >
                      <div
                        className={cn(
                          'mr-2 flex h-4 w-4 items-center justify-center rounded-sm border',
                          isSelected
                            ? 'border-primary bg-primary text-primary-foreground'
                            : 'opacity-50 [&_svg]:invisible'
                        )}
                      >
                        {isSelected && <span className="text-xs">✓</span>}
                      </div>
                      {option.label}
                    </CommandItem>
                  );
                })}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {field.multiple && selectedMultiple.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {selectedMultiple.map((option) => (
            <div
              className="flex items-center gap-1 rounded-md bg-secondary px-2 py-1 text-sm"
              key={option.value}
            >
              <span>{option.label}</span>
              <button
                className="text-muted-foreground hover:text-foreground"
                onClick={() => handleRemove(option.value)}
                type="button"
              >
                <X className="h-3 w-3" />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

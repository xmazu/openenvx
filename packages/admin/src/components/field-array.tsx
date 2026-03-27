'use client';

import { GripVertical, Plus, Trash2 } from 'lucide-react';
import { useFieldArray, useFormContext } from 'react-hook-form';
import type { FieldConfig } from '@/lib/resource-types';
import { Button } from '@/ui/shadcn/button';

type ArrayFieldConfig = FieldConfig & {
  of: { type: string; fields?: FieldConfig[] };
  minItems?: number;
  maxItems?: number;
};

export function FieldArray({ field }: { field: ArrayFieldConfig }) {
  const form = useFormContext();
  const arrayConfig = field;

  const { fields, append, remove, move } = useFieldArray({
    control: form.control,
    name: field.name,
  });

  const handleAdd = () => {
    if (arrayConfig.of.type === 'object' && arrayConfig.of.fields) {
      const defaultItem: Record<string, unknown> = {};
      for (const f of arrayConfig.of.fields) {
        defaultItem[f.name] = f.defaultValue ?? '';
      }
      append(defaultItem);
    } else {
      append('');
    }
  };

  const canAdd = !arrayConfig.maxItems || fields.length < arrayConfig.maxItems;
  const canRemove =
    !arrayConfig.minItems || fields.length > arrayConfig.minItems;

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        {fields.map((item, index) => (
          <div
            className="flex items-start gap-2 rounded-lg border bg-muted/50 p-3"
            key={item.id}
          >
            <button
              className="mt-2 cursor-grab text-muted-foreground hover:text-foreground"
              onMouseDown={(e) => {
                const startY = e.clientY;
                const handleMouseUp = (upEvent: MouseEvent) => {
                  const endY = upEvent.clientY;
                  if (endY < startY - 30 && index > 0) {
                    move(index, index - 1);
                  } else if (endY > startY + 30 && index < fields.length - 1) {
                    move(index, index + 1);
                  }
                  window.removeEventListener('mouseup', handleMouseUp);
                };
                window.addEventListener('mouseup', handleMouseUp);
              }}
              type="button"
            >
              <GripVertical className="h-4 w-4" />
            </button>

            <div className="flex-1">
              {arrayConfig.of.type === 'object' && arrayConfig.of.fields ? (
                <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                  {arrayConfig.of.fields.map((subField) => (
                    <RHFArrayField
                      field={subField}
                      key={subField.name}
                      prefix={`${field.name}.${index}`}
                    />
                  ))}
                </div>
              ) : (
                <input
                  {...form.register(`${field.name}.${index}` as const)}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  placeholder={`Item ${index + 1}`}
                />
              )}
            </div>

            {canRemove && (
              <Button
                className="text-destructive hover:text-destructive"
                onClick={() => remove(index)}
                size="icon"
                type="button"
                variant="ghost"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        ))}
      </div>

      {canAdd && (
        <Button
          className="w-full"
          onClick={handleAdd}
          size="sm"
          type="button"
          variant="outline"
        >
          <Plus className="mr-2 h-4 w-4" />
          Add Item
        </Button>
      )}
    </div>
  );
}

function RHFArrayField({
  field,
  prefix,
}: {
  field: FieldConfig;
  prefix: string;
}) {
  const form = useFormContext();
  const name = `${prefix}.${field.name}` as const;

  const fieldId = `array-${prefix}-${field.name}`;

  return (
    <div className="space-y-1">
      <label className="text-muted-foreground text-xs" htmlFor={fieldId}>
        {field.label || field.name}
      </label>
      <input
        id={fieldId}
        {...form.register(name)}
        className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
        placeholder={field.placeholder}
      />
    </div>
  );
}

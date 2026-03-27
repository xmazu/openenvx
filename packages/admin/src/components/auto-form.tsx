'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect, useMemo } from 'react';
import {
  FormProvider,
  useForm,
  useFormContext,
  useWatch,
} from 'react-hook-form';
import { FieldArray } from '@/components/field-array';
import { RelationshipField } from '@/components/relationship-field';
import type {
  ArrayFieldConfig,
  FieldConfig,
  ResourceConfig,
} from '@/lib/resource-types';
import { generateZodSchema } from '@/lib/schema-generator';
import { cn } from '@/lib/utils';
import { Alert, AlertDescription } from '@/ui/shadcn/alert';
import { Button } from '@/ui/shadcn/button';
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/ui/shadcn/card';

export interface AutoFormProps {
  className?: string;
  error?: string | null;
  initialData?: Record<string, unknown>;
  isLoading?: boolean;
  mode: 'create' | 'edit';
  onSubmit: (data: Record<string, unknown>) => Promise<void>;
  resourceConfig: ResourceConfig;
}

function ConditionalField({
  field,
  children,
}: {
  field: FieldConfig;
  children: React.ReactNode;
}) {
  const form = useFormContext();
  const watchAll = form.watch();

  const isVisible = useMemo(() => {
    if (field.hidden) {
      return false;
    }
    if (field.condition) {
      return field.condition(watchAll);
    }
    return true;
  }, [field, watchAll]);

  if (!isVisible) {
    return null;
  }

  return <>{children}</>;
}

function ComputedField({ field }: { field: FieldConfig }) {
  const form = useFormContext();

  const computedConfig = field.computed;
  const isConfig =
    computedConfig &&
    typeof computedConfig === 'object' &&
    'deps' in computedConfig;
  const deps = isConfig ? (computedConfig as { deps: string[] }).deps : [];
  const depValues = useWatch({ control: form.control, name: deps });

  useEffect(() => {
    if (!computedConfig) {
      return;
    }

    if (isConfig && 'fn' in computedConfig) {
      const data = Object.fromEntries(
        deps.map((dep: string, i: number) => [dep, depValues[i]])
      );
      const computed = (
        computedConfig as { fn: (data: Record<string, unknown>) => unknown }
      ).fn(data);
      form.setValue(field.name, computed, { shouldValidate: false });
    } else if (typeof computedConfig === 'function') {
      const allData = form.getValues();
      const computed = computedConfig(allData);
      form.setValue(field.name, computed, { shouldValidate: false });
    }
  }, [depValues, field, form, deps, isConfig, computedConfig]);

  return <input type="hidden" {...form.register(field.name)} />;
}

function FormField({ field }: { field: FieldConfig }) {
  const form = useFormContext();
  const error = form.formState.errors[field.name];

  const fieldId = `field-${field.name}`;

  return (
    <div className="space-y-2">
      <label className="font-medium text-sm" htmlFor={fieldId}>
        {field.label || field.name}
        {field.required && <span className="text-destructive">*</span>}
      </label>

      {field.computed ? (
        <ComputedField field={field} />
      ) : (
        <FormInput field={field} id={fieldId} />
      )}

      {error && (
        <p className="text-destructive text-sm">{error.message as string}</p>
      )}

      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
    </div>
  );
}

function FormInput({ field, id }: { field: FieldConfig; id?: string }) {
  const form = useFormContext();
  const { register } = form;

  const commonProps = {
    ...register(field.name),
    id,
    disabled: field.readOnly || form.formState.isSubmitting,
    className: cn(
      'flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      form.formState.errors[field.name] && 'border-destructive'
    ),
  };

  switch (field.type) {
    case 'textarea':
      return (
        <textarea
          {...commonProps}
          rows={(field as { rows?: number }).rows || 3}
        />
      );

    case 'boolean':
      return (
        <input
          type="checkbox"
          {...register(field.name)}
          className="h-4 w-4 rounded border-gray-300"
          disabled={field.readOnly || form.formState.isSubmitting}
        />
      );

    case 'select':
      return (
        <select {...commonProps}>
          <option value="">Select...</option>
          {(
            (
              field as {
                options: Array<{ value: string; label?: string } | string>;
              }
            ).options || []
          ).map((opt) => {
            const value = typeof opt === 'object' ? opt.value : opt;
            const label =
              typeof opt === 'object' ? opt.label || opt.value : opt;
            return (
              <option key={value} value={value}>
                {label}
              </option>
            );
          })}
        </select>
      );

    case 'date':
      return <input type="date" {...commonProps} />;

    case 'datetime':
      return <input type="datetime-local" {...commonProps} />;

    case 'password':
      return <input type="password" {...commonProps} />;

    case 'email':
      return <input type="email" {...commonProps} />;

    case 'number':
    case 'integer':
      return (
        <input
          type="number"
          {...commonProps}
          max={(field as { max?: number }).max}
          min={(field as { min?: number }).min}
          step={field.type === 'integer' ? 1 : 'any'}
        />
      );

    case 'color':
      return (
        <input
          type="color"
          {...commonProps}
          className={cn(commonProps.className, 'h-10 p-1')}
        />
      );

    case 'array':
      return <FieldArray field={field as ArrayFieldConfig} />;

    case 'reference':
      return <RelationshipField field={field} />;

    default:
      return <input type="text" {...commonProps} />;
  }
}

export function AutoForm({
  resourceConfig,
  initialData,
  onSubmit,
  isLoading = false,
  error: externalError,
  className,
  mode,
}: AutoFormProps) {
  const visibleFields = useMemo(
    () => resourceConfig.fields.filter((f) => !f.hidden),
    [resourceConfig.fields]
  );

  const schema = useMemo(
    () => generateZodSchema(visibleFields),
    [visibleFields]
  );

  const defaultValues = useMemo(() => {
    const values: Record<string, unknown> = {};
    for (const field of visibleFields) {
      values[field.name] =
        initialData?.[field.name] ?? field.defaultValue ?? '';
    }
    return values;
  }, [visibleFields, initialData]);

  const form = useForm({
    resolver: zodResolver(schema),
    defaultValues,
    mode: 'onBlur',
  });

  const formConfig = resourceConfig.form || {};
  const layout = formConfig.layout || 'vertical';
  const columns = formConfig.columns || 1;

  const handleSubmit = async (data: Record<string, unknown>) => {
    await onSubmit(data);
  };

  const renderFields = () => {
    if (formConfig.groups && formConfig.groups.length > 0) {
      return (
        <div className="space-y-6">
          {formConfig.groups.map((group, index) => {
            const groupFields = visibleFields.filter((f) =>
              group.fields.includes(f.name)
            );

            if (groupFields.length === 0) {
              return null;
            }

            const groupKey = group.label || `group-${index}`;

            if (group.type === 'tab' || group.type === 'collapsible') {
              return (
                <Card key={groupKey}>
                  <CardHeader>
                    <CardTitle className="text-lg">{group.label}</CardTitle>
                  </CardHeader>
                  <CardContent
                    className={cn(
                      'grid gap-4',
                      layout === 'grid' &&
                        `grid-cols-1 md:grid-cols-${group.columns || columns}`
                    )}
                  >
                    {groupFields.map((field) => (
                      <ConditionalField field={field} key={field.name}>
                        <FormField field={field} />
                      </ConditionalField>
                    ))}
                  </CardContent>
                </Card>
              );
            }

            return (
              <div className="space-y-4" key={groupKey}>
                {group.label && (
                  <h3 className="font-medium text-lg">{group.label}</h3>
                )}
                <div
                  className={cn(
                    'grid gap-4',
                    layout === 'grid' &&
                      `grid-cols-1 md:grid-cols-${group.columns || columns}`
                  )}
                >
                  {groupFields.map((field) => (
                    <ConditionalField field={field} key={field.name}>
                      <FormField field={field} />
                    </ConditionalField>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      );
    }

    return (
      <div
        className={cn(
          'space-y-4',
          layout === 'horizontal' && 'grid grid-cols-1 gap-4 md:grid-cols-2',
          layout === 'grid' && `grid grid-cols-1 md:grid-cols-${columns} gap-4`
        )}
      >
        {visibleFields.map((field) => (
          <ConditionalField field={field} key={field.name}>
            <FormField field={field} />
          </ConditionalField>
        ))}
      </div>
    );
  };

  return (
    <FormProvider {...form}>
      <form
        className={cn('space-y-6', className)}
        onSubmit={form.handleSubmit(handleSubmit)}
      >
        {externalError && (
          <Alert variant="destructive">
            <AlertDescription>{externalError}</AlertDescription>
          </Alert>
        )}

        {renderFields()}

        <CardFooter className="px-0 pt-4">
          <Button
            className="min-w-[120px]"
            disabled={isLoading || form.formState.isSubmitting}
            type="submit"
          >
            {mode === 'create' ? 'Create' : 'Save Changes'}
          </Button>
        </CardFooter>
      </form>
    </FormProvider>
  );
}

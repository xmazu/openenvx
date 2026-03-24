'use client';

import { AlertCircle, Loader2 } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import type {
  FieldConfig,
  FormViewConfig,
  ResourceConfig,
} from '@/lib/resource-protocol';
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
import { FieldRenderer } from './field-renderer';

// ============================================================================
// Form Configuration Types
// ============================================================================

export interface AutoFormProps {
  /** Additional CSS classes */
  className?: string;
  /** Error message */
  error?: string | null;
  /** Initial form data (for edit mode) */
  initialData?: Record<string, unknown>;
  /** Loading state */
  isLoading?: boolean;
  /** Form mode */
  mode: 'create' | 'edit';
  /** Form submission handler */
  onSubmit: (data: Record<string, unknown>) => Promise<void>;
  /** Resource configuration */
  resourceConfig: ResourceConfig;
}

// ============================================================================
// Form Field State
// ============================================================================

interface FieldState {
  error?: string;
  touched: boolean;
  value: unknown;
}

interface FormState {
  fields: Record<string, FieldState>;
  isSubmitting: boolean;
  submitError: string | null;
}

// ============================================================================
// Validation Utilities
// ============================================================================

function validateField(field: FieldConfig, value: unknown): string | undefined {
  // Required check
  if (
    field.required &&
    (value === undefined || value === null || value === '')
  ) {
    return `${field.label || field.name} is required`;
  }

  // Skip further validation if value is empty and not required
  if (value === undefined || value === null || value === '') {
    return undefined;
  }

  // Type-specific validation
  switch (field.type) {
    case 'email': {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (typeof value === 'string' && !emailRegex.test(value)) {
        return 'Please enter a valid email address';
      }
      break;
    }
    case 'url': {
      try {
        if (typeof value === 'string') {
          new URL(value);
        }
      } catch {
        return 'Please enter a valid URL';
      }
      break;
    }
    case 'number':
    case 'integer': {
      const numConfig = field as Extract<
        FieldConfig,
        { type: 'number' | 'integer' }
      >;
      const num =
        typeof value === 'string'
          ? Number.parseFloat(value)
          : (value as number);

      if (Number.isNaN(num)) {
        return 'Please enter a valid number';
      }

      if (numConfig.min !== undefined && num < numConfig.min) {
        return `Value must be at least ${numConfig.min}`;
      }

      if (numConfig.max !== undefined && num > numConfig.max) {
        return `Value must be at most ${numConfig.max}`;
      }
      break;
    }
    case 'text': {
      const textConfig = field as Extract<FieldConfig, { type: 'text' }>;
      const str = String(value);

      if (
        textConfig.minLength !== undefined &&
        str.length < textConfig.minLength
      ) {
        return `Must be at least ${textConfig.minLength} characters`;
      }

      if (
        textConfig.maxLength !== undefined &&
        str.length > textConfig.maxLength
      ) {
        return `Must be at most ${textConfig.maxLength} characters`;
      }
      break;
    }
    case 'textarea': {
      const textareaConfig = field as Extract<
        FieldConfig,
        { type: 'textarea' }
      >;
      const str = String(value);

      if (
        textareaConfig.minLength !== undefined &&
        str.length < textareaConfig.minLength
      ) {
        return `Must be at least ${textareaConfig.minLength} characters`;
      }

      if (
        textareaConfig.maxLength !== undefined &&
        str.length > textareaConfig.maxLength
      ) {
        return `Must be at most ${textareaConfig.maxLength} characters`;
      }
      break;
    }
  }

  // Custom validation
  if (field.validate) {
    const result = field.validate(value);
    if (result !== true) {
      return result;
    }
  }

  return undefined;
}

// ============================================================================
// AutoForm Component
// ============================================================================

export function AutoForm({
  resourceConfig,
  initialData,
  onSubmit,
  isLoading = false,
  error: externalError,
  className,
  mode,
}: AutoFormProps) {
  // Get visible fields (not hidden)
  const visibleFields = useMemo(() => {
    if (!resourceConfig.fields) {
      return [];
    }
    return resourceConfig.fields.filter((f) => !f.hidden);
  }, [resourceConfig.fields]);

  // Initialize form state
  const [formState, setFormState] = useState<FormState>(() => {
    const fields: Record<string, FieldState> = {};

    for (const field of visibleFields) {
      const initialValue =
        initialData?.[field.name] ?? field.defaultValue ?? null;
      fields[field.name] = {
        value: initialValue,
        touched: false,
        error: validateField(field, initialValue),
      };
    }

    return {
      fields,
      isSubmitting: false,
      submitError: externalError || null,
    };
  });

  // Update external error
  if (externalError && externalError !== formState.submitError) {
    setFormState((prev) => ({ ...prev, submitError: externalError }));
  }

  // Handle field change
  const handleFieldChange = useCallback(
    (fieldName: string, value: unknown) => {
      setFormState((prev) => {
        const field = visibleFields.find((f) => f.name === fieldName);
        if (!field) {
          return prev;
        }

        return {
          ...prev,
          fields: {
            ...prev.fields,
            [fieldName]: {
              ...prev.fields[fieldName],
              value,
              error: prev.fields[fieldName]?.touched
                ? validateField(field, value)
                : undefined,
            },
          },
        };
      });
    },
    [visibleFields]
  );

  // Handle field blur (mark as touched and validate)
  const handleFieldBlur = useCallback(
    (fieldName: string) => {
      setFormState((prev) => {
        const field = visibleFields.find((f) => f.name === fieldName);
        if (!field) {
          return prev;
        }

        const value = prev.fields[fieldName]?.value;
        return {
          ...prev,
          fields: {
            ...prev.fields,
            [fieldName]: {
              ...prev.fields[fieldName],
              touched: true,
              error: validateField(field, value),
            },
          },
        };
      });
    },
    [visibleFields]
  );

  // Handle form submission
  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();

      // Validate all fields
      const newFields: Record<string, FieldState> = {};
      let hasErrors = false;

      for (const field of visibleFields) {
        const fieldState = formState.fields[field.name];
        const error = validateField(field, fieldState?.value);

        newFields[field.name] = {
          ...fieldState,
          touched: true,
          error,
        };

        if (error) {
          hasErrors = true;
        }
      }

      setFormState((prev) => ({
        ...prev,
        fields: newFields,
        submitError: null,
      }));

      if (hasErrors) {
        return;
      }

      // Build submission data
      const submitData: Record<string, unknown> = {};
      for (const field of visibleFields) {
        const value = formState.fields[field.name]?.value;
        // Don't submit null values for optional fields
        if (value !== undefined && value !== null) {
          submitData[field.name] = value;
        }
      }

      // Submit
      setFormState((prev) => ({ ...prev, isSubmitting: true }));

      try {
        await onSubmit(submitData);
        setFormState((prev) => ({ ...prev, isSubmitting: false }));
      } catch (err) {
        setFormState((prev) => ({
          ...prev,
          isSubmitting: false,
          submitError: err instanceof Error ? err.message : 'An error occurred',
        }));
      }
    },
    [visibleFields, formState.fields, onSubmit]
  );

  // Get form configuration
  const formConfig: FormViewConfig = resourceConfig.form || {};
  const layout = formConfig.layout || 'vertical';
  const columns = formConfig.columns || 1;

  // Render field groups or flat list
  const renderFields = () => {
    // If groups are defined, render them
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

            return (
              <Card key={index}>
                {group.title && (
                  <CardHeader>
                    <CardTitle className="text-lg">{group.title}</CardTitle>
                  </CardHeader>
                )}
                <CardContent
                  className={cn(
                    'grid gap-4',
                    layout === 'grid' &&
                      `grid-cols-1 md:grid-cols-${group.columns || columns}`
                  )}
                >
                  {groupFields.map((field) => (
                    <FieldRenderer
                      disabled={isLoading || formState.isSubmitting}
                      error={formState.fields[field.name]?.error}
                      field={field}
                      key={field.name}
                      onBlur={() => handleFieldBlur(field.name)}
                      onChange={(value) => handleFieldChange(field.name, value)}
                      value={formState.fields[field.name]?.value}
                    />
                  ))}
                </CardContent>
              </Card>
            );
          })}
        </div>
      );
    }

    // Otherwise, render flat list of fields
    return (
      <div
        className={cn(
          'space-y-4',
          layout === 'horizontal' && 'grid grid-cols-1 gap-4 md:grid-cols-2',
          layout === 'grid' && `grid grid-cols-1 md:grid-cols-${columns} gap-4`
        )}
      >
        {visibleFields.map((field) => (
          <FieldRenderer
            disabled={isLoading || formState.isSubmitting}
            error={formState.fields[field.name]?.error}
            field={field}
            key={field.name}
            onBlur={() => handleFieldBlur(field.name)}
            onChange={(value) => handleFieldChange(field.name, value)}
            value={formState.fields[field.name]?.value}
          />
        ))}
      </div>
    );
  };

  return (
    <form className={cn('space-y-6', className)} onSubmit={handleSubmit}>
      {formState.submitError && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{formState.submitError}</AlertDescription>
        </Alert>
      )}

      {renderFields()}

      <CardFooter className="px-0 pt-4">
        <Button
          className="min-w-[120px]"
          disabled={isLoading || formState.isSubmitting}
          type="submit"
        >
          {(isLoading || formState.isSubmitting) && (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          )}
          {mode === 'create' ? 'Create' : 'Save Changes'}
        </Button>
      </CardFooter>
    </form>
  );
}

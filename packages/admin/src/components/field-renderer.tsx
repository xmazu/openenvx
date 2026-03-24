'use client';

import type { ChangeEvent } from 'react';
import { useCallback, useState } from 'react';
import type { FieldConfig, SelectOption } from '@/lib/resource-protocol';
import { cn } from '@/lib/utils';
import { Checkbox } from '@/ui/shadcn/checkbox';
import { Input } from '@/ui/shadcn/input';
import { Label } from '@/ui/shadcn/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/ui/shadcn/select';
import { Textarea } from '@/ui/shadcn/textarea';

// ============================================================================
// Regex patterns (defined at top level for performance)
// ============================================================================

const SLUG_REGEX_REPLACE = /[^a-z0-9-]/g;
const SLUG_REGEX_HYPHENS = /-+/g;
const SLUG_REGEX_TRIM = /^-|-$/g;

// Regex patterns for formatLabel
const LABEL_REGEX_UNDERSCORE = /_/g;
const LABEL_REGEX_CAMEL = /([A-Z])/g;
const LABEL_REGEX_LEADING_SPACE = /^\s+/;
const LABEL_REGEX_EXTRA_SPACES = /\s+/g;

// ============================================================================
// Field Component Props
// ============================================================================

export interface FieldRendererProps {
  disabled?: boolean;
  error?: string;
  field: FieldConfig;
  onBlur?: () => void;
  onChange: (value: unknown) => void;
  value: unknown;
}

// ============================================================================
// Utility Functions
// ============================================================================

function formatLabel(name: string): string {
  return name
    .replace(LABEL_REGEX_UNDERSCORE, ' ')
    .replace(LABEL_REGEX_CAMEL, ' $1')
    .replace(LABEL_REGEX_LEADING_SPACE, '')
    .replace(LABEL_REGEX_EXTRA_SPACES, ' ')
    .toLowerCase()
    .replace(/\b\w/g, (l) => l.toUpperCase());
}

function normalizeOptions(options: SelectOption[] | string[]): SelectOption[] {
  return options.map((opt) =>
    typeof opt === 'string' ? { label: formatLabel(opt), value: opt } : opt
  );
}

// ============================================================================
// Text Field
// ============================================================================

function TextFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const config = field as Extract<FieldConfig, { type: 'text' }>;

  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', config.className)}
      style={{ width: config.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        maxLength={config.maxLength}
        minLength={config.minLength}
        onBlur={onBlur}
        onChange={handleChange}
        pattern={config.pattern}
        placeholder={field.placeholder}
        type="text"
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Textarea Field
// ============================================================================

function TextareaFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const config = field as Extract<FieldConfig, { type: 'textarea' }>;

  const handleChange = useCallback(
    (e: ChangeEvent<HTMLTextAreaElement>) => {
      onChange(e.target.value);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', config.className)}
      style={{ width: config.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Textarea
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        maxLength={config.maxLength}
        minLength={config.minLength}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder}
        rows={config.rows}
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Number Field
// ============================================================================

function NumberFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const config = field as Extract<FieldConfig, { type: 'number' | 'integer' }>;

  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value;
      if (val === '') {
        onChange(null);
      } else {
        const num =
          config.type === 'integer'
            ? Number.parseInt(val, 10)
            : Number.parseFloat(val);
        onChange(Number.isNaN(num) ? null : num);
      }
    },
    [onChange, config.type]
  );

  return (
    <div
      className={cn('space-y-2', config.className)}
      style={{ width: config.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        max={config.max}
        min={config.min}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder}
        step={config.step}
        type="number"
        value={(value as number) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Boolean Field
// ============================================================================

function BooleanFieldRenderer({
  field,
  value,
  onChange,
  error,
  disabled,
}: FieldRendererProps) {
  const handleChange = useCallback(
    (checked: boolean) => {
      onChange(checked);
    },
    [onChange]
  );

  return (
    <div
      className={cn('flex items-start space-x-3 space-y-0', field.className)}
      style={{ width: field.width }}
    >
      <Checkbox
        checked={(value as boolean) ?? false}
        disabled={disabled || field.readOnly}
        id={field.name}
        onCheckedChange={handleChange}
      />
      <div className="space-y-1 leading-none">
        <Label htmlFor={field.name}>
          {field.label || formatLabel(field.name)}
          {field.required && <span className="ml-1 text-destructive">*</span>}
        </Label>
        {field.description && (
          <p className="text-muted-foreground text-sm">{field.description}</p>
        )}
      </div>
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Email Field
// ============================================================================

function EmailFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder}
        type="email"
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Password Field
// ============================================================================

function PasswordFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder}
        type="password"
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Select Field
// ============================================================================

function SelectFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const config = field as Extract<
    FieldConfig,
    { type: 'select' | 'multiselect' }
  >;
  const options = normalizeOptions(config.options);

  const handleChange = useCallback(
    (selectedValue: string) => {
      onChange(selectedValue);
      onBlur?.();
    },
    [onChange, onBlur]
  );

  const currentValue = (value as string) ?? '';

  return (
    <div
      className={cn('space-y-2', config.className)}
      style={{ width: config.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Select
        disabled={disabled || field.readOnly}
        onValueChange={handleChange}
        value={currentValue}
      >
        <SelectTrigger className={cn(error && 'border-destructive')}>
          <SelectValue placeholder={field.placeholder || 'Select...'} />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={String(option.value)} value={String(option.value)}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// JSON Field
// ============================================================================

function JsonFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const [textValue, setTextValue] = useState(() =>
    value !== undefined && value !== null ? JSON.stringify(value, null, 2) : ''
  );
  const [parseError, setParseError] = useState<string | null>(null);

  const handleChange = useCallback(
    (e: ChangeEvent<HTMLTextAreaElement>) => {
      const newText = e.target.value;
      setTextValue(newText);

      try {
        if (newText.trim() === '') {
          onChange(null);
          setParseError(null);
        } else {
          const parsed = JSON.parse(newText);
          onChange(parsed);
          setParseError(null);
        }
      } catch {
        setParseError('Invalid JSON');
      }
    },
    [onChange]
  );

  const handleBlur = useCallback(() => {
    onBlur?.();
  }, [onBlur]);

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Textarea
        className={cn(
          'font-mono text-sm',
          (error || parseError) && 'border-destructive'
        )}
        disabled={disabled || field.readOnly}
        id={field.name}
        onBlur={handleBlur}
        onChange={handleChange}
        placeholder={field.placeholder || '{ "key": "value" }'}
        rows={6}
        value={textValue}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {(error || parseError) && (
        <p className="text-destructive text-sm">{error || parseError}</p>
      )}
    </div>
  );
}

// ============================================================================
// URL Field
// ============================================================================

function UrlFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder || 'https://...'}
        type="url"
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Color Field
// ============================================================================

function ColorFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <div className="flex items-center gap-2">
        <Input
          className={cn('h-10 w-14 p-1', error && 'border-destructive')}
          disabled={disabled || field.readOnly}
          id={field.name}
          onBlur={onBlur}
          onChange={handleChange}
          type="color"
          value={(value as string) ?? '#000000'}
        />
        <Input
          className={cn('flex-1', error && 'border-destructive')}
          disabled={disabled || field.readOnly}
          onChange={handleChange}
          placeholder="#000000"
          type="text"
          value={(value as string) ?? ''}
        />
      </div>
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Phone Field
// ============================================================================

function PhoneFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder || '+1 (555) 000-0000'}
        type="tel"
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Slug Field
// ============================================================================

function SlugFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      // Auto-format slug: lowercase, alphanumeric with hyphens
      const val = e.target.value
        .toLowerCase()
        .replace(SLUG_REGEX_REPLACE, '-')
        .replace(SLUG_REGEX_HYPHENS, '-')
        .replace(SLUG_REGEX_TRIM, '');
      onChange(val);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder || 'my-slug'}
        type="text"
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Date Field
// ============================================================================

function DateFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const config = field as Extract<FieldConfig, { type: 'date' | 'datetime' }>;

  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value || null);
    },
    [onChange]
  );

  const inputType = config.type === 'datetime' ? 'datetime-local' : 'date';

  return (
    <div
      className={cn('space-y-2', config.className)}
      style={{ width: config.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        max={config.max}
        min={config.min}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder}
        type={inputType}
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Reference Field
// ============================================================================

function ReferenceFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  // TODO: Implement reference field with autocomplete/search
  // For now, render as text input
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value || null);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Input
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder || 'Reference ID'}
        type="text"
        value={(value as string) ?? ''}
      />
      <p className="text-muted-foreground text-sm">
        {field.description || 'Enter the ID of the referenced record'}
      </p>
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Rich Text Field (Simple Implementation)
// ============================================================================

function RichTextFieldRenderer({
  field,
  value,
  onChange,
  onBlur,
  error,
  disabled,
}: FieldRendererProps) {
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLTextAreaElement>) => {
      onChange(e.target.value);
    },
    [onChange]
  );

  return (
    <div
      className={cn('space-y-2', field.className)}
      style={{ width: field.width }}
    >
      <Label htmlFor={field.name}>
        {field.label || formatLabel(field.name)}
        {field.required && <span className="ml-1 text-destructive">*</span>}
      </Label>
      <Textarea
        className={cn(error && 'border-destructive')}
        disabled={disabled || field.readOnly}
        id={field.name}
        onBlur={onBlur}
        onChange={handleChange}
        placeholder={field.placeholder}
        rows={8}
        value={(value as string) ?? ''}
      />
      {field.description && (
        <p className="text-muted-foreground text-sm">{field.description}</p>
      )}
      {error && <p className="text-destructive text-sm">{error}</p>}
    </div>
  );
}

// ============================================================================
// Main Field Renderer
// ============================================================================

const FIELD_RENDERERS: Record<string, React.FC<FieldRendererProps>> = {
  text: TextFieldRenderer,
  textarea: TextareaFieldRenderer,
  number: NumberFieldRenderer,
  integer: NumberFieldRenderer,
  boolean: BooleanFieldRenderer,
  date: DateFieldRenderer,
  datetime: DateFieldRenderer,
  email: EmailFieldRenderer,
  password: PasswordFieldRenderer,
  select: SelectFieldRenderer,
  multiselect: SelectFieldRenderer,
  json: JsonFieldRenderer,
  reference: ReferenceFieldRenderer,
  url: UrlFieldRenderer,
  color: ColorFieldRenderer,
  phone: PhoneFieldRenderer,
  slug: SlugFieldRenderer,
  'rich-text': RichTextFieldRenderer,
};

export function FieldRenderer(
  props: FieldRendererProps
): React.ReactElement | null {
  const { field } = props;
  const Renderer = FIELD_RENDERERS[field.type];

  if (!Renderer) {
    console.warn(`Unknown field type: ${field.type}`);
    return null;
  }

  return <Renderer {...props} />;
}

// Re-export types from resource-protocol
export type { FieldConfig, SelectOption } from '@/lib/resource-protocol';

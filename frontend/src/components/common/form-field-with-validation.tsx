'use client';

import * as React from 'react';
import { CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import type { UseFormReturn } from 'react-hook-form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

type ValidationState = 'idle' | 'validating' | 'valid' | 'invalid';

interface FormFieldWithValidationProps {
  label: string;
  name: string;
  value: string;
  onChange: (value: string) => void;
  onBlur?: () => void;
  error?: string;
  validationState?: ValidationState;
  type?: 'text' | 'email' | 'password' | 'number' | 'textarea' | 'select';
  placeholder?: string;
  required?: boolean;
  description?: string;
  options?: Array<{ value: string; label: string }>;
  className?: string;
}

export function FormFieldWithValidation({
  label,
  name,
  value,
  onChange,
  onBlur,
  error,
  validationState = 'idle',
  type = 'text',
  placeholder,
  required = false,
  description,
  options,
  className,
}: FormFieldWithValidationProps) {
  const hasError = error && validationState === 'invalid';
  const isValid = validationState === 'valid' && !error;
  const isValidating = validationState === 'validating';

  const renderInput = () => {
    const baseInputClasses = cn(
      'transition-colors',
      hasError && 'border-red-500 focus-visible:ring-red-500',
      isValid && 'border-green-500 focus-visible:ring-green-500'
    );

    const iconClasses = 'absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4';

    switch (type) {
      case 'textarea':
        return (
          <div className="relative">
            <Textarea
              id={name}
              name={name}
              value={value}
              onChange={(e) => onChange(e.target.value)}
              onBlur={onBlur}
              placeholder={placeholder}
              required={required}
              className={cn(baseInputClasses, className)}
              aria-invalid={hasError ? true : undefined}
              aria-describedby={error ? `${name}-error` : description ? `${name}-description` : undefined}
            />
            {isValidating && (
              <Loader2 className={cn(iconClasses, 'animate-spin text-gray-400')} aria-hidden="true" />
            )}
            {isValid && (
              <CheckCircle2 className={cn(iconClasses, 'text-green-500')} aria-hidden="true" />
            )}
            {hasError && (
              <XCircle className={cn(iconClasses, 'text-red-500')} aria-hidden="true" />
            )}
          </div>
        );

      case 'select':
        return (
          <div className="relative">
            <Select value={value} onValueChange={onChange}>
              <SelectTrigger
                id={name}
                className={cn(baseInputClasses, className)}
                aria-invalid={hasError ? true : undefined}
                aria-describedby={error ? `${name}-error` : description ? `${name}-description` : undefined}
              >
                <SelectValue placeholder={placeholder || 'Select an option'} />
              </SelectTrigger>
              <SelectContent>
                {options?.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {isValidating && (
              <Loader2 className={cn(iconClasses, 'animate-spin text-gray-400 pointer-events-none')} aria-hidden="true" />
            )}
            {isValid && (
              <CheckCircle2 className={cn(iconClasses, 'text-green-500 pointer-events-none')} aria-hidden="true" />
            )}
            {hasError && (
              <XCircle className={cn(iconClasses, 'text-red-500 pointer-events-none')} aria-hidden="true" />
            )}
          </div>
        );

      default:
        return (
          <div className="relative">
            <Input
              id={name}
              name={name}
              type={type}
              value={value}
              onChange={(e) => onChange(e.target.value)}
              onBlur={onBlur}
              placeholder={placeholder}
              required={required}
              className={cn(baseInputClasses, 'pr-10', className)}
              aria-invalid={hasError ? true : undefined}
              aria-describedby={error ? `${name}-error` : description ? `${name}-description` : undefined}
            />
            {isValidating && (
              <Loader2 className={cn(iconClasses, 'animate-spin text-gray-400')} aria-hidden="true" />
            )}
            {isValid && (
              <CheckCircle2 className={cn(iconClasses, 'text-green-500')} aria-hidden="true" />
            )}
            {hasError && (
              <XCircle className={cn(iconClasses, 'text-red-500')} aria-hidden="true" />
            )}
          </div>
        );
    }
  };

  return (
    <div className="space-y-2">
      <Label htmlFor={name} className="flex items-center gap-1">
        {label}
        {required && <span className="text-red-500" aria-label="required">*</span>}
      </Label>
      {description && (
        <p id={`${name}-description`} className="text-sm text-gray-500">
          {description}
        </p>
      )}
      {renderInput()}
      {error && (
        <p
          id={`${name}-error`}
          className="text-sm text-red-600 flex items-center gap-1"
          role="alert"
          aria-live="polite"
        >
          <XCircle className="h-3 w-3" aria-hidden="true" />
          {error}
        </p>
      )}
      {isValid && !error && (
        <p className="text-sm text-green-600 flex items-center gap-1" aria-live="polite">
          <CheckCircle2 className="h-3 w-3" aria-hidden="true" />
          Looks good!
        </p>
      )}
    </div>
  );
}


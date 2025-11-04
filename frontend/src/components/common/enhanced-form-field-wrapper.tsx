/**
 * Enhanced Form Field Wrapper
 * use-form-with-validation과 함께 사용하는 향상된 폼 필드 컴포넌트
 */

'use client';

import * as React from 'react';
import { Controller, useFormContext, Path } from 'react-hook-form';
import { CheckCircle2, XCircle, Loader2, Info } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { ValidationState } from '@/hooks/use-form-with-validation';
import type { FieldValues } from 'react-hook-form';

interface EnhancedFormFieldWrapperProps<TFieldValues extends FieldValues = FieldValues> {
  name: Path<TFieldValues>;
  label: string;
  type?: 'text' | 'email' | 'password' | 'number' | 'textarea' | 'select';
  placeholder?: string;
  required?: boolean;
  description?: string;
  options?: Array<{ value: string; label: string }>;
  className?: string;
  getFieldError?: (fieldName: string) => string | undefined;
  getFieldValidationState?: (fieldName: string) => ValidationState;
}

function EnhancedFormFieldWrapperComponent<TFieldValues extends FieldValues = FieldValues>({
  name,
  label,
  type = 'text',
  placeholder,
  required = false,
  description,
  options,
  className,
  getFieldError,
  getFieldValidationState,
}: EnhancedFormFieldWrapperProps<TFieldValues>) {
  const { control, formState: { errors, touchedFields } } = useFormContext<TFieldValues>();
  const error = errors[name];
  const fieldError = getFieldError ? getFieldError(name as string) : (error?.message as string | undefined);
  // Check if field is touched using type assertion
  const isTouched = Boolean((touchedFields as Record<string, boolean>)[name as string]);
  const validationState = getFieldValidationState
    ? getFieldValidationState(name as string)
    : (error ? 'invalid' : isTouched ? 'valid' : 'idle');

  const hasError = !!error || validationState === 'invalid';
  const isValid = validationState === 'valid' && !error;
  const isValidating = validationState === 'validating';
  const hasValue = control._formValues[name] !== undefined && control._formValues[name] !== '';
  const showValidationIcon = (isTouched && hasValue) || hasError;

  const renderInput = () => {
    const baseInputClasses = cn(
      'transition-colors',
      hasError && 'border-red-500 focus-visible:ring-red-500',
      isValid && showValidationIcon && 'border-green-500 focus-visible:ring-green-500'
    );

    const iconClasses = 'absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 pointer-events-none';

    switch (type) {
      case 'textarea':
        return (
          <div className="relative">
            <Controller
              name={name}
              control={control}
              render={({ field }) => (
                <Textarea
                  {...field}
                  placeholder={placeholder}
                  required={required}
                  className={cn(baseInputClasses, 'pr-10', className)}
                  aria-invalid={hasError ? true : undefined}
                  aria-describedby={
                    fieldError 
                      ? `${name}-error` 
                      : description 
                        ? `${name}-description` 
                        : undefined
                  }
                />
              )}
            />
            {showValidationIcon && (
              <>
                {isValidating && (
                  <Loader2 className={cn(iconClasses, 'animate-spin text-gray-400')} aria-hidden="true" />
                )}
                {hasError && (
                  <XCircle className={cn(iconClasses, 'text-red-500')} aria-hidden="true" />
                )}
                {isValid && (
                  <CheckCircle2 className={cn(iconClasses, 'text-green-500')} aria-hidden="true" />
                )}
              </>
            )}
          </div>
        );

      case 'select':
        return (
          <div className="relative">
            <Controller
              name={name}
              control={control}
              render={({ field }) => (
                <Select value={field.value || ''} onValueChange={field.onChange}>
                  <SelectTrigger
                    className={cn(baseInputClasses, 'pr-10', className)}
                    aria-invalid={hasError}
                    aria-describedby={
                      fieldError 
                        ? `${name}-error` 
                        : description 
                          ? `${name}-description` 
                          : undefined
                    }
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
              )}
            />
            {showValidationIcon && (
              <>
                {isValidating && (
                  <Loader2 className={cn(iconClasses, 'animate-spin text-gray-400')} aria-hidden="true" />
                )}
                {hasError && (
                  <XCircle className={cn(iconClasses, 'text-red-500')} aria-hidden="true" />
                )}
                {isValid && (
                  <CheckCircle2 className={cn(iconClasses, 'text-green-500')} aria-hidden="true" />
                )}
              </>
            )}
          </div>
        );

      default:
        return (
          <div className="relative">
            <Controller
              name={name}
              control={control}
              render={({ field }) => (
                <Input
                  {...field}
                  type={type}
                  placeholder={placeholder}
                  required={required}
                  className={cn(baseInputClasses, 'pr-10', className)}
                  aria-invalid={hasError ? true : undefined}
                  aria-describedby={
                    fieldError 
                      ? `${name}-error` 
                      : description 
                        ? `${name}-description` 
                        : undefined
                  }
                />
              )}
            />
            {showValidationIcon && (
              <>
                {isValidating && (
                  <Loader2 className={cn(iconClasses, 'animate-spin text-gray-400')} aria-hidden="true" />
                )}
                {hasError && (
                  <XCircle className={cn(iconClasses, 'text-red-500')} aria-hidden="true" />
                )}
                {isValid && (
                  <CheckCircle2 className={cn(iconClasses, 'text-green-500')} aria-hidden="true" />
                )}
              </>
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
        <div id={`${name}-description`} className="text-sm text-gray-500 flex items-center gap-1">
          <Info className="h-3 w-3" aria-hidden="true" />
          {description}
        </div>
      )}
      {renderInput()}
      {fieldError && (
        <p
          id={`${name}-error`}
          className="text-sm text-red-600 flex items-center gap-1"
          role="alert"
          aria-live="polite"
        >
          <XCircle className="h-3 w-3" aria-hidden="true" />
          {fieldError}
        </p>
      )}
      {isValid && showValidationIcon && !fieldError && (
        <p className="text-sm text-green-600 flex items-center gap-1" aria-live="polite">
          <CheckCircle2 className="h-3 w-3" aria-hidden="true" />
          Looks good!
        </p>
      )}
    </div>
  );
}

export const EnhancedFormFieldWrapper = React.memo(EnhancedFormFieldWrapperComponent) as typeof EnhancedFormFieldWrapperComponent;


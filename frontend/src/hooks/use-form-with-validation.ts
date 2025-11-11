/**
 * use-form-with-validation
 * 통합 폼 훅 - react-hook-form + zod + 검증 + 에러 처리 + 로딩 상태 통합
 * 
 * 모든 폼의 공통 패턴을 추상화하여 일관된 폼 개발 경험 제공
 * 
 * 사용 예시:
 * ```tsx
 * const {
 *   form,
 *   handleSubmit,
 *   isLoading,
 *   error,
 *   success,
 *   reset,
 *   getFieldError,
 *   getFieldValidationState,
 *   EnhancedField,
 * } = useFormWithValidation({
 *   schema: createVMSchema,
 *   defaultValues: { name: '', provider: '' },
 *   onSubmit: async (data) => {
 *     await vmService.createVM(data);
 *   },
 *   onSuccess: () => {
 *     showToast('VM created successfully');
 *     reset();
 *   },
 *   onError: (error) => {
 *     showToast(error.message, 'error');
 *   },
 * });
 * 
 * return (
 *   <form onSubmit={handleSubmit} className="space-y-4">
 *     <EnhancedField
 *       name="name"
 *       label="VM Name"
 *       required
 *       placeholder="Enter VM name"
 *     />
 *     <Button type="submit" disabled={isLoading}>
 *       {isLoading ? 'Creating...' : 'Create VM'}
 *     </Button>
 *   </form>
 * );
 * ```
 */

'use client';

import { useState, useCallback, useMemo } from 'react';
import { useForm, UseFormReturn, FieldValues, Path, UseFormProps } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

// Re-export ValidationState from use-form-validation
export type ValidationState = 'idle' | 'validating' | 'valid' | 'invalid';

export interface UseFormWithValidationOptions<T extends FieldValues> {
  // Zod schema for validation
  schema: z.ZodSchema<T>;

  // Default form values
  defaultValues?: Partial<T>;

  // Form submission handler
  onSubmit: (data: T) => Promise<unknown> | void;

  // Success callback
  onSuccess?: (data: T, result?: unknown) => void;

  // Error callback
  onError?: (error: Error, data?: T) => void;

  // Form configuration
  formOptions?: Omit<UseFormProps<T>, 'resolver' | 'defaultValues'>;

  // Validation options
  validationOptions?: {
    debounceMs?: number;
    validateOnBlur?: boolean;
    validateOnChange?: boolean;
  };

  // Reset form on success
  resetOnSuccess?: boolean;

  // Custom error message handler
  getErrorMessage?: (error: unknown) => string;
}

export interface UseFormWithValidationReturn<T extends FieldValues> {
  // React Hook Form instance
  form: UseFormReturn<T>;

  // Form submission handler
  handleSubmit: (e?: React.BaseSyntheticEvent) => Promise<void>;

  // Loading state
  isLoading: boolean;

  // Global error message
  error: string | null;

  // Success state
  success: boolean;

  // Reset form
  reset: (values?: Partial<T>) => void;

  // Get field error message
  getFieldError: (fieldName: Path<T> | string) => string | undefined;

  // Get field validation state
  getFieldValidationState: (fieldName: Path<T> | string) => ValidationState;

  // Check if form is valid
  isFormValid: boolean;

  // Check if form is dirty
  isDirty: boolean;

  // Get form values
  watch: UseFormReturn<T>['watch'];

  // Set form value
  setValue: UseFormReturn<T>['setValue'];

  // Enhanced form field component props (to be used with EnhancedFormField)
  fieldProps: {
    control: UseFormReturn<T>['control'];
    formState: UseFormReturn<T>['formState'];
    getFieldError: (fieldName: Path<T> | string) => string | undefined;
    getFieldValidationState: (fieldName: Path<T> | string) => ValidationState;
  };
}

/**
 * 통합 폼 훅
 * react-hook-form + zod + 검증 + 에러 처리 + 로딩 상태를 통합한 폼 훅
 */
export function useFormWithValidation<T extends FieldValues>({
  schema,
  defaultValues,
  onSubmit,
  onSuccess,
  onError,
  formOptions,
  resetOnSuccess = false,
  getErrorMessage,
}: UseFormWithValidationOptions<T>): UseFormWithValidationReturn<T> {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  // Initialize react-hook-form with zod resolver
  // zodResolver expects z.ZodObject, but we accept z.ZodSchema<T> for flexibility
  // Type assertion is safe here as zodResolver can handle any ZodSchema
  const form = useForm<T>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(schema as any),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    defaultValues: defaultValues as any,
    mode: formOptions?.mode || 'onChange',
    ...formOptions,
  });

  // Enhanced validation state tracking
  const getValidationState = useCallback(
    (fieldName: Path<T>): ValidationState => {
      const fieldError = form.formState.errors[fieldName];
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const isTouched = (form.formState.touchedFields as any)[fieldName];
      const fieldValue = form.watch(fieldName);

      if (fieldError) {
        return 'invalid';
      }

      if (isTouched && fieldValue !== undefined && fieldValue !== '') {
        return 'valid';
      }

      return 'idle';
    },
    [form]
  );

  // Get field error message
  const getFieldError = useCallback(
    (fieldName: Path<T> | string): string | undefined => {
      const fieldError = form.formState.errors[fieldName as Path<T>];
      return fieldError?.message as string | undefined;
    },
    [form.formState.errors]
  );

  // Get field validation state
  const getFieldValidationState = useCallback(
    (fieldName: Path<T> | string): ValidationState => {
      return getValidationState(fieldName as Path<T>);
    },
    [getValidationState]
  );

  // Default error message handler
  const defaultGetErrorMessage = useCallback((err: unknown): string => {
    if (err instanceof Error) {
      return err.message;
    }
    if (typeof err === 'string') {
      return err;
    }
    if (
      err &&
      typeof err === 'object' &&
      'response' in err &&
      err.response &&
      typeof err.response === 'object' &&
      'data' in err.response
    ) {
      const data = err.response.data;
      if (data && typeof data === 'object' && 'message' in data) {
        return String(data.message);
      }
      if (data && typeof data === 'object' && 'error' in data) {
        const errorObj = data.error;
        if (errorObj && typeof errorObj === 'object' && 'message' in errorObj) {
          return String(errorObj.message);
        }
      }
    }
    return 'An error occurred. Please try again.';
  }, []);
  
  // Note: defaultGetErrorMessage는 서버에서 반환한 메시지를 그대로 사용하므로
  // 번역이 필요한 경우 useErrorHandler의 getErrorMessage를 사용하세요

  // Form submission handler
  const handleSubmit = useCallback(
    async (e?: React.BaseSyntheticEvent) => {
      e?.preventDefault();

      // Clear previous error and success
      setError(null);
      setSuccess(false);

      // Validate form before submission
      const isValid = await form.trigger();
      
      if (!isValid) {
        return;
      }

      try {
        setIsLoading(true);
        const formData = form.getValues();
        const result = await onSubmit(formData);
        setSuccess(true);

        // Reset form if configured
        if (resetOnSuccess) {
          form.reset(defaultValues as T);
        }

        // Call success callback
        onSuccess?.(formData, result);
      } catch (err: unknown) {
        const errorMessage = getErrorMessage
          ? getErrorMessage(err)
          : defaultGetErrorMessage(err);

        setError(errorMessage);
        setSuccess(false);

        // Call error callback
        onError?.(err instanceof Error ? err : new Error(errorMessage), form.getValues());
      } finally {
        setIsLoading(false);
      }
    },
    [
      form,
      onSubmit,
      onSuccess,
      onError,
      resetOnSuccess,
      defaultValues,
      getErrorMessage,
      defaultGetErrorMessage,
    ]
  );

  // Reset form
  const reset = useCallback(
    (values?: Partial<T>) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      form.reset((values || defaultValues) as any);
      setError(null);
      setSuccess(false);
    },
    [form, defaultValues]
  );

  // Field props for EnhancedFormField
  const fieldProps = useMemo(
    () => ({
      control: form.control,
      formState: form.formState,
      getFieldError,
      getFieldValidationState,
    }),
    [form.control, form.formState, getFieldError, getFieldValidationState]
  );

  return {
    form,
    handleSubmit,
    isLoading,
    error,
    success,
    reset,
    getFieldError,
    getFieldValidationState,
    isFormValid: form.formState.isValid,
    isDirty: form.formState.isDirty,
    watch: form.watch,
    setValue: form.setValue,
    fieldProps,
  };
}

// Re-export EnhancedFormFieldWrapper for convenience
export { EnhancedFormFieldWrapper as EnhancedField } from '@/components/common/enhanced-form-field-wrapper';


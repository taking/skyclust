import { useState, useEffect, useCallback } from 'react';
import { UseFormReturn } from 'react-hook-form';

type ValidationState = 'idle' | 'validating' | 'valid' | 'invalid';

interface UseFormValidationOptions {
  debounceMs?: number;
  validateOnBlur?: boolean;
  validateOnChange?: boolean;
}

/**
 * Hook to provide enhanced form validation with real-time feedback
 */
export function useFormValidation<T extends Record<string, unknown>>(
  form: UseFormReturn<T>,
  options: UseFormValidationOptions = {}
) {
  const { debounceMs = 300, validateOnBlur = true, validateOnChange = true } = options;
  const [validationStates, setValidationStates] = useState<Record<string, ValidationState>>({});

  // Watch all form values for real-time validation
  const watchedValues = form.watch();
  const { errors, trigger } = form;

  // Debounced validation function
  const debouncedValidate = useCallback(
    debounce((fieldName: string) => {
      trigger(fieldName).then((isValid) => {
        setValidationStates((prev) => ({
          ...prev,
          [fieldName]: isValid ? 'valid' : 'invalid',
        }));
      });
    }, debounceMs),
    [trigger, debounceMs]
  );

  // Update validation state based on errors
  useEffect(() => {
    const newStates: Record<string, ValidationState> = {};
    
    Object.keys(watchedValues).forEach((key) => {
      const hasValue = watchedValues[key] !== undefined && watchedValues[key] !== '';
      const hasError = errors[key];
      
      if (hasError) {
        newStates[key] = 'invalid';
      } else if (hasValue) {
        // Only show valid if field has been touched and has no errors
        const isTouched = form.formState.touchedFields[key];
        if (isTouched) {
          newStates[key] = 'valid';
        } else {
          newStates[key] = 'idle';
        }
      } else {
        newStates[key] = 'idle';
      }
    });

    setValidationStates(newStates);
  }, [errors, watchedValues, form.formState.touchedFields]);

  // Handle field change with validation
  const handleFieldChange = useCallback(
    (fieldName: keyof T, value: unknown) => {
      form.setValue(fieldName, value as T[keyof T], {
        shouldValidate: validateOnChange,
      });

      if (validateOnChange) {
        setValidationStates((prev) => ({
          ...prev,
          [fieldName as string]: 'validating',
        }));
        debouncedValidate(fieldName as string);
      }
    },
    [form, validateOnChange, debouncedValidate]
  );

  // Handle field blur with validation
  const handleFieldBlur = useCallback(
    (fieldName: keyof T) => {
      if (validateOnBlur) {
        trigger(fieldName as string).then((isValid) => {
          setValidationStates((prev) => ({
            ...prev,
            [fieldName as string]: isValid ? 'valid' : 'invalid',
          }));
        });
      }
    },
    [trigger, validateOnBlur]
  );

  // Get validation state for a field
  const getValidationState = useCallback(
    (fieldName: keyof T): ValidationState => {
      return validationStates[fieldName as string] || 'idle';
    },
    [validationStates]
  );

  // Check if form is valid
  const isFormValid = form.formState.isValid;

  return {
    validationStates,
    getValidationState,
    handleFieldChange,
    handleFieldBlur,
    isFormValid,
  };
}

// Simple debounce utility
function debounce<T extends (...args: unknown[]) => unknown>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null;
  
  return function executedFunction(...args: Parameters<T>) {
    const later = () => {
      timeout = null;
      func(...args);
    };
    
    if (timeout) {
      clearTimeout(timeout);
    }
    timeout = setTimeout(later, wait);
  };
}


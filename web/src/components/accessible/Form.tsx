import React, { useId } from 'react'
import { TextInput, TextInputProps } from '@mantine/core'

interface AccessibleFormFieldProps extends TextInputProps {
  label: string
  description?: string
  error?: string
  required?: boolean
  'aria-describedby'?: string
  'aria-invalid'?: boolean
  'aria-required'?: boolean
}

export const AccessibleFormField: React.FC<AccessibleFormFieldProps> = ({
  label,
  description,
  error,
  required = false,
  'aria-describedby': ariaDescribedBy,
  'aria-invalid': ariaInvalid,
  'aria-required': ariaRequired,
  id,
  ...props
}) => {
  const generatedId = useId()
  const fieldId = id || generatedId
  const descriptionId = `${fieldId}-description`
  const errorId = `${fieldId}-error`

  const describedBy = [
    ariaDescribedBy,
    description ? descriptionId : null,
    error ? errorId : null,
  ].filter(Boolean).join(' ')

  return (
    <div>
      <label htmlFor={fieldId} style={{ display: 'block', marginBottom: '4px' }}>
        {label}
        {required && <span aria-label="required"> *</span>}
      </label>
      
      {description && (
        <div id={descriptionId} style={{ fontSize: '0.875rem', color: '#666', marginBottom: '4px' }}>
          {description}
        </div>
      )}
      
      <TextInput
        {...props}
        id={fieldId}
        aria-describedby={describedBy || undefined}
        aria-invalid={ariaInvalid || !!error}
        aria-required={ariaRequired || required}
        error={error}
        required={required}
      />
      
      {error && (
        <div 
          id={errorId} 
          role="alert" 
          aria-live="polite"
          style={{ 
            fontSize: '0.875rem', 
            color: '#d63031', 
            marginTop: '4px' 
          }}
        >
          {error}
        </div>
      )}
    </div>
  )
}

// Form validation hook
export const useFormValidation = () => {
  const [errors, setErrors] = React.useState<Record<string, string>>({})
  const [touched, setTouched] = React.useState<Record<string, boolean>>({})

  const validateField = (name: string, value: any, rules: ValidationRule[]) => {
    for (const rule of rules) {
      const error = rule(value)
      if (error) {
        setErrors(prev => ({ ...prev, [name]: error }))
        return error
      }
    }
    
    setErrors(prev => {
      const newErrors = { ...prev }
      delete newErrors[name]
      return newErrors
    })
    
    return null
  }

  const validateForm = (values: Record<string, any>, rules: Record<string, ValidationRule[]>) => {
    const newErrors: Record<string, string> = {}
    let isValid = true

    for (const [field, fieldRules] of Object.entries(rules)) {
      const error = validateField(field, values[field], fieldRules)
      if (error) {
        newErrors[field] = error
        isValid = false
      }
    }

    setErrors(newErrors)
    return isValid
  }

  const markFieldTouched = (name: string) => {
    setTouched(prev => ({ ...prev, [name]: true }))
  }

  const reset = () => {
    setErrors({})
    setTouched({})
  }

  return {
    errors,
    touched,
    validateField,
    validateForm,
    markFieldTouched,
    reset,
  }
}

type ValidationRule = (value: any) => string | null

// Common validation rules
export const validationRules = {
  required: (message = 'This field is required'): ValidationRule => 
    (value) => !value || (typeof value === 'string' && !value.trim()) ? message : null,
  
  email: (message = 'Please enter a valid email address'): ValidationRule => 
    (value) => {
      if (!value) return null
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return !emailRegex.test(value) ? message : null
    },
  
  minLength: (min: number, message?: string): ValidationRule => 
    (value) => {
      if (!value) return null
      const msg = message || `Must be at least ${min} characters long`
      return value.length < min ? msg : null
    },
  
  maxLength: (max: number, message?: string): ValidationRule => 
    (value) => {
      if (!value) return null
      const msg = message || `Must be no more than ${max} characters long`
      return value.length > max ? msg : null
    },
  
  pattern: (regex: RegExp, message: string): ValidationRule => 
    (value) => {
      if (!value) return null
      return !regex.test(value) ? message : null
    },
}

// Accessible form component
interface AccessibleFormProps {
  onSubmit: (values: Record<string, any>) => void
  children: React.ReactNode
  'aria-label'?: string
  'aria-describedby'?: string
}

export const AccessibleForm: React.FC<AccessibleFormProps> = ({
  onSubmit,
  children,
  'aria-label': ariaLabel,
  'aria-describedby': ariaDescribedBy,
}) => {
  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    
    const formData = new FormData(event.currentTarget)
    const values: Record<string, any> = {}
    
    for (const [key, value] of formData.entries()) {
      values[key] = value
    }
    
    onSubmit(values)
  }

  return (
    <form
      onSubmit={handleSubmit}
      aria-label={ariaLabel}
      aria-describedby={ariaDescribedBy}
      noValidate // Let our custom validation handle it
    >
      {children}
    </form>
  )
}

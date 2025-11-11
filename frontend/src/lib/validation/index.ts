/**
 * Validation Module
 * 검증 관련 유틸리티 통합 모듈
 * 
 * 모든 검증 관련 기능을 중앙화하여 일관된 검증 경험 제공
 */

// Validation
export { Validation, validation, validate } from './validation';
export type { ValidationResult } from './validation';
export { useValidation } from './validation-hook';
export type { UseValidationReturn } from './validation-hook';

// Zod 스키마 팩토리
export { createValidationSchemas } from './schemas';
export type { TranslationFunction } from './schemas';

// React Hook Form 통합 훅
export { useFormWithValidation } from '@/hooks/use-form-with-validation';
export type {
  UseFormWithValidationOptions,
  UseFormWithValidationReturn,
  ValidationState,
} from '@/hooks/use-form-with-validation';

// 추가 검증 기능 (선택적)
export { useFormValidation } from '@/hooks/use-form-validation';
export type { UseFormValidationOptions } from '@/hooks/use-form-validation';

// 검증 UI 컴포넌트
export { FormFieldWithValidation } from '@/components/common/form-field-with-validation';
export { EnhancedFormField } from '@/components/common/enhanced-form-field';
export { EnhancedFormFieldWrapper } from '@/components/common/enhanced-form-field-wrapper';


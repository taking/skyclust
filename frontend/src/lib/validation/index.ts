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

/**
 * Validation Utilities
 * 공통 검증 헬퍼 함수
 */
export const ValidationUtils = {
  /**
   * 이메일 형식 검증
   */
  isValidEmail: (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  },

  /**
   * UUID 형식 검증
   */
  isValidUUID: (uuid: string): boolean => {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    return uuidRegex.test(uuid);
  },

  /**
   * CIDR 형식 검증
   */
  isValidCIDR: (cidr: string): boolean => {
    const cidrRegex = /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/;
    if (!cidrRegex.test(cidr)) return false;
    
    const [ip, prefix] = cidr.split('/');
    const prefixNum = parseInt(prefix, 10);
    
    if (prefixNum < 0 || prefixNum > 32) return false;
    
    const parts = ip.split('.').map(Number);
    return parts.every(part => part >= 0 && part <= 255);
  },

  /**
   * 비밀번호 강도 검증
   */
  validatePasswordStrength: (password: string): {
    isValid: boolean;
    errors: string[];
  } => {
    const errors: string[] = [];
    
    if (password.length < 8) {
      errors.push('Password must be at least 8 characters long');
    }
    if (!/[A-Z]/.test(password)) {
      errors.push('Password must contain at least one uppercase letter');
    }
    if (!/[a-z]/.test(password)) {
      errors.push('Password must contain at least one lowercase letter');
    }
    if (!/[0-9]/.test(password)) {
      errors.push('Password must contain at least one number');
    }
    if (!/[^A-Za-z0-9]/.test(password)) {
      errors.push('Password must contain at least one special character');
    }
    
    return {
      isValid: errors.length === 0,
      errors,
    };
  },
} as const;


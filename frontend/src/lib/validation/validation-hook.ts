/**
 * useValidation Hook
 * Validation을 React 훅으로 사용하기 위한 래퍼
 * 
 * 번역 함수를 자동으로 주입하고, validation 스키마를 쉽게 사용할 수 있도록 합니다.
 */

import { useMemo } from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { validation } from './validation';
import type { ValidationResult } from './validation';

/**
 * useValidation 반환 타입
 */
export interface UseValidationReturn {
  /**
   * Validation 스키마 (번역된)
   */
  schemas: ReturnType<typeof validation.getSchemas>;

  /**
   * Zod 스키마로 검증
   */
  validateWithSchema: <T>(
    schema: Parameters<typeof validation.validateWithSchema<T>>[0],
    data: unknown
  ) => ValidationResult<T>;

  /**
   * 안전한 파싱
   */
  safeParse: typeof validation.safeParse;

  /**
   * 이메일 검증
   */
  validateEmail: typeof validation.validateEmail;

  /**
   * UUID 검증
   */
  validateUUID: typeof validation.validateUUID;

  /**
   * CIDR 검증
   */
  validateCIDR: typeof validation.validateCIDR;

  /**
   * 비밀번호 강도 검증
   */
  validatePasswordStrength: typeof validation.validatePasswordStrength;

  /**
   * 필수 필드 검증
   */
  validateRequired: typeof validation.validateRequired;

  /**
   * 여러 필드 검증
   */
  validateFields: typeof validation.validateFields;
}

/**
 * useValidation Hook
 * 
 * Validation을 React 훅으로 사용합니다.
 * 번역 함수를 자동으로 주입하여 번역된 validation 스키마를 제공합니다.
 */
export function useValidation(): UseValidationReturn {
  const { t } = useTranslation();

  // 번역 함수 설정 (한 번만 실행)
  useMemo(() => {
    validation.setTranslationFunction(t);
  }, [t]);

  // Validation 스키마 가져오기
  const schemas = useMemo(() => validation.getSchemas(), []);

  return {
    schemas,
    validateWithSchema: validation.validateWithSchema.bind(validation),
    safeParse: validation.safeParse.bind(validation),
    validateEmail: validation.validateEmail.bind(validation),
    validateUUID: validation.validateUUID.bind(validation),
    validateCIDR: validation.validateCIDR.bind(validation),
    validatePasswordStrength: validation.validatePasswordStrength.bind(validation),
    validateRequired: validation.validateRequired.bind(validation),
    validateFields: validation.validateFields.bind(validation),
  };
}


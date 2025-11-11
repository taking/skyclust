/**
 * Validation
 * 모든 검증 로직을 통합한 단일 인터페이스
 * 
 * schemas.ts의 createValidationSchemas, ValidationUtils, use-form-with-validation을 통합
 */

import * as z from 'zod';
import { createValidationSchemas, type TranslationFunction } from './schemas';
import { ValidationUtils } from './index';

/**
 * Validation 결과
 */
export interface ValidationResult<T = unknown> {
  /**
   * 검증 성공 여부
   */
  valid: boolean;

  /**
   * 에러 메시지 (필드별)
   */
  errors: Record<string, string>;

  /**
   * 전체 에러 메시지
   */
  errorMessage?: string;

  /**
   * 검증된 데이터
   */
  data?: T;
}

/**
 * Validation 클래스
 * 모든 검증 기능을 통합한 클래스
 */
export class Validation {
  private static instance: Validation;
  private translationFunction?: TranslationFunction;
  private schemas?: ReturnType<typeof createValidationSchemas>;

  private constructor() {
    // Singleton pattern
  }

  /**
   * 인스턴스 가져오기
   */
  static getInstance(): Validation {
    if (!Validation.instance) {
      Validation.instance = new Validation();
    }
    return Validation.instance;
  }

  /**
   * 번역 함수 설정
   * 번역 함수를 설정하면 모든 스키마가 번역된 메시지를 사용합니다.
   */
  setTranslationFunction(t: TranslationFunction): void {
    this.translationFunction = t;
    this.schemas = createValidationSchemas(t);
  }

  /**
   * Validation 스키마 가져오기
   */
  getSchemas(): ReturnType<typeof createValidationSchemas> {
    if (!this.schemas) {
      // 번역 함수가 없으면 기본 영어 메시지 사용
      this.schemas = createValidationSchemas((key: string) => key);
    }
    return this.schemas;
  }

  /**
   * Zod 스키마로 검증
   */
  validateWithSchema<T>(
    schema: z.ZodSchema<T>,
    data: unknown
  ): ValidationResult<T> {
    try {
      const validated = schema.parse(data);
      return {
        valid: true,
        errors: {},
        data: validated,
      };
    } catch (error) {
      if (error instanceof z.ZodError) {
        const errors: Record<string, string> = {};
        error.issues.forEach((issue) => {
          const path = issue.path.join('.');
          errors[path] = issue.message;
        });

        return {
          valid: false,
          errors,
          errorMessage: error.issues[0]?.message || 'Validation failed',
        };
      }

      return {
        valid: false,
        errors: {},
        errorMessage: error instanceof Error ? error.message : 'Validation failed',
      };
    }
  }

  /**
   * 안전한 파싱 (에러를 throw하지 않음)
   */
  safeParse<T>(schema: z.ZodSchema<T>, data: unknown): { success: true; data: T } | { success: false; error: z.ZodError } {
    return schema.safeParse(data);
  }

  /**
   * 이메일 검증
   */
  validateEmail(email: string): boolean {
    return ValidationUtils.isValidEmail(email);
  }

  /**
   * UUID 검증
   */
  validateUUID(uuid: string): boolean {
    return ValidationUtils.isValidUUID(uuid);
  }

  /**
   * CIDR 검증
   */
  validateCIDR(cidr: string): boolean {
    return ValidationUtils.isValidCIDR(cidr);
  }

  /**
   * 비밀번호 강도 검증
   */
  validatePasswordStrength(password: string): {
    isValid: boolean;
    errors: string[];
  } {
    return ValidationUtils.validatePasswordStrength(password);
  }

  /**
   * 필수 필드 검증
   */
  validateRequired(fields: Record<string, unknown>): {
    valid: boolean;
    errors: Record<string, string>;
  } {
    const errors: Record<string, string> = {};

    Object.entries(fields).forEach(([key, value]) => {
      if (value === undefined || value === null || value === '') {
        errors[key] = `${key} is required`;
      }
    });

    return {
      valid: Object.keys(errors).length === 0,
      errors,
    };
  }

  /**
   * 여러 필드 검증
   */
  validateFields(
    data: Record<string, unknown>,
    rules: Record<string, (value: unknown) => string | null>
  ): {
    valid: boolean;
    errors: Record<string, string>;
  } {
    const errors: Record<string, string> = {};

    Object.entries(rules).forEach(([field, validator]) => {
      const value = data[field];
      const error = validator(value);
      if (error) {
        errors[field] = error;
      }
    });

    return {
      valid: Object.keys(errors).length === 0,
      errors,
    };
  }
}

/**
 * 편의를 위한 싱글톤 인스턴스 export
 */
export const validation = Validation.getInstance();

/**
 * Validation 헬퍼 함수들
 */
export const validate = {
  /**
   * Zod 스키마로 검증
   */
  withSchema: <T>(schema: z.ZodSchema<T>, data: unknown): ValidationResult<T> => {
    return validation.validateWithSchema(schema, data);
  },

  /**
   * 안전한 파싱
   */
  safeParse: <T>(schema: z.ZodSchema<T>, data: unknown): { success: true; data: T } | { success: false; error: z.ZodError } => {
    return validation.safeParse(schema, data);
  },

  /**
   * 이메일 검증
   */
  email: (email: string): boolean => {
    return validation.validateEmail(email);
  },

  /**
   * UUID 검증
   */
  uuid: (uuid: string): boolean => {
    return validation.validateUUID(uuid);
  },

  /**
   * CIDR 검증
   */
  cidr: (cidr: string): boolean => {
    return validation.validateCIDR(cidr);
  },

  /**
   * 비밀번호 강도 검증
   */
  passwordStrength: (password: string): { isValid: boolean; errors: string[] } => {
    return validation.validatePasswordStrength(password);
  },

  /**
   * 필수 필드 검증
   */
  required: (fields: Record<string, unknown>): { valid: boolean; errors: Record<string, string> } => {
    return validation.validateRequired(fields);
  },

  /**
   * 여러 필드 검증
   */
  fields: (
    data: Record<string, unknown>,
    rules: Record<string, (value: unknown) => string | null>
  ): { valid: boolean; errors: Record<string, string> } => {
    return validation.validateFields(data, rules);
  },
};


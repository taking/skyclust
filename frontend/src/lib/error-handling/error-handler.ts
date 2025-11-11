/**
 * Error Handler
 * 모든 에러 처리 로직을 통합한 중앙화된 에러 핸들러
 * 
 * BaseErrorHandler, error-translations, network-error-messages의 기능을 통합
 */

import { NetworkError, ServerError } from './types';
import { BaseErrorHandler } from './base-handler';
import { getErrorTranslationKey, getErrorCustomMessage } from './translations';
import {
  getVPCCreationErrorMessage,
  getSubnetCreationErrorMessage,
  getVPCDeletionErrorMessage,
} from './network-messages';

/**
 * ErrorHandler 클래스
 * 모든 에러 처리 로직을 통합
 */
export class ErrorHandler {
  /**
   * 에러 메시지 추출
   */
  static extractMessage(error: unknown): string {
    return BaseErrorHandler.extractMessage(error);
  }

  /**
   * 사용자 친화적 에러 메시지 가져오기
   */
  static getUserFriendlyMessage(error: unknown): string {
    return BaseErrorHandler.getUserFriendlyMessage(error);
  }

  /**
   * 재시도 가능 여부 확인
   */
  static isRetryable(error: unknown): boolean {
    return BaseErrorHandler.isRetryable(error);
  }

  /**
   * 에러 번역 키 가져오기
   */
  static getErrorTranslationKey(error: unknown): string {
    return getErrorTranslationKey(error);
  }

  /**
   * 에러 커스텀 메시지 가져오기
   */
  static getErrorCustomMessage(error: unknown): string | null {
    return getErrorCustomMessage(error);
  }

  /**
   * 네트워크 리소스 에러 메시지 가져오기
   */
  static getNetworkErrorMessage(
    error: unknown,
    operation: 'create' | 'delete',
    resourceType: 'VPC' | 'Subnet' | 'SecurityGroup',
    provider?: string
  ): string {
    if (operation === 'create') {
      if (resourceType === 'VPC') {
        return getVPCCreationErrorMessage(error, provider);
      }
      if (resourceType === 'Subnet') {
        return getSubnetCreationErrorMessage(error, provider);
      }
      if (resourceType === 'SecurityGroup') {
        return getSubnetCreationErrorMessage(error, provider); // SecurityGroup도 Subnet과 유사한 에러 처리
      }
    }

    if (operation === 'delete') {
      if (resourceType === 'VPC') {
        return getVPCDeletionErrorMessage(error, provider);
      }
      if (resourceType === 'Subnet') {
        return getVPCDeletionErrorMessage(error, provider); // Subnet 삭제도 VPC와 유사한 에러 처리
      }
      if (resourceType === 'SecurityGroup') {
        return getVPCDeletionErrorMessage(error, provider); // SecurityGroup 삭제도 VPC와 유사한 에러 처리
      }
    }

    return this.getUserFriendlyMessage(error);
  }

  /**
   * IAM 권한 에러인지 확인
   */
  static isIAMPermissionError(error: unknown): boolean {
    const message = this.extractMessage(error).toLowerCase();
    return (
      message.includes('access denied') ||
      message.includes('unauthorized') ||
      message.includes('forbidden') ||
      message.includes('permission denied') ||
      message.includes('iam') ||
      message.includes('policy')
    );
  }

  /**
   * 네트워크 에러인지 확인
   */
  static isNetworkError(error: unknown): boolean {
    return error instanceof NetworkError || BaseErrorHandler.isRetryable(error);
  }

  /**
   * 검증 에러인지 확인
   */
  static isValidationError(error: unknown): boolean {
    if (error instanceof ServerError) {
      return error.code === 'BAD_REQUEST' || error.code === 'VALIDATION_ERROR';
    }
    const message = this.extractMessage(error).toLowerCase();
    return message.includes('validation') || message.includes('invalid');
  }

  /**
   * 에러 처리 (로깅 및 메시지 반환)
   */
  static handleError(error: unknown): {
    message: string;
    translationKey: string;
    customMessage: string | null;
    isRetryable: boolean;
    isNetworkError: boolean;
    isValidationError: boolean;
    isIAMPermissionError: boolean;
  } {
    const message = this.extractMessage(error);
    const translationKey = this.getErrorTranslationKey(error);
    const customMessage = this.getErrorCustomMessage(error);
    const isRetryable = this.isRetryable(error);
    const isNetworkError = this.isNetworkError(error);
    const isValidationError = this.isValidationError(error);
    const isIAMPermissionError = this.isIAMPermissionError(error);

    return {
      message,
      translationKey,
      customMessage,
      isRetryable,
      isNetworkError,
      isValidationError,
      isIAMPermissionError,
    };
  }

  /**
   * 에러 로깅
   */
  static logError(error: unknown, context?: Record<string, unknown>): void {
    BaseErrorHandler.logError(error, context);
  }
}

/**
 * 편의를 위한 인스턴스 export
 */
export const errorHandler = ErrorHandler;


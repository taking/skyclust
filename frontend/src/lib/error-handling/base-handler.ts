/**
 * Base Error Handler
 * 기본 에러 처리 로직 제공
 */

import { NetworkError, ServerError } from './types';
import { getErrorLogger } from './logger';

/**
 * BaseErrorHandler 클래스
 * 기본 에러 처리 로직 제공
 */
export class BaseErrorHandler {
  /**
   * 에러 메시지 추출
   */
  static extractMessage(error: unknown): string {
    if (error instanceof NetworkError) {
      return 'Network connection failed. Please check your internet connection.';
    }

    if (error instanceof ServerError) {
      return error.message;
    }

    if (error instanceof Error) {
      return error.message;
    }

    if (typeof error === 'string') {
      return error;
    }

    if (error && typeof error === 'object' && 'response' in error) {
      const axiosError = error as { response?: { data?: { error?: { message?: string }; message?: string } } };
      return (
        axiosError.response?.data?.error?.message ||
        axiosError.response?.data?.message ||
        'An unexpected error occurred'
      );
    }

    return 'An unexpected error occurred';
  }

  /**
   * 사용자 친화적 에러 메시지 가져오기
   */
  static getUserFriendlyMessage(error: unknown): string {
    if (error instanceof NetworkError) {
      return 'Network connection failed. Please check your internet connection and try again.';
    }

    if (error instanceof ServerError) {
      if (error.status === 401) {
        return 'Authentication failed. Please log in again.';
      }
      if (error.status === 403) {
        return 'You do not have permission to perform this action.';
      }
      if (error.status === 404) {
        return 'The requested resource was not found.';
      }
      if (error.status === 500) {
        return 'An internal server error occurred. Please try again later.';
      }
      return error.message;
    }

    if (error instanceof Error) {
      return error.message;
    }

    return 'An unexpected error occurred. Please try again.';
  }

  /**
   * 재시도 가능 여부 확인
   */
  static isRetryable(error: unknown): boolean {
    if (error instanceof NetworkError) {
      return true;
    }

    if (error instanceof ServerError) {
      // 5xx 에러는 재시도 가능
      if (error.status && error.status >= 500 && error.status < 600) {
        return true;
      }
      // 429 Too Many Requests는 재시도 가능
      if (error.status === 429) {
        return true;
      }
      // 408 Request Timeout은 재시도 가능
      if (error.status === 408) {
        return true;
      }
    }

    return false;
  }

  /**
   * 오프라인 상태 확인
   */
  static isOffline(): boolean {
    if (typeof window === 'undefined') {
      return false;
    }
    return !navigator.onLine;
  }

  /**
   * 에러 로깅
   */
  static logError(error: unknown, context?: Record<string, unknown>): void {
    const logger = getErrorLogger();
    const message = this.extractMessage(error);
    const errorInfo: Record<string, unknown> = {
      message,
      error: error instanceof Error ? error.stack : String(error),
      ...context,
    };

    // ErrorLogger는 log() 메서드만 제공
    const errorObj = error instanceof Error ? error : new Error(message);
    
    if (error instanceof NetworkError) {
      logger.log(errorObj, {
        ...errorInfo,
        level: 'warn',
        errorType: 'NetworkError',
      });
    } else if (error instanceof ServerError) {
      logger.log(errorObj, {
        ...errorInfo,
        level: 'error',
        errorType: 'ServerError',
      });
    } else {
      logger.log(errorObj, {
        ...errorInfo,
        level: 'error',
        errorType: 'UnknownError',
      });
    }
  }
}

/**
 * @deprecated Use BaseErrorHandler.extractMessage() instead
 */
export function extractErrorMessage(error: unknown): string {
  return BaseErrorHandler.extractMessage(error);
}

/**
 * @deprecated Use BaseErrorHandler.getUserFriendlyMessage() instead
 */
export function getUserFriendlyErrorMessage(error: unknown): string {
  return BaseErrorHandler.getUserFriendlyMessage(error);
}

/**
 * @deprecated Use BaseErrorHandler.isRetryable() instead
 */
export function isRetryableError(error: unknown): boolean {
  return BaseErrorHandler.isRetryable(error);
}

/**
 * @deprecated Use BaseErrorHandler.isOffline() instead
 */
export function isOffline(): boolean {
  return BaseErrorHandler.isOffline();
}


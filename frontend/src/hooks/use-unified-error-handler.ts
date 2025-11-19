/**
 * Unified Error Handler Hook
 * 통합 에러 처리 훅
 * 
 * 모든 에러 처리를 통합하여 일관된 로깅, 사용자 알림, Sentry 전송을 제공합니다.
 */

'use client';

import { useCallback } from 'react';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { log } from '@/lib/logging';
import { useTranslation } from '@/hooks/use-translation';
import { useOffline } from './use-offline';

export interface ErrorContext {
  /**
   * 에러가 발생한 작업/기능
   */
  operation?: string;
  
  /**
   * 에러가 발생한 리소스 타입
   */
  resource?: string;
  
  /**
   * 에러가 발생한 컴포넌트/페이지
   */
  source?: string;
  
  /**
   * 추가 컨텍스트 정보
   */
  [key: string]: unknown;
}

export interface UseUnifiedErrorHandlerOptions {
  /**
   * 에러 발생 시 토스트 표시 여부
   */
  showToast?: boolean;
  
  /**
   * 에러 로깅 여부
   */
  logError?: boolean;
  
  /**
   * Sentry 전송 여부
   */
  sendToSentry?: boolean;
}

export interface UseUnifiedErrorHandlerReturn {
  /**
   * 통합 에러 처리 함수
   */
  handleError: (error: unknown, context?: ErrorContext) => void;
  
  /**
   * 사용자 친화적 에러 메시지 가져오기
   */
  getErrorMessage: (error: unknown) => string;
  
  /**
   * 재시도 가능 여부 확인
   */
  canRetry: (error: unknown) => boolean;
}

/**
 * useUnifiedErrorHandler Hook
 * 
 * 통합된 에러 처리 로직을 제공합니다.
 * - 로깅: 개발 환경에서는 console, 프로덕션에서는 Sentry
 * - 사용자 알림: Toast 메시지
 * - 에러 분류: 네트워크, 검증, 권한 등
 * 
 * @example
 * ```tsx
 * const { handleError } = useUnifiedErrorHandler();
 * 
 * try {
 *   await someAsyncOperation();
 * } catch (error) {
 *   handleError(error, { operation: 'createVM', resource: 'VM' });
 * }
 * ```
 */
export function useUnifiedErrorHandler(
  options: UseUnifiedErrorHandlerOptions = {}
): UseUnifiedErrorHandlerReturn {
  const {
    showToast = true,
    logError = true,
    sendToSentry = true,
  } = options;
  
  const { error: showErrorToast } = useToast();
  const { isOffline } = useOffline();
  const { t } = useTranslation();

  /**
   * 통합 에러 처리 함수
   */
  const handleError = useCallback((
    error: unknown,
    context: ErrorContext = {}
  ) => {
    // 1. 에러 분석
    const errorInfo = ErrorHandler.handleError(error);
    
    // 2. 로깅 (옵션이 활성화된 경우)
    if (logError) {
      const logContext: Record<string, unknown> = {
        ...context,
        isRetryable: errorInfo.isRetryable,
        isNetworkError: errorInfo.isNetworkError,
        isValidationError: errorInfo.isValidationError,
        isIAMPermissionError: errorInfo.isIAMPermissionError,
        translationKey: errorInfo.translationKey,
      };
      
      // 개발 환경: console 로깅
      // 프로덕션: Sentry 전송 (log.error가 자동으로 처리)
      if (errorInfo.isNetworkError || errorInfo.isRetryable) {
        log.warn(errorInfo.message, error, logContext);
      } else {
        log.error(errorInfo.message, error, logContext);
      }
    }

    // 3. 오프라인 상태 처리
    if (isOffline) {
      if (showToast) {
        showErrorToast(t('errors.noInternetConnection') || 'No internet connection');
      }
      return;
    }

    // 4. 사용자 알림 (옵션이 활성화된 경우)
    if (showToast) {
      // 커스텀 메시지가 있으면 우선 사용
      const message = errorInfo.customMessage || 
                     t(errorInfo.translationKey) || 
                     errorInfo.message;
      
      showErrorToast(message);
    }
  }, [showToast, logError, sendToSentry, isOffline, showErrorToast, t]);

  /**
   * 사용자 친화적 에러 메시지 가져오기
   */
  const getErrorMessage = useCallback((error: unknown): string => {
    const errorInfo = ErrorHandler.handleError(error);
    return errorInfo.customMessage || 
           t(errorInfo.translationKey) || 
           errorInfo.message;
  }, [t]);

  /**
   * 재시도 가능 여부 확인
   */
  const canRetry = useCallback((error: unknown): boolean => {
    const errorInfo = ErrorHandler.handleError(error);
    return errorInfo.isRetryable && !isOffline;
  }, [isOffline]);

  return {
    handleError,
    getErrorMessage,
    canRetry,
  };
}


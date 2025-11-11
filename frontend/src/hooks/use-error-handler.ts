'use client';

/**
 * useErrorHandler Hook
 * 에러 처리를 위한 React 훅
 * 
 * 에러 로깅, 사용자 알림, 재시도 로직 등을 통합적으로 관리
 */

import React, { useCallback } from 'react';
import { useToast } from '@/hooks/use-toast';
import { 
  ErrorHandler,
  getErrorTranslationKey,
  getErrorCustomMessage,
} from '@/lib/error-handling';
import { API } from '@/lib/constants';
import { useOffline } from './use-offline';
import { useTranslation } from './use-translation';

export interface UseErrorHandlerOptions {
  /**
   * 에러 발생 시 토스트 표시 여부
   */
  showToast?: boolean;
  
  /**
   * 자동 재시도 여부
   */
  autoRetry?: boolean;
  
  /**
   * 최대 재시도 횟수
   */
  maxRetries?: number;
  
  /**
   * 재시도 간격 (밀리초)
   */
  retryDelay?: number;
  
  /**
   * 에러 로깅 함수
   */
  onLogError?: (error: unknown, context?: Record<string, unknown>) => void;
}

export interface UseErrorHandlerReturn {
  /**
   * 에러 처리 함수
   */
  handleError: (error: unknown, context?: Record<string, unknown>) => void;
  
  /**
   * 에러 처리 및 재시도 가능 여부 반환
   */
  handleErrorWithRetry: (
    error: unknown,
    retryFn: () => Promise<unknown> | unknown,
    context?: Record<string, unknown>
  ) => Promise<boolean>;
  
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
 * useErrorHandler Hook
 * 
 * 통합된 에러 처리 로직을 제공합니다.
 * 
 * @example
 * ```tsx
 * const { handleError } = useErrorHandler();
 * 
 * try {
 *   await someAsyncOperation();
 * } catch (error) {
 *   handleError(error, { operation: 'createVM' });
 * }
 * ```
 */
export function useErrorHandler(options: UseErrorHandlerOptions = {}): UseErrorHandlerReturn {
  const {
    showToast = true,
    autoRetry = false,
    maxRetries = API.REQUEST.MAX_RETRIES,
    retryDelay = API.REQUEST.RETRY_DELAY,
    onLogError,
  } = options;
  
  const { error: showErrorToast, success: showSuccessToast } = useToast();
  const { isOffline } = useOffline();
  const { t } = useTranslation();

  /**
   * 에러 처리 함수
   */
  const handleError = useCallback((
    error: unknown,
    context?: Record<string, unknown>
  ) => {
    // 1. 에러 로깅: 커스텀 로깅 함수가 있으면 사용, 없으면 기본 ErrorHandler 사용
    if (onLogError) {
      onLogError(error, context);
    } else {
      // 기본 로깅 (ErrorHandler 사용)
      ErrorHandler.logError(error, context);
    }

    // 2. 오프라인 상태이면 별도 처리 (네트워크 에러 메시지 표시)
    if (isOffline) {
      if (showToast) {
        showErrorToast(t('errors.noInternetConnection'));
      }
      return;
    }

    // 3. 번역된 에러 메시지 가져오기
    // context에 customMessage가 있으면 우선 사용
    const contextCustomMessage = context?.customMessage as string | undefined;
    const customMessage = contextCustomMessage || getErrorCustomMessage(error);
    const translationKey = getErrorTranslationKey(error);
    const message = contextCustomMessage || customMessage || t(translationKey);

    // 4. 토스트 표시 (옵션이 활성화된 경우)
    if (showToast) {
      // customMessage가 있으면 줄바꿈을 포함한 메시지를 표시
      if (contextCustomMessage && contextCustomMessage.includes('\n')) {
        // 여러 줄 메시지는 React.createElement를 사용하여 표시
        const lines = contextCustomMessage.split('\n').filter(line => line.trim());
        const errorContent = React.createElement(
          'div',
          { className: 'space-y-1' },
          lines.map((line, index) =>
            React.createElement(
              'div',
              { key: index, className: index === 0 ? 'font-semibold' : '' },
              line
            )
          )
        );
        showErrorToast(errorContent);
      } else {
        showErrorToast(message);
      }
    }
  }, [showToast, isOffline, showErrorToast, onLogError, t]);

  /**
   * 에러 처리 및 재시도 함수
   */
  const handleErrorWithRetry = useCallback(async (
    error: unknown,
    retryFn: () => Promise<unknown> | unknown,
    context?: Record<string, unknown>
  ): Promise<boolean> => {
    // 1. 먼저 에러 처리 (로깅 및 토스트 표시)
    handleError(error, context);

    // 2. 재시도 가능 여부 확인
    // - autoRetry가 비활성화되었거나
    // - 에러가 재시도 불가능하거나
    // - 오프라인 상태인 경우 재시도하지 않음
    if (!autoRetry || !ErrorHandler.isRetryable(error) || isOffline) {
      return false;
    }

    // 3. 재시도 실행 (지수 백오프: 각 시도마다 대기 시간 증가)
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      // 지수 백오프: 첫 시도는 retryDelay, 두 번째는 retryDelay * 2, ...
      await new Promise(resolve => setTimeout(resolve, retryDelay * attempt));

      try {
        // 4. 재시도 함수 실행
        await retryFn();
        
        // 5. 성공 시 성공 메시지 표시
        if (showToast) {
          showSuccessToast(t('messages.operationSuccess'));
        }
        
        return true;
      } catch (retryError) {
        // 6. 마지막 시도에서도 실패하면 에러 처리하고 종료
        if (attempt === maxRetries) {
          handleError(retryError, { ...context, retryAttempt: attempt });
          return false;
        }
        // 중간 시도 실패는 계속 진행
      }
    }

    return false;
  }, [autoRetry, maxRetries, retryDelay, showToast, isOffline, handleError, showSuccessToast, t]);

  /**
   * 사용자 친화적 에러 메시지 가져오기 (번역 적용)
   */
  const getErrorMessage = useCallback((error: unknown): string => {
    // 1. 커스텀 메시지 확인 (에러 객체에 직접 포함된 메시지)
    const customMessage = getErrorCustomMessage(error);
    
    // 2. 번역 키 가져오기
    const translationKey = getErrorTranslationKey(error);
    
    // 3. 커스텀 메시지가 있으면 우선 사용, 없으면 번역된 메시지 반환
    return customMessage || t(translationKey);
  }, [t]);

  /**
   * 재시도 가능 여부 확인
   */
  const canRetry = useCallback((error: unknown): boolean => {
    // 1. 에러가 재시도 가능한 타입인지 확인
    // 2. 오프라인 상태가 아닌지 확인
    // 두 조건을 모두 만족해야 재시도 가능
    return ErrorHandler.isRetryable(error) && !isOffline;
  }, [isOffline]);

  return {
    handleError,
    handleErrorWithRetry,
    getErrorMessage,
    canRetry,
  };
}


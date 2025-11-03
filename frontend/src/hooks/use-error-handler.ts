/**
 * useErrorHandler Hook
 * 에러 처리를 위한 React 훅
 * 
 * 에러 로깅, 사용자 알림, 재시도 로직 등을 통합적으로 관리
 */

import { useCallback } from 'react';
import { useToast } from '@/hooks/use-toast';
import { 
  getUserFriendlyErrorMessage, 
  extractErrorMessage,
  isRetryableError,
  NetworkError,
  ServerError,
} from '@/lib/error-handler';
import { useOffline } from './use-offline';

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
    maxRetries = 3,
    retryDelay = 1000,
    onLogError,
  } = options;
  
  const { error: showErrorToast, success: showSuccessToast } = useToast();
  const { isOffline } = useOffline();

  /**
   * 에러 처리 함수
   */
  const handleError = useCallback((
    error: unknown,
    context?: Record<string, unknown>
  ) => {
    // 에러 로깅
    if (onLogError) {
      onLogError(error, context);
    } else {
      // 기본 로깅
      console.error('Error handled:', error, context);
    }

    // 오프라인 상태이면 별도 처리
    if (isOffline) {
      if (showToast) {
        showErrorToast('No internet connection. Please check your network.');
      }
      return;
    }

    // 사용자 친화적 메시지 가져오기
    const message = getUserFriendlyErrorMessage(error);

    // 토스트 표시
    if (showToast) {
      showErrorToast(message);
    }
  }, [showToast, isOffline, showErrorToast, onLogError]);

  /**
   * 에러 처리 및 재시도 함수
   */
  const handleErrorWithRetry = useCallback(async (
    error: unknown,
    retryFn: () => Promise<unknown> | unknown,
    context?: Record<string, unknown>
  ): Promise<boolean> => {
    // 에러 처리
    handleError(error, context);

    // 재시도 가능 여부 확인
    if (!autoRetry || !isRetryableError(error) || isOffline) {
      return false;
    }

    // 재시도 실행
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      await new Promise(resolve => setTimeout(resolve, retryDelay * attempt));

      try {
        await retryFn();
        
        if (showToast) {
          showSuccessToast('Operation completed successfully after retry.');
        }
        
        return true;
      } catch (retryError) {
        // 마지막 시도에서도 실패하면 에러 처리
        if (attempt === maxRetries) {
          handleError(retryError, { ...context, retryAttempt: attempt });
          return false;
        }
      }
    }

    return false;
  }, [autoRetry, maxRetries, retryDelay, showToast, isOffline, handleError, showSuccessToast]);

  /**
   * 사용자 친화적 에러 메시지 가져오기
   */
  const getErrorMessage = useCallback((error: unknown): string => {
    return getUserFriendlyErrorMessage(error);
  }, []);

  /**
   * 재시도 가능 여부 확인
   */
  const canRetry = useCallback((error: unknown): boolean => {
    return isRetryableError(error) && !isOffline;
  }, [isOffline]);

  return {
    handleError,
    handleErrorWithRetry,
    getErrorMessage,
    canRetry,
  };
}


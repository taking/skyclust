/**
 * Error Handling Module
 * 통합 에러 처리 모듈
 * 
 * 모든 에러 처리 관련 기능을 중앙화하여 export
 */

// 타입 정의
export type { ApiError, ErrorLog } from './types';
export { NetworkError, ServerError } from './types';

// 핵심 클래스
export { ErrorHandler, errorHandler } from './error-handler';
export { BaseErrorHandler } from './base-handler';
export { getErrorLogger, logError } from './logger';

// 번역 관련
export { getErrorTranslationKey, getErrorCustomMessage } from './translations';

// 네트워크 리소스 에러 메시지
export {
  getVPCCreationErrorMessage,
  getSubnetCreationErrorMessage,
  getVPCDeletionErrorMessage,
} from './network-messages';

// 하위 호환성: deprecated 함수들
export {
  extractErrorMessage,
  getUserFriendlyErrorMessage,
  isRetryableError,
  isOffline,
} from './base-handler';

/**
 * Error Message Translation Keys
 * 에러 메시지를 번역 키로 매핑하는 유틸리티
 */

import { NetworkError, ServerError } from './types';

/**
 * 에러 메시지의 번역 키를 반환
 * ErrorHandler.getUserFriendlyMessage()의 반환값을 번역 키로 매핑
 */
export function getErrorTranslationKey(error: unknown): string {
  if (error instanceof NetworkError) {
    return 'errors.unableToConnect';
  }

  if (error instanceof ServerError) {
    switch (error.status) {
      case 400:
        return error.message ? 'errors.generic' : 'errors.invalidRequest';
      case 401:
        return 'errors.sessionExpired';
      case 403:
        return 'errors.noPermission';
      case 404:
        return 'errors.resourceNotFound';
      case 409:
        return 'errors.resourceConflict';
      case 422:
        return error.message ? 'errors.generic' : 'errors.validationFailed';
      case 429:
        return 'errors.tooManyRequests';
      case 500:
        return 'errors.serverErrorTryLater';
      case 502:
      case 503:
      case 504:
        return 'errors.serviceUnavailable';
      default:
        return error.message ? 'errors.generic' : 'errors.errorProcessingRequest';
    }
  }

  // Check if it's an axios error with response
  if (error && typeof error === 'object' && 'response' in error) {
    const axiosError = error as { response?: { status?: number; data?: { error?: { message?: string }; message?: string } } };
    
    if (axiosError.response?.status === 401) {
      return 'errors.sessionExpired';
    }

    if (axiosError.response?.status === 403) {
      return 'errors.noPermission';
    }

    if (axiosError.response?.status === 404) {
      return 'errors.resourceNotFound';
    }

    if (axiosError.response?.status === 429) {
      return 'errors.tooManyRequests';
    }

    if (axiosError.response?.status && axiosError.response.status >= 500) {
      return 'errors.serverErrorTryLater';
    }

    // 서버에서 반환한 메시지가 있으면 그대로 사용 (이미 번역되어 있을 수 있음)
    if (axiosError.response?.data?.error?.message || axiosError.response?.data?.message) {
      return 'errors.generic';
    }

    return 'errors.errorProcessingRequest';
  }

  // Check if it's a network error (no response)
  if (error && typeof error === 'object' && 'message' in error && !('response' in error)) {
    const err = error as { message?: string };
    if (err.message?.includes('Network Error') || err.message?.includes('timeout')) {
      return 'errors.unableToConnect';
    }
  }

  return 'errors.unexpectedError';
}

/**
 * 에러에서 직접 번역이 필요한 메시지를 추출
 * 서버에서 반환한 커스텀 메시지가 있으면 그대로 반환
 */
export function getErrorCustomMessage(error: unknown): string | null {
  if (error instanceof ServerError) {
    // 400, 422 등은 서버 메시지를 그대로 사용할 수 있음
    if (error.status === 400 || error.status === 422) {
      return error.message || null;
    }
  }

  if (error && typeof error === 'object' && 'response' in error) {
    const axiosError = error as { response?: { data?: { error?: { message?: string }; message?: string } } };
    return axiosError.response?.data?.error?.message || axiosError.response?.data?.message || null;
  }

  return null;
}


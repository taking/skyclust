/**
 * Error handling utilities for consistent error messages and handling
 */

import { getErrorLogger } from './error-logger';

export interface ApiError {
  message: string;
  code?: string;
  status?: number;
  details?: Record<string, unknown>;
}

export class NetworkError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'NetworkError';
  }
}

export class ServerError extends Error {
  status: number;
  code?: string;
  details?: Record<string, unknown>;

  constructor(message: string, status: number, code?: string, details?: Record<string, unknown>) {
    super(message);
    this.name = 'ServerError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

/**
 * ErrorHandler 클래스
 * 통합된 에러 처리 로직 제공
 */
export class ErrorHandler {
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
      return 'Unable to connect to the server. Please check your internet connection and try again.';
    }

    if (error instanceof ServerError) {
      switch (error.status) {
        case 400:
          return error.message || 'Invalid request. Please check your input and try again.';
        case 401:
          return 'Your session has expired. Please log in again.';
        case 403:
          return 'You do not have permission to perform this action.';
        case 404:
          return 'The requested resource was not found.';
        case 409:
          return 'This resource already exists or conflicts with another resource.';
        case 422:
          return error.message || 'Validation failed. Please check your input.';
        case 429:
          return 'Too many requests. Please wait a moment and try again.';
        case 500:
          return 'Server error. Please try again later.';
        case 502:
        case 503:
        case 504:
          return 'Service temporarily unavailable. Please try again later.';
        default:
          return error.message || 'An error occurred while processing your request.';
      }
    }

    // Check if it's an axios error with response
    if (error && typeof error === 'object' && 'response' in error) {
      const axiosError = error as { response?: { status?: number; data?: { error?: { message?: string }; message?: string } } };
      
      if (axiosError.response?.status === 401) {
        return 'Your session has expired. Please log in again.';
      }

      if (axiosError.response?.status === 403) {
        return 'You do not have permission to perform this action.';
      }

      if (axiosError.response?.status === 404) {
        return 'The requested resource was not found.';
      }

      if (axiosError.response?.status === 429) {
        return 'Too many requests. Please wait a moment and try again.';
      }

      if (axiosError.response?.status && axiosError.response.status >= 500) {
        return 'Server error. Please try again later.';
      }

      return (
        axiosError.response?.data?.error?.message ||
        axiosError.response?.data?.message ||
        'An error occurred while processing your request.'
      );
    }

    // Check if it's a network error (no response)
    if (error && typeof error === 'object' && 'message' in error && !('response' in error)) {
      const err = error as { message?: string };
      if (err.message?.includes('Network Error') || err.message?.includes('timeout')) {
        return 'Unable to connect to the server. Please check your internet connection.';
      }
    }

    return this.extractMessage(error);
  }

  /**
   * 재시도 가능 여부 확인
   */
  static isRetryable(error: unknown): boolean {
    if (error instanceof NetworkError) {
      return true;
    }

    if (error instanceof ServerError) {
      return error.status >= 500 || error.status === 429;
    }

    if (error && typeof error === 'object' && 'response' in error) {
      const axiosError = error as { response?: { status?: number } };
      const status = axiosError.response?.status;

      if (!status) {
        return true;
      }

      return status >= 500 || status === 429;
    }

    if (error && typeof error === 'object' && 'message' in error && !('response' in error)) {
      return true;
    }

    return false;
  }

  /**
   * 에러 로깅
   */
  static logError(error: unknown, context?: Record<string, unknown>): void {
    getErrorLogger().log(error, context);
  }
}

/**
 * @deprecated Use ErrorHandler.extractMessage() instead
 */
export function extractErrorMessage(error: unknown): string {
  return ErrorHandler.extractMessage(error);
}

/**
 * @deprecated Use ErrorHandler.getUserFriendlyMessage() instead
 */
export function getUserFriendlyErrorMessage(error: unknown): string {
  return ErrorHandler.getUserFriendlyMessage(error);
}

/**
 * @deprecated Use ErrorHandler.isRetryable() instead
 */
export function isRetryableError(error: unknown): boolean {
  return ErrorHandler.isRetryable(error);
}

/**
 * Check if user is offline
 */
export function isOffline(): boolean {
  return typeof navigator !== 'undefined' && !navigator.onLine;
}


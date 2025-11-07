/**
 * Base Service
 * 모든 서비스 레이어의 공통 기능을 제공하는 베이스 클래스
 * 통일된 에러 처리 및 API 호출 패턴 제공
 */

import api from './api';
import type { ApiResponse } from './types/api';
import { ServiceError, type ServiceRequestOptions } from './types/service';
import { getApiUrl, getApiVersion } from './api-config';
import { ErrorHandler } from './error-handler';

export abstract class BaseService {
  /**
   * 공통 API 요청 메서드
   * 모든 HTTP 메서드에 대한 통일된 에러 처리 및 응답 파싱
   */
  protected async request<T>(
    method: 'get' | 'post' | 'put' | 'delete' | 'patch',
    url: string,
    data?: unknown,
    options?: ServiceRequestOptions
  ): Promise<T> {
    try {
      const config = {
        timeout: options?.timeout,
        headers: options?.headers,
        signal: options?.signal,
      };

      let response;
      switch (method) {
        case 'get':
          response = await api.get<ApiResponse<T>>(url, config);
          break;
        case 'post':
          response = await api.post<ApiResponse<T>>(url, data, config);
          break;
        case 'put':
          response = await api.put<ApiResponse<T>>(url, data, config);
          break;
        case 'patch':
          response = await api.patch<ApiResponse<T>>(url, data, config);
          break;
        case 'delete':
          // DELETE 요청도 body를 받을 수 있도록 config에 data 포함
          if (data !== undefined) {
            response = await api.delete<ApiResponse<T>>(url, { ...config, data });
          } else {
            response = await api.delete<ApiResponse<T>>(url, config);
          }
          break;
      }

      // 응답 검증
      if (!response.data) {
        throw new ServiceError(
          'No response data received from server',
          'NO_DATA',
          response.status
        );
      }

      // 성공 응답 처리
      if (response.data.success) {
        if (response.data.data === undefined) {
          throw new ServiceError(
            'Success response but no data field',
            'NO_DATA_FIELD',
            response.status
          );
        }
        return response.data.data;
      }

      // 실패 응답 처리
      const errorMessage =
        response.data.error ||
        response.data.message ||
        `Request failed with status ${response.status}`;
      const errorCode = response.data.code;

      throw new ServiceError(errorMessage, errorCode, response.status, response.data);
    } catch (error) {
      // ServiceError는 그대로 전파
      if (error instanceof ServiceError) {
        // ServiceError는 이미 처리된 에러이므로 로깅만 수행
        ErrorHandler.logError(error, { method, url, service: this.constructor.name });
        throw error;
      }

      // Axios 에러 처리
      if (error && typeof error === 'object' && 'response' in error) {
        const axiosError = error as { response?: { status?: number; data?: unknown } };
        const status = axiosError.response?.status;
        const responseData = axiosError.response?.data as
          | { error?: { message?: string; code?: string }; message?: string }
          | undefined;

        const message =
          responseData?.error?.message ||
          responseData?.message ||
          `Request failed with status ${status || 'unknown'}`;
        const code = responseData?.error?.code;

        const serviceError = new ServiceError(message, code, status, responseData);
        ErrorHandler.logError(serviceError, { method, url, service: this.constructor.name });
        throw serviceError;
      }

      // 네트워크 에러 등 기타 에러
      if (error instanceof Error) {
        const serviceError = new ServiceError(error.message, 'NETWORK_ERROR', undefined, error);
        ErrorHandler.logError(serviceError, { method, url, service: this.constructor.name });
        throw serviceError;
      }

      const unknownError = new ServiceError('Unknown error occurred', 'UNKNOWN_ERROR');
      ErrorHandler.logError(unknownError, { method, url, service: this.constructor.name, originalError: error });
      throw unknownError;
    }
  }

  /**
   * API URL 생성 헬퍼
   * 버전 관리를 포함한 URL 생성
   */
  protected buildApiUrl(endpoint: string, version?: string): string {
    return getApiUrl(endpoint, version);
  }

  /**
   * 현재 API 버전 가져오기
   */
  protected getApiVersion(): string {
    return getApiVersion();
  }

  /**
   * GET 요청 헬퍼
   */
  protected async get<T>(endpoint: string, options?: ServiceRequestOptions & { version?: string }): Promise<T> {
    const url = this.buildApiUrl(endpoint, options?.version);
    const { version: _version, ...requestOptions } = options || {};
    return this.request<T>('get', url, undefined, requestOptions);
  }

  /**
   * POST 요청 헬퍼
   */
  protected async post<T>(
    endpoint: string,
    data?: unknown,
    options?: ServiceRequestOptions & { version?: string }
  ): Promise<T> {
    const url = this.buildApiUrl(endpoint, options?.version);
    const { version: _version, ...requestOptions } = options || {};
    return this.request<T>('post', url, data, requestOptions);
  }

  /**
   * PUT 요청 헬퍼
   */
  protected async put<T>(
    endpoint: string,
    data?: unknown,
    options?: ServiceRequestOptions & { version?: string }
  ): Promise<T> {
    const url = this.buildApiUrl(endpoint, options?.version);
    const { version: _version, ...requestOptions } = options || {};
    return this.request<T>('put', url, data, requestOptions);
  }

  /**
   * PATCH 요청 헬퍼
   */
  protected async patch<T>(
    endpoint: string,
    data?: unknown,
    options?: ServiceRequestOptions & { version?: string }
  ): Promise<T> {
    const url = this.buildApiUrl(endpoint, options?.version);
    const { version: _version, ...requestOptions } = options || {};
    return this.request<T>('patch', url, data, requestOptions);
  }

  /**
   * DELETE 요청 헬퍼
   */
  protected async delete<T>(
    endpoint: string,
    data?: unknown,
    options?: ServiceRequestOptions & { version?: string }
  ): Promise<T> {
    const url = this.buildApiUrl(endpoint, options?.version);
    const { version: _version, ...requestOptions } = options || {};
    return this.request<T>('delete', url, data, requestOptions);
  }
}


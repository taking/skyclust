/**
 * Base Service
 * 모든 서비스 레이어의 공통 기능을 제공하는 베이스 클래스
 * 통일된 에러 처리 및 API 호출 패턴 제공
 */

import api from './api';
import type { ApiResponse } from './types/api';
import { ServiceError, type ServiceRequestOptions } from './types/service';

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
          response = await api.delete<ApiResponse<T>>(url, config);
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

        throw new ServiceError(message, code, status, responseData);
      }

      // 네트워크 에러 등 기타 에러
      if (error instanceof Error) {
        throw new ServiceError(error.message, 'NETWORK_ERROR', undefined, error);
      }

      throw new ServiceError('Unknown error occurred', 'UNKNOWN_ERROR');
    }
  }

  /**
   * GET 요청 헬퍼
   */
  protected async get<T>(url: string, options?: ServiceRequestOptions): Promise<T> {
    return this.request<T>('get', url, undefined, options);
  }

  /**
   * POST 요청 헬퍼
   */
  protected async post<T>(
    url: string,
    data?: unknown,
    options?: ServiceRequestOptions
  ): Promise<T> {
    return this.request<T>('post', url, data, options);
  }

  /**
   * PUT 요청 헬퍼
   */
  protected async put<T>(
    url: string,
    data?: unknown,
    options?: ServiceRequestOptions
  ): Promise<T> {
    return this.request<T>('put', url, data, options);
  }

  /**
   * PATCH 요청 헬퍼
   */
  protected async patch<T>(
    url: string,
    data?: unknown,
    options?: ServiceRequestOptions
  ): Promise<T> {
    return this.request<T>('patch', url, data, options);
  }

  /**
   * DELETE 요청 헬퍼
   */
  protected async delete<T>(url: string, options?: ServiceRequestOptions): Promise<T> {
    return this.request<T>('delete', url, undefined, options);
  }
}


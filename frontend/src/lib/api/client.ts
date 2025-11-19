import axios, { AxiosError } from 'axios';
import { NetworkError, ServerError, BaseErrorHandler } from '../error-handling';
import { getOfflineQueue } from '../offline/queue';
import { useAuthStore } from '@/store/auth';
import { API_CONFIG } from './config';
import { authService } from '@/services/auth';

export const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: API_CONFIG.TIMEOUT,
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    // Check if offline
    if (BaseErrorHandler.isOffline()) {
      // 오프라인 상태면 큐에 추가 (GET 요청 제외, mutation만 큐에 추가)
      const method = (config.method?.toUpperCase() || 'GET') as 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
      
      if (method !== 'GET') {
        const queue = getOfflineQueue();
        queue.addRequest({
          method,
          url: config.url || '',
          data: config.data,
          headers: config.headers as Record<string, string>,
        });
        
        // 오프라인 상태이므로 요청 실패 처리
        return Promise.reject(new NetworkError('No internet connection. Request queued for retry.'));
      }
      
      return Promise.reject(new NetworkError('No internet connection'));
    }

    // Remove Content-Type header for FormData to let Axios set it automatically with boundary
    if (config.data instanceof FormData) {
      delete config.headers['Content-Type'];
    }

    // Get token from Zustand persist storage (auth-storage)
    // Fallback to legacy token for backward compatibility
    let token: string | null = null;
    if (typeof window !== 'undefined') {
      try {
        const authStorage = localStorage.getItem('auth-storage');
        if (authStorage) {
          const parsed = JSON.parse(authStorage);
          token = parsed?.state?.token || null;
        }
        // Fallback to legacy token if auth-storage doesn't have it
        if (!token) {
          token = localStorage.getItem('token');
        }
      } catch {
        // If parse fails, try legacy token
        token = localStorage.getItem('token');
      }
    }
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (_error) => {
    // Request interceptor error - network failure
    return Promise.reject(new NetworkError('Failed to send request'));
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    // Check if offline
    if (BaseErrorHandler.isOffline()) {
      return Promise.reject(new NetworkError('No internet connection'));
    }

    // Network error (no response)
    if (!error.response) {
      if (error.code === 'ECONNABORTED') {
        return Promise.reject(new NetworkError('Request timeout. Please try again.'));
      }
      return Promise.reject(new NetworkError('Network error. Please check your connection.'));
    }

    // Server error
    const status = error.response.status;
    const data = error.response.data as { error?: { message?: string; code?: string }; message?: string };

    // Handle 401 - Unauthorized
    if (status === 401) {
      // Try to refresh token if we have a refresh token
      const authState = useAuthStore.getState();
      const refreshToken = authState.refreshToken;
      const currentPath = typeof window !== 'undefined' ? window.location.pathname : '';
      
      // If we have a refresh token and this is not a refresh request, try to refresh
      if (refreshToken && !error.config?.url?.includes('/auth/refresh')) {
        try {
          const refreshResponse = await authService.refreshToken(refreshToken);
          
          // Update tokens in store
          authState.setTokens(refreshResponse.access_token, refreshResponse.refresh_token);
          
          // Retry the original request with new token
          if (error.config) {
            error.config.headers.Authorization = `Bearer ${refreshResponse.access_token}`;
            return api.request(error.config);
          }
        } catch (refreshError) {
          // Refresh failed, logout and redirect
          // 현재 페이지가 로그인 페이지가 아닌 경우에만 리다이렉트
          useAuthStore.getState().logout();
          if (currentPath && !currentPath.includes('/login')) {
            // 현재 페이지를 유지하기 위해 router를 사용하지 않고
            // 로그인 페이지로만 리다이렉트 (필요한 경우)
            // 로그인 모달이나 토스트 메시지로 처리하는 것이 더 나을 수 있음
            setTimeout(() => {
              // 세션이 만료되었으므로 로그인 페이지로 이동
              // 하지만 사용자가 보고 있던 페이지 정보는 유지 (쿼리 파라미터 등)
              const returnUrl = encodeURIComponent(currentPath + (window.location.search || ''));
              window.location.href = `/login?returnUrl=${returnUrl}`;
            }, 100);
          }
          return Promise.reject(new ServerError('Session expired. Please login again.', 401));
        }
      } else {
        // No refresh token or refresh request failed, logout and redirect
        // 현재 페이지가 로그인 페이지가 아닌 경우에만 리다이렉트
        useAuthStore.getState().logout();
        if (currentPath && !currentPath.includes('/login')) {
          // 현재 페이지 정보를 유지하여 로그인 후 복귀 가능하도록
          const returnUrl = encodeURIComponent(currentPath + (window.location.search || ''));
          setTimeout(() => {
            window.location.href = `/login?returnUrl=${returnUrl}`;
          }, 100);
        }
        return Promise.reject(new ServerError('Session expired. Please login again.', 401));
      }
    }

    const message = data?.error?.message || data?.message || `Server error (${status})`;
    const code = data?.error?.code;

    return Promise.reject(new ServerError(message, status, code, data));
  }
);

export default api;


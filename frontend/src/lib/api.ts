import axios, { AxiosError } from 'axios';
import { NetworkError, ServerError, isOffline } from './error-handler';
import { getOfflineQueue } from './offline-queue';
import { useAuthStore } from '@/store/auth';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000, // 30 seconds
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    // Check if offline
    if (isOffline()) {
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
  (error) => {
    // Request interceptor error - network failure
    return Promise.reject(new NetworkError('Failed to send request'));
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    // Check if offline
    if (isOffline()) {
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
    // Only redirect if not already on login page
    if (status === 401 && !window.location.pathname.includes('/login')) {
      // Clear auth-storage (Zustand persist) and legacy token
      useAuthStore.getState().logout();
      // Delay redirect to allow error message to be shown
      setTimeout(() => {
        window.location.href = '/login';
      }, 100);
    }

    const message = data?.error?.message || data?.message || `Server error (${status})`;
    const code = data?.error?.code;

    return Promise.reject(new ServerError(message, status, code, data));
  }
);

export default api;

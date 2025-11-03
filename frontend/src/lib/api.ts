import axios, { AxiosError } from 'axios';
import { NetworkError, ServerError, isOffline } from './error-handler';

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
      return Promise.reject(new NetworkError('No internet connection'));
    }

    // Remove Content-Type header for FormData to let Axios set it automatically with boundary
    if (config.data instanceof FormData) {
      delete config.headers['Content-Type'];
    }

    // Try to get token from localStorage first, then from Zustand store
    const token = localStorage.getItem('token') || 
                  (typeof window !== 'undefined' && window.localStorage?.getItem('auth-storage') 
                    ? JSON.parse(window.localStorage.getItem('auth-storage') || '{}').state?.token 
                    : null);
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
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
      localStorage.removeItem('token');
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

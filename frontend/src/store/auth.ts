/**
 * Auth Store
 * 인증 상태 관리 스토어
 * 
 * Zustand Devtools 지원 (개발 환경에서만 활성화)
 */

import { create } from 'zustand';
import { persist, createJSONStorage, devtools } from 'zustand/middleware';
import { User, AuthResponse } from '@/lib/types';

/**
 * 개발 환경 확인
 */
const isDevelopment = process.env.NODE_ENV === 'development';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (authResponse: AuthResponse) => void;
  logout: () => void;
  updateUser: (user: User) => void;
  initialize: () => void;
}

// Persist 미들웨어 설정
const persistedStore = persist<AuthState>(
  (set, get) => ({
    user: null,
    token: null,
    isAuthenticated: false,
    login: (authResponse: AuthResponse) => {
      set({
        user: authResponse.user,
        token: authResponse.token,
        isAuthenticated: true,
      });
      localStorage.setItem('token', authResponse.token);
    },
    logout: () => {
      set({
        user: null,
        token: null,
        isAuthenticated: false,
      });
      localStorage.removeItem('token');
    },
    updateUser: (user: User) => {
      set({ user });
    },
    initialize: () => {
      // Check localStorage for token on initialization
      if (typeof window !== 'undefined') {
        const storedToken = localStorage.getItem('token');
        const state = get();
        
        // If token exists in localStorage but not in state, restore from localStorage
        if (storedToken && !state.token) {
          // Try to get from persist storage
          try {
            const authStorage = localStorage.getItem('auth-storage');
            if (authStorage) {
              const parsed = JSON.parse(authStorage);
              if (parsed?.state?.token) {
                set({
                  token: parsed.state.token,
                  user: parsed.state.user || null,
                  isAuthenticated: !!(parsed.state.token && parsed.state.user),
                });
              }
            } else if (storedToken) {
              // Token exists but persist storage is not set, sync it
              set({
                token: storedToken,
                isAuthenticated: !!storedToken,
              });
            }
          } catch (_e) {
            // If parse fails but token exists, set it
            if (storedToken) {
              set({
                token: storedToken,
                isAuthenticated: true,
              });
            }
          }
        }
      }
    },
  }),
  {
    name: 'auth-storage',
    storage: createJSONStorage(() => localStorage),
    partialize: (state) => ({
      user: state.user,
      token: state.token,
      isAuthenticated: state.isAuthenticated,
    }),
    onRehydrateStorage: () => (state) => {
      // After rehydration, sync with localStorage token
      if (state && typeof window !== 'undefined') {
        const storedToken = localStorage.getItem('token');
        if (storedToken && !state.token) {
          state.token = storedToken;
          state.isAuthenticated = !!state.user && !!storedToken;
        } else if (state.token && !storedToken) {
          // If state has token but localStorage doesn't, sync it
          localStorage.setItem('token', state.token);
        }
      }
    },
  }
);

// 개발 환경에서만 devtools 적용
export const useAuthStore = create<AuthState>()(
  isDevelopment 
    ? devtools(persistedStore, { name: 'AuthStore' })
    : persistedStore
);

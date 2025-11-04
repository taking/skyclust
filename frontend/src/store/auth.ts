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
      // Zustand persist will automatically save to auth-storage
      // No need to manually save token to localStorage
    },
    logout: () => {
      set({
        user: null,
        token: null,
        isAuthenticated: false,
      });
      // Zustand persist will automatically remove from auth-storage
      // Also remove legacy token if it exists
      if (typeof window !== 'undefined') {
        localStorage.removeItem('token');
      }
    },
    updateUser: (user: User) => {
      set({ user });
    },
    initialize: () => {
      // Zustand persist automatically handles rehydration
      // This method is kept for backward compatibility but should rely on persist
      if (typeof window !== 'undefined') {
        const state = get();
        
        // If state is not hydrated yet, try to migrate from legacy token
        if (!state.token) {
          try {
            const authStorage = localStorage.getItem('auth-storage');
            if (authStorage) {
              const parsed = JSON.parse(authStorage);
              if (parsed?.state?.token && !state.token) {
                // State will be rehydrated by persist, but we can ensure it here
                set({
                  token: parsed.state.token,
                  user: parsed.state.user || null,
                  isAuthenticated: !!(parsed.state.token && parsed.state.user),
                });
              }
            }
            
            // Migrate legacy token to auth-storage if it exists
            const legacyToken = localStorage.getItem('token');
            if (legacyToken && !state.token) {
              set({
                token: legacyToken,
                isAuthenticated: !!legacyToken,
              });
              // Remove legacy token after migration
              localStorage.removeItem('token');
            }
          } catch (_e) {
            // Ignore parse errors
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
      // After rehydration, migrate legacy token if it exists
      if (state && typeof window !== 'undefined') {
        const legacyToken = localStorage.getItem('token');
        
        // If we have a legacy token but no state token, migrate it
        if (legacyToken && !state.token) {
          state.token = legacyToken;
          state.isAuthenticated = !!state.user && !!legacyToken;
          // Remove legacy token after migration
          localStorage.removeItem('token');
        }
        // If state has token, ensure legacy token is removed (we use auth-storage only)
        else if (state.token && legacyToken) {
          // Remove legacy token to avoid conflicts
          localStorage.removeItem('token');
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

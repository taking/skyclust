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

// Persist 미들웨어 설정 (localStorage에 자동 저장)
const persistedStore = persist<AuthState>(
  (set, get) => ({
    // 초기 상태
    user: null,
    token: null,
    isAuthenticated: false,
    
    /**
     * 로그인 처리
     * 인증 응답을 받아서 사용자 정보와 토큰을 저장합니다.
     * Zustand persist가 자동으로 localStorage에 저장하므로 수동 저장 불필요
     */
    login: (authResponse: AuthResponse) => {
      set({
        user: authResponse.user,
        token: authResponse.token,
        isAuthenticated: true,
      });
      // Zustand persist가 자동으로 auth-storage에 저장함
      // localStorage에 토큰을 수동으로 저장할 필요 없음
    },
    
    /**
     * 로그아웃 처리
     * 사용자 정보와 토큰을 초기화합니다.
     * Zustand persist가 자동으로 localStorage에서 제거하며, 레거시 토큰도 함께 제거합니다.
     */
    logout: () => {
      set({
        user: null,
        token: null,
        isAuthenticated: false,
      });
      // Zustand persist가 자동으로 auth-storage에서 제거함
      // 레거시 토큰이 있으면 함께 제거
      if (typeof window !== 'undefined') {
        localStorage.removeItem('token');
      }
    },
    
    /**
     * 사용자 정보 업데이트
     * 로그인한 사용자의 정보를 업데이트합니다.
     */
    updateUser: (user: User) => {
      set({ user });
    },
    
    /**
     * 스토어 초기화
     * Zustand persist가 자동으로 rehydration을 처리하지만,
     * 하위 호환성을 위해 레거시 토큰 마이그레이션을 수행합니다.
     */
    initialize: () => {
      // Zustand persist가 자동으로 rehydration을 처리함
      // 이 메서드는 하위 호환성을 위해 유지되지만 persist에 의존해야 함
      if (typeof window !== 'undefined') {
        const state = get();
        
        // 1. 토큰이 없으면 레거시 토큰에서 마이그레이션 시도
        if (!state.token) {
          try {
            // 2. auth-storage에서 토큰 확인
            const authStorage = localStorage.getItem('auth-storage');
            if (authStorage) {
              const parsed = JSON.parse(authStorage);
              if (parsed?.state?.token && !state.token) {
                // persist에 의해 상태가 rehydrated되지만, 여기서도 확인 가능
                set({
                  token: parsed.state.token,
                  user: parsed.state.user || null,
                  isAuthenticated: !!(parsed.state.token && parsed.state.user),
                });
              }
            }
            
            // 3. 레거시 토큰('token')이 있으면 auth-storage로 마이그레이션
            const legacyToken = localStorage.getItem('token');
            if (legacyToken && !state.token) {
              set({
                token: legacyToken,
                isAuthenticated: !!legacyToken,
              });
              // 마이그레이션 후 레거시 토큰 제거
              localStorage.removeItem('token');
            }
          } catch (_e) {
            // 파싱 에러 무시
          }
        }
      }
    },
  }),
  {
    name: 'auth-storage',
    storage: createJSONStorage(() => localStorage),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    partialize: (state: AuthState): any => ({
      user: state.user,
      token: state.token,
      isAuthenticated: state.isAuthenticated,
    }),
    /**
     * Rehydration 후 콜백
     * localStorage에서 상태를 복원한 후 레거시 토큰을 마이그레이션합니다.
     */
    onRehydrateStorage: () => (state) => {
      // rehydration 후 레거시 토큰이 있으면 마이그레이션
      if (state && typeof window !== 'undefined') {
        const legacyToken = localStorage.getItem('token');
        
        // 1. 레거시 토큰이 있고 상태에 토큰이 없으면 마이그레이션
        if (legacyToken && !state.token) {
          state.token = legacyToken;
          state.isAuthenticated = !!state.user && !!legacyToken;
          // 마이그레이션 후 레거시 토큰 제거
          localStorage.removeItem('token');
        }
        // 2. 상태에 토큰이 있으면 레거시 토큰 제거 (auth-storage만 사용)
        else if (state.token && legacyToken) {
          // 충돌 방지를 위해 레거시 토큰 제거
          localStorage.removeItem('token');
        }
      }
    },
  }
);

// 개발 환경에서만 devtools 적용
export const useAuthStore = (isDevelopment 
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ? create<AuthState>()(devtools(persistedStore as any, { name: 'AuthStore' }) as any)
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  : create<AuthState>()(persistedStore as any)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
) as any;

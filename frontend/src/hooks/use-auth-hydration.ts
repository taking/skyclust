/**
 * useAuthHydration Hook
 * Zustand persist hydration 대기 및 인증 상태 체크
 * 
 * Zustand persist의 hydration 타이밍 이슈를 해결하고,
 * 인증 상태를 정확하게 확인할 수 있도록 돕습니다.
 * Legacy token fallback도 지원합니다.
 * 
 * @example
 * ```tsx
 * const { isHydrated, isAuthenticated, token, isLoading } = useAuthHydration();
 * 
 * if (!isHydrated || isLoading) {
 *   return <Spinner />;
 * }
 * 
 * if (!isAuthenticated) {
 *   router.push('/login');
 * }
 * ```
 */

import { useState, useEffect, useCallback } from 'react';
import { useAuthStore } from '@/store/auth';
import { TIME, STORAGE_KEYS } from '@/lib/constants';

export interface UseAuthHydrationReturn {
  /**
   * Zustand persist hydration 완료 여부
   */
  isHydrated: boolean;
  
  /**
   * 인증 여부 (토큰이 있는지)
   */
  isAuthenticated: boolean;
  
  /**
   * 인증 토큰
   */
  token: string | null;
  
  /**
   * 체크 중인지 여부
   */
  isLoading: boolean;
  
  /**
   * 인증 상태 재확인
   */
  recheck: () => void;
}

export interface UseAuthHydrationOptions {
  /**
   * Hydration 대기 시간 (ms)
   * @default 300
   */
  hydrationDelay?: number;
  
  /**
   * Legacy token도 체크할지 여부
   * @default true
   */
  checkLegacyToken?: boolean;
}

/**
 * Zustand persist에서 토큰 가져오기
 */
function getTokenFromStorage(checkLegacyToken: boolean = true): string | null {
  if (typeof window === 'undefined') {
    return null;
  }

  try {
    // 1. auth-storage (Zustand persist)에서 체크
    const authStorage = localStorage.getItem(STORAGE_KEYS.AUTH_STORAGE);
    if (authStorage) {
      const parsedAuth = JSON.parse(authStorage) as {
        state?: {
          isAuthenticated?: boolean;
          token?: string;
        };
      };
      
      if (parsedAuth?.state?.token) {
        return parsedAuth.state.token;
      }
    }
    
    // 2. Legacy token fallback (backward compatibility)
    if (checkLegacyToken) {
      const legacyToken = localStorage.getItem('token');
      if (legacyToken) {
        return legacyToken;
      }
    }
  } catch {
    // Parse 실패 시 legacy token 체크
    if (checkLegacyToken) {
      try {
        return localStorage.getItem('token');
      } catch {
        return null;
      }
    }
  }

  return null;
}


/**
 * useAuthHydration Hook
 * 
 * Zustand persist hydration 대기 및 인증 상태 체크
 */
export function useAuthHydration(
  options: UseAuthHydrationOptions = {}
): UseAuthHydrationReturn {
  const {
    hydrationDelay = TIME.DELAY.AUTH_HYDRATION,
    checkLegacyToken = true,
  } = options;

  const { initialize, token: storeToken, isAuthenticated: storeIsAuthenticated } = useAuthStore();
  const [isHydrated, setIsHydrated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [token, setToken] = useState<string | null>(null);

  /**
   * 인증 상태 재확인
   */
  const recheck = useCallback(() => {
    const storedToken = getTokenFromStorage(checkLegacyToken);
    const authState = useAuthStore.getState();
    const finalToken = authState.token || storedToken;
    const finalAuth = authState.isAuthenticated || !!finalToken;
    
    setIsAuthenticated(finalAuth);
    setToken(finalToken);
  }, [checkLegacyToken]);

  /**
   * Hydration 및 인증 상태 체크
   */
  useEffect(() => {
    // 1. Auth store 초기화
    initialize();

    // 2. Hydration 대기 후 상태 체크
    const timer = setTimeout(() => {
      setIsHydrated(true);
      
      // 3. 인증 상태 확인
      const storedToken = getTokenFromStorage(checkLegacyToken);
      const authState = useAuthStore.getState();
      
      // Store 상태와 로컬 스토리지 상태를 모두 고려
      const finalToken = authState.token || storeToken || storedToken;
      const finalAuth = authState.isAuthenticated || storeIsAuthenticated || !!finalToken;
      
      setIsAuthenticated(finalAuth);
      setToken(finalToken);
      setIsLoading(false);
    }, hydrationDelay);

    return () => clearTimeout(timer);
  }, [initialize, hydrationDelay, checkLegacyToken, storeIsAuthenticated, storeToken]);

  /**
   * Store 상태 변경 감지
   */
  useEffect(() => {
    if (!isHydrated) {
      return;
    }

    // Store 상태가 변경되면 로컬 상태도 업데이트
    const finalAuth = storeIsAuthenticated || !!storeToken;
    const finalToken = storeToken || token;
    
    setIsAuthenticated(finalAuth);
    setToken(finalToken);
  }, [isHydrated, storeIsAuthenticated, storeToken, token]);

  return {
    isHydrated,
    isAuthenticated,
    token,
    isLoading,
    recheck,
  };
}


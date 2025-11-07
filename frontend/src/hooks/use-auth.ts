import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth';

/**
 * useAuth 훅
 * 
 * 사용자 인증 상태를 확인하고, 인증되지 않은 경우 리다이렉트합니다.
 * 
 * @param redirectTo - 인증되지 않은 경우 리다이렉트할 경로 (기본값: '/login')
 * @returns 인증 상태, 사용자 정보, 토큰
 * 
 * @example
 * ```tsx
 * function ProtectedPage() {
 *   const { isAuthenticated, user, token } = useAuth();
 *   
 *   if (!isAuthenticated) {
 *     return <div>Loading...</div>;
 *   }
 *   
 *   return <div>Welcome, {user?.name}!</div>;
 * }
 * ```
 */
export const useAuth = (redirectTo: string = '/login') => {
  const router = useRouter();
  const { isAuthenticated, token, user } = useAuthStore();

  useEffect(() => {
    // 1. 기본 인증 상태 확인: 인증되지 않았거나 토큰/사용자 정보가 없으면 리다이렉트
    if (!isAuthenticated || !token || !user) {
      router.push(redirectTo);
      return;
    }

    // 2. localStorage에서 토큰 검증 (서버 사이드 렌더링 대응)
    let storedToken: string | null = null;
    try {
      // 2-1. Zustand persist storage에서 토큰 가져오기
      const authStorage = typeof window !== 'undefined' ? localStorage.getItem('auth-storage') : null;
      if (authStorage) {
        const parsed = JSON.parse(authStorage);
        storedToken = parsed?.state?.token || null;
      }
      // 2-2. 레거시 토큰 fallback (하위 호환성)
      if (!storedToken && typeof window !== 'undefined') {
        storedToken = localStorage.getItem('token');
      }
    } catch {
      // 2-3. 파싱 에러 발생 시 레거시 토큰으로 fallback
      storedToken = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
    }
    
    // 3. 저장된 토큰과 현재 토큰 비교: 불일치하면 로그아웃 처리
    if (!storedToken || storedToken !== token) {
      useAuthStore.getState().logout();
      router.push(redirectTo);
      return;
    }
  }, [isAuthenticated, token, user, router, redirectTo]);

  return { isAuthenticated, user, token };
};

/**
 * useRequireAuth 훅
 * 
 * useAuth의 래퍼로, 로딩 상태를 추가로 제공합니다.
 * 
 * @param redirectTo - 인증되지 않은 경우 리다이렉트할 경로 (기본값: '/login')
 * @returns 인증 상태, 사용자 정보, 토큰, 로딩 상태
 * 
 * @example
 * ```tsx
 * function ProtectedPage() {
 *   const { isLoading, user } = useRequireAuth();
 *   
 *   if (isLoading) {
 *     return <div>Loading...</div>;
 *   }
 *   
 *   return <div>Welcome, {user?.name}!</div>;
 * }
 * ```
 */
export const useRequireAuth = (redirectTo: string = '/login') => {
  // 1. useAuth 훅 호출하여 기본 인증 정보 가져오기
  const { isAuthenticated, user, token } = useAuth(redirectTo);
  
  // 2. 로딩 상태 계산: 인증 정보가 모두 없으면 로딩 중으로 간주
  return {
    isAuthenticated,
    user,
    token,
    isLoading: !isAuthenticated && !user && !token,
  };
};


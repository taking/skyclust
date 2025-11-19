'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuthHydration } from '@/hooks/use-auth-hydration';
import { useSystemInitialized } from '@/hooks/use-system-initialized';
import { useAuthStore } from '@/store/auth';
import { Spinner } from '@/components/ui/spinner';

export default function HomePage() {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated } = useAuthStore();
  const { isHydrated, isLoading: isAuthLoading } = useAuthHydration({
    hydrationDelay: 300,
    checkLegacyToken: true,
  });

  // 시스템 초기화 상태 확인
  const { data: initStatus, isLoading: isLoadingStatus } = useSystemInitialized({
    enabled: isHydrated && !isAuthLoading,
  });

  useEffect(() => {
    // 루트 경로(`/`)가 아닌 경우 리다이렉트 스킵
    // 새로고침 시 현재 페이지를 유지하기 위함
    if (pathname !== '/') {
      return;
    }

    // hydration 또는 초기화 상태 확인 중이면 대기
    if (!isHydrated || isAuthLoading || isLoadingStatus || !initStatus) {
      return;
    }

    // 루트 경로에서만 리다이렉트 수행
    // 1. 초기화되지 않음 → /setup으로 리다이렉트
    if (!initStatus.initialized) {
      router.replace('/setup');
      return;
    }

    // 2. 초기화됨 + 인증됨 → workspace가 있으면 해당 workspace의 dashboard로, 없으면 /workspaces로 리다이렉트
    if (isAuthenticated) {
      // Workspace Store에서 현재 workspace 확인 (동적 import로 클라이언트 사이드에서만 실행)
      if (typeof window !== 'undefined') {
        import('@/store/workspace').then(({ useWorkspaceStore }) => {
          const { currentWorkspace } = useWorkspaceStore.getState();
          
          if (currentWorkspace?.id) {
            import('@/lib/routing/helpers').then(({ buildManagementPath }) => {
              router.replace(buildManagementPath(currentWorkspace.id, 'dashboard'));
            });
          } else {
            router.replace('/workspaces');
          }
        });
      } else {
        router.replace('/dashboard');
      }
      return;
    }

    // 3. 초기화됨 + 미인증 → /login으로 리다이렉트
    router.replace('/login');
  }, [router, pathname, isHydrated, isAuthLoading, isLoadingStatus, initStatus, isAuthenticated]);

  if (!isHydrated || isAuthLoading || isLoadingStatus || !initStatus) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center space-y-4">
          <h1 className="text-4xl font-bold text-gray-900">SkyClust</h1>
          <Spinner size="lg" label="Loading..." />
        </div>
      </div>
    );
  }

  return null;
}
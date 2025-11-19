'use client';

import { Header } from './header';
import { Sidebar } from './sidebar';
import { useAuthStore } from '@/store/auth';
import { useRouter, usePathname } from 'next/navigation';
import { Suspense, useEffect, useRef, useMemo } from 'react';
import { SkipLink } from '@/components/accessibility/skip-link';
import { LiveRegion } from '@/components/accessibility/live-region';
import { OfflineBanner } from '@/components/common/offline-banner';
import { Breadcrumb } from '@/components/common/breadcrumb';
import { useSSEEvents } from '@/hooks/use-sse-events';
import { useAuthHydration } from '@/hooks/use-auth-hydration';
import { STORAGE_KEYS } from '@/lib/constants';

interface LayoutProps {
  children: React.ReactNode;
}

/**
 * Layout 컴포넌트
 * 
 * 애플리케이션의 메인 레이아웃을 제공합니다.
 * Header, Sidebar, 메인 콘텐츠 영역을 포함하며, 인증 상태 관리 및 접근성 기능을 제공합니다.
 * 
 * @example
 * ```tsx
 * <Layout>
 *   <YourPageContent />
 * </Layout>
 * ```
 */
export function Layout({ children }: LayoutProps) {
  const { isAuthenticated, token } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();
  
  // 인증 상태 hydration 대기 (Zustand persist rehydration 완료까지)
  const { isHydrated, isAuthenticated: hydratedIsAuthenticated, isLoading: isAuthLoading } = useAuthHydration({
    hydrationDelay: 300,
    checkLegacyToken: true,
  });

  // 최종 인증 상태: hydration된 인증 상태 또는 store의 인증 상태
  const finalIsAuthenticated = isHydrated ? (hydratedIsAuthenticated || isAuthenticated) : false;
  const hasRedirectedRef = useRef(false);
  const authCheckTimerRef = useRef<NodeJS.Timeout | null>(null);

  // 1. SSE 이벤트 구독 및 React Query 통합 (실시간 데이터 업데이트)
  useSSEEvents(token);

  // 2. 마지막 방문 경로 추적 (공개 경로 제외)
  useEffect(() => {
    const publicPaths = ['/login', '/register', '/setup'];
    if (!publicPaths.includes(pathname) && typeof window !== 'undefined') {
      sessionStorage.setItem('lastPath', pathname);
    }
  }, [pathname]);

  // 3. 인증되지 않은 사용자 리다이렉트 (hydration 완료 후에만 실행)
  // hydration이 완료되고 인증 상태가 확정된 후에만 리다이렉트 수행
  useEffect(() => {
    // 이전 타이머 정리
    if (authCheckTimerRef.current) {
      clearTimeout(authCheckTimerRef.current);
      authCheckTimerRef.current = null;
    }

    // hydration이 완료되지 않았거나 로딩 중이면 스킵
    if (!isHydrated || isAuthLoading) {
      return;
    }

    // 이미 리다이렉트했으면 스킵 (무한 루프 방지)
    if (hasRedirectedRef.current) {
      return;
    }

    // 세션 스토리지에서 리다이렉트 플래그 확인
    const redirectedPath = typeof window !== 'undefined' ? sessionStorage.getItem('authRedirected') : null;
    if (redirectedPath === pathname) {
      // 이미 이 경로로 리다이렉트했으면 스킵
      return;
    }

    // 공개 경로 확인 (로그인/회원가입/초기 설정 페이지)
    const publicPaths = ['/login', '/register', '/setup'];
    const isPublicPath = publicPaths.includes(pathname);
    
    // 인증되지 않은 경우
    if (!finalIsAuthenticated && !isPublicPath) {
      // 추가 안전 검사: localStorage에서 직접 토큰 확인
      const hasToken = typeof window !== 'undefined' && (
        localStorage.getItem(STORAGE_KEYS.AUTH_STORAGE) || 
        localStorage.getItem('token')
      );

      // 토큰이 있는데 finalIsAuthenticated가 false면 hydration 대기
      if (hasToken && !finalIsAuthenticated) {
        // 추가 대기 후 재확인 (100ms)
        authCheckTimerRef.current = setTimeout(() => {
          const authState = useAuthStore.getState();
          if (authState.isAuthenticated || authState.token) {
            // 인증 상태가 복원되었으므로 리다이렉트 스킵
            return;
          }
          
          // 여전히 인증되지 않았으면 리다이렉트
          if (!hasRedirectedRef.current) {
            hasRedirectedRef.current = true;
            if (typeof window !== 'undefined') {
              sessionStorage.setItem('authRedirected', pathname);
              const returnUrl = encodeURIComponent(pathname + window.location.search);
              router.replace(`/login?returnUrl=${returnUrl}`);
            }
          }
        }, 100);
        
        return;
      }

      // 토큰이 없거나 확실히 인증되지 않은 경우 리다이렉트
      hasRedirectedRef.current = true;
      if (typeof window !== 'undefined') {
        sessionStorage.setItem('authRedirected', pathname);
        const returnUrl = encodeURIComponent(pathname + window.location.search);
        router.replace(`/login?returnUrl=${returnUrl}`);
      }
    }

    // Cleanup: 타이머 정리
    return () => {
      if (authCheckTimerRef.current) {
        clearTimeout(authCheckTimerRef.current);
        authCheckTimerRef.current = null;
      }
    };
  }, [isHydrated, isAuthLoading, finalIsAuthenticated, pathname, router]);

  // 경로 변경 시 리다이렉트 플래그 리셋 (세션 스토리지도 정리)
  useEffect(() => {
    hasRedirectedRef.current = false;
    if (typeof window !== 'undefined') {
      const publicPaths = ['/login', '/register', '/setup'];
      if (publicPaths.includes(pathname)) {
        // 공개 경로로 이동하면 리다이렉트 플래그 제거
        sessionStorage.removeItem('authRedirected');
      }
    }
  }, [pathname]);

  // Breadcrumb 표시 여부 계산 (hooks는 항상 early return 이전에 호출되어야 함)
  const shouldShowBreadcrumb = useMemo(() => {
    const segments = pathname.split('/').filter(Boolean);
    const meaningfulSegments = segments.filter(
      segment => !segment.match(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i)
    );
    return meaningfulSegments.length >= 3;
  }, [pathname]);

  // 4. hydration이 완료될 때까지 아무것도 렌더링하지 않음 (깜빡임 방지)
  if (!isHydrated || isAuthLoading) {
    return null;
  }

  // 5. 로그인/회원가입/초기 설정 페이지에서는 레이아웃 없이 children만 렌더링
  const publicPaths = ['/login', '/register', '/setup'];
  if (publicPaths.includes(pathname)) {
    return <>{children}</>;
  }

  // 6. 인증되지 않았으면 아무것도 렌더링하지 않음 (리다이렉트 대기)
  if (!finalIsAuthenticated) {
    return null;
  }

  return (
    <div className="flex h-screen bg-background">
      {/* 7. 접근성: 메인 콘텐츠로 건너뛰기 링크 */}
      <SkipLink href="#main-content">Skip to main content</SkipLink>
      
      {/* 8. 오프라인 배너 (네트워크 상태 표시) */}
      <OfflineBanner position="top" autoHide showRefreshButton />
      
      {/* 9. 사이드바 네비게이션 */}
      <Sidebar />
      
      {/* 10. 메인 콘텐츠 영역 */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* 10-1. 헤더 (Suspense로 감싸서 비동기 로딩 지원) */}
        <Suspense fallback={<div className="h-16 border-b bg-background" />}>
          <Header />
        </Suspense>
        
        {/* 10-2. Breadcrumb (Header 아래에 배치) */}
        {shouldShowBreadcrumb && (
          <div className="sticky top-16 z-30 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4 sm:px-6 py-2">
            <Breadcrumb className="text-sm" />
          </div>
        )}
        
        {/* 10-3. 메인 콘텐츠 영역 (스크롤 가능) */}
        <main 
          id="main-content"
          className="flex-1 overflow-y-auto bg-muted/30 p-3 sm:p-4 md:p-6"
          role="main"
          tabIndex={-1}
        >
          {children}
        </main>
      </div>
      
      {/* 11. 접근성: Live Region (스크린 리더용 동적 메시지) */}
      <LiveRegion message="" />
    </div>
  );
}

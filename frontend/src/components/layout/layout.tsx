'use client';

import { Header } from './header';
import { Sidebar } from './sidebar';
import { useAuthStore } from '@/store/auth';
import { useRouter, usePathname } from 'next/navigation';
import { useEffect, useState, Suspense } from 'react';
import { SkipLink } from '@/components/accessibility/skip-link';
import { LiveRegion } from '@/components/accessibility/live-region';
import { GlobalKeyboardShortcuts } from '@/components/common/global-keyboard-shortcuts';
import { OfflineBanner } from '@/components/common/offline-banner';
import { KeyboardShortcutsHelp } from '@/components/common/keyboard-shortcuts-help';
import { KeyboardShortcut } from '@/hooks/use-keyboard-shortcuts';
import { HelpCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useSSEEvents } from '@/hooks/use-sse-events';

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
  const { isAuthenticated, initialize, token } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();
  const [isHydrated, setIsHydrated] = useState(false);

  // 1. SSE 이벤트 구독 및 React Query 통합 (실시간 데이터 업데이트)
  useSSEEvents(token);

  // 2. 인증 스토어 초기화 및 hydration 대기
  useEffect(() => {
    // 2-1. 인증 스토어 초기화 (레거시 토큰 마이그레이션 등)
    initialize();
    
    // 2-2. Zustand persist의 rehydration 완료 대기
    const checkHydration = () => {
      if (typeof window !== 'undefined') {
        // 2-2-1. localStorage에서 auth-storage 확인 (주요 소스)
        const authStorage = localStorage.getItem('auth-storage');
        // 2-2-2. 레거시 토큰 fallback (하위 호환성)
        const legacyToken = localStorage.getItem('token');
        
        // 2-2-3. 인증 데이터가 있으면 Zustand rehydration을 위해 잠시 대기
        if (authStorage || legacyToken) {
          // Zustand rehydration을 위한 시간 제공
          setTimeout(() => {
            setIsHydrated(true);
          }, 100);
        } else {
          // 2-2-4. 인증 데이터가 없으면 즉시 hydration 완료로 처리
          setIsHydrated(true);
        }
      } else {
        // 서버 사이드에서는 즉시 hydration 완료
        setIsHydrated(true);
      }
    };

    checkHydration();
  }, [initialize]);

  // 3. 인증되지 않은 사용자 리다이렉트 (hydration 완료 후에만 실행)
  useEffect(() => {
    // hydration이 완료되지 않았으면 스킵
    if (!isHydrated) return;
    
    // 3-1. 공개 경로 확인 (로그인/회원가입/초기 설정 페이지)
    const publicPaths = ['/login', '/register', '/setup'];
    const isPublicPath = publicPaths.includes(pathname);
    
    // 3-2. 인증되지 않았고 공개 경로가 아니면 로그인 페이지로 리다이렉트
    if (!isAuthenticated && !isPublicPath) {
      router.replace('/login');
    }
  }, [isAuthenticated, isHydrated, router, pathname]);

  // 4. hydration이 완료될 때까지 아무것도 렌더링하지 않음 (깜빡임 방지)
  if (!isHydrated) {
    return null;
  }

  // 5. 로그인/회원가입/초기 설정 페이지에서는 레이아웃 없이 children만 렌더링
  const publicPaths = ['/login', '/register', '/setup'];
  if (publicPaths.includes(pathname)) {
    return <>{children}</>;
  }

  // 6. 인증되지 않았으면 아무것도 렌더링하지 않음 (리다이렉트 대기)
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="flex h-screen bg-background">
      {/* 7. 전역 키보드 단축키 등록 */}
      <GlobalKeyboardShortcuts />
      
      {/* 8. 접근성: 메인 콘텐츠로 건너뛰기 링크 */}
      <SkipLink href="#main-content">Skip to main content</SkipLink>
      
      {/* 9. 오프라인 배너 (네트워크 상태 표시) */}
      <OfflineBanner position="top" autoHide showRefreshButton />
      
      {/* 10. 사이드바 네비게이션 */}
      <Sidebar />
      
      {/* 11. 메인 콘텐츠 영역 */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* 11-1. 헤더 (Suspense로 감싸서 비동기 로딩 지원) */}
        <Suspense fallback={<div className="h-16 border-b bg-background" />}>
          <Header />
        </Suspense>
        
        {/* 11-2. 메인 콘텐츠 영역 (스크롤 가능) */}
        <main 
          id="main-content"
          className="flex-1 overflow-y-auto bg-muted/30 p-3 sm:p-4 md:p-6"
          role="main"
          tabIndex={-1}
        >
          {children}
        </main>
      </div>
      
      {/* 12. 접근성: Live Region (스크린 리더용 동적 메시지) */}
      <LiveRegion message="" />
      
      {/* 13. 플로팅 도움말 버튼 (키보드 단축키 도움말) - Safari 호환성 최적화 */}
      <div 
        className="fixed bottom-6 right-6 z-50" 
        style={{ 
          position: 'fixed', 
          bottom: '1.5rem', 
          right: '1.5rem', 
          zIndex: 50,
          WebkitTransform: 'translateZ(0)', // Safari 하드웨어 가속
          transform: 'translateZ(0)',
          willChange: 'transform', // Safari 최적화
        }}
      >
        <KeyboardShortcutsHelp
          shortcuts={[]}
          trigger={
            <Button
              variant="default"
              size="icon"
              className="h-12 w-12 rounded-full shadow-lg hover:shadow-xl transition-shadow"
              aria-label="Show keyboard shortcuts"
            >
              <HelpCircle className="h-6 w-6" />
            </Button>
          }
        />
      </div>
    </div>
  );
}

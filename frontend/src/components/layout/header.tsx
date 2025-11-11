/**
 * Header 컴포넌트
 * 
 * 애플리케이션 상단 헤더를 담당하는 컴포넌트입니다.
 * 
 * @example
 * ```tsx
 * // Layout에서 자동으로 사용됨
 * <Layout>
 *   <Header />  // 자동으로 렌더링됨
 *   <main>...</main>
 * </Layout>
 * ```
 * 
 * 기능:
 * - 사용자 인증 상태 표시
 * - Workspace 선택
 * - 언어 선택
 * - 테마 토글
 * - URL과 자격 증명/리전 상태 동기화
 * - Breadcrumb 네비게이션
 */
'use client';

import * as React from 'react';
import { usePathname, useSearchParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useAuthStore } from '@/store/auth';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useRouter } from 'next/navigation';
import { LogOut, User, Settings, Plus } from 'lucide-react';
import { useCredentials } from '@/hooks/use-credentials';
import { MobileNav } from './mobile-nav';
import { ThemeToggle } from '@/components/theme/theme-toggle';
import { ScreenReaderOnly } from '@/components/accessibility/screen-reader-only';
import { getActionAriaLabel } from '@/lib/accessibility';
import { Breadcrumb } from '@/components/common/breadcrumb';
import { getRegionsByProvider, supportsRegionSelection, getDefaultRegionForProvider } from '@/lib/regions';
import type { CloudProvider } from '@/lib/types/kubernetes';
import { useTranslation } from '@/hooks/use-translation';
import { locales, localeNames, type Locale } from '@/i18n/config';

function HeaderComponent() {
  const { user, logout } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion } = useCredentialContextStore();
  const { t, locale, setLocale } = useTranslation();

  // 자격 증명/리전 선택기를 표시할 경로인지 확인 (URL 동기화용)
  const shouldShowSelectors = React.useMemo(() => {
    return pathname.startsWith('/compute') || 
           pathname.startsWith('/kubernetes') || 
           pathname.startsWith('/networks') ||
           pathname.startsWith('/dashboard');
  }, [pathname]);

  // 현재 워크스페이스의 자격 증명 조회 (URL 동기화용)
  const { credentials, isLoading: isLoadingCredentials } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: !!currentWorkspace && shouldShowSelectors,
  });

  // 이전 워크스페이스 ID 추적 (변경 감지용)
  const prevWorkspaceIdRef = React.useRef<string | undefined>(currentWorkspace?.id);
  const prevCredentialIdRef = React.useRef<string | null>(selectedCredentialId);
  
  // 워크스페이스 변경 시 자격 증명 선택 초기화
  React.useEffect(() => {
    // 1. 초기 조건 확인: 선택기 표시 여부, 워크스페이스 존재, 자격 증명 로딩 상태
    if (!shouldShowSelectors || !currentWorkspace || isLoadingCredentials) {
      prevWorkspaceIdRef.current = currentWorkspace?.id;
      return;
    }

    // 2. 변경 사항 감지: 워크스페이스 ID와 자격 증명 ID 비교
    const workspaceChanged = prevWorkspaceIdRef.current !== currentWorkspace.id;
    const credentialChanged = prevCredentialIdRef.current !== selectedCredentialId;
    
    // 3. 워크스페이스가 변경된 경우 자격 증명 및 리전 초기화
    if (workspaceChanged) {
      prevWorkspaceIdRef.current = currentWorkspace.id;
      if (selectedCredentialId) {
        // 전역 상태 초기화
        setSelectedCredential(null);
        setSelectedRegion(null);
        // URL에서도 제거
        const params = new URLSearchParams(searchParams.toString());
        params.delete('credentialId');
        params.delete('region');
        router.replace(`${pathname}?${params.toString()}`, { scroll: false });
      }
      return;
    }

    // 4. 선택된 자격 증명이 현재 워크스페이스에 속하는지 검증
    if (selectedCredentialId && credentials.length > 0 && !credentialChanged) {
      const credentialExists = credentials.some(c => c.id === selectedCredentialId);
      if (!credentialExists) {
        // 자격 증명이 현재 워크스페이스에 없으면 초기화
        setSelectedCredential(null);
        setSelectedRegion(null);
        const urlCredentialId = searchParams.get('credentialId');
        if (urlCredentialId) {
          // URL에도 반영
          const params = new URLSearchParams(searchParams.toString());
          params.delete('credentialId');
          params.delete('region');
          router.replace(`${pathname}?${params.toString()}`, { scroll: false });
        }
      }
    } else if (selectedCredentialId && credentials.length === 0 && !isLoadingCredentials && !credentialChanged) {
      // 5. 자격 증명 목록이 비어있는데 선택된 자격 증명이 있는 경우 초기화
      setSelectedCredential(null);
      setSelectedRegion(null);
      const urlCredentialId = searchParams.get('credentialId');
      if (urlCredentialId) {
        const params = new URLSearchParams(searchParams.toString());
        params.delete('credentialId');
        params.delete('region');
        router.replace(`${pathname}?${params.toString()}`, { scroll: false });
      }
    }
    
    // 6. 현재 자격 증명 ID를 이전 값으로 저장 (다음 렌더링에서 변경 감지용)
    prevCredentialIdRef.current = selectedCredentialId;
  }, [currentWorkspace, isLoadingCredentials, credentials, selectedCredentialId, shouldShowSelectors, pathname, router, searchParams, setSelectedCredential, setSelectedRegion]);

  // 이전 URL 파라미터 추적 (불필요한 업데이트 방지)
  const prevUrlCredentialIdRef = React.useRef<string | null>(null);
  const prevUrlRegionRef = React.useRef<string | null>(null);
  
  // URL 쿼리 파라미터와 양방향 동기화
  // URL이 있으면 URL을 우선하고, URL이 없으면 Store의 값을 URL에 반영
  React.useEffect(() => {
    // 1. 초기 조건 확인
    if (!shouldShowSelectors || !currentWorkspace || isLoadingCredentials) return;

    // 2. URL에서 자격 증명 ID와 리전 가져오기
    const urlCredentialId = searchParams.get('credentialId');
    const urlRegion = searchParams.get('region');

    // 3. URL이 변경되지 않았으면 스킵 (무한 루프 방지)
    if (urlCredentialId === prevUrlCredentialIdRef.current && urlRegion === prevUrlRegionRef.current) {
      return;
    }

    // ===== 우선순위 1: URL → Store (URL이 있으면 Store 업데이트) =====
    if (urlCredentialId && urlCredentialId !== selectedCredentialId) {
      if (credentials.length > 0) {
        // 4. URL의 자격 증명이 현재 워크스페이스에 속하는지 확인
        const credentialExists = credentials.some(c => c.id === urlCredentialId);
        if (credentialExists) {
          // 5. 유효한 자격 증명이면 Store 업데이트
          setSelectedCredential(urlCredentialId);
          prevUrlCredentialIdRef.current = urlCredentialId;
          
          // 6. 자격 증명 변경 시 리전 처리
          const newCredential = credentials.find(c => c.id === urlCredentialId);
          if (newCredential) {
            if (supportsRegionSelection(newCredential.provider as CloudProvider)) {
              // 6-1. 프로바이더가 리전을 지원하고 URL에 리전이 없으면 기본 리전 설정
              if (!urlRegion) {
                const defaultRegion = getDefaultRegionForProvider(newCredential.provider);
                if (defaultRegion) {
                  setSelectedRegion(defaultRegion);
                  prevUrlRegionRef.current = defaultRegion;
                  const params = new URLSearchParams(searchParams.toString());
                  params.set('region', defaultRegion);
                  router.replace(`${pathname}?${params.toString()}`, { scroll: false });
                }
              }
            } else {
              // 6-2. 프로바이더가 리전을 지원하지 않으면 리전 제거
              setSelectedRegion(null);
              prevUrlRegionRef.current = null;
              const params = new URLSearchParams(searchParams.toString());
              params.delete('region');
              router.replace(`${pathname}?${params.toString()}`, { scroll: false });
            }
          }
        } else {
          // 7. URL의 자격 증명이 현재 워크스페이스에 없으면 URL에서 제거
          const params = new URLSearchParams(searchParams.toString());
          params.delete('credentialId');
          params.delete('region');
          prevUrlCredentialIdRef.current = null;
          prevUrlRegionRef.current = null;
          router.replace(`${pathname}?${params.toString()}`, { scroll: false });
        }
      } else {
        // 8. 자격 증명 목록이 아직 로드 중이면 URL 값만 저장 (나중에 처리)
        prevUrlCredentialIdRef.current = urlCredentialId;
      }
    } else if (!urlCredentialId) {
      prevUrlCredentialIdRef.current = null;
    }
    
    // 9. URL의 리전을 Store에 동기화
    if (urlRegion !== null && urlRegion !== selectedRegion) {
      setSelectedRegion(urlRegion || null);
      prevUrlRegionRef.current = urlRegion;
    } else if (urlRegion === null) {
      prevUrlRegionRef.current = null;
    }

    // ===== 우선순위 2: Store → URL (URL이 없고 Store에 값이 있으면 URL에 반영) =====
    if (!urlCredentialId && selectedCredentialId && credentials.length > 0) {
      // 10. Store에 자격 증명이 있고 URL에 없으면 URL에 추가
      const credentialExists = credentials.some(c => c.id === selectedCredentialId);
      if (credentialExists) {
        const params = new URLSearchParams(searchParams.toString());
        params.set('credentialId', selectedCredentialId);
        
        const credential = credentials.find(c => c.id === selectedCredentialId);
        
        // 11. 리전 처리: Store에 리전이 있으면 추가, 없으면 기본 리전 설정
        if (selectedRegion) {
          if (credential && supportsRegionSelection(credential.provider as CloudProvider)) {
            params.set('region', selectedRegion);
          }
        } else if (credential && supportsRegionSelection(credential.provider as CloudProvider)) {
          // Store에 리전이 없으면 기본 리전으로 설정
          const defaultRegion = getDefaultRegionForProvider(credential.provider);
          if (defaultRegion) {
            setSelectedRegion(defaultRegion);
            params.set('region', defaultRegion);
            prevUrlRegionRef.current = defaultRegion;
          }
        }
        
        prevUrlCredentialIdRef.current = selectedCredentialId;
        if (selectedRegion || (credential && getDefaultRegionForProvider(credential.provider))) {
          prevUrlRegionRef.current = selectedRegion || (credential ? getDefaultRegionForProvider(credential.provider) : null);
        }
        router.replace(`${pathname}?${params.toString()}`, { scroll: false });
      } else {
        // 12. Store의 자격 증명이 현재 워크스페이스에 없으면 Store 초기화
        setSelectedCredential(null);
        setSelectedRegion(null);
      }
    } else if (urlCredentialId && !urlRegion && credentials.length > 0) {
      // 13. URL에 자격 증명은 있지만 리전이 없는 경우 기본 리전 설정
      const credential = credentials.find(c => c.id === urlCredentialId);
      if (credential && supportsRegionSelection(credential.provider as CloudProvider)) {
        const defaultRegion = getDefaultRegionForProvider(credential.provider);
        if (defaultRegion) {
          setSelectedRegion(defaultRegion);
          const params = new URLSearchParams(searchParams.toString());
          params.set('region', defaultRegion);
          prevUrlRegionRef.current = defaultRegion;
          router.replace(`${pathname}?${params.toString()}`, { scroll: false });
        }
      }
    } else if (urlCredentialId && !urlRegion && selectedRegion && credentials.length > 0) {
      // 14. URL에 자격 증명은 있지만 리전이 없고 Store에 리전이 있는 경우 URL에 추가
      const credential = credentials.find(c => c.id === urlCredentialId);
      if (credential && supportsRegionSelection(credential.provider as CloudProvider)) {
        const params = new URLSearchParams(searchParams.toString());
        params.set('region', selectedRegion);
        prevUrlRegionRef.current = selectedRegion;
        router.replace(`${pathname}?${params.toString()}`, { scroll: false });
      }
    }
  }, [searchParams, shouldShowSelectors, isLoadingCredentials, selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion, pathname, router, credentials, currentWorkspace]);

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <header className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60" role="banner">
      <div className="flex flex-col">
        {/* Main Header Row */}
        <div className="flex h-16 items-center justify-between px-4 sm:px-6">
          <div className="flex items-center space-x-2 sm:space-x-4 flex-1 min-w-0 overflow-hidden">
            <MobileNav />
            <div className="flex-1 min-w-0 overflow-hidden">
              <Breadcrumb className="text-sm truncate" />
            </div>
          </div>

          <div className="flex items-center space-x-1 sm:space-x-2 md:space-x-4 flex-shrink-0">

            {/* Language Selector */}
            <Select
              value={locale}
              onValueChange={(value) => setLocale(value as Locale)}
            >
              <SelectTrigger className="w-[120px] h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {locales.map((loc) => (
                  <SelectItem key={loc} value={loc} className="text-xs">
                    {localeNames[loc]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <ThemeToggle />
            {user ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button 
                  variant="ghost" 
                  className="relative h-8 w-8 rounded-full"
                  aria-label={`User menu for ${user.username}`}
                  aria-haspopup="menu"
                >
                  <Avatar className="h-8 w-8">
                    <AvatarImage src="" alt={`${user.username}'s avatar`} />
                    <AvatarFallback>
                      {user.username.charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56" align="end" forceMount role="menu">
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">{user.username}</p>
                    <p className="text-xs leading-none text-muted-foreground">
                      {user.email}
                    </p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={() => router.push('/profile')}
                  role="menuitem"
                  aria-label="Go to profile page"
                >
                  <User className="mr-2 h-4 w-4" aria-hidden="true" />
                  <span>{typeof t === 'function' ? t('user.profile') : 'Profile'}</span>
                </DropdownMenuItem>
                <DropdownMenuItem 
                  onClick={() => router.push('/settings')}
                  role="menuitem"
                  aria-label="Go to settings page"
                >
                  <Settings className="mr-2 h-4 w-4" aria-hidden="true" />
                  <span>{typeof t === 'function' ? t('user.settings') : 'Settings'}</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={handleLogout}
                  role="menuitem"
                  aria-label="Log out of account"
                >
                  <LogOut className="mr-2 h-4 w-4" aria-hidden="true" />
                  <span>{typeof t === 'function' ? t('user.logout') : 'Log out'}</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <div className="flex items-center space-x-2">
              <Button variant="ghost" onClick={() => router.push('/login')}>
                {typeof t === 'function' ? t('user.login') : 'Login'}
              </Button>
              <Button onClick={() => router.push('/register')}>
                {typeof t === 'function' ? t('user.signUp') : 'Sign Up'}
              </Button>
            </div>
          )}
          </div>
        </div>
      </div>
    </header>
  );
}

export const Header = React.memo(HeaderComponent);

'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useRouter, usePathname, useSearchParams } from 'next/navigation';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Home,
  Server,
  Key,
  Container,
  Network,
  ChevronDown,
  Image,
  HardDrive,
  Layers,
  Shield,
  Building2,
  FolderTree,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { ScreenReaderOnly } from '@/components/accessibility/screen-reader-only';
import { workspaceService } from '@/features/workspaces';
import { useToast } from '@/hooks/use-toast';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { CredentialSelector } from './credential-selector';
import { useCredentials } from '@/hooks/use-credentials';
import { buildCredentialResourcePath, buildWorkspaceManagementPath, buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { useResourceContext } from '@/hooks/use-resource-context';

// 계층적 메뉴 구조의 네비게이션
interface NavigationItem {
  name: string;
  href?: string;
  icon: React.ComponentType<{ className?: string }>;
  children?: NavigationItem[];
}

// 네비게이션은 번역을 사용하여 동적으로 생성됨
// selectedProvider를 파라미터로 받아 Azure인 경우 IAM 메뉴 추가
// workspaceId와 credentialId를 받아 새로운 라우팅 구조에 맞게 경로 생성
export function getNavigation(
  t: (key: string) => string, 
  selectedProvider?: string,
  workspaceId?: string,
  credentialId?: string
): NavigationItem[] {
  const navItems: NavigationItem[] = [
    { 
      name: t('nav.dashboard'), 
      href: workspaceId ? buildWorkspaceManagementPath(workspaceId, 'dashboard') : '/dashboard', 
      icon: Home 
    },
    {
      name: t('nav.compute'),
      icon: Server,
      children: [
        { 
          name: t('nav.vms'), 
          href: workspaceId && credentialId 
            ? buildCredentialResourcePath(workspaceId, credentialId, 'compute', '/vms')
            : '/compute/vms', 
          icon: Server 
        },
        { 
          name: t('nav.images'), 
          href: workspaceId && credentialId 
            ? buildCredentialResourcePath(workspaceId, credentialId, 'compute', '/images')
            : '/compute/images', 
          icon: Image 
        },
        { 
          name: t('nav.snapshots'), 
          href: workspaceId && credentialId 
            ? buildCredentialResourcePath(workspaceId, credentialId, 'compute', '/snapshots')
            : '/compute/snapshots', 
          icon: HardDrive 
        },
      ],
    },
    {
      name: t('nav.kubernetes'),
      icon: Container,
      children: [
        { 
          name: t('nav.clusters'), 
          href: workspaceId && credentialId 
            ? buildCredentialResourcePath(workspaceId, credentialId, 'k8s', '/clusters')
            : '/kubernetes/clusters', 
          icon: Container 
        },
        // AWS인 경우 Node Groups, 그 외는 Node Pools
        ...(selectedProvider === 'aws'
          ? [{ 
              name: t('nav.nodeGroups'), 
              href: workspaceId && credentialId 
                ? buildCredentialResourcePath(workspaceId, credentialId, 'k8s', '/node-groups')
                : '/kubernetes/node-groups', 
              icon: Layers 
            }]
          : [{ 
              name: t('nav.nodePools'), 
              href: workspaceId && credentialId 
                ? buildCredentialResourcePath(workspaceId, credentialId, 'k8s', '/node-pools')
                : '/kubernetes/node-pools', 
              icon: Layers 
            }]
        ),
        { 
          name: t('nav.nodes'), 
          href: workspaceId && credentialId
            ? buildCredentialResourcePath(workspaceId, credentialId, 'k8s', '/nodes')
            : '/kubernetes/nodes',
          icon: Building2 
        },
      ],
    },
    {
      name: t('nav.networks'),
      icon: Network,
      children: [
        { 
          name: t('nav.vpcs'), 
          href: workspaceId && credentialId 
            ? buildCredentialResourcePath(workspaceId, credentialId, 'networks', '/vpcs')
            : '/networks/vpcs', 
          icon: Network 
        },
        { 
          name: t('nav.subnets'), 
          href: workspaceId && credentialId 
            ? buildCredentialResourcePath(workspaceId, credentialId, 'networks', '/subnets')
            : '/networks/subnets', 
          icon: Layers 
        },
        { 
          name: t('nav.securityGroups'), 
          href: workspaceId && credentialId 
            ? buildCredentialResourcePath(workspaceId, credentialId, 'networks', '/security-groups')
            : '/networks/security-groups', 
          icon: Shield 
        },
      ],
    },
  ];

  // Azure인 경우 IAM 메뉴 추가
  if (selectedProvider === 'azure') {
    navItems.push({
      name: t('nav.iam'),
      icon: Shield,
      children: [
        { 
          name: t('nav.resourceGroups'), 
          href: workspaceId && credentialId 
            ? buildCredentialResourcePath(workspaceId, credentialId, 'azure', '/iam/resource-groups')
            : '/azure/iam/resource-groups', 
          icon: FolderTree 
        },
      ],
    });
  }

  // Credentials는 Workspace 상세 페이지로 통합되었으므로 사이드바에서 제거

  return navItems;
}

/**
 * Sidebar 컴포넌트
 * 
 * 메인 사이드바 네비게이션을 제공합니다.
 * Workspace 선택, Credential 선택, 네비게이션 메뉴를 포함합니다.
 * 
 * @example
 * ```tsx
 * // Layout에서 자동으로 사용됨
 * <Layout>
 *   <Sidebar />  // 자동으로 렌더링됨
 * </Layout>
 * ```
 */
function SidebarComponent() {
  const { currentWorkspace, setCurrentWorkspace, workspaces, setWorkspaces } = useWorkspaceStore();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const { t } = useTranslation();
  
  // 현재 선택된 credential의 provider 확인 (Azure IAM 메뉴 표시용)
  const { selectedCredentialId } = useCredentialContextStore();
  const { selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: !!currentWorkspace,
  });

  // Path Parameter에서 컨텍스트 추출 (새로운 라우팅 구조)
  const resourceContext = useResourceContext();
  const finalWorkspaceId = resourceContext.workspaceId || currentWorkspace?.id;
  const finalCredentialId = resourceContext.credentialId || selectedCredentialId || undefined;
  
  // 1. 네비게이션 메뉴 생성: 번역 함수를 사용하여 다국어 지원
  // t가 함수인지 확인하여 안전하게 처리
  const navigation = React.useMemo(() => {
    if (typeof t === 'function') {
      return getNavigation(t, selectedProvider, finalWorkspaceId, finalCredentialId);
    }
    // Fallback: 기본 영어 메뉴 (번역 함수가 없을 경우)
    const fallbackNav: NavigationItem[] = [
      { name: 'Dashboard', href: '/dashboard', icon: Home },
      {
        name: 'Compute',
        icon: Server,
        children: [
          { name: 'VMs', href: '/compute/vms', icon: Server },
          { name: 'Images', href: '/compute/images', icon: Image },
          { name: 'Snapshots', href: '/compute/snapshots', icon: HardDrive },
        ],
      },
      {
        name: 'Kubernetes',
        icon: Container,
        children: [
          { name: 'Clusters', href: '/kubernetes/clusters', icon: Container },
          { name: 'Node Pools', href: '/kubernetes/node-pools', icon: Layers },
          { name: 'Nodes', href: '/kubernetes/nodes', icon: Building2 },
        ],
      },
      {
        name: 'Networks',
        icon: Network,
        children: [
          { name: 'VPCs', href: '/networks/vpcs', icon: Network },
          { name: 'Subnets', href: '/networks/subnets', icon: Layers },
          { name: 'Security Groups', href: '/networks/security-groups', icon: Shield },
        ],
      },
    ];

    // Azure인 경우 IAM 메뉴 추가
    if (selectedProvider === 'azure') {
      fallbackNav.push({
        name: 'IAM',
        icon: Shield,
        children: [
          { name: 'Resource Groups', href: '/azure/iam/resource-groups', icon: FolderTree },
        ],
      });
    }

    fallbackNav.push({ name: 'Credentials', href: '/credentials', icon: Key });

    return fallbackNav;
  }, [t, selectedProvider]);

  // 2. 워크스페이스 목록 조회 (React Query 사용)
  const { data: fetchedWorkspaces = [] } = useQuery({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    staleTime: CACHE_TIMES.STABLE,
    gcTime: GC_TIMES.LONG,
    retry: 3,
    retryDelay: 1000,
  });

  // 3. 워크스페이스 목록이 조회되면 스토어에 업데이트
  // 이전 워크스페이스 목록을 추적하여 변경 시에만 업데이트
  const prevWorkspacesRef = React.useRef<string>('');
  const prevCurrentWorkspaceIdRef = React.useRef<string | undefined>(currentWorkspace?.id);
  
  React.useEffect(() => {
    // 워크스페이스 목록의 ID 문자열을 생성하여 변경 감지
    const currentWorkspacesIds = fetchedWorkspaces.map(w => w.id).sort().join(',');
    
    // 목록이 변경된 경우에만 스토어 업데이트
    if (prevWorkspacesRef.current !== currentWorkspacesIds) {
      prevWorkspacesRef.current = currentWorkspacesIds;
      
      if (fetchedWorkspaces.length > 0) {
        setWorkspaces(fetchedWorkspaces);
        
        // 현재 워크스페이스가 없거나 삭제된 경우에만 첫 번째 워크스페이스로 자동 선택
        // removeWorkspace가 이미 자동 전환을 처리하므로, 여기서는 유효성 검사만 수행
        const currentWorkspaceId = currentWorkspace?.id;
        const isCurrentWorkspaceValid = currentWorkspaceId && fetchedWorkspaces.find(w => w.id === currentWorkspaceId);
        
        // 이전 값과 비교하여 실제로 변경이 필요한 경우에만 업데이트
        if ((!currentWorkspaceId || !isCurrentWorkspaceValid) && prevCurrentWorkspaceIdRef.current !== fetchedWorkspaces[0]?.id) {
          setCurrentWorkspace(fetchedWorkspaces[0]);
          prevCurrentWorkspaceIdRef.current = fetchedWorkspaces[0]?.id;
        } else if (currentWorkspaceId) {
          prevCurrentWorkspaceIdRef.current = currentWorkspaceId;
        }
      } else {
        // 워크스페이스가 없으면 스토어 초기화
        setWorkspaces([]);
        if (currentWorkspace !== null && prevCurrentWorkspaceIdRef.current !== undefined) {
          setCurrentWorkspace(null);
          prevCurrentWorkspaceIdRef.current = undefined;
        }
      }
    }
    // fetchedWorkspaces만 의존성으로 사용하여 무한 루프 방지
    // currentWorkspace는 useRef로 추적하므로 의존성에 포함하지 않음
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fetchedWorkspaces]);
  
  // currentWorkspace 변경 시 ref 업데이트 (렌더링 트리거 없이)
  React.useEffect(() => {
    prevCurrentWorkspaceIdRef.current = currentWorkspace?.id;
  }, [currentWorkspace?.id]);

  // 4. Settings/Members 페이지에서 URL 파라미터로부터 현재 워크스페이스 동기화
  React.useEffect(() => {
    if (fetchedWorkspaces.length === 0) return;
    
    // 4-1. Settings 또는 Members 페이지인지 확인
    if (pathname.startsWith('/workspaces/') && (pathname.includes('/settings') || pathname.includes('/members'))) {
      // 4-2. URL 경로에서 워크스페이스 ID 추출
      const match = pathname.match(/\/workspaces\/([^/]+)\/(settings|members)/);
      if (match && match[1]) {
        const urlWorkspaceId = match[1];
        const workspaceFromUrl = fetchedWorkspaces.find(w => w.id === urlWorkspaceId);
        
        // 4-3. 워크스페이스가 존재하고 현재 워크스페이스와 다르면 업데이트
        if (workspaceFromUrl && currentWorkspace?.id !== urlWorkspaceId) {
          setCurrentWorkspace(workspaceFromUrl);
        }
      }
    }
  }, [pathname, fetchedWorkspaces, currentWorkspace?.id, setCurrentWorkspace]);


  // 7. 현재 경로에 따라 열려야 할 Accordion 항목 결정
  // useState로 관리하여 pathname 변경 시 자동 업데이트
  const getOpenItemsFromPathname = React.useCallback((currentPathname: string): string[] => {
    const items: string[] = [];
    // 새로운 라우팅 구조: /w/{workspaceId}/c/{credentialId}/k8s/... 또는 /w/{workspaceId}/c/{credentialId}/compute/...
    // 기존 구조: /{workspaceId}/{credentialId}/kubernetes/... 또는 /{workspaceId}/{credentialId}/compute/...
    
    // Compute 체크: /compute 또는 /c/.../compute 포함
    if (currentPathname.includes('/compute')) {
      items.push('compute');
    }
    
    // Kubernetes 체크: /kubernetes, /k8s, 또는 /c/.../k8s 포함
    if (currentPathname.includes('/kubernetes') || currentPathname.includes('/k8s')) {
      items.push('kubernetes');
    }
    
    // Networks 체크: /networks 또는 /c/.../networks 포함
    if (currentPathname.includes('/networks')) {
      items.push('networks');
    }
    
    // Azure IAM 체크: /azure/iam, /azure, 또는 /c/.../azure 포함
    if (currentPathname.includes('/azure/iam') || currentPathname.includes('/azure')) {
      items.push('iam');
    }
    
    return items;
  }, []);

  const [openItems, setOpenItems] = React.useState<string[]>(() => {
    return getOpenItemsFromPathname(pathname);
  });

  // 8. pathname이 변경될 때마다 openItems 업데이트 (현재 페이지에 맞는 섹션 자동 열기)
  React.useEffect(() => {
    const items = getOpenItemsFromPathname(pathname);
    setOpenItems(items);
  }, [pathname, getOpenItemsFromPathname]);

  /**
   * 네비게이션 항목이 활성화되어 있는지 확인
   * 
   * @param item - 확인할 네비게이션 항목
   * @returns 활성화 여부
   */
  const isItemActive = (item: NavigationItem): boolean => {
    // 1. href가 있으면 현재 경로와 비교
    if (item.href) {
      // 새로운 라우팅 구조에서는 경로가 다를 수 있으므로 포함 여부로 확인
      return pathname === item.href || pathname.includes(item.href.split('?')[0]);
    }
    // 2. 자식 항목이 있으면 자식 중 하나라도 활성화되어 있는지 확인
    if (item.children) {
      return item.children.some(child => isItemActive(child));
    }
    return false;
  };

  /**
   * 자식 네비게이션 항목이 활성화되어 있는지 확인
   * 
   * @param item - 확인할 네비게이션 항목
   * @returns 활성화 여부
   */
  const isChildActive = (item: NavigationItem): boolean => {
    if (item.href) {
      // 새로운 라우팅 구조에서는 경로가 다를 수 있으므로 포함 여부로 확인
      return pathname === item.href || pathname.includes(item.href.split('?')[0]);
    }
    return false;
  };

  return (
    <nav className="flex h-full w-64 flex-col bg-card border-r lg:block hidden" role="navigation" aria-label="Main navigation">
      <div className="flex flex-col flex-1 overflow-y-auto">
        <div className="flex flex-col flex-1 px-3 py-4">
          {/* Logo */}
          <div className="mb-6 pt-4">
            <Button
              variant="ghost"
              className="w-full justify-start px-2 h-auto hover:bg-transparent"
              onClick={() => {
                if (finalWorkspaceId) {
                  router.push(buildWorkspaceManagementPath(finalWorkspaceId, 'dashboard'));
                } else {
                  router.push('/dashboard');
                }
              }}
              aria-label="Go to dashboard"
            >
              <h1 className="text-lg sm:text-xl font-bold text-foreground">
                SkyClust
                <ScreenReaderOnly>Multi-Cloud Management Platform</ScreenReaderOnly>
              </h1>
            </Button>
          </div>

          {/* Credential Selector */}
          <CredentialSelector />

          {/* Navigation Menu */}
          <div className="mt-4 flex-1" role="list">
            <Accordion 
              type="multiple" 
              value={openItems} 
              onValueChange={setOpenItems}
              className="w-full"
            >
              {navigation.map((item) => {
                if (item.children) {
                  // 자식이 있는 부모 항목 (accordion)
                  const isActive = isItemActive(item);
                  
                  // pathname 기반으로 accordion 값 결정
                  // openItems 상태와 일관된 매칭 보장
                  // 새로운 라우팅 구조와 기존 구조 모두 지원
                  let accordionValue = '';
                  
                  // Compute 체크
                  if (item.children.some(child => {
                    const href = child.href || '';
                    return href.includes('/compute');
                  })) {
                    accordionValue = 'compute';
                  } 
                  // Kubernetes 체크
                  else if (item.children.some(child => {
                    const href = child.href || '';
                    return href.includes('/kubernetes') || href.includes('/k8s');
                  })) {
                    accordionValue = 'kubernetes';
                  } 
                  // Networks 체크
                  else if (item.children.some(child => {
                    const href = child.href || '';
                    return href.includes('/networks');
                  })) {
                    accordionValue = 'networks';
                  } 
                  // Azure IAM 체크
                  else if (item.children.some(child => {
                    const href = child.href || '';
                    return href.includes('/azure/iam') || href.includes('/azure');
                  })) {
                    accordionValue = 'iam';
                  }
                  
                  return (
                    <AccordionItem key={item.name} value={accordionValue} className="border-none">
                      <AccordionTrigger
                        className={cn(
                          'py-2 px-3 hover:no-underline',
                          isActive && 'bg-accent'
                        )}
                      >
                        <div className="flex items-center w-full">
                          <item.icon className="mr-3 h-5 w-5 flex-shrink-0" aria-hidden="true" />
                          <span className="text-sm font-medium">{item.name}</span>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent className="pb-1 pt-0">
                        <div className="ml-4 space-y-1">
                          {item.children.map((child) => {
                            // 9. 자식 항목의 활성화 상태 확인
                            const isChildItemActive = isChildActive(child);
                            
                            /**
                             * 네비게이션 핸들러
                             * URL 파라미터를 유지하면서 페이지 이동
                             */
                            const handleNavigation = () => {
                              if (!child.href) return;
                              
                              // 9-1. 현재 URL의 파라미터 유지 (workspaceId, credentialId, region)
                              const params = new URLSearchParams(searchParams.toString());
                              
                              // 9-2. workspaceId는 항상 유지
                              if (currentWorkspace?.id) {
                                params.set('workspaceId', currentWorkspace.id);
                              }
                              
                              // 9-3. credentialId와 region은 compute/kubernetes/networks/azure 경로에서만 유지
                              const shouldKeepParams = child.href.startsWith('/compute') || 
                                                      child.href.startsWith('/kubernetes') || 
                                                      child.href.startsWith('/networks') ||
                                                      child.href.startsWith('/azure');
                              
                              if (!shouldKeepParams) {
                                params.delete('credentialId');
                                params.delete('region');
                              }
                              
                              // 9-4. 쿼리 스트링이 있으면 URL에 추가
                              const queryString = params.toString();
                              const url = queryString ? `${child.href}?${queryString}` : child.href;
                              router.push(url);
                            };
                            
                            return (
                              <Button
                                key={child.name}
                                onClick={handleNavigation}
                                variant={isChildItemActive ? 'secondary' : 'ghost'}
                                size="sm"
                                className={cn(
                                  'w-full justify-start text-sm',
                                  isChildItemActive && 'bg-accent font-medium'
                                )}
                                role="listitem"
                                aria-current={isChildItemActive ? 'page' : undefined}
                                aria-label={`Navigate to ${child.name} page`}
                              >
                                <child.icon className="mr-2 h-4 w-4" aria-hidden="true" />
                                {child.name}
                              </Button>
                            );
                          })}
                        </div>
                      </AccordionContent>
                    </AccordionItem>
                  );
                } else {
                  // 10. 자식이 없는 단일 항목
                  const isActive = pathname === item.href;
                  
                  /**
                   * 네비게이션 핸들러
                   * 새로운 라우팅 구조에 맞게 페이지 이동
                   */
                  const handleNavigation = () => {
                    if (!item.href) return;
                    
                    // 새로운 라우팅 구조에서는 href가 이미 올바른 경로를 포함하므로 직접 사용
                    router.push(item.href);
                  };
                  
                  return (
                    <Button
                      key={item.name}
                      onClick={handleNavigation}
                      variant={isActive ? 'secondary' : 'ghost'}
                      className={cn(
                        'w-full justify-start mb-1',
                        isActive && 'bg-accent font-medium'
                      )}
                      role="listitem"
                      aria-current={isActive ? 'page' : undefined}
                      aria-label={`Navigate to ${item.name} page`}
                    >
                      <item.icon className="mr-3 h-5 w-5" aria-hidden="true" />
                      {item.name}
                    </Button>
                  );
                }
              })}
            </Accordion>
          </div>
        </div>
      </div>
    </nav>
  );
}

export const Sidebar = React.memo(SidebarComponent);

'use client';

import * as React from 'react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Form } from '@/components/ui/form';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useRouter, usePathname, useSearchParams } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Home,
  Server,
  Key,
  Plus,
  Container,
  Network,
  Settings,
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
import { ErrorHandler } from '@/lib/error-handling';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { CreateWorkspaceForm } from '@/lib/types';
import * as z from 'zod';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { CredentialSelector } from './credential-selector';
import { useCredentials } from '@/hooks/use-credentials';

const createWorkspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters'),
});

// 계층적 메뉴 구조의 네비게이션
interface NavigationItem {
  name: string;
  href?: string;
  icon: React.ComponentType<{ className?: string }>;
  children?: NavigationItem[];
}

// 네비게이션은 번역을 사용하여 동적으로 생성됨
// selectedProvider를 파라미터로 받아 Azure인 경우 IAM 메뉴 추가
function getNavigation(t: (key: string) => string, selectedProvider?: string): NavigationItem[] {
  const navItems: NavigationItem[] = [
    { name: t('nav.dashboard'), href: '/dashboard', icon: Home },
    {
      name: t('nav.compute'),
      icon: Server,
      children: [
        { name: t('nav.vms'), href: '/compute/vms', icon: Server },
        { name: t('nav.images'), href: '/compute/images', icon: Image },
        { name: t('nav.snapshots'), href: '/compute/snapshots', icon: HardDrive },
      ],
    },
    {
      name: t('nav.kubernetes'),
      icon: Container,
      children: [
        { name: t('nav.clusters'), href: '/kubernetes/clusters', icon: Container },
        // AWS인 경우 Node Groups, 그 외는 Node Pools
        ...(selectedProvider === 'aws'
          ? [{ name: t('nav.nodeGroups'), href: '/kubernetes/node-groups', icon: Layers }]
          : [{ name: t('nav.nodePools'), href: '/kubernetes/node-pools', icon: Layers }]
        ),
        { name: t('nav.nodes'), href: '/kubernetes/nodes', icon: Building2 },
      ],
    },
    {
      name: t('nav.networks'),
      icon: Network,
      children: [
        { name: t('nav.vpcs'), href: '/networks/vpcs', icon: Network },
        { name: t('nav.subnets'), href: '/networks/subnets', icon: Layers },
        { name: t('nav.securityGroups'), href: '/networks/security-groups', icon: Shield },
      ],
    },
  ];

  // Azure인 경우 IAM 메뉴 추가
  if (selectedProvider === 'azure') {
    navItems.push({
      name: t('nav.iam'),
      icon: Shield,
      children: [
        { name: t('nav.resourceGroups'), href: '/azure/iam/resource-groups', icon: FolderTree },
      ],
    });
  }

  navItems.push({ name: t('nav.credentials'), href: '/credentials', icon: Key });

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
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const { t } = useTranslation();
  
  // 현재 선택된 credential의 provider 확인 (Azure IAM 메뉴 표시용)
  const { selectedCredentialId } = useCredentialContextStore();
  const { selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: !!currentWorkspace,
  });
  
  // 1. 네비게이션 메뉴 생성: 번역 함수를 사용하여 다국어 지원
  // t가 함수인지 확인하여 안전하게 처리
  const navigation = React.useMemo(() => {
    if (typeof t === 'function') {
      return getNavigation(t, selectedProvider);
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

  // 워크스페이스 생성 폼
  const {
    form,
    handleSubmit,
    isLoading: isFormLoading,
    error: formError,
    reset,
    getFieldError,
    getFieldValidationState,
  } = useFormWithValidation<CreateWorkspaceForm>({
    schema: createWorkspaceSchema,
    defaultValues: {
      name: '',
      description: '',
    },
    onSubmit: async (data) => {
      await workspaceService.createWorkspace(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      setIsCreateDialogOpen(false);
      reset();
      // t 함수를 안전하게 사용
      const successMessage = typeof t === 'function' 
        ? t('messages.created', { resource: t('workspace.title') })
        : 'Workspace created successfully';
      success(successMessage);
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'createWorkspace', source: 'sidebar' });
      // t 함수를 안전하게 사용
      const errorMessage = typeof t === 'function'
        ? t('messages.operationFailed')
        : 'Failed to create workspace';
      showError(errorMessage);
    },
    resetOnSuccess: true,
  });

  /**
   * 워크스페이스 변경 핸들러
   * 
   * 워크스페이스 변경 시:
   * - 자격 증명 및 리전 선택 초기화
   * - 관련 쿼리 무효화
   * - URL 파라미터 업데이트
   * 
   * @param workspaceId - 선택할 워크스페이스 ID
   */
  const handleWorkspaceChange = (workspaceId: string) => {
    // 1. 유효하지 않은 워크스페이스 ID면 스킵
    if (workspaceId === 'all' || !workspaceId) return;
    
    // 2. 워크스페이스 찾기
    const workspace = fetchedWorkspaces.find(w => w.id === workspaceId);
    if (!workspace) return;
    
    // 3. 워크스페이스 변경 여부 확인
    const previousWorkspaceId = currentWorkspace?.id;
    const isWorkspaceChanged = previousWorkspaceId !== workspace.id;
    
    // 4. 현재 워크스페이스 업데이트
    setCurrentWorkspace(workspace);
    
    if (isWorkspaceChanged) {
      // 5. 워크스페이스가 변경된 경우
      // 5-1. 자격 증명 및 리전 선택 초기화
      const { clearSelection } = useCredentialContextStore.getState();
      clearSelection();
      
      // 5-2. URL 파라미터 업데이트 (workspaceId는 유지, credentialId와 region은 제거)
      const params = new URLSearchParams(window.location.search);
      params.set('workspaceId', workspace.id);
      params.delete('credentialId');
      params.delete('region');
      
      // 5-3. 관련 쿼리 무효화 (이전 워크스페이스의 데이터는 더 이상 유효하지 않음)
      queryClient.invalidateQueries({
        predicate: (query) => {
          const key = query.queryKey;
          if (!Array.isArray(key)) return false;
          
          const keyString = key.join('-').toLowerCase();
          
          // 무효화할 쿼리 키 패턴 확인
          const shouldInvalidate = 
            (previousWorkspaceId && key.includes(previousWorkspaceId)) ||
            keyString.includes('vms') ||
            keyString.includes('credentials') ||
            keyString.includes('kubernetes') ||
            keyString.includes('clusters') ||
            keyString.includes('node-pools') ||
            keyString.includes('node-groups') ||
            keyString.includes('nodes') ||
            keyString.includes('vpcs') ||
            keyString.includes('subnets') ||
            keyString.includes('security-groups');
          
          return shouldInvalidate;
        },
      });
      
      // 5-4. Settings/Members 페이지인 경우 해당 페이지로 리다이렉트
      if (pathname.startsWith('/workspaces/') && (pathname.includes('/settings') || pathname.includes('/members'))) {
        const pageType = pathname.includes('/settings') ? 'settings' : 'members';
        router.push(`/workspaces/${workspaceId}/${pageType}`);
        return;
      }
      
      // 5-5. 현재 경로에 업데이트된 파라미터 적용
      const currentPath = pathname;
      router.replace(`${currentPath}?${params.toString()}`, { scroll: false });
    } else {
      // 6. 같은 워크스페이스인 경우 URL에 workspaceId만 업데이트
      const currentPath = pathname;
      const params = new URLSearchParams(window.location.search);
      params.set('workspaceId', workspace.id);
      router.replace(`${currentPath}?${params.toString()}`, { scroll: false });
    }
  };

  const displayWorkspaces = fetchedWorkspaces.length > 0 ? fetchedWorkspaces : workspaces;

  // 7. 현재 경로에 따라 열려야 할 Accordion 항목 결정
  // useState로 관리하여 pathname 변경 시 자동 업데이트
  const [openItems, setOpenItems] = React.useState<string[]>(() => {
    const items: string[] = [];
    // 현재 경로에 따라 해당 섹션을 열림 상태로 설정
    if (pathname.startsWith('/compute')) items.push('compute');
    if (pathname.startsWith('/kubernetes')) items.push('kubernetes');
    if (pathname.startsWith('/networks')) items.push('networks');
    if (pathname.startsWith('/azure/iam')) items.push('iam');
    return items;
  });

  // 8. pathname이 변경될 때마다 openItems 업데이트 (현재 페이지에 맞는 섹션 자동 열기)
  React.useEffect(() => {
    const items: string[] = [];
    if (pathname.startsWith('/compute')) items.push('compute');
    if (pathname.startsWith('/kubernetes')) items.push('kubernetes');
    if (pathname.startsWith('/networks')) items.push('networks');
    if (pathname.startsWith('/azure/iam')) items.push('iam');
    setOpenItems(items);
  }, [pathname]);

  /**
   * 네비게이션 항목이 활성화되어 있는지 확인
   * 
   * @param item - 확인할 네비게이션 항목
   * @returns 활성화 여부
   */
  const isItemActive = (item: NavigationItem): boolean => {
    // 1. href가 있으면 현재 경로와 비교
    if (item.href) {
      return pathname === item.href || pathname.startsWith(item.href + '/');
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
      return pathname === item.href || pathname.startsWith(item.href + '/');
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
              onClick={() => router.push('/dashboard')}
              aria-label="Go to dashboard"
            >
              <h1 className="text-lg sm:text-xl font-bold text-foreground">
                SkyClust
                <ScreenReaderOnly>Multi-Cloud Management Platform</ScreenReaderOnly>
              </h1>
            </Button>
          </div>

          {/* Workspace Selector */}
          <div className="mb-4 space-y-2">
            <label className="text-sm font-medium text-muted-foreground">
              <ScreenReaderOnly>{t('workspace.select')}</ScreenReaderOnly>
              {t('workspace.title')}
            </label>
            <Select
              value={currentWorkspace?.id || 'all'}
              onValueChange={handleWorkspaceChange}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder={t('workspace.select')} />
              </SelectTrigger>
              <SelectContent>
                {displayWorkspaces.map((workspace) => (
                  <SelectItem key={workspace.id} value={workspace.id} className="truncate">
                    {workspace.name}
                  </SelectItem>
                ))}
                <div className="h-px bg-border my-1" />
                {currentWorkspace && (
                  <div className="px-2 py-1.5">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start"
                      onClick={(e) => {
                        e.stopPropagation();
                        router.push(`/workspaces/${currentWorkspace.id}/settings`);
                      }}
                      aria-label="Open workspace settings"
                    >
                      <Settings className="mr-2 h-4 w-4" aria-hidden="true" />
                      {t('workspace.settings')}
                    </Button>
                  </div>
                )}
                <div className="px-2 py-1.5 border-t border-border">
                  <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
                    <DialogTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start"
                        onClick={(e) => {
                          e.stopPropagation();
                        }}
                        aria-label="Create a new workspace"
                      >
                        <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                        {t('workspace.createNew')}
                      </Button>
                        </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>{t('workspace.createNew')}</DialogTitle>
                        <DialogDescription>
                          {t('workspace.createNewDescription')}
                        </DialogDescription>
                      </DialogHeader>
                      <Form {...form}>
                        <form onSubmit={handleSubmit} className="space-y-4">
                          <EnhancedField
                            name="name"
                            label={t('workspace.name')}
                            type="text"
                            placeholder={t('workspace.name')}
                            required
                            getFieldError={getFieldError}
                            getFieldValidationState={getFieldValidationState}
                          />
                          <EnhancedField
                            name="description"
                            label={t('workspace.description')}
                            type="textarea"
                            placeholder={t('workspace.description')}
                            required
                            getFieldError={getFieldError}
                            getFieldValidationState={getFieldValidationState}
                          />
                          
                          {formError && (
                            <div className="text-sm text-red-600 text-center" role="alert">
                              {formError}
                            </div>
                          )}

                          <div className="flex justify-end space-x-2">
                            <Button
                              type="button"
                              variant="outline"
                              onClick={() => {
                                reset();
                                setIsCreateDialogOpen(false);
                              }}
                            >
                              {t('common.cancel')}
                            </Button>
                            <Button type="submit" disabled={isFormLoading}>
                              {isFormLoading ? t('common.loading') : t('workspace.create')}
                            </Button>
                          </div>
                        </form>
                      </Form>
                    </DialogContent>
                  </Dialog>
                </div>
              </SelectContent>
            </Select>
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
                  let accordionValue = '';
                  if (item.children.some(child => child.href?.startsWith('/compute'))) {
                    accordionValue = 'compute';
                  } else if (item.children.some(child => child.href?.startsWith('/kubernetes'))) {
                    accordionValue = 'kubernetes';
                  } else if (item.children.some(child => child.href?.startsWith('/networks'))) {
                    accordionValue = 'networks';
                  } else if (item.children.some(child => child.href?.startsWith('/azure/iam'))) {
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
                   * URL 파라미터를 유지하면서 페이지 이동
                   */
                  const handleNavigation = () => {
                    if (!item.href) return;
                    
                    // 10-1. 현재 URL의 파라미터 유지 (workspaceId, credentialId, region)
                    const params = new URLSearchParams(searchParams.toString());
                    
                    // 10-2. workspaceId는 항상 유지
                    if (currentWorkspace?.id) {
                      params.set('workspaceId', currentWorkspace.id);
                    }
                    
                    // 10-3. credentialId와 region은 compute/kubernetes/networks 경로에서만 유지
                    const shouldKeepParams = item.href.startsWith('/compute') || 
                                            item.href.startsWith('/kubernetes') || 
                                            item.href.startsWith('/networks');
                    
                    if (!shouldKeepParams) {
                      params.delete('credentialId');
                      params.delete('region');
                    }
                    
                    // 10-4. 쿼리 스트링이 있으면 URL에 추가
                    const queryString = params.toString();
                    const url = queryString ? `${item.href}?${queryString}` : item.href;
                    router.push(url);
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

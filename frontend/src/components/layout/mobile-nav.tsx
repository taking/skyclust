'use client';

import * as React from 'react';
import { useState } from 'react';
import { useRouter, usePathname, useSearchParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Form } from '@/components/ui/form';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { CredentialSelector } from './credential-selector';
import { Menu, Home, Server, Key, Plus, Container, Network, Settings, Image, HardDrive, Layers, Shield, Building2 } from 'lucide-react';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { workspaceService } from '@/features/workspaces';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { CreateWorkspaceForm } from '@/lib/types';
import * as z from 'zod';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';

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

const navigation: NavigationItem[] = [
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
  { name: 'Credentials', href: '/credentials', icon: Key },
];

/**
 * MobileNav 컴포넌트
 * 
 * 모바일 환경에서 사용되는 사이드바 네비게이션입니다.
 * Sheet 컴포넌트를 사용하여 슬라이드 메뉴를 제공합니다.
 * 
 * @example
 * ```tsx
 * // Header에서 자동으로 사용됨
 * <Header>
 *   <MobileNav />  // 모바일에서만 표시됨
 * </Header>
 * ```
 */
export function MobileNav() {
  const [open, setOpen] = useState(false);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const { currentWorkspace, setCurrentWorkspace, workspaces, setWorkspaces } = useWorkspaceStore();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();

  // 1. 키보드 단축키 이벤트 리스너 (Shift + M으로 사이드바 토글)
  React.useEffect(() => {
    const handleToggleSidebar = () => {
      setOpen((prev) => !prev);
    };
    window.addEventListener('toggle-sidebar', handleToggleSidebar);
    return () => {
      window.removeEventListener('toggle-sidebar', handleToggleSidebar);
    };
  }, []);

  // 워크스페이스 조회
  const { data: fetchedWorkspaces = [] } = useQuery({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    staleTime: CACHE_TIMES.STABLE,
    gcTime: GC_TIMES.LONG,
    retry: 3,
    retryDelay: 1000,
  });

  // 워크스페이스 조회 시 스토어 업데이트
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

  // Settings/Members 페이지에서 URL 파라미터로부터 현재 워크스페이스 동기화
  React.useEffect(() => {
    if (fetchedWorkspaces.length === 0) return;
    
    // Settings 또는 Members 페이지인지 확인
    if (pathname.startsWith('/workspaces/') && (pathname.includes('/settings') || pathname.includes('/members'))) {
      // URL 경로에서 워크스페이스 ID 추출
      const match = pathname.match(/\/workspaces\/([^/]+)\/(settings|members)/);
      if (match && match[1]) {
        const urlWorkspaceId = match[1];
        const workspaceFromUrl = fetchedWorkspaces.find(w => w.id === urlWorkspaceId);
        
        // 워크스페이스가 존재하고 현재 워크스페이스와 다르면 업데이트
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
      success('Workspace created successfully');
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'createWorkspace', source: 'mobile-nav' });
      showError('Failed to create workspace');
    },
    resetOnSuccess: true,
  });

  /**
   * 네비게이션 핸들러
   * 
   * URL 파라미터를 유지하면서 페이지 이동하고, 모바일 메뉴를 닫습니다.
   * 
   * @param href - 이동할 경로
   */
  const handleNavigation = (href: string) => {
    // 1. 현재 URL의 파라미터 유지 (workspaceId, credentialId, region)
    const params = new URLSearchParams(searchParams.toString());
    
    // 2. workspaceId는 항상 유지
    if (currentWorkspace?.id) {
      params.set('workspaceId', currentWorkspace.id);
    }
    
    // 3. credentialId와 region은 compute/kubernetes/networks 경로에서만 유지
    const shouldKeepParams = href.startsWith('/compute') || 
                            href.startsWith('/kubernetes') || 
                            href.startsWith('/networks');
    
    if (!shouldKeepParams) {
      params.delete('credentialId');
      params.delete('region');
    }
    
    // 4. 쿼리 스트링이 있으면 URL에 추가
    const queryString = params.toString();
    const url = queryString ? `${href}?${queryString}` : href;
    
    // 5. 페이지 이동 및 모바일 메뉴 닫기
    router.push(url);
    setOpen(false);
  };

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
      
      // 5-2. URL 파라미터 업데이트
      const params = new URLSearchParams(window.location.search);
      params.set('workspaceId', workspace.id);
      params.delete('credentialId');
      params.delete('region');
      
      // 5-3. 관련 쿼리 무효화
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
        handleNavigation(`/workspaces/${workspaceId}/${pageType}`);
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

  // 현재 경로에 따라 열려야 할 accordion 항목 결정
  const getDefaultOpenItems = () => {
    const openItems: string[] = [];
    if (pathname.startsWith('/compute')) openItems.push('compute');
    if (pathname.startsWith('/kubernetes')) openItems.push('kubernetes');
    if (pathname.startsWith('/networks')) openItems.push('networks');
    return openItems;
  };

  // 네비게이션 항목이 활성화되어 있는지 확인
  const isItemActive = (item: NavigationItem): boolean => {
    if (item.href) {
      return pathname === item.href || pathname.startsWith(item.href + '/');
    }
    if (item.children) {
      return item.children.some(child => isItemActive(child));
    }
    return false;
  };

  // 자식 항목이 활성화되어 있는지 확인
  const isChildActive = (item: NavigationItem): boolean => {
    if (item.href) {
      return pathname === item.href || pathname.startsWith(item.href + '/');
    }
    return false;
  };

  return (
    <>
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild>
          <Button variant="ghost" size="sm" className="lg:hidden">
            <Menu className="h-5 w-5" />
          </Button>
        </SheetTrigger>
        <SheetContent side="left" className="w-64">
          <div className="flex flex-col h-full">
            <div className="flex items-center space-x-2 mb-6">
              <h2 className="text-lg font-semibold text-gray-900">SkyClust</h2>
            </div>
            
            {/* Workspace Selector */}
            <div className="mb-4 space-y-2">
              <label className="text-sm font-medium text-gray-500 uppercase tracking-wider">
                Workspace
              </label>
              <Select
                value={currentWorkspace?.id || 'all'}
                onValueChange={handleWorkspaceChange}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select Workspace" />
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
                          handleNavigation(`/workspaces/${currentWorkspace.id}/settings`);
                        }}
                        aria-label="Open workspace settings"
                      >
                        <Settings className="mr-2 h-4 w-4" />
                        Settings
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
                          <Plus className="mr-2 h-4 w-4" />
                          Create Workspace
                        </Button>
                      </DialogTrigger>
                      <DialogContent>
                        <DialogHeader>
                          <DialogTitle>Create New Workspace</DialogTitle>
                          <DialogDescription>
                            Create a new workspace to organize your resources and collaborate with your team.
                          </DialogDescription>
                        </DialogHeader>
                        <Form {...form}>
                          <form onSubmit={handleSubmit} className="space-y-4">
                            <EnhancedField
                              name="name"
                              label="Workspace Name"
                              type="text"
                              placeholder="Enter workspace name"
                              required
                              getFieldError={getFieldError}
                              getFieldValidationState={getFieldValidationState}
                            />
                            <EnhancedField
                              name="description"
                              label="Description"
                              type="textarea"
                              placeholder="Enter workspace description"
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
                                Cancel
                              </Button>
                              <Button type="submit" disabled={isFormLoading}>
                                {isFormLoading ? 'Creating...' : 'Create Workspace'}
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
            <nav className="flex-1 overflow-y-auto">
              <Accordion type="multiple" defaultValue={getDefaultOpenItems()} className="w-full">
                {navigation.map((item) => {
                  if (item.children) {
                    // 자식이 있는 부모 항목 (accordion)
                    const isActive = isItemActive(item);
                    return (
                      <AccordionItem key={item.name} value={item.name.toLowerCase()} className="border-none">
                        <AccordionTrigger
                          className={`py-2 px-3 hover:no-underline ${isActive ? 'bg-accent' : ''}`}
                        >
                          <div className="flex items-center w-full">
                            <item.icon className="mr-3 h-5 w-5 flex-shrink-0" />
                            <span className="text-sm font-medium">{item.name}</span>
                          </div>
                        </AccordionTrigger>
                        <AccordionContent className="pb-1 pt-0">
                          <div className="ml-4 space-y-1">
                            {item.children.map((child) => {
                              const isChildItemActive = isChildActive(child);
                              return (
                                <Button
                                  key={child.name}
                                  onClick={() => {
                                    if (child.href) {
                                      handleNavigation(child.href);
                                    }
                                  }}
                                  variant={isChildItemActive ? 'secondary' : 'ghost'}
                                  size="sm"
                                  className={`w-full justify-start text-sm ${isChildItemActive ? 'bg-accent font-medium' : ''}`}
                                >
                                  <child.icon className="mr-2 h-4 w-4" />
                                  {child.name}
                                </Button>
                              );
                            })}
                          </div>
                        </AccordionContent>
                      </AccordionItem>
                    );
                  } else {
                    // 자식이 없는 단일 항목
                    const isActive = pathname === item.href;
                    return (
                      <Button
                        key={item.name}
                        onClick={() => item.href && handleNavigation(item.href)}
                        variant={isActive ? 'secondary' : 'ghost'}
                        className={`w-full justify-start mb-1 ${isActive ? 'bg-accent font-medium' : ''}`}
                      >
                        <item.icon className="mr-3 h-5 w-5" />
                        {item.name}
                      </Button>
                    );
                  }
                })}
              </Accordion>
            </nav>
          </div>
        </SheetContent>
      </Sheet>
    </>
  );
}

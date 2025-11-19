/**
 * Workspace Switcher Component
 * Header에 표시되는 Workspace 전환 컴포넌트
 * 
 * 기능:
 * - 현재 Workspace 표시
 * - Workspace 목록 드롭다운
 * - 빠른 전환
 */

'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import { Check, ChevronsUpDown, Building2, Settings, Plus, Search } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Form } from '@/components/ui/form';
import { useWorkspaceStore } from '@/store/workspace';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { workspaceService } from '@/features/workspaces';
import { queryKeys } from '@/lib/query';
import { API } from '@/lib/constants';
import { buildWorkspaceManagementPath, buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { getWorkspaceRedirectPath } from '@/lib/routing/workspace-redirect';
import { credentialService } from '@/services/credential';
import { useCredentialContextStore } from '@/store/credential-context';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { log } from '@/lib/logging';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { useTranslation } from '@/hooks/use-translation';
import type { CreateWorkspaceForm } from '@/lib/types';
import * as z from 'zod';

const createWorkspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters'),
});

export function WorkspaceSwitcher() {
  const [open, setOpen] = React.useState(false);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = React.useState(false);
  const [searchQuery, setSearchQuery] = React.useState('');
  const router = useRouter();
  const queryClient = useQueryClient();
  const { currentWorkspace, setCurrentWorkspace, setWorkspaces } = useWorkspaceStore();
  const { clearSelection } = useCredentialContextStore();
  const { success, error: showError } = useToast();
  const { t } = useTranslation();
  const isChangingWorkspaceRef = React.useRef(false);
  
  // Workspace 목록 조회
  const { data: workspaces = [], isLoading } = useQuery({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    retry: API.REQUEST.MAX_RETRIES,
    retryDelay: API.REQUEST.RETRY_DELAY,
  });
  
  // Workspace 목록 업데이트
  React.useEffect(() => {
    if (!isLoading && workspaces.length > 0) {
      setWorkspaces(workspaces);
    }
  }, [workspaces, isLoading, setWorkspaces]);

  // 검색 필터링된 workspace 목록
  const filteredWorkspaces = React.useMemo(() => {
    if (!searchQuery.trim()) {
      return workspaces;
    }
    const query = searchQuery.toLowerCase();
    return workspaces.filter(workspace => {
      const nameMatch = workspace.name.toLowerCase().includes(query);
      const descMatch = workspace.description?.toLowerCase().includes(query);
      return nameMatch || descMatch;
    });
  }, [workspaces, searchQuery]);
  
  // Workspace 생성 폼
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
      const successMessage = typeof t === 'function' 
        ? t('messages.created', { resource: t('workspace.title') })
        : 'Workspace created successfully';
      success(successMessage);
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'createWorkspace', source: 'workspace-switcher' });
      const errorMessage = typeof t === 'function'
        ? t('messages.operationFailed')
        : 'Failed to create workspace';
      showError(errorMessage);
    },
    resetOnSuccess: true,
  });
  
  // Workspace 전환 핸들러
  const handleWorkspaceChange = React.useCallback(async (workspaceId: string) => {
    // 중복 호출 방지
    if (isChangingWorkspaceRef.current) return;
    isChangingWorkspaceRef.current = true;
    
    const workspace = workspaces.find(w => w.id === workspaceId);
    if (!workspace) {
      isChangingWorkspaceRef.current = false;
      return;
    }
    
    const previousWorkspaceId = currentWorkspace?.id;
    const isWorkspaceChanged = previousWorkspaceId !== workspace.id;
    
    // Workspace 변경
    setCurrentWorkspace(workspace);
    setOpen(false);
    
    if (isWorkspaceChanged) {
      // Credential 선택 초기화
      clearSelection();
      
      // 쿼리 캐시 무효화
      queryClient.invalidateQueries({
        predicate: (query) => {
          const key = query.queryKey;
          if (!Array.isArray(key)) return false;
          return previousWorkspaceId ? key.includes(previousWorkspaceId) : false;
        },
      });
      
      // 스마트 리다이렉트: 현재 경로 분석하여 적절한 경로로 이동
      try {
        const currentPath = window.location.pathname + window.location.search;
        
        // Resource List 페이지인 경우 첫 번째 credential 가져오기
        let firstCredential: { id: string; provider: string } | null = null;
        try {
          const credentials = await credentialService.getCredentials(workspace.id);
          if (credentials.length > 0) {
            firstCredential = {
              id: credentials[0].id,
              provider: credentials[0].provider,
            };
          }
        } catch (error) {
          // Credential 조회 실패 시 무시 (Dashboard로 이동)
          log.error('Failed to fetch credentials for redirect', error, {
            service: 'WorkspaceSwitcher',
            action: 'getCredentials',
            workspaceId: workspace.id,
          });
        }
        
        // 스마트 리다이렉트 경로 생성
        const redirectPath = getWorkspaceRedirectPath(currentPath, workspace.id, firstCredential);
        router.push(redirectPath);
      } catch (error) {
        // 에러 발생 시 Dashboard로 이동 (Fallback)
        log.error('Failed to determine redirect path', error, {
          service: 'WorkspaceSwitcher',
          action: 'getWorkspaceRedirectPath',
          workspaceId: workspace.id,
        });
        router.push(buildWorkspaceManagementPath(workspace.id, 'dashboard'));
      }
    }
    
    // 다음 클릭을 위해 ref 초기화 (약간의 지연)
    setTimeout(() => {
      isChangingWorkspaceRef.current = false;
    }, 100);
  }, [workspaces, currentWorkspace, setCurrentWorkspace, clearSelection, queryClient, router]);
  
  // Workspace 설정 열기
  const handleOpenSettings = React.useCallback(() => {
    if (!currentWorkspace?.id) return;
    setOpen(false);
    router.push(buildWorkspaceResourcePath(currentWorkspace.id, currentWorkspace.id, 'settings'));
  }, [currentWorkspace, router]);
  
  if (isLoading || workspaces.length === 0) {
    return (
      <Button variant="ghost" size="sm" disabled>
        <Building2 className="mr-2 h-4 w-4" />
        Loading...
      </Button>
    );
  }
  
  return (
    <>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-[200px] justify-between"
          >
            <div className="flex items-center">
              <Building2 className="mr-2 h-4 w-4 shrink-0" />
              <span className="truncate">
                {currentWorkspace?.name || 'Select workspace...'}
              </span>
            </div>
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[300px] p-0">
          <div className="flex flex-col">
            {/* 검색 입력 */}
            <div className="flex items-center border-b px-3">
              <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
              <input
                type="text"
                placeholder={t('workspace.searchPlaceholder') || 'Search workspace...'}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="flex h-11 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>

            {/* Workspace 목록 */}
            <div className="max-h-[300px] overflow-y-auto">
              {filteredWorkspaces.length === 0 ? (
                <div className="py-6 text-center text-sm text-muted-foreground">
                  {t('workspace.noWorkspaceFound') || 'No workspace found.'}
                </div>
              ) : (
                <div className="p-1">
                  {filteredWorkspaces.map((workspace) => {
                    const isSelected = currentWorkspace?.id === workspace.id;
                    
                    return (
                      <div
                        key={workspace.id}
                        onClick={() => {
                          handleWorkspaceChange(workspace.id);
                        }}
                        className={cn(
                          "relative flex cursor-pointer select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none",
                          "hover:bg-accent hover:text-accent-foreground",
                          "focus:bg-accent focus:text-accent-foreground",
                          "active:bg-accent active:text-accent-foreground",
                          isSelected && "bg-accent text-accent-foreground",
                          "transition-colors"
                        )}
                        role="button"
                        tabIndex={0}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            handleWorkspaceChange(workspace.id);
                          }
                        }}
                      >
                        <Check
                          className={cn(
                            "mr-2 h-4 w-4 shrink-0",
                            isSelected ? "opacity-100" : "opacity-0"
                          )}
                        />
                        <div className="flex flex-col flex-1 min-w-0">
                          <span className="truncate font-medium text-foreground">
                            {workspace.name}
                          </span>
                          {workspace.description && (
                            <span className="text-xs text-muted-foreground truncate">
                              {workspace.description}
                            </span>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            {/* 구분선 및 액션 버튼 */}
            <div className="border-t">
              {/* 현재 Workspace 설정 버튼 */}
              {currentWorkspace && (
                <div className="p-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleOpenSettings();
                    }}
                    aria-label="Open workspace settings"
                  >
                    <Settings className="mr-2 h-4 w-4" aria-hidden="true" />
                    {t('workspace.settings') || 'Settings'}
                  </Button>
                </div>
              )}
              
              {/* 새 Workspace 생성 */}
              <div className="p-1">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full justify-start"
                  onClick={(e) => {
                    e.stopPropagation();
                    setOpen(false);
                    setIsCreateDialogOpen(true);
                  }}
                  aria-label="Create a new workspace"
                >
                  <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                  {t('workspace.createNew') || 'Create New Workspace'}
                </Button>
              </div>
            </div>
          </div>
        </PopoverContent>
      </Popover>
      
      {/* Workspace 생성 Dialog */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('workspace.createNew') || 'Create New Workspace'}</DialogTitle>
            <DialogDescription>
              {t('workspace.createNewDescription') || 'Create a new workspace to organize your resources and collaborate with your team.'}
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={handleSubmit} className="space-y-4">
              <EnhancedField
                name="name"
                label={t('workspace.name') || 'Name'}
                placeholder={t('workspace.namePlaceholder') || 'Enter workspace name'}
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
              <EnhancedField
                name="description"
                label={t('workspace.description') || 'Description'}
                type="textarea"
                placeholder={t('workspace.descriptionPlaceholder') || 'Enter workspace description'}
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
              {formError && (
                <div className="text-sm text-red-600" role="alert">
                  {formError}
                </div>
              )}
              <div className="flex justify-end space-x-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setIsCreateDialogOpen(false);
                    reset();
                  }}
                >
                  {t('common.cancel') || 'Cancel'}
                </Button>
                <Button type="submit" disabled={isFormLoading}>
                  {isFormLoading
                    ? (t('common.creating') || 'Creating...')
                    : (t('common.create') || 'Create')}
                </Button>
              </div>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </>
  );
}


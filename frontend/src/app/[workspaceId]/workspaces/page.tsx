/**
 * Workspaces Page
 * 워크스페이스 관리 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/workspaces
 */

'use client';

import { useState, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Form } from '@/components/ui/form';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useWorkspaces, useWorkspaceActions } from '@/features/workspaces';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useRouter } from 'next/navigation';
import { Plus, Users, Calendar, Trash2, Home, Key } from 'lucide-react';
import { CreateWorkspaceForm, Workspace } from '@/lib/types';
import { useRequireAuth } from '@/hooks/use-auth';
import { useToast } from '@/hooks/use-toast';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { ErrorHandler } from '@/lib/error-handling';
import { queryKeys } from '@/lib/query';
import { createValidationSchemas } from '@/lib/validation';
import { useTranslation } from '@/hooks/use-translation';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { useResourceContext } from '@/hooks/use-resource-context';
import { buildWorkspaceManagementPath, buildWorkspaceDetailPath, buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { getWorkspaceRedirectPath } from '@/lib/routing/workspace-redirect';
import { credentialService } from '@/services/credential';
import { Layout } from '@/components/layout/layout';
import { toLocaleDateString } from '@/lib/utils/date-format';
import { log } from '@/lib/logging';

export default function WorkspacesPage() {
  const { t, locale } = useTranslation();
  const { workspaceId: pathWorkspaceId } = useResourceContext();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { currentWorkspace, setCurrentWorkspace } = useWorkspaceStore();
  const { clearSelection } = useCredentialContextStore();
  const { success, error: showError } = useToast();
  const { isLoading: authLoading } = useRequireAuth();

  const {
    workspaces,
    isLoading,
  } = useWorkspaces();

  const {
    createWorkspaceMutation,
    deleteWorkspaceMutation,
  } = useWorkspaceActions({
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
    },
  });

  // Delete dialog state must be declared before handlers that use it
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    workspaceId: string | null;
    workspaceName?: string;
  }>({
    open: false,
    workspaceId: null,
    workspaceName: undefined,
  });

  const handleSelectWorkspace = useCallback(async (workspace: Workspace) => {
    const previousWorkspaceId = currentWorkspace?.id;
    const isWorkspaceChanged = previousWorkspaceId !== workspace.id;
    
    setCurrentWorkspace(workspace);
    
    if (isWorkspaceChanged) {
      clearSelection();
      
      queryClient.invalidateQueries({
        predicate: (query) => {
          const key = query.queryKey;
          if (!Array.isArray(key)) return false;
          
          const keyString = key.join('-').toLowerCase();
          
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
            service: 'WorkspacesPage',
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
          service: 'WorkspacesPage',
          action: 'getWorkspaceRedirectPath',
          workspaceId: workspace.id,
        });
        router.push(buildWorkspaceManagementPath(workspace.id, 'dashboard'));
      }
    }
  }, [currentWorkspace?.id, setCurrentWorkspace, router, queryClient, clearSelection]);

  const handleDeleteWorkspace = useCallback((workspaceId: string) => {
    const workspace = workspaces.find(w => w.id === workspaceId);
    setDeleteDialogState({ open: true, workspaceId, workspaceName: workspace?.name });
  }, [workspaces, setDeleteDialogState]);

  const handleConfirmDelete = useCallback(() => {
    if (!deleteDialogState.workspaceId) return;
    
    deleteWorkspaceMutation.mutate(deleteDialogState.workspaceId, {
      onSuccess: () => {
        setDeleteDialogState({ open: false, workspaceId: null, workspaceName: undefined });
        if (currentWorkspace?.id === deleteDialogState.workspaceId) {
          setCurrentWorkspace(null);
          clearSelection();
        }
      },
    });
  }, [deleteDialogState.workspaceId, deleteWorkspaceMutation, setDeleteDialogState, currentWorkspace?.id, setCurrentWorkspace, clearSelection]);

  const {
    form,
    handleSubmit,
    isLoading: isFormLoading,
    error: formError,
    reset,
    getFieldError,
    getFieldValidationState,
  } = useFormWithValidation<CreateWorkspaceForm>({
    schema: createValidationSchemas(t).createWorkspaceSchema,
    defaultValues: {
      name: '',
      description: '',
    },
    onSubmit: async (data) => {
      await createWorkspaceMutation.mutateAsync(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      setIsCreateDialogOpen(false);
      reset();
      // toast는 useWorkspaceActions의 successMessage에서 처리되므로 여기서는 제거
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'createWorkspace', source: 'workspaces-page' });
      showError(t('workspace.createFailed') || 'Failed to create workspace');
    },
    resetOnSuccess: true,
  });

  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  if (authLoading) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{t('workspace.title')}</h1>
            <p className="text-gray-600">{t('workspace.description')}</p>
          </div>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            {t('workspace.create')}
          </Button>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          </div>
        ) : workspaces.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Home className="h-12 w-12 text-gray-400 mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">{t('workspace.noWorkspaces')}</h3>
              <p className="text-sm text-gray-600 mb-4">{t('workspace.createFirst')}</p>
              <Button onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                {t('workspace.create')}
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {workspaces.map((workspace) => (
              <Card
                key={workspace.id}
                className={`cursor-pointer hover:shadow-lg transition-shadow ${
                  currentWorkspace?.id === workspace.id ? 'ring-2 ring-blue-500' : ''
                }`}
                onClick={() => handleSelectWorkspace(workspace)}
              >
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">{workspace.name}</CardTitle>
                    {currentWorkspace?.id === workspace.id && (
                      <Badge variant="default">Current</Badge>
                    )}
                  </div>
                  <CardDescription className="line-clamp-2">
                    {workspace.description}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="grid grid-cols-2 gap-4">
                      <div 
                        className="flex items-center gap-2 cursor-pointer hover:bg-accent rounded-md p-2 -m-2 transition-colors"
                        onClick={(e) => {
                          e.stopPropagation();
                          router.push(`${buildWorkspaceDetailPath(workspace.id, workspace.id)}?tab=credentials`);
                        }}
                      >
                        <Key className="h-4 w-4 text-muted-foreground" />
                        <div>
                          <div className="text-2xl font-bold">{workspace.credential_count ?? 0}</div>
                          <div className="text-xs text-muted-foreground">{t('workspace.credentials') || 'Credentials'}</div>
                        </div>
                      </div>
                      <div 
                        className="flex items-center gap-2 cursor-pointer hover:bg-accent rounded-md p-2 -m-2 transition-colors"
                        onClick={(e) => {
                          e.stopPropagation();
                          router.push(`${buildWorkspaceDetailPath(workspace.id, workspace.id)}?tab=members`);
                        }}
                      >
                        <Users className="h-4 w-4 text-muted-foreground" />
                        <div>
                          <div className="text-2xl font-bold">{workspace.member_count ?? 0}</div>
                          <div className="text-xs text-muted-foreground">{t('workspace.members') || 'Members'}</div>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center text-sm text-gray-600">
                      <Calendar className="mr-1 h-4 w-4" />
                      <span>{toLocaleDateString(workspace.created_at, locale as 'ko' | 'en')}</span>
                    </div>
                  </div>
                  <div className="mt-4 flex justify-end gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        router.push(buildWorkspaceDetailPath(workspace.id, workspace.id));
                      }}
                    >
                      {t('common.view') || 'View'}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        router.push(buildWorkspaceResourcePath(workspace.id, workspace.id, 'settings'));
                      }}
                    >
                      {t('common.settings') || 'Settings'}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteWorkspace(workspace.id);
                      }}
                    >
                      <Trash2 className="h-4 w-4 text-red-600" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Create Workspace Dialog */}
        {isCreateDialogOpen && (
          <Card className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <CardContent className="bg-white p-6 rounded-lg max-w-md w-full m-4">
              <CardHeader>
                <CardTitle>{t('workspace.create')}</CardTitle>
                <CardDescription>{t('workspace.createDescription')}</CardDescription>
              </CardHeader>
              <Form {...form}>
                <form onSubmit={handleSubmit} className="space-y-4">
                  <EnhancedField
                    control={form.control}
                    name="name"
                    label={t('workspace.name')}
                    placeholder={t('workspace.namePlaceholder')}
                    error={getFieldError('name')}
                    validationState={getFieldValidationState('name')}
                  />

                  <EnhancedField
                    control={form.control}
                    name="description"
                    label={t('workspace.description')}
                    placeholder={t('workspace.descriptionPlaceholder')}
                    error={getFieldError('description')}
                    validationState={getFieldValidationState('description')}
                    multiline
                    rows={4}
                  />

                  {formError && (
                    <div className="text-sm text-red-600">{formError}</div>
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
                      {t('common.cancel')}
                    </Button>
                    <Button type="submit" disabled={isFormLoading || createWorkspaceMutation.isPending}>
                      {isFormLoading || createWorkspaceMutation.isPending
                        ? t('common.creating')
                        : t('common.create')}
                    </Button>
                  </div>
                </form>
              </Form>
            </CardContent>
          </Card>
        )}

        <DeleteConfirmationDialog
          open={deleteDialogState.open}
          onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
          onConfirm={handleConfirmDelete}
          title={t('workspace.deleteWorkspace')}
          description={deleteDialogState.workspaceName 
            ? t('workspace.confirmDelete', { workspaceName: deleteDialogState.workspaceName })
            : t('workspace.confirmDeleteGeneric')}
          isLoading={deleteWorkspaceMutation.isPending}
          resourceName={deleteDialogState.workspaceName}
          resourceNameLabel={t('workspace.name')}
        />
      </div>
    </Layout>
  );
}


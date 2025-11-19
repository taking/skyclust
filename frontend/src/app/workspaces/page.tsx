/**
 * Workspaces Page (Legacy Route)
 * /workspaces -> /{workspaceId}/workspaces로 리다이렉트하거나
 * workspaces가 없을 때는 빈 상태와 생성 폼을 보여줌
 */

'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { workspaceService, useWorkspaceActions } from '@/features/workspaces';
import { queryKeys } from '@/lib/query';
import { API } from '@/lib/constants';
import { buildManagementPath } from '@/lib/routing/helpers';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Home, Plus } from 'lucide-react';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { CreateWorkspaceForm } from '@/lib/types';
import { createValidationSchemas } from '@/lib/validation';
import { useTranslation } from '@/hooks/use-translation';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { useRequireAuth } from '@/hooks/use-auth';

export default function WorkspacesRedirectPage() {
  const router = useRouter();
  const pathname = usePathname();
  const { currentWorkspace, setCurrentWorkspace, setWorkspaces } = useWorkspaceStore();
  const hasRedirectedRef = useRef(false);
  const { t } = useTranslation();
  const { success, error: showError } = useToast();
  const queryClient = useQueryClient();
  const { isLoading: authLoading } = useRequireAuth();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  // Fetch workspaces
  const { data: fetchedWorkspaces = [], isLoading } = useQuery({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    retry: API.REQUEST.MAX_RETRIES,
    retryDelay: API.REQUEST.RETRY_DELAY,
  });

  const {
    createWorkspaceMutation,
  } = useWorkspaceActions({
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
    },
  });

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
      const createdWorkspace = await createWorkspaceMutation.mutateAsync(data);
      
      // 새로 생성된 workspace로 리다이렉트
      if (createdWorkspace?.id) {
        setCurrentWorkspace(createdWorkspace);
        const newPath = buildManagementPath(createdWorkspace.id, 'workspaces');
        router.replace(newPath);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      setIsCreateDialogOpen(false);
      reset();
      // toast는 useWorkspaceActions의 successMessage에서 처리되므로 여기서는 제거
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'createWorkspace', source: 'workspaces-redirect-page' });
      showError(t('workspace.createFailed') || 'Failed to create workspace');
    },
    resetOnSuccess: true,
  });

  useEffect(() => {
    // 이미 리다이렉트했거나 로딩 중이면 스킵
    if (hasRedirectedRef.current || isLoading) return;

    // 이미 올바른 경로에 있으면 스킵 (예: /w/{workspaceId}/workspaces)
    if (pathname?.startsWith('/w/') && pathname.includes('/workspaces')) {
      hasRedirectedRef.current = true;
      return;
    }

    if (fetchedWorkspaces.length > 0) {
      setWorkspaces(fetchedWorkspaces);
      
      // 현재 선택된 workspace가 있으면 해당 workspace의 workspaces 페이지로
      if (currentWorkspace?.id) {
        const newPath = buildManagementPath(currentWorkspace.id, 'workspaces');
        hasRedirectedRef.current = true;
        router.replace(newPath);
      } else {
        // 첫 번째 workspace 선택
        const firstWorkspace = fetchedWorkspaces[0];
        setCurrentWorkspace(firstWorkspace);
        const newPath = buildManagementPath(firstWorkspace.id, 'workspaces');
        hasRedirectedRef.current = true;
        router.replace(newPath);
      }
    } else {
      // Workspace가 없으면 빈 상태 표시 (리다이렉트하지 않음)
      setWorkspaces([]);
      hasRedirectedRef.current = true;
    }
  }, [isLoading, fetchedWorkspaces, currentWorkspace?.id, setCurrentWorkspace, setWorkspaces, router, pathname]);

  if (authLoading) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </div>
      </Layout>
    );
  }

  // 로딩 중이거나 workspaces가 있는 경우 리다이렉트 중
  if (isLoading || (fetchedWorkspaces.length > 0 && !hasRedirectedRef.current)) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-screen">
          <div className="text-center">
            <Spinner size="lg" label="Redirecting..." />
          </div>
        </div>
      </Layout>
    );
  }

  // Workspaces가 없을 때 빈 상태와 생성 폼 표시
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
      </div>
    </Layout>
  );
}

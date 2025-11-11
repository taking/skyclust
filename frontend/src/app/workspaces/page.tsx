/**
 * Workspaces Page
 * 워크스페이스 관리 페이지
 * 
 * use-form-with-validation 훅을 사용한 리팩토링 버전
 */

'use client';

import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Form } from '@/components/ui/form';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { useWorkspaces, useWorkspaceActions } from '@/features/workspaces';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useRouter } from 'next/navigation';
import { Plus, Users, Calendar, Trash2, Home } from 'lucide-react';
import { CreateWorkspaceForm, Workspace } from '@/lib/types';
import { useRequireAuth } from '@/hooks/use-auth';
import { useToast } from '@/hooks/use-toast';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { ErrorHandler } from '@/lib/error-handling';
import { queryKeys } from '@/lib/query';
import { createValidationSchemas } from '@/lib/validation';
import { useTranslation } from '@/hooks/use-translation';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';

export default function WorkspacesPage() {
  const { t } = useTranslation();
  const { createWorkspaceSchema } = createValidationSchemas(t);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    workspaceId: string | null;
    workspaceName?: string;
  }>({
    open: false,
    workspaceId: null,
    workspaceName: undefined,
  });
  const { currentWorkspace, setCurrentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { isLoading: authLoading } = useRequireAuth();
  const { success } = useToast();

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
      await createWorkspaceMutation.mutateAsync(data);
    },
    onSuccess: () => {
      setIsCreateDialogOpen(false);
      reset();
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'createWorkspace' });
    },
    resetOnSuccess: true,
  });

  // Fetch workspaces
  const { workspaces, isLoading, error } = useWorkspaces();

  // Workspace actions
  const {
    createWorkspaceMutation,
    deleteWorkspaceMutation,
  } = useWorkspaceActions({
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
    },
  });

  const handleSelectWorkspace = (workspace: Workspace) => {
    const previousWorkspaceId = currentWorkspace?.id;
    const isWorkspaceChanged = previousWorkspaceId !== workspace.id;
    
    setCurrentWorkspace(workspace);
    
    if (isWorkspaceChanged) {
      const { clearSelection } = useCredentialContextStore.getState();
      
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
    }
    
    router.push('/dashboard');
  };

  const handleDeleteWorkspace = (workspaceId: string) => {
    const workspace = workspaces.find(w => w.id === workspaceId);
    setDeleteDialogState({ open: true, workspaceId, workspaceName: workspace?.name });
  };

  const handleConfirmDelete = () => {
    if (deleteDialogState.workspaceId) {
      const deletedWorkspaceId = deleteDialogState.workspaceId;
      deleteWorkspaceMutation.mutate(deletedWorkspaceId, {
        onSuccess: () => {
          // Store에서 워크스페이스 제거 (자동으로 다른 워크스페이스로 전환됨)
          const { removeWorkspace } = useWorkspaceStore.getState();
          removeWorkspace(deletedWorkspaceId);
          
          // 쿼리 무효화로 최신 워크스페이스 목록 가져오기
          queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
          
          setDeleteDialogState({ open: false, workspaceId: null, workspaceName: undefined });
        },
      });
    }
  };

  if (authLoading || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">
            {authLoading ? t('workspace.checkingAuthentication') : t('workspace.loadingWorkspaces')}
          </p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="mx-auto h-12 w-12 text-red-400">
            <Users className="h-12 w-12" />
          </div>
          <h3 className="mt-2 text-sm font-medium text-gray-900">{t('workspace.errorLoadingWorkspaces')}</h3>
          <p className="mt-1 text-sm text-gray-500">
            {error instanceof Error ? error.message : t('errors.generic')}
          </p>
          <div className="mt-6">
            <Button onClick={() => window.location.reload()}>
              {t('workspace.tryAgain')}
            </Button>
        </div>
      </div>

      {/* Delete Workspace Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={handleConfirmDelete}
        title={t('workspace.deleteWorkspace')}
        description={t('workspace.confirmDeleteWorkspaceMessage')}
        isLoading={deleteWorkspaceMutation.isPending}
        resourceName={deleteDialogState.workspaceName}
        resourceNameLabel={t('workspace.workspaceNameLabelForDelete')}
      />
    </div>
  );
}

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{t('workspace.workspacesPageTitle')}</h1>
            <p className="text-gray-600">{t('workspace.workspacesPageDescription')}</p>
          </div>
          <div className="flex items-center space-x-2">
            <Button variant="outline" onClick={() => router.push('/dashboard')}>
              <Home className="mr-2 h-4 w-4" />
              {t('workspace.home')}
            </Button>
            <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  {t('workspace.createWorkspace')}
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>{t('workspace.createNewWorkspace')}</DialogTitle>
                  <DialogDescription>
                    {t('workspace.createNewWorkspaceDescription')}
                  </DialogDescription>
                </DialogHeader>
                <Form {...form}>
                  <form onSubmit={handleSubmit} className="space-y-4">
                    <EnhancedField
                      name="name"
                      label={t('workspace.workspaceNameLabel')}
                      type="text"
                      placeholder={t('workspace.workspaceNamePlaceholder')}
                      required
                      getFieldError={getFieldError}
                      getFieldValidationState={getFieldValidationState}
                    />
                    <EnhancedField
                      name="description"
                      label={t('workspace.descriptionLabel')}
                      type="textarea"
                      placeholder={t('workspace.descriptionPlaceholder')}
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
                        {t('workspace.cancel')}
                      </Button>
                      <Button type="submit" disabled={isFormLoading}>
                        {isFormLoading ? t('workspace.creating') : t('workspace.createWorkspace')}
                      </Button>
                    </div>
                  </form>
                </Form>
              </DialogContent>
            </Dialog>
          </div>
        </div>

        {/* Workspaces Grid */}
        {workspaces.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <div className="text-center py-12">
                <Users className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">{t('workspace.noWorkspaces')}</h3>
                <p className="mt-1 text-sm text-gray-500">
                  {t('workspace.noWorkspacesDescription')}
                </p>
              </div>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {workspaces.map((workspace) => (
              <Card
                key={workspace.id}
                className="cursor-pointer hover:shadow-lg transition-shadow"
                onClick={() => handleSelectWorkspace(workspace)}
              >
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">{workspace.name}</CardTitle>
                    <Badge variant={workspace.is_active ? 'default' : 'secondary'}>
                      {workspace.is_active ? t('workspace.active') : t('workspace.inactive')}
                    </Badge>
                  </div>
                  <CardDescription className="mt-2">
                    {workspace.description || t('workspace.noDescription')}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center justify-between text-sm text-gray-500">
                    <div className="flex items-center gap-1">
                      <Calendar className="h-4 w-4" />
                      <span>
                        {new Date(workspace.created_at).toLocaleDateString()}
                      </span>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteWorkspace(workspace.id);
                      }}
                      disabled={deleteWorkspaceMutation.isPending}
                    >
                      <Trash2 className="h-4 w-4 text-red-500" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Delete Workspace Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={handleConfirmDelete}
        title={t('workspace.deleteWorkspace')}
        description="이 워크스페이스를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        isLoading={deleteWorkspaceMutation.isPending}
        resourceName={deleteDialogState.workspaceName}
        resourceNameLabel="워크스페이스 이름"
      />
    </div>
  );
}


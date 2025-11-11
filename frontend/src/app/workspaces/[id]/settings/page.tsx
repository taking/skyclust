/**
 * Workspace Settings Page
 * 워크스페이스 설정 페이지
 */

'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { workspaceService } from '@/features/workspaces';
import { useWorkspaceStore } from '@/store/workspace';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { UpdateWorkspaceForm } from '@/lib/types';
import { Layout } from '@/components/layout/layout';
import { ArrowLeft, Settings, Users, Trash2, AlertTriangle } from 'lucide-react';
import * as React from 'react';
// import * as z from 'zod'; // Not used directly
import { Form } from '@/components/ui/form';
import { queryKeys } from '@/lib/query';
import { createValidationSchemas } from '@/lib/validation';
import { useTranslation } from '@/hooks/use-translation';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';

export default function WorkspaceSettingsPage() {
  const { t } = useTranslation();
  const { updateWorkspaceSchema } = createValidationSchemas(t);
  const params = useParams();
  const router = useRouter();
  const workspaceId = params.id as string;
  const { currentWorkspace, setCurrentWorkspace } = useWorkspaceStore();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
  }>({
    open: false,
  });

  // Fetch workspace details
  const { data: workspace, isLoading } = useQuery({
    queryKey: queryKeys.workspaces.detail(workspaceId),
    queryFn: () => workspaceService.getWorkspace(workspaceId),
    enabled: !!workspaceId,
  });

  // Sync currentWorkspace with URL parameter on mount and when workspace loads
  React.useEffect(() => {
    if (workspace && currentWorkspace?.id !== workspace.id) {
      setCurrentWorkspace(workspace);
    }
  }, [workspace, currentWorkspace?.id, setCurrentWorkspace]);

  // Update workspace form
  const {
    form,
    handleSubmit,
    isLoading: isFormLoading,
    error: formError,
    getFieldError,
    getFieldValidationState,
  } = useFormWithValidation<UpdateWorkspaceForm>({
    schema: updateWorkspaceSchema,
    defaultValues: {
      name: workspace?.name || '',
      description: workspace?.description || '',
    },
    onSubmit: async (data) => {
      await workspaceService.updateWorkspace(workspaceId, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.detail(workspaceId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      
      // Update current workspace if it's the same
      if (currentWorkspace?.id === workspaceId) {
        const updatedWorkspace = { ...currentWorkspace };
        const name = form.getValues('name');
        const description = form.getValues('description');
        if (name) updatedWorkspace.name = name;
        if (description) updatedWorkspace.description = description;
        setCurrentWorkspace(updatedWorkspace);
      }
      
      success(t('workspace.workspaceUpdatedSuccessfully'));
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'updateWorkspace' });
      showError(t('workspace.failedToUpdateWorkspace'));
    },
  });

  // Update form when workspace data loads
  React.useEffect(() => {
    if (workspace) {
      form.reset({
        name: workspace.name,
        description: workspace.description,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [workspace]);

  // Delete workspace mutation
  const deleteWorkspaceMutation = useMutation({
    mutationFn: (id: string) => workspaceService.deleteWorkspace(id),
    onSuccess: () => {
      // Store에서 워크스페이스 제거 (자동으로 다른 워크스페이스로 전환됨)
      const { removeWorkspace } = useWorkspaceStore.getState();
      removeWorkspace(workspaceId);
      
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      success(t('workspace.workspaceDeletedSuccessfully'));
      router.push('/workspaces');
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'deleteWorkspace' });
      showError(t('workspace.failedToDeleteWorkspace'));
    },
  });

  const handleDeleteWorkspace = () => {
    setDeleteDialogState({ open: true });
  };

  const handleConfirmDelete = () => {
    deleteWorkspaceMutation.mutate(workspaceId);
    setDeleteDialogState({ open: false });
  };

  if (isLoading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">{t('workspace.loadingWorkspaceSettings')}</p>
          </div>
        </div>
      </Layout>
    );
  }

  if (!workspace) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">{t('workspace.workspaceNotFound')}</h3>
            <p className="mt-1 text-sm text-gray-500">{t('workspace.workspaceNotFoundDescription')}</p>
            <Button onClick={() => router.push('/dashboard')} className="mt-4">
              {t('workspace.goToDashboard')}
            </Button>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <Button
              variant="ghost"
              onClick={() => router.back()}
              className="mb-4"
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              {t('workspace.back')}
            </Button>
            <div className="flex items-center space-x-2 mb-2">
              <Settings className="h-6 w-6 text-gray-600" />
              <h1 className="text-3xl font-bold text-gray-900">{t('workspace.workspaceSettings')}</h1>
            </div>
            <p className="text-gray-600">
              {t('workspace.manageWorkspaceSettings')}
            </p>
          </div>

          {/* Workspace Information */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle>{t('workspace.workspaceInformation')}</CardTitle>
              <CardDescription>
                {t('workspace.updateWorkspaceNameDescription')}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={handleSubmit} className="space-y-4">
                  <EnhancedField
                    name="name"
                    label={t('workspace.workspaceName')}
                    type="text"
                    placeholder={t('workspace.enterWorkspaceName')}
                    required
                    getFieldError={getFieldError}
                    getFieldValidationState={getFieldValidationState}
                  />
                  <EnhancedField
                    name="description"
                    label={t('workspace.description')}
                    type="textarea"
                    placeholder={t('workspace.enterWorkspaceDescription')}
                    required
                    getFieldError={getFieldError}
                    getFieldValidationState={getFieldValidationState}
                  />
                  
                  {formError && (
                    <Alert variant="destructive">
                      <AlertTriangle className="h-4 w-4" />
                      <AlertDescription>{formError}</AlertDescription>
                    </Alert>
                  )}

                  <div className="flex justify-end">
                    <Button type="submit" disabled={isFormLoading}>
                      {isFormLoading ? t('workspace.saving') : t('workspace.saveChanges')}
                    </Button>
                  </div>
                </form>
              </Form>
            </CardContent>
          </Card>

          {/* Members Management */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle>{t('workspace.members')}</CardTitle>
              <CardDescription>
                {t('workspace.manageWorkspaceMembersPermissions')}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Users className="h-5 w-5 text-gray-500" />
                  <span className="text-sm text-gray-600">
                    {t('workspace.manageWorkspaceMembers')}
                  </span>
                </div>
                <Button
                  variant="outline"
                  onClick={() => router.push(`/workspaces/${workspaceId}/members`)}
                >
                  <Users className="mr-2 h-4 w-4" />
                  {t('workspace.manageMembers')}
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Danger Zone */}
          <Card className="border-red-200">
            <CardHeader>
              <CardTitle className="text-red-600">{t('workspace.dangerZone')}</CardTitle>
              <CardDescription>
                {t('workspace.irreversibleDestructiveActions')}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <h3 className="text-sm font-medium text-gray-900">{t('workspace.deleteWorkspace')}</h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {t('workspace.deleteWorkspaceDescription')}
                  </p>
                </div>
                <Button 
                  variant="destructive"
                  onClick={handleDeleteWorkspace}
                  disabled={deleteWorkspaceMutation.isPending}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  {t('workspace.deleteWorkspace')}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Delete Workspace Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ open })}
        onConfirm={handleConfirmDelete}
        title={t('workspace.deleteWorkspace')}
        description="이 워크스페이스를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        isLoading={deleteWorkspaceMutation.isPending}
        resourceName={workspace?.name}
        resourceNameLabel="워크스페이스 이름"
      />
    </Layout>
  );
}


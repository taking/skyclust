/**
 * Workspace Settings Tab
 * Workspace 상세 페이지의 Settings 탭
 * 
 * 기존 WorkspaceSettingsPage의 내용을 재사용
 */

'use client';

import * as React from 'react';
import { Suspense } from 'react';
import { useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { workspaceService } from '../services/workspace';
import { useWorkspaceStore } from '@/store/workspace';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { UpdateWorkspaceForm } from '@/lib/types';
import { Settings, Trash2, AlertTriangle } from 'lucide-react';
import { Form } from '@/components/ui/form';
import { queryKeys } from '@/lib/query';
import { createValidationSchemas } from '@/lib/validation';
import { useTranslation } from '@/hooks/use-translation';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { buildWorkspaceManagementPath } from '@/lib/routing/helpers';
import { useState } from 'react';
import { Spinner } from '@/components/ui/spinner';

interface WorkspaceSettingsTabProps {
  workspaceId: string;
}

export function WorkspaceSettingsTab({ workspaceId }: WorkspaceSettingsTabProps) {
  const { t } = useTranslation();
  const { updateWorkspaceSchema } = createValidationSchemas(t);
  const router = useRouter();
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

  // Update workspace mutation
  const updateWorkspaceMutation = useMutation({
    mutationFn: (data: UpdateWorkspaceForm) => workspaceService.updateWorkspace(workspaceId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.detail(workspaceId) });
      success(t('workspace.updated') || 'Workspace updated successfully');
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'updateWorkspace', source: 'workspace-settings-tab' });
      showError(t('workspace.updateFailed') || 'Failed to update workspace');
    },
  });

  // Delete workspace mutation
  const deleteWorkspaceMutation = useMutation({
    mutationFn: () => workspaceService.deleteWorkspace(workspaceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      if (currentWorkspace?.id === workspaceId) {
        setCurrentWorkspace(null);
      }
      if (currentWorkspace?.id) {
        router.push(buildWorkspaceManagementPath(currentWorkspace.id, 'workspaces'));
      } else {
        router.push('/workspaces');
      }
      success(t('workspace.deleted') || 'Workspace deleted successfully');
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'deleteWorkspace', source: 'workspace-settings-tab' });
      showError(t('workspace.deleteFailed') || 'Failed to delete workspace');
    },
  });

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
      updateWorkspaceMutation.mutate({
        name: form.getValues('name'),
        description: form.getValues('description'),
      });
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'updateWorkspace', source: 'workspace-settings-tab' });
      showError(t('workspace.updateFailed') || 'Failed to update workspace');
    },
    resetOnSuccess: false,
  });

  // Update form when workspace data changes
  React.useEffect(() => {
    if (workspace) {
      form.reset({
        name: workspace.name,
        description: workspace.description,
      });
    }
  }, [workspace, form]);

  const handleDelete = () => {
    setDeleteDialogState({ open: true });
  };

  const handleConfirmDelete = () => {
    deleteWorkspaceMutation.mutate();
    setDeleteDialogState({ open: false });
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" label="Loading settings..." />
      </div>
    );
  }

  if (!workspace) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <p className="text-muted-foreground">{t('workspace.workspaceNotFound') || 'Workspace not found'}</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Settings className="mr-2 h-5 w-5" />
            {t('workspace.title') || 'Workspace'}
          </CardTitle>
          <CardDescription>
            {t('workspace.settingsDescription') || 'Update workspace name and description'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={handleSubmit} className="space-y-4">
              <EnhancedField
                name="name"
                label={t('workspace.name') || 'Name'}
                placeholder={t('workspace.namePlaceholder') || 'Enter workspace name'}
                description={t('workspace.nameDescription') || 'Workspace name'}
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />

              <EnhancedField
                name="description"
                label={t('workspace.description') || 'Description'}
                placeholder={t('workspace.descriptionPlaceholder') || 'Enter workspace description'}
                description={t('workspace.descriptionDescription') || 'Workspace description'}
                type="textarea"
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />

              {formError && (
                <Alert variant="destructive">
                  <AlertTriangle className="h-4 w-4" />
                  <AlertDescription>{formError}</AlertDescription>
                </Alert>
              )}

              <div className="flex justify-end space-x-2">
                <Button type="submit" disabled={isFormLoading || updateWorkspaceMutation.isPending}>
                  {isFormLoading || updateWorkspaceMutation.isPending
                    ? t('common.saving') || 'Saving...'
                    : t('common.save') || 'Save'}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center text-red-600">
            <Trash2 className="mr-2 h-5 w-5" />
            {t('workspace.dangerZone') || 'Danger Zone'}
          </CardTitle>
          <CardDescription>
            {t('workspace.dangerZoneDescription') || 'Irreversible and destructive actions'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-medium">{t('workspace.deleteWorkspace') || 'Delete Workspace'}</h3>
              <p className="text-sm text-gray-600">{t('workspace.deleteDescription') || 'Deleting a workspace cannot be undone'}</p>
            </div>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleteWorkspaceMutation.isPending}
            >
              {t('workspace.delete') || 'Delete'}
            </Button>
          </div>
        </CardContent>
      </Card>

      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ open })}
        onConfirm={handleConfirmDelete}
        title={t('workspace.deleteWorkspace') || 'Delete Workspace'}
        description={t('workspace.confirmDelete', { workspaceName: workspace.name }) || `Are you sure you want to delete "${workspace.name}"?`}
        isLoading={deleteWorkspaceMutation.isPending}
        resourceName={workspace.name}
        resourceNameLabel={t('workspace.name') || 'Workspace Name'}
      />
    </div>
  );
}


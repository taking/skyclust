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
import { workspaceService, useWorkspaceActions } from '@/features/workspaces';
import { useWorkspaceStore } from '@/store/workspace';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handler';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { CreateWorkspaceForm } from '@/lib/types';
import { Layout } from '@/components/layout/layout';
import { ArrowLeft, Settings, Users, Trash2, AlertTriangle } from 'lucide-react';
import * as React from 'react';
import * as z from 'zod';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Form } from '@/components/ui/form';
import { queryKeys } from '@/lib/query-keys';

const updateWorkspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters'),
});

export default function WorkspaceSettingsPage() {
  const params = useParams();
  const router = useRouter();
  const workspaceId = params.id as string;
  const { currentWorkspace, setCurrentWorkspace } = useWorkspaceStore();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

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
    reset,
    getFieldError,
    getFieldValidationState,
  } = useFormWithValidation<CreateWorkspaceForm>({
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
        if (form.getValues('name')) updatedWorkspace.name = form.getValues('name');
        if (form.getValues('description')) updatedWorkspace.description = form.getValues('description');
        setCurrentWorkspace(updatedWorkspace);
      }
      
      success('Workspace updated successfully');
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'updateWorkspace' });
      showError('Failed to update workspace');
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
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      success('Workspace deleted successfully');
      router.push('/workspaces');
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'deleteWorkspace' });
      showError('Failed to delete workspace');
    },
  });

  const handleDeleteWorkspace = () => {
    if (confirm('Are you sure you want to delete this workspace? This action cannot be undone.')) {
      deleteWorkspaceMutation.mutate(workspaceId);
    }
  };

  if (isLoading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">Loading workspace settings...</p>
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
            <h3 className="text-lg font-medium text-gray-900">Workspace not found</h3>
            <p className="mt-1 text-sm text-gray-500">The workspace you&apos;re looking for doesn&apos;t exist.</p>
            <Button onClick={() => router.push('/dashboard')} className="mt-4">
              Go to Dashboard
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
              Back
            </Button>
            <div className="flex items-center space-x-2 mb-2">
              <Settings className="h-6 w-6 text-gray-600" />
              <h1 className="text-3xl font-bold text-gray-900">Workspace Settings</h1>
            </div>
            <p className="text-gray-600">
              Manage your workspace settings and preferences
            </p>
          </div>

          {/* Workspace Information */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle>Workspace Information</CardTitle>
              <CardDescription>
                Update your workspace name and description
              </CardDescription>
            </CardHeader>
            <CardContent>
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
                    <Alert variant="destructive">
                      <AlertTriangle className="h-4 w-4" />
                      <AlertDescription>{formError}</AlertDescription>
                    </Alert>
                  )}

                  <div className="flex justify-end">
                    <Button type="submit" disabled={isFormLoading}>
                      {isFormLoading ? 'Saving...' : 'Save Changes'}
                    </Button>
                  </div>
                </form>
              </Form>
            </CardContent>
          </Card>

          {/* Members Management */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle>Members</CardTitle>
              <CardDescription>
                Manage workspace members and their permissions
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Users className="h-5 w-5 text-gray-500" />
                  <span className="text-sm text-gray-600">
                    Manage workspace members
                  </span>
                </div>
                <Button
                  variant="outline"
                  onClick={() => router.push(`/workspaces/${workspaceId}/members`)}
                >
                  <Users className="mr-2 h-4 w-4" />
                  Manage Members
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Danger Zone */}
          <Card className="border-red-200">
            <CardHeader>
              <CardTitle className="text-red-600">Danger Zone</CardTitle>
              <CardDescription>
                Irreversible and destructive actions
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <h3 className="text-sm font-medium text-gray-900">Delete Workspace</h3>
                  <p className="text-sm text-gray-500 mt-1">
                    Once you delete a workspace, there is no going back. Please be certain.
                  </p>
                </div>
                <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
                  <DialogTrigger asChild>
                    <Button variant="destructive">
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete Workspace
                    </Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Delete Workspace</DialogTitle>
                      <DialogDescription>
                        Are you sure you want to delete &quot;{workspace.name}&quot;? This action cannot be undone.
                        All resources in this workspace will be permanently deleted.
                      </DialogDescription>
                    </DialogHeader>
                    <div className="flex justify-end space-x-2 mt-4">
                      <Button
                        variant="outline"
                        onClick={() => setIsDeleteDialogOpen(false)}
                      >
                        Cancel
                      </Button>
                      <Button
                        variant="destructive"
                        onClick={() => {
                          handleDeleteWorkspace();
                          setIsDeleteDialogOpen(false);
                        }}
                        disabled={deleteWorkspaceMutation.isPending}
                      >
                        {deleteWorkspaceMutation.isPending ? 'Deleting...' : 'Delete Workspace'}
                      </Button>
                    </div>
                  </DialogContent>
                </Dialog>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </Layout>
  );
}


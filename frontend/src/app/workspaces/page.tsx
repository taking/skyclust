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
import { workspaceService, useWorkspaces, useWorkspaceActions } from '@/features/workspaces';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useRouter } from 'next/navigation';
import { Plus, Users, Calendar, Trash2, Home } from 'lucide-react';
import { CreateWorkspaceForm, Workspace } from '@/lib/types';
import { useRequireAuth } from '@/hooks/use-auth';
import { useToast } from '@/hooks/use-toast';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { ErrorHandler } from '@/lib/error-handler';
import * as z from 'zod';
import { queryKeys } from '@/lib/query-keys';

const createWorkspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters'),
});

export default function WorkspacesPage() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const { currentWorkspace, setCurrentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { isLoading: authLoading } = useRequireAuth();
  const { success, error: showError } = useToast();

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
    if (confirm('Are you sure you want to delete this workspace?')) {
      deleteWorkspaceMutation.mutate(workspaceId);
    }
  };

  if (authLoading || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">
            {authLoading ? 'Checking authentication...' : 'Loading workspaces...'}
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
          <h3 className="mt-2 text-sm font-medium text-gray-900">Error loading workspaces</h3>
          <p className="mt-1 text-sm text-gray-500">
            {error instanceof Error ? error.message : 'Something went wrong'}
          </p>
          <div className="mt-6">
            <Button onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Workspaces</h1>
            <p className="text-gray-600">Manage your workspaces and collaborate with your team</p>
          </div>
          <div className="flex items-center space-x-2">
            <Button variant="outline" onClick={() => router.push('/dashboard')}>
              <Home className="mr-2 h-4 w-4" />
              Home
            </Button>
            <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
              <DialogTrigger asChild>
                <Button>
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
        </div>

        {/* Workspaces Grid */}
        {workspaces.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <div className="text-center py-12">
                <Users className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">No workspaces</h3>
                <p className="mt-1 text-sm text-gray-500">
                  Get started by creating a new workspace.
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
                      {workspace.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </div>
                  <CardDescription className="mt-2">
                    {workspace.description || 'No description'}
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
    </div>
  );
}


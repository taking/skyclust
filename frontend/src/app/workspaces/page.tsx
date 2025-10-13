'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { workspaceService } from '@/services/workspace';
import { useWorkspaceStore } from '@/store/workspace';
import { useRouter } from 'next/navigation';
import { Plus, Users, Calendar, Trash2 } from 'lucide-react';
import { CreateWorkspaceForm, Workspace } from '@/lib/types';
import { useRequireAuth } from '@/hooks/useAuth';
import { useToast } from '@/hooks/useToast';

const createWorkspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters'),
});

export default function WorkspacesPage() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const { setCurrentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { isLoading: authLoading } = useRequireAuth();
  const { success, error: showError } = useToast();

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<CreateWorkspaceForm>({
    resolver: zodResolver(createWorkspaceSchema),
  });

  // Fetch workspaces
  const { data: workspaces = [], isLoading, error } = useQuery({
    queryKey: ['workspaces'],
    queryFn: workspaceService.getWorkspaces,
    retry: 3,
    retryDelay: 1000,
  });

  // Create workspace mutation
  const createWorkspaceMutation = useMutation({
    mutationFn: workspaceService.createWorkspace,
    onSuccess: (newWorkspace) => {
      queryClient.invalidateQueries({ queryKey: ['workspaces'] });
      setIsCreateDialogOpen(false);
      reset();
      success('Workspace created successfully');
    },
    onError: (error) => {
      console.error('Failed to create workspace:', error);
      showError('Failed to create workspace');
    },
  });

  // Delete workspace mutation
  const deleteWorkspaceMutation = useMutation({
    mutationFn: workspaceService.deleteWorkspace,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workspaces'] });
      success('Workspace deleted successfully');
    },
    onError: (error) => {
      console.error('Failed to delete workspace:', error);
      showError('Failed to delete workspace');
    },
  });

  const handleCreateWorkspace = (data: CreateWorkspaceForm) => {
    createWorkspaceMutation.mutate(data);
  };

  const handleSelectWorkspace = (workspace: Workspace) => {
    setCurrentWorkspace(workspace);
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
                  Create a new workspace to organize your cloud resources and collaborate with your team.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit(handleCreateWorkspace)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Workspace Name</Label>
                  <Input
                    id="name"
                    placeholder="Enter workspace name"
                    {...register('name')}
                  />
                  {errors.name && (
                    <p className="text-sm text-red-600">{errors.name.message}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    placeholder="Enter workspace description"
                    {...register('description')}
                  />
                  {errors.description && (
                    <p className="text-sm text-red-600">{errors.description.message}</p>
                  )}
                </div>
                <div className="flex justify-end space-x-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setIsCreateDialogOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createWorkspaceMutation.isPending}>
                    {createWorkspaceMutation.isPending ? 'Creating...' : 'Create'}
                  </Button>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {workspaces.length === 0 ? (
          <div className="text-center py-12">
            <div className="mx-auto h-12 w-12 text-gray-400">
              <Users className="h-12 w-12" />
            </div>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No workspaces</h3>
            <p className="mt-1 text-sm text-gray-500">
              Get started by creating a new workspace.
            </p>
            <div className="mt-6">
              <Button onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Create Workspace
              </Button>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {workspaces.map((workspace) => (
              <Card key={workspace.id} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div>
                      <CardTitle className="text-lg">{workspace.name}</CardTitle>
                      <CardDescription className="mt-1">
                        {workspace.description}
                      </CardDescription>
                    </div>
                    <div className="flex space-x-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteWorkspace(workspace.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex items-center text-sm text-gray-500">
                      <Calendar className="mr-2 h-4 w-4" />
                      Created {new Date(workspace.created_at).toLocaleDateString()}
                    </div>
                    <div className="flex items-center justify-between">
                      <Badge variant="secondary">Owner</Badge>
                      <div className="flex space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => router.push(`/workspaces/${workspace.id}/members`)}
                        >
                          <Users className="mr-2 h-4 w-4" />
                          Members
                        </Button>
                        <Button
                          size="sm"
                          onClick={() => handleSelectWorkspace(workspace)}
                        >
                          Open
                        </Button>
                      </div>
                    </div>
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

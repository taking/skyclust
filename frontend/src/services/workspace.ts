import api from '@/lib/api';
import { ApiResponse, Workspace, CreateWorkspaceForm } from '@/lib/types';

export const workspaceService = {
  // Get all workspaces
  getWorkspaces: async (): Promise<Workspace[]> => {
    const response = await api.get<ApiResponse<{ workspaces: Workspace[] }>>('/api/v1/workspaces');
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to fetch workspaces');
    }
    return response.data.data?.workspaces || [];
  },

  // Get workspace by ID
  getWorkspace: async (id: string): Promise<Workspace> => {
    const response = await api.get<ApiResponse<{ workspace: Workspace }>>(`/api/v1/workspaces/${id}`);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to fetch workspace');
    }
    if (!response.data.data?.workspace) {
      throw new Error('Workspace not found');
    }
    return response.data.data.workspace;
  },

  // Create workspace
  createWorkspace: async (data: CreateWorkspaceForm): Promise<Workspace> => {
    const response = await api.post<ApiResponse<{ workspace: Workspace }>>('/api/v1/workspaces', data);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to create workspace');
    }
    if (!response.data.data?.workspace) {
      throw new Error('Failed to create workspace');
    }
    return response.data.data.workspace;
  },

  // Update workspace
  updateWorkspace: async (id: string, data: Partial<CreateWorkspaceForm>): Promise<Workspace> => {
    const response = await api.put<ApiResponse<{ workspace: Workspace }>>(`/api/v1/workspaces/${id}`, data);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to update workspace');
    }
    if (!response.data.data?.workspace) {
      throw new Error('Failed to update workspace');
    }
    return response.data.data.workspace;
  },

  // Delete workspace
  deleteWorkspace: async (id: string): Promise<void> => {
    const response = await api.delete(`/api/v1/workspaces/${id}`);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to delete workspace');
    }
  },

  // Add member to workspace
  addMember: async (workspaceId: string, userId: string, role: string = 'member'): Promise<void> => {
    await api.post(`/api/v1/workspaces/${workspaceId}/members`, {
      user_id: userId,
      role,
    });
  },

  // Remove member from workspace
  removeMember: async (workspaceId: string, userId: string): Promise<void> => {
    await api.delete(`/api/v1/workspaces/${workspaceId}/members/${userId}`);
  },
};

/**
 * Workspace Service
 * Workspace 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import type { Workspace, CreateWorkspaceForm } from '@/lib/types';

class WorkspaceService extends BaseService {
  // Get all workspaces
  async getWorkspaces(): Promise<Workspace[]> {
    const data = await this.get<{ workspaces: Workspace[] }>('/api/v1/workspaces');
    return data.workspaces || [];
  }

  // Get workspace by ID
  async getWorkspace(id: string): Promise<Workspace> {
    const data = await this.get<{ workspace: Workspace }>(`/api/v1/workspaces/${id}`);
    if (!data.workspace) {
      throw new Error('Workspace not found');
    }
    return data.workspace;
  }

  // Create workspace
  async createWorkspace(data: CreateWorkspaceForm): Promise<Workspace> {
    const result = await this.post<{ workspace: Workspace }>('/api/v1/workspaces', data);
    if (!result.workspace) {
      throw new Error('Failed to create workspace');
    }
    return result.workspace;
  }

  // Update workspace
  async updateWorkspace(id: string, data: Partial<CreateWorkspaceForm>): Promise<Workspace> {
    const result = await this.put<{ workspace: Workspace }>(`/api/v1/workspaces/${id}`, data);
    if (!result.workspace) {
      throw new Error('Failed to update workspace');
    }
    return result.workspace;
  }

  // Delete workspace
  async deleteWorkspace(id: string): Promise<void> {
    return this.delete<void>(`/api/v1/workspaces/${id}`);
  }

  // Add member to workspace
  async addMember(workspaceId: string, userId: string, role: string = 'member'): Promise<void> {
    return this.post<void>(`/api/v1/workspaces/${workspaceId}/members`, {
      user_id: userId,
      role,
    });
  }

  // Remove member from workspace
  async removeMember(workspaceId: string, userId: string): Promise<void> {
    return this.delete<void>(`/api/v1/workspaces/${workspaceId}/members/${userId}`);
  }
}

export const workspaceService = new WorkspaceService();

/**
 * Workspace Service
 * Workspace 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import type { Workspace, CreateWorkspaceForm } from '@/lib/types';

class WorkspaceService extends BaseService {
  // Get all workspaces
  async getWorkspaces(): Promise<Workspace[]> {
    const data = await this.get<{ workspaces: Workspace[] }>('workspaces');
    return data.workspaces || [];
  }

  // Get workspace by ID
  async getWorkspace(id: string): Promise<Workspace> {
    // Backend returns { success: true, data: Workspace, ... }
    // The BaseService.get already extracts the data field
    const workspace = await this.get<Workspace>(`workspaces/${id}`);
    if (!workspace) {
      throw new Error('Workspace not found');
    }
    return workspace;
  }

  // Create workspace
  async createWorkspace(data: CreateWorkspaceForm): Promise<Workspace> {
    // Backend returns { success: true, data: Workspace, ... }
    const workspace = await this.post<Workspace>('workspaces', data);
    if (!workspace) {
      throw new Error('Failed to create workspace');
    }
    return workspace;
  }

  // Update workspace
  async updateWorkspace(id: string, data: Partial<CreateWorkspaceForm>): Promise<Workspace> {
    // Backend returns { success: true, data: Workspace, ... }
    const workspace = await this.put<Workspace>(`workspaces/${id}`, data);
    if (!workspace) {
      throw new Error('Failed to update workspace');
    }
    return workspace;
  }

  // Delete workspace
  async deleteWorkspace(id: string): Promise<void> {
    return this.delete<void>(`workspaces/${id}`);
  }

  // Add member to workspace
  async addMember(workspaceId: string, userId: string, role: string = 'member'): Promise<void> {
    return this.post<void>(`workspaces/${workspaceId}/members`, {
      user_id: userId,
      role,
    });
  }

  // Remove member from workspace
  async removeMember(workspaceId: string, userId: string): Promise<void> {
    return this.delete<void>(`workspaces/${workspaceId}/members/${userId}`);
  }
}

export const workspaceService = new WorkspaceService();


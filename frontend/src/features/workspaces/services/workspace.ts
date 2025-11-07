/**
 * Workspace Service
 * Clean Architecture: Application 계층 - Repository를 통한 데이터 접근
 */

import { workspaceRepository } from '@/infrastructure/repositories';
import type { Workspace, CreateWorkspaceForm, WorkspaceMember } from '@/lib/types';

class WorkspaceService {
  // Get all workspaces
  async getWorkspaces(): Promise<Workspace[]> {
    return workspaceRepository.list();
  }

  // Get workspace by ID
  async getWorkspace(id: string): Promise<Workspace> {
    return workspaceRepository.getById(id);
  }

  // Create workspace
  async createWorkspace(data: CreateWorkspaceForm): Promise<Workspace> {
    return workspaceRepository.create(data);
  }

  // Update workspace
  async updateWorkspace(id: string, data: Partial<CreateWorkspaceForm>): Promise<Workspace> {
    return workspaceRepository.update(id, data);
  }

  // Delete workspace
  async deleteWorkspace(id: string): Promise<void> {
    return workspaceRepository.delete(id);
  }

  // Get workspace members
  async getMembers(workspaceId: string): Promise<WorkspaceMember[]> {
    return workspaceRepository.getMembers(workspaceId);
  }

  // Add member to workspace by email
  async addMember(workspaceId: string, email: string, role: string = 'member'): Promise<void> {
    return workspaceRepository.addMember(workspaceId, email, role);
  }

  // Remove member from workspace
  async removeMember(workspaceId: string, userId: string): Promise<void> {
    return workspaceRepository.removeMember(workspaceId, userId);
  }
}

export const workspaceService = new WorkspaceService();


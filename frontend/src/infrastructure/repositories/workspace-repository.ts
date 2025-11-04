/**
 * Workspace Repository Implementation
 * Workspace Service를 Repository 패턴으로 래핑
 */

import { workspaceService } from '@/features/workspaces';
import type { IWorkspaceRepository } from '@/lib/types/repository';
import type { Workspace, CreateWorkspaceForm } from '@/lib/types';

export class WorkspaceRepository implements IWorkspaceRepository {
  async findAll(): Promise<Workspace[]> {
    return this.list();
  }

  async findById(id: string): Promise<Workspace | null> {
    const workspace = await workspaceService.getWorkspace(id);
    return workspace || null;
  }

  async list(): Promise<Workspace[]> {
    return workspaceService.getWorkspaces();
  }

  async create(data: CreateWorkspaceForm): Promise<Workspace> {
    return workspaceService.createWorkspace(data);
  }

  async update(id: string, data: Partial<CreateWorkspaceForm>): Promise<Workspace> {
    return workspaceService.updateWorkspace(id, data);
  }

  async delete(id: string): Promise<void> {
    return workspaceService.deleteWorkspace(id);
  }

  async addMember(workspaceId: string, userId: string, role: string = 'member'): Promise<void> {
    return workspaceService.addMember(workspaceId, userId, role);
  }

  async removeMember(workspaceId: string, userId: string): Promise<void> {
    return workspaceService.removeMember(workspaceId, userId);
  }
}

export const workspaceRepository = new WorkspaceRepository();


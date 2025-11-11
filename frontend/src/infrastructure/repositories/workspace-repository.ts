/**
 * Workspace Repository Implementation
 * Clean Architecture: Infrastructure 계층 - API 직접 호출
 */

import { BaseRepository } from './base-repository';
import { API_ENDPOINTS } from '@/lib/api';
import type { IWorkspaceRepository } from '@/lib/types/repository';
import type { Workspace, CreateWorkspaceForm, WorkspaceMember } from '@/lib/types';

export class WorkspaceRepository extends BaseRepository implements IWorkspaceRepository {
  async findAll(): Promise<Workspace[]> {
    return this.list();
  }

  async findById(id: string): Promise<Workspace | null> {
    try {
      const workspace = await this.getById(id);
      return workspace;
    } catch {
      return null;
    }
  }

  async getById(id: string): Promise<Workspace> {
    const workspace = await this.get<Workspace>(API_ENDPOINTS.workspaces.detail(id));
    if (!workspace) {
      throw new Error(`Workspace with id ${id} not found`);
    }
    return workspace;
  }

  async list(): Promise<Workspace[]> {
    const data = await this.get<Workspace[]>(API_ENDPOINTS.workspaces.list());
    return Array.isArray(data) ? data : [];
  }

  async create(data: CreateWorkspaceForm): Promise<Workspace> {
    const workspace = await this.post<Workspace>(API_ENDPOINTS.workspaces.create(), data);
    if (!workspace) {
      throw new Error('Failed to create workspace');
    }
    return workspace;
  }

  async update(id: string, data: Partial<CreateWorkspaceForm>): Promise<Workspace> {
    const workspace = await this.put<Workspace>(API_ENDPOINTS.workspaces.update(id), data);
    if (!workspace) {
      throw new Error('Failed to update workspace');
    }
    return workspace;
  }

  // IWorkspaceRepository 인터페이스의 delete 메서드 구현
  // BaseService의 protected delete와 이름이 같아 TypeScript 타입 충돌 발생
  // @ts-expect-error - BaseService의 protected delete와 이름이 같지만 인터페이스 구현을 위해 필요
  async delete(id: string): Promise<void> {
    const url = this.buildApiUrl(API_ENDPOINTS.workspaces.delete(id));
    await this.request<void>('delete', url, undefined);
  }

  async getMembers(workspaceId: string): Promise<WorkspaceMember[]> {
    return this.get<{ members: WorkspaceMember[] }>(API_ENDPOINTS.workspaces.members.list(workspaceId)).then(
      (response) => response.members || []
    );
  }

  async addMember(workspaceId: string, email: string, role: string = 'member'): Promise<void> {
    return this.post<void>(API_ENDPOINTS.workspaces.members.add(workspaceId), {
      email,
      role,
    });
  }

  async removeMember(workspaceId: string, userId: string): Promise<void> {
    const url = this.buildApiUrl(API_ENDPOINTS.workspaces.members.remove(workspaceId, userId));
    await this.request<void>('delete', url, undefined);
  }
}

export const workspaceRepository = new WorkspaceRepository();


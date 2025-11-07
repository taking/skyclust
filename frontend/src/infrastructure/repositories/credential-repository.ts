/**
 * Credential Repository Implementation
 * Clean Architecture: Infrastructure 계층 - API 직접 호출
 */

import { BaseRepository } from './base-repository';
import { API_ENDPOINTS } from '@/lib/api-endpoints';
import type { ICredentialRepository } from '@/lib/types/repository';
import type { Credential, CreateCredentialForm } from '@/lib/types';

export class CredentialRepository extends BaseRepository implements ICredentialRepository {
  async findAll(): Promise<Credential[]> {
    throw new Error('Use list() method with workspaceId');
  }

  async findById(id: string): Promise<Credential | null> {
    try {
      const credential = await this.getById(id);
      return credential;
    } catch {
      return null;
    }
  }

  async getById(id: string): Promise<Credential> {
    const credential = await this.get<Credential>(API_ENDPOINTS.credentials.detail(id));
    if (!credential) {
      throw new Error(`Credential with id ${id} not found`);
    }
    return credential;
  }

  async list(workspaceId: string): Promise<Credential[]> {
    const data = await this.get<{ credentials: Credential[] } | Credential[]>(
      API_ENDPOINTS.credentials.list(workspaceId)
    );
    
    if (Array.isArray(data)) {
      return data;
    }
    
    if (data && typeof data === 'object' && 'credentials' in data) {
      return (data as { credentials: Credential[] }).credentials || [];
    }
    
    return [];
  }

  async create(data: CreateCredentialForm & { workspace_id: string; name?: string }): Promise<Credential> {
    return this.post<Credential>(API_ENDPOINTS.credentials.create(), {
      workspace_id: data.workspace_id,
      name: data.name || `${data.provider.toUpperCase()} Credential`,
      provider: data.provider,
      data: data.credentials || {},
    });
  }

  async createFromFile(workspaceId: string, name: string, provider: string, file: File): Promise<Credential> {
    const formData = new FormData();
    formData.append('workspace_id', workspaceId);
    formData.append('name', name);
    formData.append('provider', provider);
    formData.append('file', file);
    
    const url = this.buildApiUrl(API_ENDPOINTS.credentials.upload());
    return this.request<Credential>('post', url, formData);
  }

  async update(id: string, data: Partial<CreateCredentialForm>): Promise<Credential> {
    return this.put<Credential>(API_ENDPOINTS.credentials.update(id), data);
  }

  // ICredentialRepository 인터페이스의 delete 메서드 구현
  // BaseService의 protected delete와 이름이 같아 TypeScript 타입 충돌 발생
  // @ts-expect-error - BaseService의 protected delete와 이름이 같지만 인터페이스 구현을 위해 필요
  async delete(id: string): Promise<void> {
    const url = this.buildApiUrl(API_ENDPOINTS.credentials.delete(id));
    // BaseService의 protected request 메서드를 직접 사용하여 delete 요청
    await this.request<void>('delete', url, undefined);
  }
}

export const credentialRepository = new CredentialRepository();


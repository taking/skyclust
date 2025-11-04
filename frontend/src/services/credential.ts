/**
 * Credential Service
 * Credential 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import type { Credential, CreateCredentialForm } from '@/lib/types';

class CredentialService extends BaseService {
  // Get credentials by workspace
  async getCredentials(workspaceId: string): Promise<Credential[]> {
    const data = await this.get<{ credentials: Credential[] } | Credential[]>(
      `credentials?workspace_id=${workspaceId}`
    );
    
    // Backend returns { credentials: [...] } inside data field
    if (Array.isArray(data)) {
      return data;
    }
    
    if (data && typeof data === 'object' && 'credentials' in data) {
      return (data as { credentials: Credential[] }).credentials || [];
    }
    
    return [];
  }

  // Get credential by ID
  async getCredential(id: string): Promise<Credential> {
    return this.get<Credential>(`credentials/${id}`);
  }

  // Create credential
  async createCredential(data: CreateCredentialForm & { workspace_id: string; name?: string }): Promise<Credential> {
    return this.post<Credential>('credentials', {
      workspace_id: data.workspace_id,
      name: data.name || `${data.provider.toUpperCase()} Credential`,
      provider: data.provider,
      data: data.credentials || {},
    });
  }

  // Create credential from file (multipart/form-data)
  // FormData는 BaseService의 request를 직접 사용해야 함
  async createCredentialFromFile(workspaceId: string, name: string, provider: string, file: File): Promise<Credential> {
    const formData = new FormData();
    formData.append('workspace_id', workspaceId);
    formData.append('name', name);
    formData.append('provider', provider);
    formData.append('file', file);
    
    // FormData는 BaseService의 request 메서드를 직접 사용
    const url = this.buildApiUrl('credentials/upload');
    return this.request<Credential>('post', url, formData);
  }

  // Update credential
  async updateCredential(id: string, data: Partial<CreateCredentialForm>): Promise<Credential> {
    return this.put<Credential>(`credentials/${id}`, data);
  }

  // Delete credential
  async deleteCredential(id: string): Promise<void> {
    return this.delete<void>(`credentials/${id}`);
  }
}

export const credentialService = new CredentialService();

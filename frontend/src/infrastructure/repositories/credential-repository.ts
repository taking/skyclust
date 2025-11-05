/**
 * Credential Repository Implementation
 * Credential Service를 Repository 패턴으로 래핑
 */

import { credentialService } from '@/services/credential';
import type { ICredentialRepository } from '@/lib/types/repository';
import type { Credential, CreateCredentialForm } from '@/lib/types';

export class CredentialRepository implements ICredentialRepository {
  async findAll(): Promise<Credential[]> {
    throw new Error('Use list() method with workspaceId');
  }

  async findById(id: string): Promise<Credential | null> {
    const credential = await credentialService.getCredential(id);
    return credential || null;
  }

  async getById(id: string): Promise<Credential> {
    const credential = await credentialService.getCredential(id);
    if (!credential) {
      throw new Error(`Credential with id ${id} not found`);
    }
    return credential;
  }

  async list(workspaceId: string): Promise<Credential[]> {
    return credentialService.getCredentials(workspaceId);
  }

  async create(data: CreateCredentialForm & { workspace_id: string; name?: string }): Promise<Credential> {
    return credentialService.createCredential(data);
  }

  async createFromFile(workspaceId: string, name: string, provider: string, file: File): Promise<Credential> {
    return credentialService.createCredentialFromFile(workspaceId, name, provider, file);
  }

  async update(id: string, data: Partial<CreateCredentialForm>): Promise<Credential> {
    return credentialService.updateCredential(id, data);
  }

  async delete(id: string): Promise<void> {
    return credentialService.deleteCredential(id);
  }
}

export const credentialRepository = new CredentialRepository();


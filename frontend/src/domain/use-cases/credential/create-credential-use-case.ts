/**
 * Create Credential Use Case
 * Credential 생성 비즈니스 로직
 */

import type { ICredentialRepository } from '@/lib/types/repository';
import type { Credential, CreateCredentialForm } from '@/lib/types';

export interface CreateCredentialUseCaseInput {
  workspaceId: string;
  data: CreateCredentialForm;
  name?: string;
}

export class CreateCredentialUseCase {
  constructor(private credentialRepository: ICredentialRepository) {}

  async execute(input: CreateCredentialUseCaseInput): Promise<Credential> {
    const { workspaceId, data, name } = input;

    // Validation
    if (!workspaceId || workspaceId.trim().length === 0) {
      throw new Error('Workspace ID is required');
    }

    if (!data.provider) {
      throw new Error('Provider is required');
    }

    // Business logic: Provider별 validation
    this.validateProviderData(data.provider, data.credentials || {});

    // Business logic: 기본 이름 생성
    const credentialName = name || `${data.provider.toUpperCase()} Credential`;

    // Repository 호출
    return this.credentialRepository.create({
      ...data,
      workspace_id: workspaceId,
      name: credentialName,
    });
  }

  private validateProviderData(provider: string, credentials: Record<string, unknown>): void {
    switch (provider.toLowerCase()) {
      case 'aws':
        if (!credentials.access_key || !credentials.secret_key) {
          throw new Error('AWS credentials require access_key and secret_key');
        }
        break;
      case 'gcp':
        if (!credentials.project_id) {
          throw new Error('GCP credentials require project_id');
        }
        break;
      case 'azure':
        if (!credentials.subscription_id || !credentials.client_id || !credentials.client_secret || !credentials.tenant_id) {
          throw new Error('Azure credentials require subscription_id, client_id, client_secret, and tenant_id');
        }
        break;
      default:
        throw new Error(`Unsupported provider: ${provider}`);
    }
  }
}


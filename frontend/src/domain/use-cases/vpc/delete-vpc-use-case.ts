/**
 * Delete VPC Use Case
 * VPC 삭제 비즈니스 로직
 */

import type { IVPCRepository } from '@/lib/types/repository';
import type { CloudProvider } from '@/lib/types';

export interface DeleteVPCUseCaseInput {
  provider: CloudProvider;
  vpcId: string;
  credentialId: string;
  region: string;
}

export class DeleteVPCUseCase {
  constructor(private vpcRepository: IVPCRepository) {}

  async execute(input: DeleteVPCUseCaseInput): Promise<void> {
    const { provider, vpcId, credentialId, region } = input;

    // Validation
    if (!provider) {
      throw new Error('Provider is required');
    }

    if (!vpcId || vpcId.trim().length === 0) {
      throw new Error('VPC ID is required');
    }

    if (!credentialId || credentialId.trim().length === 0) {
      throw new Error('Credential ID is required');
    }

    if (!region || region.trim().length === 0) {
      throw new Error('Region is required');
    }

    // Business logic: VPC 삭제 전 확인 (Repository에서 처리)
    return this.vpcRepository.delete(provider, vpcId, credentialId, region);
  }
}


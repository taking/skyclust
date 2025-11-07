/**
 * Bulk Delete VPCs Use Case
 * 여러 VPC 일괄 삭제 비즈니스 로직
 */

import type { IVPCRepository } from '@/lib/types/repository';
import type { CloudProvider } from '@/lib/types';
import { DeleteVPCUseCase } from './delete-vpc-use-case';
import { logger } from '@/lib/logger';

export interface BulkDeleteVPCsUseCaseInput {
  provider: CloudProvider;
  vpcIds: string[];
  vpcs: Array<{ id: string; region?: string }>;
  credentialId: string;
  defaultRegion: string;
}

export class BulkDeleteVPCsUseCase {
  private deleteVPCUseCase: DeleteVPCUseCase;

  constructor(vpcRepository: IVPCRepository) {
    this.deleteVPCUseCase = new DeleteVPCUseCase(vpcRepository);
  }

  async execute(input: BulkDeleteVPCsUseCaseInput): Promise<void> {
    const { provider, vpcIds, vpcs, credentialId, defaultRegion } = input;

    // Validation
    if (!provider) {
      throw new Error('Provider is required');
    }

    if (!vpcIds || vpcIds.length === 0) {
      throw new Error('At least one VPC ID is required');
    }

    if (!credentialId || credentialId.trim().length === 0) {
      throw new Error('Credential ID is required');
    }

    // Business logic: 각 VPC를 순차적으로 삭제
    const vpcsToDelete = vpcs.filter(v => vpcIds.includes(v.id));

    if (vpcsToDelete.length === 0) {
      throw new Error('No VPCs found to delete');
    }

    // 병렬 처리로 삭제 (하나 실패해도 다른 것은 계속 진행)
    const deletePromises = vpcsToDelete.map(vpc =>
      this.deleteVPCUseCase.execute({
        provider,
        vpcId: vpc.id,
        credentialId,
        region: vpc.region || defaultRegion,
      }).catch(error => {
        // 개별 삭제 실패는 로깅만 하고 계속 진행
        logger.error(`Failed to delete VPC ${vpc.id}`, error instanceof Error ? error : new Error(String(error)), {
          vpcId: vpc.id,
          provider,
          credentialId,
        });
        throw error;
      })
    );

    await Promise.all(deletePromises);
  }
}


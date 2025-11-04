/**
 * Create VPC Use Case
 * VPC 생성 비즈니스 로직
 */

import type { IVPCRepository } from '@/lib/types/repository';
import type { VPC, CreateVPCForm, CloudProvider } from '@/lib/types';

export interface CreateVPCUseCaseInput {
  provider: CloudProvider;
  data: CreateVPCForm;
}

export class CreateVPCUseCase {
  constructor(private vpcRepository: IVPCRepository) {}

  async execute(input: CreateVPCUseCaseInput): Promise<VPC> {
    const { provider, data } = input;

    // Validation
    if (!provider) {
      throw new Error('Provider is required');
    }

    if (!data.name || data.name.trim().length === 0) {
      throw new Error('VPC name is required');
    }

    // Business logic: Validate CIDR block if provided
    if (data.cidr_block) {
      if (!this.isValidCIDR(data.cidr_block)) {
        throw new Error('Invalid CIDR block format');
      }
    }

    // Repository 호출
    return this.vpcRepository.create(provider, data);
  }

  private isValidCIDR(cidr: string): boolean {
    const cidrRegex = /^([0-9]{1,3}\.){3}[0-9]{1,3}\/([0-9]|[1-2][0-9]|3[0-2])$/;
    if (!cidrRegex.test(cidr)) {
      return false;
    }

    const parts = cidr.split('/');
    const ip = parts[0].split('.');
    const mask = parseInt(parts[1], 10);

    return ip.every(octet => {
      const num = parseInt(octet, 10);
      return num >= 0 && num <= 255;
    }) && mask >= 0 && mask <= 32;
  }
}


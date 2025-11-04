/**
 * VPC Repository Implementation
 * Network Service를 Repository 패턴으로 래핑
 */

import { networkService } from '@/services/network';
import type { IVPCRepository } from '@/lib/types/repository';
import type { VPC, CreateVPCForm, CloudProvider } from '@/lib/types';

export class VPCRepository implements IVPCRepository {
  async findAll(): Promise<VPC[]> {
    throw new Error('Use list() method with provider and credential');
  }

  async findById(): Promise<VPC | null> {
    throw new Error('Use getById() method with provider and credential');
  }

  async create(): Promise<VPC> {
    throw new Error('Use create() method with provider');
  }

  async update(): Promise<VPC> {
    throw new Error('Use update() method with provider');
  }

  async delete(): Promise<void> {
    throw new Error('Use delete() method with provider');
  }

  async list(provider: CloudProvider, credentialId: string, region?: string): Promise<VPC[]> {
    return networkService.listVPCs(provider, credentialId, region);
  }

  async getById(provider: CloudProvider, vpcId: string, credentialId: string, region: string): Promise<VPC> {
    return networkService.getVPC(provider, vpcId, credentialId, region);
  }

  async create(provider: CloudProvider, data: CreateVPCForm): Promise<VPC> {
    return networkService.createVPC(provider, data);
  }

  async update(provider: CloudProvider, vpcId: string, data: Partial<CreateVPCForm>, credentialId: string, region: string): Promise<VPC> {
    return networkService.updateVPC(provider, vpcId, data, credentialId, region);
  }

  async delete(provider: CloudProvider, vpcId: string, credentialId: string, region: string): Promise<void> {
    return networkService.deleteVPC(provider, vpcId, credentialId, region);
  }
}

export const vpcRepository = new VPCRepository();


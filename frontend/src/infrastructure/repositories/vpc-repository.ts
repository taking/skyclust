/**
 * VPC Repository Implementation
 * Clean Architecture: Infrastructure 계층 - API 직접 호출
 */

import { BaseRepository } from './base-repository';
import { API_ENDPOINTS } from '@/lib/api-endpoints';
import type { IVPCRepository } from '@/lib/types/repository';
import type { VPC, CreateVPCForm, CloudProvider } from '@/lib/types';

export class VPCRepository extends BaseRepository implements IVPCRepository {
  async findAll(): Promise<VPC[]> {
    throw new Error('Use list() method with provider and credential');
  }

  async findById(): Promise<VPC | null> {
    throw new Error('Use getById() method with provider and credential');
  }

  async list(provider: CloudProvider, credentialId: string, region?: string): Promise<VPC[]> {
    const data = await this.get<{ vpcs: VPC[] }>(
      API_ENDPOINTS.network.vpcs.list(provider, credentialId, region)
    );
    return data.vpcs || [];
  }

  async getById(provider: CloudProvider, vpcId: string, credentialId: string, region: string): Promise<VPC> {
    return this.get<VPC>(
      API_ENDPOINTS.network.vpcs.detail(provider, vpcId, credentialId, region)
    );
  }

  async create(provider: CloudProvider, data: CreateVPCForm): Promise<VPC> {
    return this.post<VPC>(API_ENDPOINTS.network.vpcs.create(provider), data);
  }

  async update(provider: CloudProvider, vpcId: string, data: Partial<CreateVPCForm>, credentialId: string, region: string): Promise<VPC> {
    return this.put<VPC>(
      API_ENDPOINTS.network.vpcs.update(provider, vpcId, credentialId, region),
      data
    );
  }

  // IVPCRepository 인터페이스의 delete 메서드 구현
  // BaseService의 protected delete와 이름이 같아 TypeScript 타입 충돌 발생
  // @ts-expect-error - BaseService의 protected delete와 이름이 같지만 인터페이스 구현을 위해 필요
  async delete(provider: CloudProvider, vpcId: string, credentialId: string, region: string): Promise<void> {
    const url = this.buildApiUrl(API_ENDPOINTS.network.vpcs.delete(provider, vpcId, credentialId, region));
    await this.request<void>('delete', url, undefined);
  }
}

export const vpcRepository = new VPCRepository();


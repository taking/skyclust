/**
 * Network Service
 * Network 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import { API_ENDPOINTS } from '@/lib/api-endpoints';
import type {
  VPC,
  Subnet,
  SecurityGroup,
  CreateVPCForm,
  CreateSubnetForm,
  CreateSecurityGroupForm,
  SecurityGroupRule,
  CloudProvider,
} from '@/lib/types';

class NetworkService extends BaseService {
  // ===== VPC Management =====
  
  /**
   * VPC 목록 조회
   * 
   * @param provider - 클라우드 프로바이더 (aws, gcp, azure 등)
   * @param credentialId - 자격 증명 ID
   * @param region - 리전 (선택사항)
   * @returns VPC 배열
   * 
   * @example
   * ```tsx
   * const vpcs = await networkService.listVPCs('aws', 'credential-id', 'ap-northeast-2');
   * ```
   */
  async listVPCs(
    provider: CloudProvider,
    credentialId: string,
    region?: string
  ): Promise<VPC[]> {
    // 1. API 호출하여 VPC 목록 가져오기
    const data = await this.get<{ vpcs: VPC[] }>(
      API_ENDPOINTS.network.vpcs.list(provider, credentialId, region)
    );
    // 2. 응답 데이터에서 vpcs 배열 추출 (없으면 빈 배열)
    return data.vpcs || [];
  }

  /**
   * 특정 VPC 조회
   * 
   * @param provider - 클라우드 프로바이더
   * @param vpcId - VPC ID
   * @param credentialId - 자격 증명 ID
   * @param region - 리전
   * @returns VPC 정보
   * 
   * @example
   * ```tsx
   * const vpc = await networkService.getVPC('aws', 'vpc-123', 'credential-id', 'ap-northeast-2');
   * ```
   */
  async getVPC(
    provider: CloudProvider,
    vpcId: string,
    credentialId: string,
    region: string
  ): Promise<VPC> {
    return this.get<VPC>(
      API_ENDPOINTS.network.vpcs.detail(provider, vpcId, credentialId, region)
    );
  }

  /**
   * VPC 생성
   * 
   * @param provider - 클라우드 프로바이더
   * @param data - VPC 생성 데이터 (name, cidr_block, region 등)
   * @returns 생성된 VPC 정보
   * 
   * @example
   * ```tsx
   * const vpc = await networkService.createVPC('aws', {
   *   credential_id: 'credential-id',
   *   name: 'my-vpc',
   *   cidr_block: '10.0.0.0/16',
   *   region: 'ap-northeast-2',
   * });
   * ```
   */
  async createVPC(
    provider: CloudProvider,
    data: CreateVPCForm
  ): Promise<VPC> {
    return this.post<VPC>(
      API_ENDPOINTS.network.vpcs.create(provider),
      data
    );
  }

  async updateVPC(
    provider: CloudProvider,
    vpcId: string,
    data: Partial<CreateVPCForm>,
    credentialId: string,
    region: string
  ): Promise<VPC> {
    return this.put<VPC>(
      API_ENDPOINTS.network.vpcs.update(provider, vpcId, credentialId, region),
      data
    );
  }

  async deleteVPC(
    provider: CloudProvider,
    vpcId: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    return this.delete<void>(
      API_ENDPOINTS.network.vpcs.delete(provider, vpcId, credentialId, region)
    );
  }

  // Subnet 관리
  async listSubnets(
    provider: CloudProvider,
    credentialId: string,
    vpcId: string,
    region: string
  ): Promise<Subnet[]> {
    const data = await this.get<{ subnets: Subnet[] }>(
      API_ENDPOINTS.network.subnets.list(provider, credentialId, vpcId, region)
    );
    return data.subnets || [];
  }

  async getSubnet(
    provider: CloudProvider,
    subnetId: string,
    credentialId: string,
    region: string
  ): Promise<Subnet> {
    return this.get<Subnet>(
      API_ENDPOINTS.network.subnets.detail(provider, subnetId, credentialId, region)
    );
  }

  async createSubnet(
    provider: CloudProvider,
    data: CreateSubnetForm
  ): Promise<Subnet> {
    return this.post<Subnet>(
      API_ENDPOINTS.network.subnets.create(provider),
      data
    );
  }

  async updateSubnet(
    provider: CloudProvider,
    subnetId: string,
    data: Partial<CreateSubnetForm>,
    credentialId: string,
    region: string
  ): Promise<Subnet> {
    return this.put<Subnet>(
      API_ENDPOINTS.network.subnets.update(provider, subnetId, credentialId, region),
      data
    );
  }

  async deleteSubnet(
    provider: CloudProvider,
    subnetId: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    return this.delete<void>(
      API_ENDPOINTS.network.subnets.delete(provider, subnetId, credentialId, region)
    );
  }

  // Security Group 관리
  async listSecurityGroups(
    provider: CloudProvider,
    credentialId: string,
    vpcId: string,
    region: string
  ): Promise<SecurityGroup[]> {
    const data = await this.get<{ security_groups: SecurityGroup[] }>(
      API_ENDPOINTS.network.securityGroups.list(provider, credentialId, vpcId, region)
    );
    return data.security_groups || [];
  }

  async getSecurityGroup(
    provider: CloudProvider,
    securityGroupId: string,
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> {
    return this.get<SecurityGroup>(
      API_ENDPOINTS.network.securityGroups.detail(provider, securityGroupId, credentialId, region)
    );
  }

  async createSecurityGroup(
    provider: CloudProvider,
    data: CreateSecurityGroupForm
  ): Promise<SecurityGroup> {
    return this.post<SecurityGroup>(
      API_ENDPOINTS.network.securityGroups.create(provider),
      data
    );
  }

  async updateSecurityGroup(
    provider: CloudProvider,
    securityGroupId: string,
    data: Partial<CreateSecurityGroupForm>,
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> {
    return this.put<SecurityGroup>(
      API_ENDPOINTS.network.securityGroups.update(provider, securityGroupId, credentialId, region),
      data
    );
  }

  async deleteSecurityGroup(
    provider: CloudProvider,
    securityGroupId: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    return this.delete<void>(
      API_ENDPOINTS.network.securityGroups.delete(provider, securityGroupId, credentialId, region)
    );
  }

  // Security Group Rule 관리
  async addSecurityGroupRule(
    provider: CloudProvider,
    securityGroupId: string,
    rule: SecurityGroupRule,
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> {
    return this.post<SecurityGroup>(
      API_ENDPOINTS.network.securityGroups.rules.add(provider, securityGroupId, credentialId, region),
      rule
    );
  }

  async removeSecurityGroupRule(
    provider: CloudProvider,
    securityGroupId: string,
    ruleId: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    return this.delete<void>(
      API_ENDPOINTS.network.securityGroups.rules.remove(provider, securityGroupId, credentialId, region, ruleId)
    );
  }

  async updateSecurityGroupRules(
    provider: CloudProvider,
    securityGroupId: string,
    rules: SecurityGroupRule[],
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> {
    return this.put<SecurityGroup>(
      API_ENDPOINTS.network.securityGroups.rules.update(provider, securityGroupId, credentialId, region),
      { rules }
    );
  }
}

export const networkService = new NetworkService();

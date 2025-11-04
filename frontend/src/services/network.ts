/**
 * Network Service
 * Network 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
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
  // VPC management
  async listVPCs(
    provider: CloudProvider,
    credentialId: string,
    region?: string
  ): Promise<VPC[]> {
    const params = new URLSearchParams({ credential_id: credentialId });
    if (region) params.append('region', region);
    
    const data = await this.get<{ vpcs: VPC[] }>(
      `${provider}/network/vpcs?${params.toString()}`
    );
    return data.vpcs || [];
  }

  async getVPC(
    provider: CloudProvider,
    vpcId: string,
    credentialId: string,
    region: string
  ): Promise<VPC> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.get<VPC>(
      `${provider}/network/vpcs/${encodeURIComponent(vpcId)}?${params.toString()}`
    );
  }

  async createVPC(
    provider: CloudProvider,
    data: CreateVPCForm
  ): Promise<VPC> {
    return this.post<VPC>(
      `${provider}/network/vpcs`,
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
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.put<VPC>(
      `${provider}/network/vpcs/${encodeURIComponent(vpcId)}?${params.toString()}`,
      data
    );
  }

  async deleteVPC(
    provider: CloudProvider,
    vpcId: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.delete<void>(
      `${provider}/network/vpcs/${encodeURIComponent(vpcId)}?${params.toString()}`
    );
  }

  // Subnet management
  async listSubnets(
    provider: CloudProvider,
    credentialId: string,
    vpcId: string,
    region: string
  ): Promise<Subnet[]> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      vpc_id: vpcId,
      region,
    });
    
    const data = await this.get<{ subnets: Subnet[] }>(
      `${provider}/network/subnets?${params.toString()}`
    );
    return data.subnets || [];
  }

  async getSubnet(
    provider: CloudProvider,
    subnetId: string,
    credentialId: string,
    region: string
  ): Promise<Subnet> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.get<Subnet>(
      `${provider}/network/subnets/${encodeURIComponent(subnetId)}?${params.toString()}`
    );
  }

  async createSubnet(
    provider: CloudProvider,
    data: CreateSubnetForm
  ): Promise<Subnet> {
    return this.post<Subnet>(
      `${provider}/network/subnets`,
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
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.put<Subnet>(
      `${provider}/network/subnets/${encodeURIComponent(subnetId)}?${params.toString()}`,
      data
    );
  }

  async deleteSubnet(
    provider: CloudProvider,
    subnetId: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.delete<void>(
      `${provider}/network/subnets/${encodeURIComponent(subnetId)}?${params.toString()}`
    );
  }

  // Security Group management
  async listSecurityGroups(
    provider: CloudProvider,
    credentialId: string,
    vpcId: string,
    region: string
  ): Promise<SecurityGroup[]> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      vpc_id: vpcId,
      region,
    });
    
    const data = await this.get<{ security_groups: SecurityGroup[] }>(
      `${provider}/network/security-groups?${params.toString()}`
    );
    return data.security_groups || [];
  }

  async getSecurityGroup(
    provider: CloudProvider,
    securityGroupId: string,
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.get<SecurityGroup>(
      `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}?${params.toString()}`
    );
  }

  async createSecurityGroup(
    provider: CloudProvider,
    data: CreateSecurityGroupForm
  ): Promise<SecurityGroup> {
    return this.post<SecurityGroup>(
      `${provider}/network/security-groups`,
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
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.put<SecurityGroup>(
      `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}?${params.toString()}`,
      data
    );
  }

  async deleteSecurityGroup(
    provider: CloudProvider,
    securityGroupId: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.delete<void>(
      `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}?${params.toString()}`
    );
  }

  // Security Group Rule management
  async addSecurityGroupRule(
    provider: CloudProvider,
    securityGroupId: string,
    rule: SecurityGroupRule,
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.post<SecurityGroup>(
      `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules?${params.toString()}`,
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
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
      rule_id: ruleId,
    });
    
    return this.delete<void>(
      `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules?${params.toString()}`
    );
  }

  async updateSecurityGroupRules(
    provider: CloudProvider,
    securityGroupId: string,
    rules: SecurityGroupRule[],
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.put<SecurityGroup>(
      `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules?${params.toString()}`,
      { rules }
    );
  }
}

export const networkService = new NetworkService();

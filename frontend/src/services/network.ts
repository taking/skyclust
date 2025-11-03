import api from '@/lib/api';
import { 
  ApiResponse, 
  VPC,
  Subnet,
  SecurityGroup,
  CreateVPCForm,
  CreateSubnetForm,
  CreateSecurityGroupForm,
  SecurityGroupRule,
  CloudProvider
} from '@/lib/types';

export const networkService = {
  // VPC management
  listVPCs: async (
    provider: CloudProvider,
    credentialId: string,
    region?: string
  ): Promise<VPC[]> => {
    const params = new URLSearchParams({ credential_id: credentialId });
    if (region) params.append('region', region);
    
    const response = await api.get<ApiResponse<{ vpcs: VPC[] }>>(
      `/api/v1/${provider}/network/vpcs?${params.toString()}`
    );
    return response.data.data?.vpcs || [];
  },

  getVPC: async (
    provider: CloudProvider,
    vpcId: string,
    credentialId: string,
    region: string
  ): Promise<VPC> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<VPC>>(
      `/api/v1/${provider}/network/vpcs/${encodeURIComponent(vpcId)}?${params.toString()}`
    );
    return response.data.data!;
  },

  createVPC: async (
    provider: CloudProvider,
    data: CreateVPCForm
  ): Promise<VPC> => {
    const response = await api.post<ApiResponse<VPC>>(
      `/api/v1/${provider}/network/vpcs`,
      data
    );
    return response.data.data!;
  },

  updateVPC: async (
    provider: CloudProvider,
    vpcId: string,
    data: Partial<CreateVPCForm>,
    credentialId: string,
    region: string
  ): Promise<VPC> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.put<ApiResponse<VPC>>(
      `/api/v1/${provider}/network/vpcs/${encodeURIComponent(vpcId)}?${params.toString()}`,
      data
    );
    return response.data.data!;
  },

  deleteVPC: async (
    provider: CloudProvider,
    vpcId: string,
    credentialId: string,
    region: string
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    await api.delete(
      `/api/v1/${provider}/network/vpcs/${encodeURIComponent(vpcId)}?${params.toString()}`
    );
  },

  // Subnet management
  listSubnets: async (
    provider: CloudProvider,
    credentialId: string,
    vpcId: string,
    region: string
  ): Promise<Subnet[]> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      vpc_id: vpcId,
      region,
    });
    
    const response = await api.get<ApiResponse<{ subnets: Subnet[] }>>(
      `/api/v1/${provider}/network/subnets?${params.toString()}`
    );
    return response.data.data?.subnets || [];
  },

  getSubnet: async (
    provider: CloudProvider,
    subnetId: string,
    credentialId: string,
    region: string
  ): Promise<Subnet> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<Subnet>>(
      `/api/v1/${provider}/network/subnets/${encodeURIComponent(subnetId)}?${params.toString()}`
    );
    return response.data.data!;
  },

  createSubnet: async (
    provider: CloudProvider,
    data: CreateSubnetForm
  ): Promise<Subnet> => {
    const response = await api.post<ApiResponse<Subnet>>(
      `/api/v1/${provider}/network/subnets`,
      data
    );
    return response.data.data!;
  },

  updateSubnet: async (
    provider: CloudProvider,
    subnetId: string,
    data: Partial<CreateSubnetForm>,
    credentialId: string,
    region: string
  ): Promise<Subnet> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.put<ApiResponse<Subnet>>(
      `/api/v1/${provider}/network/subnets/${encodeURIComponent(subnetId)}?${params.toString()}`,
      data
    );
    return response.data.data!;
  },

  deleteSubnet: async (
    provider: CloudProvider,
    subnetId: string,
    credentialId: string,
    region: string
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    await api.delete(
      `/api/v1/${provider}/network/subnets/${encodeURIComponent(subnetId)}?${params.toString()}`
    );
  },

  // Security Group management
  listSecurityGroups: async (
    provider: CloudProvider,
    credentialId: string,
    vpcId: string,
    region: string
  ): Promise<SecurityGroup[]> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      vpc_id: vpcId,
      region,
    });
    
    const response = await api.get<ApiResponse<{ security_groups: SecurityGroup[] }>>(
      `/api/v1/${provider}/network/security-groups?${params.toString()}`
    );
    return response.data.data?.security_groups || [];
  },

  getSecurityGroup: async (
    provider: CloudProvider,
    securityGroupId: string,
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<SecurityGroup>>(
      `/api/v1/${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}?${params.toString()}`
    );
    return response.data.data!;
  },

  createSecurityGroup: async (
    provider: CloudProvider,
    data: CreateSecurityGroupForm
  ): Promise<SecurityGroup> => {
    const response = await api.post<ApiResponse<SecurityGroup>>(
      `/api/v1/${provider}/network/security-groups`,
      data
    );
    return response.data.data!;
  },

  updateSecurityGroup: async (
    provider: CloudProvider,
    securityGroupId: string,
    data: Partial<CreateSecurityGroupForm>,
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.put<ApiResponse<SecurityGroup>>(
      `/api/v1/${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}?${params.toString()}`,
      data
    );
    return response.data.data!;
  },

  deleteSecurityGroup: async (
    provider: CloudProvider,
    securityGroupId: string,
    credentialId: string,
    region: string
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    await api.delete(
      `/api/v1/${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}?${params.toString()}`
    );
  },

  // Security Group Rule management
  addSecurityGroupRule: async (
    provider: CloudProvider,
    securityGroupId: string,
    rule: SecurityGroupRule,
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.post<ApiResponse<SecurityGroup>>(
      `/api/v1/${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules?${params.toString()}`,
      rule
    );
    return response.data.data!;
  },

  removeSecurityGroupRule: async (
    provider: CloudProvider,
    securityGroupId: string,
    ruleId: string,
    credentialId: string,
    region: string
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
      rule_id: ruleId,
    });
    
    await api.delete(
      `/api/v1/${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules?${params.toString()}`
    );
  },

  updateSecurityGroupRules: async (
    provider: CloudProvider,
    securityGroupId: string,
    rules: SecurityGroupRule[],
    credentialId: string,
    region: string
  ): Promise<SecurityGroup> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.put<ApiResponse<SecurityGroup>>(
      `/api/v1/${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules?${params.toString()}`,
      { rules }
    );
    return response.data.data!;
  },
};


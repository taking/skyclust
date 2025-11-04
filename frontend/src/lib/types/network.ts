/**
 * Network 관련 타입 정의
 */

export interface VPC {
  id: string;
  name: string;
  state: string;
  is_default: boolean;
  region?: string;
  network_mode?: string;
  routing_mode?: string;
  mtu?: number;
  auto_subnets?: boolean;
  description?: string;
  firewall_rule_count?: number;
  gateway?: GatewayInfo;
  creation_timestamp?: string;
  tags?: Record<string, string>;
}

export interface GatewayInfo {
  type?: string;
  ip_address?: string;
  name?: string;
}

export interface Subnet {
  id: string;
  name: string;
  vpc_id: string;
  cidr_block: string;
  availability_zone: string;
  state: string;
  is_public: boolean;
  region: string;
  description?: string;
  gateway_address?: string;
  private_ip_google_access?: boolean;
  flow_logs?: boolean;
  creation_timestamp?: string;
  tags?: Record<string, string>;
}

export interface SecurityGroup {
  id: string;
  name: string;
  description: string;
  vpc_id: string;
  region: string;
  rules?: SecurityGroupRule[];
  tags?: Record<string, string>;
}

export interface SecurityGroupRule {
  id: string;
  type: 'ingress' | 'egress';
  protocol: string;
  from_port?: number;
  to_port?: number;
  cidr_blocks?: string[];
  source_groups?: string[];
  description?: string;
}

export interface CreateVPCForm {
  credential_id: string;
  name: string;
  description?: string;
  cidr_block?: string;
  region?: string;
  project_id?: string;
  auto_create_subnets?: boolean;
  routing_mode?: string;
  mtu?: number;
  tags?: Record<string, string>;
}

export interface CreateSubnetForm {
  credential_id: string;
  name: string;
  vpc_id: string;
  cidr_block: string;
  availability_zone: string;
  region: string;
  description?: string;
  private_ip_google_access?: boolean;
  flow_logs?: boolean;
  tags?: Record<string, string>;
}

export interface CreateSecurityGroupForm {
  credential_id: string;
  name: string;
  description: string;
  vpc_id: string;
  region: string;
  project_id?: string;
  direction?: string;
  priority?: number;
  action?: string;
  protocol?: string;
  ports?: string[];
  source_ranges?: string[];
  target_tags?: string[];
  tags?: Record<string, string>;
}


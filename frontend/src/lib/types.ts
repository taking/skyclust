// API Response types
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
  code?: string;
}

// User types
export interface User {
  id: string;
  username: string;
  email: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  user: User;
}

// Workspace types
export interface Workspace {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

export interface WorkspaceMember {
  user_id: string;
  workspace_id: string;
  role: 'owner' | 'admin' | 'member';
  joined_at: string;
  user: User;
}

// VM types
export interface VM {
  id: string;
  workspace_id: string;
  name: string;
  provider: string;
  instance_id: string;
  status: string;
  instance_type: string;
  region: string;
  public_ip?: string;
  private_ip?: string;
  created_at: string;
  updated_at: string;
}

// Credential types
export interface Credential {
  id: string;
  workspace_id: string;
  provider: string;
  name?: string; // Credential name for display
  created_at: string;
  updated_at: string;
}

// Provider types
export interface Provider {
  name: string;
  version: string;
}

export interface Instance {
  id: string;
  name: string;
  status: string;
  type: string;
  region: string;
  public_ip?: string;
  private_ip?: string;
  created_at: string;
  tags?: Record<string, string>;
}

export interface Region {
  name: string;
  display_name: string;
}

// Form types
export interface LoginForm {
  email: string;
  password: string;
}

export interface RegisterForm {
  email: string;
  password: string;
  name: string;
}

export interface CreateWorkspaceForm {
  name: string;
  description: string;
}

export interface CreateVMForm {
  name: string;
  provider: string;
  instance_type: string;
  region: string;
  image_id: string;
}

export interface CreateCredentialForm {
  name?: string;
  provider: string;
  credentials: Record<string, unknown>;
}

// Kubernetes types
export type CloudProvider = 'aws' | 'gcp' | 'azure' | 'ncp';

export interface KubernetesCluster {
  id: string;
  name: string;
  version: string;
  status: string;
  region: string;
  zone?: string;
  endpoint?: string;
  project_id?: string;
  created_at?: string;
  updated_at?: string;
  tags?: Record<string, string>;
  network_config?: NetworkConfigInfo;
  node_pool_info?: NodePoolSummaryInfo;
  security_config?: SecurityConfigInfo;
}

export interface NetworkConfigInfo {
  vpc_id?: string;
  subnet_id?: string;
  pod_cidr?: string;
  service_cidr?: string;
  private_nodes?: boolean;
  private_endpoint?: boolean;
}

export interface NodePoolSummaryInfo {
  total_node_pools: number;
  total_nodes: number;
  min_nodes: number;
  max_nodes: number;
}

export interface SecurityConfigInfo {
  workload_identity?: boolean;
  binary_authorization?: boolean;
  network_policy?: boolean;
  pod_security_policy?: boolean;
}

export interface NodePool {
  id: string;
  name: string;
  cluster_name: string;
  version?: string;
  status: string;
  node_count: number;
  min_nodes: number;
  max_nodes: number;
  instance_type: string;
  disk_size_gb?: number;
  disk_type?: string;
  region: string;
  zone?: string;
  auto_scaling?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface NodeGroup {
  id: string;
  name: string;
  cluster_name: string;
  status: string;
  node_count: number;
  min_size: number;
  max_size: number;
  instance_type: string;
  disk_size_gb?: number;
  region: string;
  created_at?: string;
  updated_at?: string;
}

export interface Node {
  id: string;
  name: string;
  cluster_name: string;
  node_pool_name?: string;
  node_group_name?: string;
  status: string;
  instance_type: string;
  zone?: string;
  private_ip?: string;
  public_ip?: string;
  created_at?: string;
}

export interface CreateClusterForm {
  credential_id: string;
  name: string;
  version: string;
  region: string;
  zone?: string;
  subnet_ids: string[];
  vpc_id?: string;
  role_arn?: string;
  tags?: Record<string, string>;
  access_config?: {
    authentication_mode?: string;
    bootstrap_cluster_creator_admin_permissions?: boolean;
  };
}

export interface CreateNodePoolForm {
  credential_id: string;
  name: string;
  cluster_name: string;
  version?: string;
  region: string;
  zone?: string;
  instance_type: string;
  disk_size_gb?: number;
  disk_type?: string;
  min_nodes: number;
  max_nodes: number;
  node_count: number;
  auto_scaling?: boolean;
  tags?: Record<string, string>;
}

export interface CreateNodeGroupForm {
  credential_id: string;
  name: string;
  cluster_name: string;
  instance_type: string;
  disk_size_gb?: number;
  min_size: number;
  max_size: number;
  desired_size: number;
  region: string;
  tags?: Record<string, string>;
}

// Network types
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

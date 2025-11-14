/**
 * Kubernetes 관련 타입 정의
 */

export type CloudProvider = 'aws' | 'gcp' | 'azure' | 'ncp';

// Base cluster interface with common fields
export interface BaseCluster {
  id: string;
  name: string;
  version: string;
  status: string;
  region: string;
  zone?: string;
  endpoint?: string;
  resource_group?: string; // Azure-specific, but included in base for listing compatibility
  project_id?: string; // GCP-specific, but included in base for listing compatibility
  created_at?: string;
  updated_at?: string;
  tags?: Record<string, string>;
  network_config?: NetworkConfigInfo;
  node_pool_info?: NodePoolSummaryInfo;
  security_config?: SecurityConfigInfo;
}

// AWS EKS Cluster
export interface AWSResourcesVPCConfig {
  subnet_ids: string[];
  security_group_ids: string[];
  cluster_security_group_id?: string;
  vpc_id: string;
  endpoint_public_access: boolean;
  endpoint_private_access: boolean;
  public_access_cidrs: string[];
}

export interface AWSElasticLoadBalancing {
  enabled: boolean;
}

export interface AWSKubernetesNetworkConfig {
  service_ipv4_cidr?: string;
  service_ipv6_cidr?: string;
  ip_family?: string; // "ipv4" or "ipv6"
  elastic_load_balancing?: AWSElasticLoadBalancing;
}

export interface AWSAccessConfig {
  bootstrap_cluster_creator_admin_permissions?: boolean;
  authentication_mode?: string; // "API", "CONFIG_MAP", "API_AND_CONFIG_MAP"
}

export interface AWSUpgradePolicy {
  support_type?: string; // "EXTENDED", "STANDARD"
}

export interface AWSCluster extends BaseCluster {
  resources_vpc_config?: AWSResourcesVPCConfig;
  kubernetes_network_config?: AWSKubernetesNetworkConfig;
  access_config?: AWSAccessConfig;
  upgrade_policy?: AWSUpgradePolicy;
  role_arn?: string;
  platform_version?: string;
  deletion_protection?: boolean;
}

// GCP GKE Cluster
export interface GCPNetworkConfig {
  network?: string; // VPC network name
  subnetwork?: string; // Subnet name
  pod_cidr?: string; // Pod CIDR range
  service_cidr?: string; // Service CIDR range
  private_nodes?: boolean;
  private_endpoint?: boolean;
}

export interface GCPSecurityConfig {
  workload_identity?: boolean;
  binary_authorization?: boolean;
  network_policy?: boolean;
  pod_security_policy?: boolean;
}

export interface GCPWorkloadIdentityConfig {
  workload_pool?: string;
}

export interface GCPPrivateClusterConfig {
  enable_private_nodes: boolean;
  enable_private_endpoint: boolean;
  master_ipv4_cidr?: string;
}

export interface GCPMasterAuthorizedNetworksConfig {
  enabled: boolean;
  cidr_blocks?: string[];
}

export interface GCPCluster extends BaseCluster {
  project_id?: string;
  network_config?: GCPNetworkConfig;
  security_config?: GCPSecurityConfig;
  workload_identity_config?: GCPWorkloadIdentityConfig;
  private_cluster_config?: GCPPrivateClusterConfig;
  master_authorized_networks_config?: GCPMasterAuthorizedNetworksConfig;
}

// Azure AKS Cluster
export interface AzureNetworkProfile {
  network_plugin?: string; // "azure" or "kubenet"
  network_policy?: string; // "azure" or "calico"
  pod_cidr?: string;
  service_cidr?: string;
  dns_service_ip?: string;
  docker_bridge_cidr?: string;
  load_balancer_sku?: string;
  network_mode?: string;
}

export interface AzureServicePrincipal {
  client_id?: string;
}

export interface AzureCluster extends BaseCluster {
  resource_group?: string;
  network_profile?: AzureNetworkProfile;
  service_principal?: AzureServicePrincipal;
  addon_profiles?: Record<string, unknown>;
  enable_rbac?: boolean;
  enable_pod_security_policy?: boolean;
  api_server_authorized_ip_ranges?: string[];
}

// Union type for provider-specific clusters
export type ProviderCluster = AWSCluster | GCPCluster | AzureCluster;

// Backward compatibility: KubernetesCluster is an alias for BaseCluster
export type KubernetesCluster = BaseCluster;

// Type guards
export function isAWSCluster(cluster: ProviderCluster | BaseCluster): cluster is AWSCluster {
  return 'resources_vpc_config' in cluster || 'role_arn' in cluster;
}

export function isGCPCluster(cluster: ProviderCluster | BaseCluster): cluster is GCPCluster {
  return 'project_id' in cluster && !('resource_group' in cluster);
}

export function isAzureCluster(cluster: ProviderCluster | BaseCluster): cluster is AzureCluster {
  return 'resource_group' in cluster || 'network_profile' in cluster;
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
  node_count?: number;
  min_size: number;
  max_size: number;
  desired_size?: number;
  instance_type?: string;
  instance_types?: string[];
  disk_size_gb?: number;
  disk_size?: number;
  region: string;
  version?: string;
  capacity_type?: string;
  created_at?: string;
  updated_at?: string;
}

// AWS EKS Node Group detailed information
export interface AWSRemoteAccessConfig {
  ec2_ssh_key?: string;
  source_security_groups?: string[];
}

export interface AWSAutoScalingGroup {
  name: string;
}

export interface AWSNodeGroupResources {
  auto_scaling_groups?: AWSAutoScalingGroup[];
  remote_access_security_group?: string;
}

export interface AWSNodeGroupHealthIssue {
  code: string;
  message: string;
  resource_ids?: string[];
}

export interface AWSNodeGroupHealth {
  issues?: AWSNodeGroupHealthIssue[];
}

export interface AWSLaunchTemplateSpec {
  id?: string;
  name?: string;
  version?: string;
}

export interface AWSUpdateConfig {
  max_unavailable?: number;
  max_unavailable_percentage?: number;
}

export interface AWSNodeGroup extends NodeGroup {
  node_role_arn?: string;
  ami_type?: string;
  release_version?: string;
  subnets?: string[];
  remote_access_config?: AWSRemoteAccessConfig;
  resources?: AWSNodeGroupResources;
  health?: AWSNodeGroupHealth;
  launch_template?: AWSLaunchTemplateSpec;
  update_config?: AWSUpdateConfig;
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
  // Azure AKS specific fields
  location?: string; // Azure uses 'location' instead of 'region'
  resource_group?: string;
  network?: {
    virtual_network_id: string;
    subnet_id: string;
    network_plugin?: string; // "azure" or "kubenet"
    network_policy?: string; // "azure" or "calico"
    pod_cidr?: string;
    service_cidr?: string;
    dns_service_ip?: string;
    docker_bridge_cidr?: string;
  };
  node_pool?: {
    name: string;
    vm_size: string; // e.g., "Standard_D2s_v3"
    os_disk_size_gb?: number;
    os_disk_type?: string; // "Managed" or "Ephemeral"
    os_type?: string; // "Linux" or "Windows"
    os_sku?: string; // "Ubuntu" or "CBLMariner"
    node_count: number;
    min_count?: number;
    max_count?: number;
    enable_auto_scaling?: boolean;
    max_pods?: number;
    vnet_subnet_id?: string;
    availability_zones?: string[];
    labels?: Record<string, string>;
    taints?: string[];
    mode?: string; // "System" or "User"
  };
  security?: {
    enable_rbac?: boolean;
    enable_pod_security_policy?: boolean;
    enable_private_cluster?: boolean;
    api_server_authorized_ip_ranges?: string[];
    enable_azure_policy?: boolean;
    enable_workload_identity?: boolean;
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
  instance_types: string[]; // Changed from instance_type to instance_types (array)
  ami_type?: string;
  disk_size?: number; // Changed from disk_size_gb to disk_size
  min_size: number;
  max_size: number;
  desired_size: number;
  region: string;
  subnet_ids?: string[];
  capacity_type?: string; // ON_DEMAND, SPOT
  tags?: Record<string, string>;
}


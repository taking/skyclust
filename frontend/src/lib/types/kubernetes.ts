/**
 * Kubernetes 관련 타입 정의
 */

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
  instance_type: string;
  disk_size_gb?: number;
  min_size: number;
  max_size: number;
  desired_size: number;
  region: string;
  tags?: Record<string, string>;
}


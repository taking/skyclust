/**
 * GCP GKE Cluster Creation Payload Builder
 * GCP GKE 클러스터 생성 API 요청 payload 생성
 */

import type { CreateClusterForm } from '@/lib/types';

export function buildGCPPayload(data: CreateClusterForm): Record<string, unknown> {
  const payload: Record<string, unknown> = {};

  // 1. cluster_mode
  if (data.cluster_mode) {
    payload.cluster_mode = {
      remove_default_node_pool: data.cluster_mode.remove_default_node_pool || false,
      type: data.cluster_mode.type || 'standard',
    };
  }

  // 2. credential_id
  if (data.credential_id) {
    payload.credential_id = data.credential_id;
  }

  // 3. name
  if (data.name) {
    payload.name = data.name;
  }

  // 4. network
  const networkObj: Record<string, unknown> = {};
  if (data.network?.pod_cidr) networkObj.pod_cidr = data.network.pod_cidr;
  if (data.network?.service_cidr) networkObj.service_cidr = data.network.service_cidr;
  
  /**
   * Extract network name from VPC ID
   * Supports two formats:
   * 1. Full format: projects/{project}/global/networks/{network_name}
   * 2. Simple format: {network_name}
   */
  const extractNetworkNameFromVPCID = (vpcID: string): string => {
    if (!vpcID) return vpcID;
    
    // Check if it's a full format
    const parts = vpcID.split('/');
    if (parts.length >= 4 && parts[parts.length - 2] === 'networks') {
      return parts[parts.length - 1];
    }
    
    // If it's a simple format, return as is
    return vpcID;
  };

  /**
   * Extract subnetwork name from Subnet ID
   * Supports two formats:
   * 1. Full format: projects/{project}/regions/{region}/subnetworks/{subnet_name}
   * 2. Simple format: {subnet_name}
   */
  const extractSubnetworkNameFromSubnetID = (subnetID: string): string => {
    if (!subnetID) return subnetID;
    
    // Handle full GCP subnet path: projects/{project}/regions/{region}/subnetworks/{subnet_name}
    if (subnetID.includes('/subnetworks/')) {
      const parts = subnetID.split('/subnetworks/');
      if (parts.length === 2) {
        return parts[1];
      }
    }
    
    // Handle simple subnet name
    return subnetID;
  };
  
  // subnet_id: network.subnet_id 우선, 없으면 subnet_ids[0]
  // Extract subnetwork name from full resource path format
  const subnetId = data.network?.subnet_id || (data.subnet_ids && data.subnet_ids.length > 0 ? data.subnet_ids[0] : undefined);
  if (subnetId) {
    networkObj.subnet_id = extractSubnetworkNameFromSubnetID(subnetId);
  }
  
  // vpc_id: network.virtual_network_id 우선, 없으면 vpc_id
  // Extract network name from full resource path format
  const vpcId = data.network?.virtual_network_id || data.vpc_id;
  if (vpcId) {
    networkObj.vpc_id = extractNetworkNameFromVPCID(vpcId);
  }
  
  if (data.network?.master_authorized_networks && data.network.master_authorized_networks.length > 0) {
    networkObj.master_authorized_networks = data.network.master_authorized_networks;
  }
  if (data.network?.private_endpoint !== undefined) {
    networkObj.private_endpoint = data.network.private_endpoint;
  }
  if (data.network?.private_nodes !== undefined) {
    networkObj.private_nodes = data.network.private_nodes;
  }
  
  if (Object.keys(networkObj).length > 0) {
    payload.network = networkObj;
  }

  // 5. node_pool
  if (data.node_pool) {
    const nodePoolObj: Record<string, unknown> = {};
    
    if (data.node_pool.auto_scaling) {
      nodePoolObj.auto_scaling = {
        enabled: data.node_pool.auto_scaling.enabled || false,
        ...(data.node_pool.auto_scaling.min_node_count !== undefined && {
          min_node_count: data.node_pool.auto_scaling.min_node_count,
        }),
        ...(data.node_pool.auto_scaling.max_node_count !== undefined && {
          max_node_count: data.node_pool.auto_scaling.max_node_count,
        }),
      };
    }
    if (data.node_pool.disk_size_gb !== undefined) nodePoolObj.disk_size_gb = data.node_pool.disk_size_gb;
    if (data.node_pool.disk_type) nodePoolObj.disk_type = data.node_pool.disk_type;
    if (data.node_pool.labels && Object.keys(data.node_pool.labels).length > 0) {
      nodePoolObj.labels = data.node_pool.labels;
    }
    if (data.node_pool.machine_type) nodePoolObj.machine_type = data.node_pool.machine_type;
    if (data.node_pool.name) nodePoolObj.name = data.node_pool.name;
    if (data.node_pool.node_count !== undefined) nodePoolObj.node_count = data.node_pool.node_count;
    if (data.node_pool.preemptible !== undefined) nodePoolObj.preemptible = data.node_pool.preemptible;
    if (data.node_pool.spot !== undefined) nodePoolObj.spot = data.node_pool.spot;
    
    if (Object.keys(nodePoolObj).length > 0) {
      payload.node_pool = nodePoolObj;
    }
  }

  // 6. project_id (있을 때만 포함)
  if (data.project_id && data.project_id.trim() !== '') {
    payload.project_id = data.project_id;
  }

  // 7. region
  if (data.region) {
    payload.region = data.region;
  }

  // 8. zone
  if (data.zone) {
    payload.zone = data.zone;
  }

  // 9. version
  if (data.version) {
    payload.version = data.version;
  }

  // 10. security
  if (data.security) {
    const securityObj: Record<string, unknown> = {};
    if (data.security.binary_authorization !== undefined) {
      securityObj.binary_authorization = data.security.binary_authorization;
    }
    if (data.security.network_policy !== undefined) {
      securityObj.network_policy = data.security.network_policy;
    }
    if (data.security.pod_security_policy !== undefined) {
      securityObj.pod_security_policy = data.security.pod_security_policy;
    }
    if (data.security.enable_workload_identity !== undefined) {
      securityObj.workload_identity = data.security.enable_workload_identity;
    }
    
    if (Object.keys(securityObj).length > 0) {
      payload.security = securityObj;
    }
  }

  // 11. tags
  if (data.tags && Object.keys(data.tags).length > 0) {
    payload.tags = data.tags;
  }

  return payload;
}


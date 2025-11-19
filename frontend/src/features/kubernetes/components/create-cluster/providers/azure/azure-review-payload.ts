/**
 * Azure AKS Cluster Creation Payload Builder
 * Azure AKS 클러스터 생성 API 요청 payload 생성
 */

import type { CreateClusterForm } from '@/lib/types';

export function buildAzurePayload(data: CreateClusterForm): Record<string, unknown> {
  const payload: Record<string, unknown> = {};

  if (data.credential_id) payload.credential_id = data.credential_id;
  if (data.name) payload.name = data.name;
  if (data.version) payload.version = data.version;
  if (data.location) payload.location = data.location;
  if (data.region && !data.location) payload.region = data.region;
  if (data.resource_group) payload.resource_group = data.resource_group;
  if (data.zone) payload.zone = data.zone;

  // Network configuration
  if (data.network) {
    const networkObj: Record<string, unknown> = {};
    if (data.network.virtual_network_id) networkObj.virtual_network_id = data.network.virtual_network_id;
    if (data.network.subnet_id) networkObj.subnet_id = data.network.subnet_id;
    if (data.network.network_plugin) networkObj.network_plugin = data.network.network_plugin;
    if (data.network.network_policy) networkObj.network_policy = data.network.network_policy;
    if (data.network.pod_cidr) networkObj.pod_cidr = data.network.pod_cidr;
    if (data.network.service_cidr) networkObj.service_cidr = data.network.service_cidr;
    if (data.network.dns_service_ip) networkObj.dns_service_ip = data.network.dns_service_ip;
    if (data.network.docker_bridge_cidr) networkObj.docker_bridge_cidr = data.network.docker_bridge_cidr;
    
    if (Object.keys(networkObj).length > 0) {
      payload.network = networkObj;
    }
  }

  // Node Pool configuration
  if (data.node_pool) {
    const nodePoolObj: Record<string, unknown> = {};
    if (data.node_pool.name) nodePoolObj.name = data.node_pool.name;
    if (data.node_pool.vm_size) nodePoolObj.vm_size = data.node_pool.vm_size;
    if (data.node_pool.os_disk_size_gb !== undefined) nodePoolObj.os_disk_size_gb = data.node_pool.os_disk_size_gb;
    if (data.node_pool.os_disk_type) nodePoolObj.os_disk_type = data.node_pool.os_disk_type;
    if (data.node_pool.os_type) nodePoolObj.os_type = data.node_pool.os_type;
    if (data.node_pool.os_sku) nodePoolObj.os_sku = data.node_pool.os_sku;
    if (data.node_pool.node_count !== undefined) nodePoolObj.node_count = data.node_pool.node_count;
    if (data.node_pool.min_count !== undefined) nodePoolObj.min_count = data.node_pool.min_count;
    if (data.node_pool.max_count !== undefined) nodePoolObj.max_count = data.node_pool.max_count;
    if (data.node_pool.enable_auto_scaling !== undefined) nodePoolObj.enable_auto_scaling = data.node_pool.enable_auto_scaling;
    if (data.node_pool.max_pods !== undefined) nodePoolObj.max_pods = data.node_pool.max_pods;
    if (data.node_pool.vnet_subnet_id) nodePoolObj.vnet_subnet_id = data.node_pool.vnet_subnet_id;
    if (data.node_pool.availability_zones && data.node_pool.availability_zones.length > 0) {
      nodePoolObj.availability_zones = data.node_pool.availability_zones;
    }
    if (data.node_pool.labels && Object.keys(data.node_pool.labels).length > 0) {
      nodePoolObj.labels = data.node_pool.labels;
    }
    if (data.node_pool.taints && data.node_pool.taints.length > 0) {
      nodePoolObj.taints = data.node_pool.taints;
    }
    if (data.node_pool.mode) nodePoolObj.mode = data.node_pool.mode;
    
    if (Object.keys(nodePoolObj).length > 0) {
      payload.node_pool = nodePoolObj;
    }
  }

  // Security configuration
  if (data.security) {
    const securityObj: Record<string, unknown> = {};
    if (data.security.enable_rbac !== undefined) securityObj.enable_rbac = data.security.enable_rbac;
    if (data.security.enable_pod_security_policy !== undefined) {
      securityObj.enable_pod_security_policy = data.security.enable_pod_security_policy;
    }
    if (data.security.enable_private_cluster !== undefined) {
      securityObj.enable_private_cluster = data.security.enable_private_cluster;
    }
    if (data.security.api_server_authorized_ip_ranges && data.security.api_server_authorized_ip_ranges.length > 0) {
      securityObj.api_server_authorized_ip_ranges = data.security.api_server_authorized_ip_ranges;
    }
    if (data.security.enable_azure_policy !== undefined) {
      securityObj.enable_azure_policy = data.security.enable_azure_policy;
    }
    if (data.security.enable_workload_identity !== undefined) {
      securityObj.enable_workload_identity = data.security.enable_workload_identity;
    }
    
    if (Object.keys(securityObj).length > 0) {
      payload.security = securityObj;
    }
  }

  // Tags (Azure는 map[string]*string 형식)
  if (data.tags && Object.keys(data.tags).length > 0) {
    const tagsObj: Record<string, string> = {};
    Object.entries(data.tags).forEach(([key, value]) => {
      if (value) tagsObj[key] = value;
    });
    if (Object.keys(tagsObj).length > 0) {
      payload.tags = tagsObj;
    }
  }

  return payload;
}


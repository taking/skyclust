import api from '@/lib/api';
import { 
  ApiResponse, 
  KubernetesCluster, 
  NodePool,
  NodeGroup,
  Node,
  CreateClusterForm,
  CreateNodePoolForm,
  CreateNodeGroupForm,
  CloudProvider
} from '@/lib/types';

export const kubernetesService = {
  // Cluster management
  listClusters: async (
    provider: CloudProvider,
    credentialId: string,
    region?: string
  ): Promise<KubernetesCluster[]> => {
    const params = new URLSearchParams({ credential_id: credentialId });
    if (region) params.append('region', region);
    
    const response = await api.get<ApiResponse<{ clusters: KubernetesCluster[] }>>(
      `/api/v1/${provider}/kubernetes/clusters?${params.toString()}`
    );
    return response.data.data?.clusters || [];
  },

  getCluster: async (
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<KubernetesCluster> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<KubernetesCluster>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}?${params.toString()}`
    );
    return response.data.data!;
  },

  createCluster: async (
    provider: CloudProvider,
    data: CreateClusterForm
  ): Promise<KubernetesCluster> => {
    const response = await api.post<ApiResponse<KubernetesCluster>>(
      `/api/v1/${provider}/kubernetes/clusters`,
      data
    );
    return response.data.data!;
  },

  deleteCluster: async (
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    await api.delete(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}?${params.toString()}`
    );
  },

  getKubeconfig: async (
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<string> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<{ kubeconfig: string }>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/kubeconfig?${params.toString()}`
    );
    return response.data.data?.kubeconfig || '';
  },

  upgradeCluster: async (
    provider: CloudProvider,
    clusterName: string,
    version: string,
    credentialId: string,
    region: string
  ): Promise<{ message: string; upgrade_id?: string }> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.post<ApiResponse<{ message: string; upgrade_id?: string }>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/upgrade?${params.toString()}`,
      { version }
    );
    return response.data.data!;
  },

  getUpgradeStatus: async (
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<{ status: string; current_version?: string; target_version?: string; progress?: number; error?: string }> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<{ status: string; current_version?: string; target_version?: string; progress?: number; error?: string }>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/upgrade/status?${params.toString()}`
    );
    return response.data.data!;
  },

  // Update cluster tags
  updateClusterTags: async (
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string,
    tags: Record<string, string>
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    await api.put(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/tags?${params.toString()}`,
      { tags }
    );
  },

  // Bulk update tags for multiple clusters
  bulkUpdateTags: async (
    provider: CloudProvider,
    clusterUpdates: Array<{
      clusterName: string;
      credentialId: string;
      region: string;
      tags: Record<string, string>;
    }>
  ): Promise<void> => {
    // Note: This would require a backend bulk API endpoint
    // For now, we'll call individual updates
    await Promise.all(
      clusterUpdates.map(({ clusterName, credentialId, region, tags }) =>
        kubernetesService.updateClusterTags(provider, clusterName, credentialId, region, tags)
      )
    );
  },

  // Node Pool management (GKE, AKS, NKS)
  listNodePools: async (
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<NodePool[]> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<{ node_pools: NodePool[] }>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools?${params.toString()}`
    );
    return response.data.data?.node_pools || [];
  },

  getNodePool: async (
    provider: CloudProvider,
    clusterName: string,
    nodePoolName: string,
    credentialId: string,
    region: string
  ): Promise<NodePool> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<NodePool>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools/${nodePoolName}?${params.toString()}`
    );
    return response.data.data!;
  },

  createNodePool: async (
    provider: CloudProvider,
    clusterName: string,
    data: CreateNodePoolForm
  ): Promise<NodePool> => {
    const response = await api.post<ApiResponse<NodePool>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools`,
      data
    );
    return response.data.data!;
  },

  deleteNodePool: async (
    provider: CloudProvider,
    clusterName: string,
    nodePoolName: string,
    credentialId: string,
    region: string
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    await api.delete(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools/${nodePoolName}?${params.toString()}`
    );
  },

  scaleNodePool: async (
    provider: CloudProvider,
    clusterName: string,
    nodePoolName: string,
    nodeCount: number,
    credentialId: string,
    region: string
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    await api.put(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools/${nodePoolName}/scale?${params.toString()}`,
      { node_count: nodeCount }
    );
  },

  // Node Group management (EKS specific)
  listNodeGroups: async (
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<NodeGroup[]> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<{ node_groups: NodeGroup[] }>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/node-groups?${params.toString()}`
    );
    return response.data.data?.node_groups || [];
  },

  getNodeGroup: async (
    provider: CloudProvider,
    clusterName: string,
    nodeGroupName: string,
    credentialId: string,
    region: string
  ): Promise<NodeGroup> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<NodeGroup>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/node-groups/${nodeGroupName}?${params.toString()}`
    );
    return response.data.data!;
  },

  createNodeGroup: async (
    provider: CloudProvider,
    clusterName: string,
    data: CreateNodeGroupForm
  ): Promise<NodeGroup> => {
    const response = await api.post<ApiResponse<NodeGroup>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/node-groups`,
      data
    );
    return response.data.data!;
  },

  deleteNodeGroup: async (
    provider: CloudProvider,
    clusterName: string,
    nodeGroupName: string,
    credentialId: string,
    region: string
  ): Promise<void> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    await api.delete(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/node-groups/${nodeGroupName}?${params.toString()}`
    );
  },

  // Node management
  listNodes: async (
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<Node[]> => {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const response = await api.get<ApiResponse<{ nodes: Node[] }>>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodes?${params.toString()}`
    );
    return response.data.data?.nodes || [];
  },
};


/**
 * Kubernetes Service
 * Kubernetes 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import type {
  KubernetesCluster,
  NodePool,
  NodeGroup,
  Node,
  CreateClusterForm,
  CreateNodePoolForm,
  CreateNodeGroupForm,
  CloudProvider,
} from '@/lib/types';

class KubernetesService extends BaseService {
  // Cluster management
  async listClusters(
    provider: CloudProvider,
    credentialId: string,
    region?: string
  ): Promise<KubernetesCluster[]> {
    const params = new URLSearchParams({ credential_id: credentialId });
    if (region) params.append('region', region);
    
    const data = await this.get<{ clusters: KubernetesCluster[] }>(
      `/api/v1/${provider}/kubernetes/clusters?${params.toString()}`
    );
    return data.clusters || [];
  }

  async getCluster(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<KubernetesCluster> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.get<KubernetesCluster>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}?${params.toString()}`
    );
  }

  async createCluster(
    provider: CloudProvider,
    data: CreateClusterForm
  ): Promise<KubernetesCluster> {
    return this.post<KubernetesCluster>(
      `/api/v1/${provider}/kubernetes/clusters`,
      data
    );
  }

  async deleteCluster(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.delete<void>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}?${params.toString()}`
    );
  }

  async getKubeconfig(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<string> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const data = await this.get<{ kubeconfig: string }>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/kubeconfig?${params.toString()}`
    );
    return data.kubeconfig || '';
  }

  async upgradeCluster(
    provider: CloudProvider,
    clusterName: string,
    version: string,
    credentialId: string,
    region: string
  ): Promise<{ message: string; upgrade_id?: string }> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.post<{ message: string; upgrade_id?: string }>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/upgrade?${params.toString()}`,
      { version }
    );
  }

  async getUpgradeStatus(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<{ status: string; current_version?: string; target_version?: string; progress?: number; error?: string }> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.get<{ status: string; current_version?: string; target_version?: string; progress?: number; error?: string }>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/upgrade/status?${params.toString()}`
    );
  }

  // Update cluster tags
  async updateClusterTags(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string,
    tags: Record<string, string>
  ): Promise<void> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.put<void>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/tags?${params.toString()}`,
      { tags }
    );
  }

  // Bulk update tags for multiple clusters
  async bulkUpdateTags(
    provider: CloudProvider,
    clusterUpdates: Array<{
      clusterName: string;
      credentialId: string;
      region: string;
      tags: Record<string, string>;
    }>
  ): Promise<void> {
    // Note: This would require a backend bulk API endpoint
    // For now, we'll call individual updates
    await Promise.all(
      clusterUpdates.map(({ clusterName, credentialId, region, tags }) =>
        this.updateClusterTags(provider, clusterName, credentialId, region, tags)
      )
    );
  }

  // Node Pool management (GKE, AKS, NKS)
  async listNodePools(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<NodePool[]> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const data = await this.get<{ node_pools: NodePool[] }>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools?${params.toString()}`
    );
    return data.node_pools || [];
  }

  async getNodePool(
    provider: CloudProvider,
    clusterName: string,
    nodePoolName: string,
    credentialId: string,
    region: string
  ): Promise<NodePool> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.get<NodePool>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools/${nodePoolName}?${params.toString()}`
    );
  }

  async createNodePool(
    provider: CloudProvider,
    clusterName: string,
    data: CreateNodePoolForm
  ): Promise<NodePool> {
    return this.post<NodePool>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools`,
      data
    );
  }

  async deleteNodePool(
    provider: CloudProvider,
    clusterName: string,
    nodePoolName: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.delete<void>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools/${nodePoolName}?${params.toString()}`
    );
  }

  async scaleNodePool(
    provider: CloudProvider,
    clusterName: string,
    nodePoolName: string,
    nodeCount: number,
    credentialId: string,
    region: string
  ): Promise<void> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.put<void>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodepools/${nodePoolName}/scale?${params.toString()}`,
      { node_count: nodeCount }
    );
  }

  // Node Group management (EKS specific)
  async listNodeGroups(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<NodeGroup[]> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const data = await this.get<{ node_groups: NodeGroup[] }>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/node-groups?${params.toString()}`
    );
    return data.node_groups || [];
  }

  async getNodeGroup(
    provider: CloudProvider,
    clusterName: string,
    nodeGroupName: string,
    credentialId: string,
    region: string
  ): Promise<NodeGroup> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.get<NodeGroup>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/node-groups/${nodeGroupName}?${params.toString()}`
    );
  }

  async createNodeGroup(
    provider: CloudProvider,
    clusterName: string,
    data: CreateNodeGroupForm
  ): Promise<NodeGroup> {
    return this.post<NodeGroup>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/node-groups`,
      data
    );
  }

  async deleteNodeGroup(
    provider: CloudProvider,
    clusterName: string,
    nodeGroupName: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    return this.delete<void>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/node-groups/${nodeGroupName}?${params.toString()}`
    );
  }

  // Node management
  async listNodes(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<Node[]> {
    const params = new URLSearchParams({
      credential_id: credentialId,
      region,
    });
    
    const data = await this.get<{ nodes: Node[] }>(
      `/api/v1/${provider}/kubernetes/clusters/${clusterName}/nodes?${params.toString()}`
    );
    return data.nodes || [];
  }
}

export const kubernetesService = new KubernetesService();

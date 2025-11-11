/**
 * Kubernetes Service
 * Kubernetes 관련 API 호출
 */

import { BaseService } from '@/lib/api';
import { API_ENDPOINTS } from '@/lib/api';
import api from '@/lib/api';
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
    const data = await this.get<{ clusters: KubernetesCluster[] }>(
      API_ENDPOINTS.kubernetes.clusters.list(provider, credentialId, region)
    );
    return data.clusters || [];
  }

  async getCluster(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<KubernetesCluster> {
    return this.get<KubernetesCluster>(
      API_ENDPOINTS.kubernetes.clusters.detail(provider, clusterName, credentialId, region)
    );
  }

  async createCluster(
    provider: CloudProvider,
    data: CreateClusterForm
  ): Promise<KubernetesCluster> {
    return this.post<KubernetesCluster>(
      API_ENDPOINTS.kubernetes.clusters.create(provider),
      data
    );
  }

  async deleteCluster(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<void> {
    return this.delete<void>(
      API_ENDPOINTS.kubernetes.clusters.delete(provider, clusterName, credentialId, region)
    );
  }

  async getKubeconfig(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<string> {
    // Kubeconfig는 YAML 텍스트로 직접 반환되므로 BaseService의 JSON 파싱을 우회
    const url = this.buildApiUrl(
      API_ENDPOINTS.kubernetes.clusters.kubeconfig(provider, clusterName, credentialId, region)
    );
    
    // 직접 api를 사용하여 텍스트 응답 처리
    const response = await api.get<string>(url, {
      responseType: 'text',
    });
    
    return response.data || '';
  }

  async upgradeCluster(
    provider: CloudProvider,
    clusterName: string,
    version: string,
    credentialId: string,
    region: string
  ): Promise<{ message: string; upgrade_id?: string }> {
    return this.post<{ message: string; upgrade_id?: string }>(
      API_ENDPOINTS.kubernetes.clusters.upgrade(provider, clusterName, credentialId, region),
      { version }
    );
  }

  async getUpgradeStatus(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<{ status: string; current_version?: string; target_version?: string; progress?: number; error?: string }> {
    return this.get<{ status: string; current_version?: string; target_version?: string; progress?: number; error?: string }>(
      API_ENDPOINTS.kubernetes.clusters.upgradeStatus(provider, clusterName, credentialId, region)
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
    return this.put<void>(
      API_ENDPOINTS.kubernetes.clusters.tags(provider, clusterName, credentialId, region),
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
    const data = await this.get<{ node_pools: NodePool[] }>(
      API_ENDPOINTS.kubernetes.nodePools.list(provider, clusterName, credentialId, region)
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
    return this.get<NodePool>(
      API_ENDPOINTS.kubernetes.nodePools.detail(provider, clusterName, nodePoolName, credentialId, region)
    );
  }

  async createNodePool(
    provider: CloudProvider,
    clusterName: string,
    data: CreateNodePoolForm
  ): Promise<NodePool> {
    return this.post<NodePool>(
      API_ENDPOINTS.kubernetes.nodePools.create(provider, clusterName),
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
    return this.delete<void>(
      API_ENDPOINTS.kubernetes.nodePools.delete(provider, clusterName, nodePoolName, credentialId, region)
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
    return this.put<void>(
      API_ENDPOINTS.kubernetes.nodePools.scale(provider, clusterName, nodePoolName, credentialId, region),
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
    const data = await this.get<{ node_groups: NodeGroup[] }>(
      API_ENDPOINTS.kubernetes.nodeGroups.list(provider, clusterName, credentialId, region)
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
    return this.get<NodeGroup>(
      API_ENDPOINTS.kubernetes.nodeGroups.detail(provider, clusterName, nodeGroupName, credentialId, region)
    );
  }

  async createNodeGroup(
    provider: CloudProvider,
    clusterName: string,
    data: CreateNodeGroupForm
  ): Promise<NodeGroup> {
    return this.post<NodeGroup>(
      API_ENDPOINTS.kubernetes.nodeGroups.create(provider, clusterName),
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
    return this.delete<void>(
      API_ENDPOINTS.kubernetes.nodeGroups.delete(provider, clusterName, nodeGroupName, credentialId, region)
    );
  }

  // Node management
  async listNodes(
    provider: CloudProvider,
    clusterName: string,
    credentialId: string,
    region: string
  ): Promise<Node[]> {
    const data = await this.get<{ nodes: Node[] }>(
      API_ENDPOINTS.kubernetes.clusters.nodes(provider, clusterName, credentialId, region)
    );
    return data.nodes || [];
  }

  // Metadata endpoints (AWS only)
  async getEKSVersions(
    provider: CloudProvider,
    credentialId: string,
    region: string
  ): Promise<string[]> {
    const data = await this.get<{ versions: string[] }>(
      API_ENDPOINTS.kubernetes.metadata.versions(provider, credentialId, region)
    );
    return data.versions || [];
  }

  async getAWSRegions(
    provider: CloudProvider,
    credentialId: string
  ): Promise<string[]> {
    const data = await this.get<{ regions: string[] }>(
      API_ENDPOINTS.kubernetes.metadata.regions(provider, credentialId)
    );
    return data.regions || [];
  }

  async getAvailabilityZones(
    provider: CloudProvider,
    credentialId: string,
    region: string
  ): Promise<string[]> {
    const data = await this.get<{ zones: string[] }>(
      API_ENDPOINTS.kubernetes.metadata.availabilityZones(provider, credentialId, region)
    );
    return data.zones || [];
  }
}

export const kubernetesService = new KubernetesService();

/**
 * Kubernetes Service
 * Kubernetes 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import { API_ENDPOINTS } from '@/lib/api-endpoints';
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
  // ===== Cluster Management =====
  
  /**
   * Kubernetes 클러스터 목록 조회
   * 
   * @param provider - 클라우드 프로바이더 (aws, gcp, azure 등)
   * @param credentialId - 자격 증명 ID
   * @param region - 리전 (선택사항)
   * @returns 클러스터 배열
   * 
   * @example
   * ```tsx
   * const clusters = await kubernetesService.listClusters('aws', 'credential-id', 'ap-northeast-2');
   * ```
   */
  async listClusters(
    provider: CloudProvider,
    credentialId: string,
    region?: string
  ): Promise<KubernetesCluster[]> {
    // 1. API 호출하여 클러스터 목록 가져오기
    const data = await this.get<{ clusters: KubernetesCluster[] }>(
      API_ENDPOINTS.kubernetes.clusters.list(provider, credentialId, region)
    );
    // 2. 응답 데이터에서 clusters 배열 추출 (없으면 빈 배열)
    return data.clusters || [];
  }

  /**
   * 특정 Kubernetes 클러스터 조회
   * 
   * @param provider - 클라우드 프로바이더
   * @param clusterName - 클러스터 이름
   * @param credentialId - 자격 증명 ID
   * @param region - 리전
   * @returns 클러스터 정보
   * 
   * @example
   * ```tsx
   * const cluster = await kubernetesService.getCluster('aws', 'my-cluster', 'credential-id', 'ap-northeast-2');
   * ```
   */
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

  /**
   * Kubernetes 클러스터 생성
   * 
   * @param provider - 클라우드 프로바이더
   * @param data - 클러스터 생성 데이터 (name, version, region, subnet_ids 등)
   * @returns 생성된 클러스터 정보
   * 
   * @example
   * ```tsx
   * const cluster = await kubernetesService.createCluster('aws', {
   *   credential_id: 'credential-id',
   *   name: 'my-cluster',
   *   version: '1.31',
   *   region: 'ap-northeast-2',
   *   subnet_ids: ['subnet-123'],
   * });
   * ```
   */
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
    const data = await this.get<{ kubeconfig: string }>(
      API_ENDPOINTS.kubernetes.clusters.kubeconfig(provider, clusterName, credentialId, region)
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

  // 클러스터 태그 업데이트
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

  // 여러 클러스터의 태그 일괄 업데이트
  async bulkUpdateTags(
    provider: CloudProvider,
    clusterUpdates: Array<{
      clusterName: string;
      credentialId: string;
      region: string;
      tags: Record<string, string>;
    }>
  ): Promise<void> {
    // 참고: 백엔드에 일괄 업데이트 API 엔드포인트가 필요함
    // 현재는 개별 업데이트를 호출함
    await Promise.all(
      clusterUpdates.map(({ clusterName, credentialId, region, tags }) =>
        this.updateClusterTags(provider, clusterName, credentialId, region, tags)
      )
    );
  }

  // Node Pool 관리 (GKE, AKS, NKS)
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

  // Node Group 관리 (EKS 전용)
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

  // Node 관리
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
}

export const kubernetesService = new KubernetesService();

/**
 * Kubernetes Cluster Detail Hook
 * 클러스터 상세 페이지의 데이터 fetching 및 mutations 관리
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useWorkspaceStore } from '@/store/workspace';
import { useProviderStore } from '@/store/provider';
import { useCredentials } from '@/hooks/use-credentials';
import { kubernetesService } from '../services/kubernetes';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { downloadKubeconfig } from '@/utils/kubeconfig';
import type { CreateNodePoolForm, CreateNodeGroupForm, CloudProvider } from '@/lib/types';

export interface UseClusterDetailOptions {
  clusterName: string;
}

export function useClusterDetail({ clusterName }: UseClusterDetailOptions) {
  const _router = useRouter();
  const queryClient = useQueryClient();
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedProvider } = useProviderStore();
  const [selectedRegion, setSelectedRegion] = useState<string>('ap-northeast-2');
  const [selectedCredentialId, setSelectedCredentialId] = useState<string>('');

  // Fetch credentials using unified hook
  const { credentials } = useCredentials({
    workspaceId: currentWorkspace?.id,
  });

  const filteredCredentials = selectedProvider
    ? credentials.filter((c) => c.provider === selectedProvider)
    : [];

  // Fetch cluster details
  const { data: cluster, isLoading: isLoadingCluster } = useQuery({
    queryKey: queryKeys.kubernetesClusters.detail(clusterName),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getCluster(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 5000,
  });

  // Fetch node pools (GKE, AKS, NKS)
  const { data: nodePools = [], isLoading: isLoadingNodePools } = useQuery({
    queryKey: queryKeys.kubernetesClusters.nodePools(clusterName, selectedProvider || undefined, selectedCredentialId || undefined, selectedRegion || undefined),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listNodePools(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 10000,
  });

  // Fetch node groups (EKS)
  const { data: nodeGroups = [], isLoading: isLoadingNodeGroups } = useQuery({
    queryKey: queryKeys.kubernetesClusters.nodePools(clusterName, selectedProvider || undefined, selectedCredentialId || undefined, selectedRegion || undefined),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listNodeGroups(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace && selectedProvider === 'aws',
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 10000,
  });

  // Fetch nodes
  const { data: nodes = [], isLoading: isLoadingNodes } = useQuery({
    queryKey: queryKeys.kubernetesClusters.nodes(clusterName, selectedProvider || undefined, selectedCredentialId || undefined, selectedRegion || undefined),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listNodes(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 10000,
  });

  // Fetch upgrade status
  const { data: upgradeStatus, isLoading: isLoadingUpgradeStatus } = useQuery({
    queryKey: ['upgrade-status', selectedProvider, clusterName, selectedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getUpgradeStatus(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace,
    refetchInterval: (query) => {
      const data = query.state.data as { status?: string } | undefined;
      return data?.status === 'IN_PROGRESS' || data?.status === 'PENDING' ? 5000 : 30000;
    },
  });

  // Create node pool mutation
  const createNodePoolMutation = useMutation({
    mutationFn: (data: CreateNodePoolForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return kubernetesService.createNodePool(selectedProvider, clusterName, data);
    },
    onSuccess: () => {
      success('Node pool creation initiated');
      queryClient.invalidateQueries({ queryKey: queryKeys.nodePools.all });
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'createNodePool', resource: 'NodePool' });
    },
  });

  // Create node group mutation
  const createNodeGroupMutation = useMutation({
    mutationFn: (data: CreateNodeGroupForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return kubernetesService.createNodeGroup(selectedProvider, clusterName, data);
    },
    onSuccess: () => {
      success('Node group creation initiated');
      queryClient.invalidateQueries({ queryKey: queryKeys.nodePools.all });
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'createNodeGroup', resource: 'NodeGroup' });
    },
  });

  // Delete node pool mutation
  const deleteNodePoolMutation = useMutation({
    mutationFn: async ({ nodePoolName }: { nodePoolName: string }) => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.deleteNodePool(selectedProvider, clusterName, nodePoolName, selectedCredentialId, selectedRegion);
    },
    onSuccess: () => {
      success('Node pool deletion initiated');
      queryClient.invalidateQueries({ queryKey: queryKeys.nodePools.all });
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'deleteNodePool', resource: 'NodePool' });
    },
  });

  // Delete node group mutation
  const deleteNodeGroupMutation = useMutation({
    mutationFn: async ({ nodeGroupName }: { nodeGroupName: string }) => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.deleteNodeGroup(selectedProvider, clusterName, nodeGroupName, selectedCredentialId, selectedRegion);
    },
    onSuccess: () => {
      success('Node group deletion initiated');
      queryClient.invalidateQueries({ queryKey: queryKeys.nodePools.all });
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'deleteNodeGroup', resource: 'NodeGroup' });
    },
  });

  // Scale node pool mutation
  const scaleNodePoolMutation = useMutation({
    mutationFn: async ({ nodePoolName, nodeCount }: { nodePoolName: string; nodeCount: number }) => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.scaleNodePool(selectedProvider, clusterName, nodePoolName, nodeCount, selectedCredentialId, selectedRegion);
    },
    onSuccess: () => {
      success('Node pool scaling initiated');
      queryClient.invalidateQueries({ queryKey: queryKeys.nodePools.all });
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'scaleNodePool', resource: 'NodePool' });
    },
  });

  // Upgrade cluster mutation
  const upgradeClusterMutation = useMutation({
    mutationFn: async (version: string) => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.upgradeCluster(selectedProvider, clusterName, version, selectedCredentialId, selectedRegion);
    },
    onSuccess: () => {
      success('Cluster upgrade initiated');
      queryClient.invalidateQueries({ queryKey: ['kubernetes-cluster'] });
      queryClient.invalidateQueries({ queryKey: ['upgrade-status'] });
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'upgradeCluster', resource: 'Cluster' });
    },
  });

  // Download kubeconfig mutation
  const downloadKubeconfigMutation = useMutation({
    mutationFn: async () => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getKubeconfig(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    onSuccess: (kubeconfig) => {
      downloadKubeconfig(kubeconfig, clusterName);
      success('Kubeconfig downloaded');
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'downloadKubeconfig', resource: 'Cluster' });
    },
  });

  return {
    // State
    selectedRegion,
    setSelectedRegion,
    selectedCredentialId,
    setSelectedCredentialId,
    
    // Data
    credentials: filteredCredentials,
    cluster,
    nodePools,
    nodeGroups,
    nodes,
    upgradeStatus,
    
    // Loading states
    isLoadingCluster,
    isLoadingNodePools,
    isLoadingNodeGroups,
    isLoadingNodes,
    isLoadingUpgradeStatus,
    
    // Mutations
    createNodePoolMutation,
    createNodeGroupMutation,
    deleteNodePoolMutation,
    deleteNodeGroupMutation,
    scaleNodePoolMutation,
    upgradeClusterMutation,
    downloadKubeconfigMutation,
    
    // Computed
    selectedProvider: selectedProvider as CloudProvider | undefined,
    currentWorkspace,
  };
}


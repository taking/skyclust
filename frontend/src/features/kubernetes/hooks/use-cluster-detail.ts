/**
 * Kubernetes Cluster Detail Hook
 * 클러스터 상세 페이지의 데이터 fetching 및 mutations 관리
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContext } from '@/hooks/use-credential-context';
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
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion: contextRegion, setSelectedCredential, setSelectedRegion: setContextRegion } = useCredentialContext();
  
  // Get provider from URL query parameter (for backward compatibility)
  const urlProvider = searchParams?.get('provider') as CloudProvider | null;
  
  // Fetch credentials using unified hook
  const { credentials, selectedCredential, selectedProvider: credentialProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });
  
  // Priority: credential provider > URL provider
  // Credential에서 추출한 provider를 우선 사용, 없으면 URL provider 사용
  const effectiveProvider = credentialProvider || urlProvider;
  
  const [selectedRegion, setSelectedRegion] = useState<string>(contextRegion || 'ap-northeast-2');
  
  // Use credential context values directly
  const activeCredentialId = selectedCredentialId || '';
  
  // Handle credential change - update credential context
  const handleCredentialChange = (credentialId: string) => {
    setSelectedCredential(credentialId);
  };
  
  // Handle region change - update both local state and context
  const handleRegionChange = (region: string) => {
    setSelectedRegion(region);
    setContextRegion(region);
  };

  const filteredCredentials = effectiveProvider
    ? credentials.filter((c) => c.provider === effectiveProvider)
    : credentials;

  // Fetch cluster details
  const { data: cluster, isLoading: isLoadingCluster } = useQuery({
    queryKey: queryKeys.kubernetesClusters.detail(clusterName),
    queryFn: async () => {
      if (!effectiveProvider || !activeCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getCluster(effectiveProvider, clusterName, activeCredentialId, selectedRegion);
    },
    enabled: !!effectiveProvider && !!activeCredentialId && !!clusterName && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 5000,
  });

  // Fetch node pools (GKE, AKS, NKS)
  const { data: nodePools = [], isLoading: isLoadingNodePools } = useQuery({
    queryKey: queryKeys.kubernetesClusters.nodePools(clusterName, effectiveProvider || undefined, activeCredentialId || undefined, selectedRegion || undefined),
    queryFn: async () => {
      if (!effectiveProvider || !activeCredentialId) return [];
      return kubernetesService.listNodePools(effectiveProvider, clusterName, activeCredentialId, selectedRegion);
    },
    enabled: !!effectiveProvider && !!activeCredentialId && !!clusterName && !!currentWorkspace && !!cluster,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 10000,
  });

  // Fetch node groups (EKS)
  const { data: nodeGroups = [], isLoading: isLoadingNodeGroups } = useQuery({
    queryKey: queryKeys.kubernetesClusters.nodePools(clusterName, effectiveProvider || undefined, activeCredentialId || undefined, selectedRegion || undefined),
    queryFn: async () => {
      if (!effectiveProvider || !activeCredentialId) return [];
      return kubernetesService.listNodeGroups(effectiveProvider, clusterName, activeCredentialId, selectedRegion);
    },
    enabled: !!effectiveProvider && !!activeCredentialId && !!clusterName && !!currentWorkspace && effectiveProvider === 'aws' && !!cluster,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 10000,
  });

  // Fetch nodes
  const { data: nodes = [], isLoading: isLoadingNodes } = useQuery({
    queryKey: queryKeys.kubernetesClusters.nodes(clusterName, effectiveProvider || undefined, activeCredentialId || undefined, selectedRegion || undefined),
    queryFn: async () => {
      if (!effectiveProvider || !activeCredentialId) return [];
      return kubernetesService.listNodes(effectiveProvider, clusterName, activeCredentialId, selectedRegion);
    },
    enabled: !!effectiveProvider && !!activeCredentialId && !!clusterName && !!currentWorkspace && !!cluster,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 10000,
  });

  // Fetch upgrade status
  const { data: upgradeStatus, isLoading: isLoadingUpgradeStatus } = useQuery({
    queryKey: ['upgrade-status', effectiveProvider, clusterName, activeCredentialId, selectedRegion],
    queryFn: async () => {
      if (!effectiveProvider || !activeCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getUpgradeStatus(effectiveProvider, clusterName, activeCredentialId, selectedRegion);
    },
    enabled: !!effectiveProvider && !!activeCredentialId && !!clusterName && !!currentWorkspace && !!cluster,
    refetchInterval: (query) => {
      const data = query.state.data as { status?: string } | undefined;
      return data?.status === 'IN_PROGRESS' || data?.status === 'PENDING' ? 5000 : 30000;
    },
  });

  // Create node pool mutation
  const createNodePoolMutation = useMutation({
    mutationFn: (data: CreateNodePoolForm) => {
      if (!effectiveProvider) throw new Error('Provider not selected');
      return kubernetesService.createNodePool(effectiveProvider, clusterName, data);
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
      if (!effectiveProvider) throw new Error('Provider not selected');
      return kubernetesService.createNodeGroup(effectiveProvider, clusterName, data);
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
      if (!effectiveProvider || !activeCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.deleteNodePool(effectiveProvider, clusterName, nodePoolName, activeCredentialId, selectedRegion);
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
      if (!effectiveProvider || !activeCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.deleteNodeGroup(effectiveProvider, clusterName, nodeGroupName, activeCredentialId, selectedRegion);
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
      if (!effectiveProvider || !activeCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.scaleNodePool(effectiveProvider, clusterName, nodePoolName, nodeCount, activeCredentialId, selectedRegion);
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
      if (!effectiveProvider || !activeCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.upgradeCluster(effectiveProvider, clusterName, version, activeCredentialId, selectedRegion);
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
      if (!effectiveProvider || !activeCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getKubeconfig(effectiveProvider, clusterName, activeCredentialId, selectedRegion);
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
    setSelectedRegion: handleRegionChange,
    selectedCredentialId: activeCredentialId,
    setSelectedCredentialId: handleCredentialChange,
    
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
    selectedProvider: effectiveProvider as CloudProvider | undefined,
    currentWorkspace,
  };
}


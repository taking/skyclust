/**
 * Kubernetes Clusters Hook
 * 클러스터 데이터 fetching 및 mutations 관리
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { kubernetesService } from '../services/kubernetes';
import { credentialService } from '@/services/credential';
import type { KubernetesCluster, CreateClusterForm, CloudProvider } from '@/lib/types';

interface UseKubernetesClustersOptions {
  workspaceId?: string;
  selectedCredentialId?: string;
  selectedRegion?: string;
}

export function useKubernetesClusters({
  workspaceId,
  selectedCredentialId,
  selectedRegion,
}: UseKubernetesClustersOptions) {
  const queryClient = useQueryClient();

  // Fetch credentials (자주 변경되지 않으므로 긴 staleTime)
  const { data: credentials = [] } = useQuery({
    queryKey: ['credentials', workspaceId],
    queryFn: () => workspaceId ? credentialService.getCredentials(workspaceId) : Promise.resolve([]),
    enabled: !!workspaceId,
    staleTime: 10 * 60 * 1000, // 10분 - 자격 증명은 자주 변경되지 않음
    gcTime: 30 * 60 * 1000, // 30분 - GC 시간
  });

  // Get selected credential to determine provider
  const selectedCredential = credentials.find(c => c.id === selectedCredentialId);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  // Fetch clusters (클러스터 상태 변화를 반영하기 위해 짧은 staleTime과 polling)
  const { data: clusters = [], isLoading } = useQuery({
    queryKey: ['kubernetes-clusters', selectedProvider, selectedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) {
        return [];
      }
      return kubernetesService.listClusters(selectedProvider, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId,
    staleTime: 60 * 1000, // 1분 - 클러스터 상태는 비교적 안정적이지만 변경 가능
    gcTime: 10 * 60 * 1000, // 10분 - GC 시간
    refetchInterval: 30000, // Poll every 30 seconds
    refetchIntervalInBackground: false, // 백그라운드 polling 비활성화
  });

  // Create cluster mutation
  const createClusterMutation = useMutation({
    mutationFn: ({ provider, data }: { provider: CloudProvider; data: CreateClusterForm }) => {
      return kubernetesService.createCluster(provider, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['kubernetes-clusters'] });
    },
  });

  // Delete cluster mutation
  const deleteClusterMutation = useMutation({
    mutationFn: async ({ 
      provider, 
      clusterName, 
      credentialId, 
      region 
    }: { 
      provider: CloudProvider; 
      clusterName: string; 
      credentialId: string; 
      region: string;
    }) => {
      return kubernetesService.deleteCluster(provider, clusterName, credentialId, region);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['kubernetes-clusters'] });
    },
  });

  // Download kubeconfig mutation
  const downloadKubeconfigMutation = useMutation({
    mutationFn: async ({
      provider,
      clusterName,
      credentialId,
      region,
    }: {
      provider: CloudProvider;
      clusterName: string;
      credentialId: string;
      region: string;
    }) => {
      return kubernetesService.getKubeconfig(provider, clusterName, credentialId, region);
    },
  });

  return {
    credentials,
    clusters,
    isLoading,
    createClusterMutation,
    deleteClusterMutation,
    downloadKubeconfigMutation,
  };
}


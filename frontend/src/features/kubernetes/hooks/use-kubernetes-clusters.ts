/**
 * Kubernetes Clusters Hook
 * 클러스터 데이터 fetching 및 mutations 관리
 */

import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query';
import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { kubernetesService } from '../services/kubernetes';
import type { KubernetesCluster, CreateClusterForm, CloudProvider } from '@/lib/types';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import { useCredentials } from '@/hooks/use-credentials';

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
  const { handleError } = useErrorHandler();

  // Fetch credentials using unified hook
  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId,
    selectedCredentialId,
  });

  // Fetch clusters (클러스터 상태 변화를 반영하기 위해 짧은 staleTime과 polling)
  const { data: clusters = [], isLoading } = useQuery({
    queryKey: queryKeys.kubernetesClusters.list(workspaceId, selectedProvider, selectedCredentialId, selectedRegion),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) {
        return [];
      }
      return kubernetesService.listClusters(selectedProvider, selectedCredentialId, selectedRegion);
    },
    enabled: !!workspaceId && !!selectedProvider && !!selectedCredentialId,
    staleTime: CACHE_TIMES.MONITORING, // 1분 - 클러스터 상태는 비교적 안정적이지만 변경 가능
    gcTime: GC_TIMES.MEDIUM, // 10분 - GC 시간
    refetchInterval: 30000, // Poll every 30 seconds
    refetchIntervalInBackground: false, // 백그라운드 polling 비활성화
  });

  // Create cluster mutation
  const createClusterMutation = useStandardMutation({
    mutationFn: ({ provider, data }: { provider: CloudProvider; data: CreateClusterForm }) => {
      return kubernetesService.createCluster(provider, data);
    },
    invalidateQueries: [queryKeys.kubernetesClusters.all],
    successMessage: 'Cluster creation initiated',
    errorContext: { operation: 'createCluster', resource: 'Cluster' },
  });

  // Delete cluster mutation
  const deleteClusterMutation = useStandardMutation({
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
    invalidateQueries: [queryKeys.kubernetesClusters.all],
    successMessage: 'Cluster deletion initiated',
    errorContext: { operation: 'deleteCluster', resource: 'Cluster' },
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


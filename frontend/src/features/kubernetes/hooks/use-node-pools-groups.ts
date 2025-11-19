/**
 * Node Pools/Groups Hook
 * Node Pool/Group 데이터 fetching 및 mutations 관리
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { kubernetesService } from '../services/kubernetes';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';

interface UseNodePoolsGroupsOptions {
  workspaceId: string;
  selectedCredentialId: string;
  selectedRegion?: string;
  provider?: CloudProvider;
}

export function useNodePoolsGroups({
  workspaceId,
  selectedCredentialId,
  selectedRegion,
  provider,
}: UseNodePoolsGroupsOptions) {
  const { success, error: showError } = useToast();
  const { handleError } = useErrorHandler();
  const queryClient = useQueryClient();

  // Delete mutation
  const deleteNodePoolGroupMutation = useMutation({
    mutationFn: async ({
      provider: p,
      clusterName,
      name,
      isNodeGroup,
      credentialId,
      region,
    }: {
      provider: CloudProvider;
      clusterName: string;
      name: string;
      isNodeGroup: boolean;
      credentialId: string;
      region: string;
    }) => {
      if (isNodeGroup) {
        return kubernetesService.deleteNodeGroup(p, clusterName, name, credentialId, region);
      }
      return kubernetesService.deleteNodePool(p, clusterName, name, credentialId, region);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.kubernetesClusters.all,
      });
      success('Node Pool/Group deletion initiated');
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'deleteNodePoolGroup', resource: 'NodePoolGroup' });
    },
  });

  return {
    deleteNodePoolGroupMutation,
  };
}


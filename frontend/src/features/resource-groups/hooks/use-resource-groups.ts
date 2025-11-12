/**
 * Resource Groups Hook
 * Azure Resource Groups 데이터 fetching 및 상태 관리
 */

import { useQuery } from '@tanstack/react-query';
import { resourceGroupService, type ResourceGroupInfo } from '@/services/resource-group';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useCredentials } from '@/hooks/use-credentials';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useWorkspaceStore } from '@/store/workspace';

export interface UseResourceGroupsOptions {
  limit?: number;
  enabled?: boolean;
}

export interface UseResourceGroupsReturn {
  resourceGroups: ResourceGroupInfo[];
  isLoadingResourceGroups: boolean;
  error: Error | null;
  credentials: ReturnType<typeof useCredentials>['credentials'];
  selectedProvider: string | undefined;
  selectedCredentialId: string;
}

export function useResourceGroups(
  options: UseResourceGroupsOptions = {}
): UseResourceGroupsReturn {
  const { limit, enabled = true } = options;
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId } = useCredentialContext();

  const { credentials, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: enabled && !!currentWorkspace,
  });

  const { data: resourceGroups = [], isLoading: isLoadingResourceGroups, error } = useQuery<ResourceGroupInfo[]>({
    queryKey: queryKeys.azureResourceGroups.list(selectedCredentialId || undefined, limit),
    queryFn: () => {
      if (!selectedCredentialId) {
        return Promise.resolve([]);
      }
      return resourceGroupService.listResourceGroups(selectedCredentialId, limit);
    },
    enabled: enabled && !!currentWorkspace && !!selectedCredentialId && selectedProvider === 'azure',
    staleTime: CACHE_TIMES.STABLE,
    gcTime: GC_TIMES.LONG,
  });

  return {
    resourceGroups,
    isLoadingResourceGroups,
    error: error as Error | null,
    credentials,
    selectedProvider,
    selectedCredentialId: selectedCredentialId || '',
  };
}


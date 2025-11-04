/**
 * useWorkspaces Hook
 * Workspaces 데이터 fetching hook
 */

import { useQuery } from '@tanstack/react-query';
import { workspaceService } from '../services/workspace';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import type { Workspace } from '@/lib/types';

interface UseWorkspacesOptions {
  enabled?: boolean;
}

export function useWorkspaces(options: UseWorkspacesOptions = {}) {
  const { enabled = true } = options;

  const {
    data: workspaces = [],
    isLoading,
    error,
    refetch,
  } = useQuery<Workspace[]>({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    enabled,
    staleTime: CACHE_TIMES.STABLE,
    gcTime: GC_TIMES.LONG,
    retry: 3,
    retryDelay: 1000,
  });

  return {
    workspaces,
    isLoading,
    error,
    refetch,
  };
}


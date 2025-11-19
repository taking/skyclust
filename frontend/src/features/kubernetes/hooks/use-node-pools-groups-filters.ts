/**
 * Node Pools/Groups Filters Hook
 * Node Pool/Group 필터링 및 검색 로직
 */

import { useState, useMemo, useCallback } from 'react';
import { DataProcessor } from '@/lib/data';
import type { NodePool, NodeGroup, CloudProvider } from '@/lib/types/kubernetes';
import type { FilterValue } from '@/components/ui/filter-panel';

type NodePoolOrGroup = (NodePool | NodeGroup) & {
  cluster_name: string;
  cluster_id: string;
  provider: CloudProvider;
  resource_type: 'node-pool' | 'node-group';
};

interface UseNodePoolsGroupsFiltersOptions {
  items: NodePoolOrGroup[];
  filters: FilterValue;
}

export function useNodePoolsGroupsFilters({
  items,
  filters,
}: UseNodePoolsGroupsFiltersOptions) {
  const [searchQuery, setSearchQuery] = useState('');

  // Custom filter function for node pool/group-specific filtering (memoized)
  const filterFn = useCallback((item: NodePoolOrGroup, filters: FilterValue): boolean => {
    if (filters.status && item.status !== filters.status) return false;
    if (filters.region && item.region !== filters.region) return false;
    if (filters.provider && item.provider !== filters.provider) return false;
    if (filters.cluster && item.cluster_name !== filters.cluster) return false;
    
    return true;
  }, []);

  // Apply search and filter using DataProcessor (memoized)
  const filteredItems = useMemo(() => {
    let result = DataProcessor.search(items, searchQuery, {
      keys: ['name', 'cluster_name', 'instance_type', 'status', 'region'],
      threshold: 0.3,
    });

    // Apply filters
    result = DataProcessor.filter(result, filters, filterFn);
    
    return result;
  }, [items, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  return {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filteredItems,
  };
}




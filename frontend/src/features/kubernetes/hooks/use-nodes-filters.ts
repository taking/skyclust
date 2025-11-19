/**
 * Nodes Filters Hook
 * 노드 필터링 및 검색 로직
 */

import { useState, useMemo, useCallback } from 'react';
import { DataProcessor } from '@/lib/data';
import type { Node, CloudProvider } from '@/lib/types';
import type { FilterValue } from '@/components/ui/filter-panel';

interface UseNodesFiltersOptions {
  nodes: Array<Node & { cluster_name: string; cluster_id: string; provider?: CloudProvider }>;
  filters: FilterValue;
}

export function useNodesFilters({
  nodes,
  filters,
}: UseNodesFiltersOptions) {
  const [searchQuery, setSearchQuery] = useState('');

  const filterFn = useCallback((node: Node & { cluster_name: string; cluster_id: string; provider?: CloudProvider }, filters: FilterValue): boolean => {
    if (filters.status && node.status !== filters.status) return false;
    if (filters.cluster && node.cluster_name !== filters.cluster) return false;
    if (filters.provider && node.provider !== filters.provider) return false;
    if (filters.instance_type && node.instance_type !== filters.instance_type) return false;
    return true;
  }, []);

  const filteredNodes = useMemo(() => {
    let result = DataProcessor.search(nodes, searchQuery, {
      keys: ['name', 'cluster_name', 'instance_type', 'status', 'zone', 'private_ip'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result;
  }, [nodes, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  return {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filteredNodes,
  };
}


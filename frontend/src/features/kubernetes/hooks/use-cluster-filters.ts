/**
 * Cluster Filters Hook
 * 클러스터 필터링 및 검색 로직
 */

import { useState, useMemo, useCallback } from 'react';
import { DataProcessor } from '@/lib/data-processor';
import type { KubernetesCluster } from '@/lib/types';
import type { FilterValue } from '@/components/ui/filter-panel';

interface UseClusterFiltersOptions {
  clusters: KubernetesCluster[];
  filters: FilterValue;
  tagFilters: Record<string, string[]>;
}

export function useClusterFilters({
  clusters,
  filters,
  tagFilters,
}: UseClusterFiltersOptions) {
  const [searchQuery, setSearchQuery] = useState('');

  // Extract available tags from clusters
  const availableTags = useMemo(() => {
    const tagMap: Record<string, Set<string>> = {};
    clusters.forEach(cluster => {
      if (cluster.tags) {
        Object.entries(cluster.tags).forEach(([key, value]) => {
          if (!tagMap[key]) {
            tagMap[key] = new Set();
          }
          tagMap[key].add(value);
        });
      }
    });
    
    const result: Record<string, string[]> = {};
    Object.entries(tagMap).forEach(([key, valueSet]) => {
      result[key] = Array.from(valueSet).sort();
    });
    return result;
  }, [clusters]);

  // Custom filter function for cluster-specific filtering including tags (memoized)
  const filterFn = useCallback((cluster: KubernetesCluster, filters: FilterValue): boolean => {
    if (filters.status && cluster.status !== filters.status) return false;
    if (filters.region && cluster.region !== filters.region) return false;
    
    // Apply tag filters
    if (Object.keys(tagFilters).length > 0 && cluster.tags) {
      for (const [tagKey, tagValues] of Object.entries(tagFilters)) {
        if (tagValues && tagValues.length > 0) {
          const clusterTagValue = cluster.tags[tagKey];
          if (!clusterTagValue || !tagValues.includes(clusterTagValue)) {
            return false;
          }
        }
      }
    }
    
    return true;
  }, [tagFilters]);

  // Apply search and filter using DataProcessor (memoized)
  const filteredClusters = useMemo(() => {
    let result = DataProcessor.search(clusters, searchQuery, {
      keys: ['name', 'version', 'status', 'region'],
      threshold: 0.3,
    });

    // Apply filters including tag filters
    result = DataProcessor.filter(result, filters, filterFn);
    
    return result;
  }, [clusters, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, [setSearchQuery]);

  return {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    availableTags,
    filteredClusters,
  };
}

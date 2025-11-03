/**
 * Cluster Filters Hook
 * 클러스터 필터링 및 검색 로직
 */

import { useMemo } from 'react';
import { useSearch } from '@/hooks/use-search';
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
  // Search functionality
  const {
    query: searchQuery,
    setQuery: setSearchQuery,
    results: searchResults,
    isSearching,
    clearSearch,
  } = useSearch(clusters, {
    keys: ['name', 'version', 'status', 'region'],
    threshold: 0.3,
  });

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

  // Apply filters including tag filters
  const filteredClusters = useMemo(() => {
    const result = searchResults.filter((cluster: KubernetesCluster) => {
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
    });
    
    return result;
  }, [searchResults, filters, tagFilters]);

  return {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    availableTags,
    filteredClusters,
  };
}


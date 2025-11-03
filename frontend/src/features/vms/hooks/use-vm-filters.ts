/**
 * VM Filters Hook
 * VM 필터링 및 검색 로직
 */

import { useMemo } from 'react';
import { useSearch } from '@/hooks/use-search';
import { useAdvancedFiltering } from '@/hooks/use-advanced-filtering';
import { multiSort } from '@/utils/sort-utils';
import type { VM } from '@/lib/types';
import type { FilterValue } from '@/components/ui/filter-panel';

interface UseVMFiltersOptions {
  vms: VM[];
  advancedFilters: FilterValue;
  sortConfig: Array<{ field: string; direction: 'asc' | 'desc' }>;
}

export function useVMFilters({
  vms,
  advancedFilters,
  sortConfig,
}: UseVMFiltersOptions) {
  // Search functionality
  const {
    query: searchQuery,
    setQuery: setSearchQuery,
    results: searchResults,
    isSearching,
    clearSearch,
  } = useSearch(vms, {
    keys: ['name', 'provider', 'instance_type', 'region', 'status'],
    threshold: 0.3,
  });

  // Apply filters and sorting to search results
  const filteredVMs = useMemo(() => {
    let result = searchResults.filter((vm) => {
      // Status filter
      if (advancedFilters.status && Array.isArray(advancedFilters.status) && advancedFilters.status.length > 0) {
        if (!advancedFilters.status.includes(vm.status)) return false;
      }

      // Provider filter
      if (advancedFilters.provider && Array.isArray(advancedFilters.provider) && advancedFilters.provider.length > 0) {
        if (!advancedFilters.provider.includes(vm.provider)) return false;
      }

      // Region filter
      if (advancedFilters.region && advancedFilters.region !== vm.region) {
        return false;
      }

      return true;
    });
    
    // Apply multi-sort
    if (sortConfig.length > 0) {
      result = multiSort(result, sortConfig, (vm, field) => {
        switch (field) {
          case 'name': return vm.name;
          case 'provider': return vm.provider;
          case 'status': return vm.status;
          case 'region': return vm.region;
          case 'instance_type': return vm.instance_type;
          case 'created_at': return vm.created_at ? new Date(vm.created_at) : null;
          default: return null;
        }
      });
    }
    
    return result;
  }, [searchResults, advancedFilters, sortConfig]);

  return {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filteredVMs,
  };
}


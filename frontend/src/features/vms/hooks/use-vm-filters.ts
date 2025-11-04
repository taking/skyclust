/**
 * VM Filters Hook
 * VM 필터링 및 검색 로직
 */

import { useState, useMemo } from 'react';
import { DataProcessor } from '@/lib/data-processor';
import type { SortConfig } from '@/hooks/use-advanced-filtering';
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
  const [searchQuery, setSearchQuery] = useState('');

  // Get value function for sorting
  const getValue = (vm: VM, field: string): unknown => {
    switch (field) {
      case 'name': return vm.name;
      case 'provider': return vm.provider;
      case 'status': return vm.status;
      case 'region': return vm.region;
      case 'instance_type': return vm.instance_type;
      case 'created_at': return vm.created_at ? new Date(vm.created_at) : null;
      default: return null;
    }
  };

  // Custom filter function for VM-specific filtering
  const filterFn = (vm: VM, filters: FilterValue): boolean => {
    // Status filter
    if (filters.status && Array.isArray(filters.status) && filters.status.length > 0) {
      if (!filters.status.includes(vm.status)) return false;
    }

    // Provider filter
    if (filters.provider && Array.isArray(filters.provider) && filters.provider.length > 0) {
      if (!filters.provider.includes(vm.provider)) return false;
    }

    // Region filter
    if (filters.region && filters.region !== vm.region) {
      return false;
    }

    return true;
  };

  // Apply search, filter, and sort using DataProcessor
  const filteredVMs = useMemo(() => {
    return DataProcessor.combine(vms, {
      searchQuery,
      searchKeys: ['name', 'provider', 'instance_type', 'region', 'status'],
      filters: advancedFilters,
      sortConfig: sortConfig as SortConfig[],
      getValue,
    });
  }, [vms, searchQuery, advancedFilters, sortConfig]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = () => {
    setSearchQuery('');
  };

  return {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filteredVMs,
  };
}

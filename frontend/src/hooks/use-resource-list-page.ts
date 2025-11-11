/**
 * useResourceListPage Hook
 * 리소스 리스트 페이지의 공통 로직을 통합한 훅
 * 
 * 검색, 필터, 정렬, 페이지네이션을 통합하여 제공합니다.
 * 모든 리소스 리스트 페이지에서 일관된 패턴을 사용할 수 있도록 합니다.
 * 
 * @example
 * ```tsx
 * const {
 *   searchQuery,
 *   setSearchQuery,
 *   filters,
 *   setAllFilters,
 *   sortConfig,
 *   toggleSort,
 *   clearSort,
 *   page,
 *   pageSize,
 *   setPage,
 *   setPageSize,
 *   paginatedData,
 *   totalPages,
 *   totalItems,
 *   processedData,
 * } = useResourceListPage({
 *   resourceName: 'vms',
 *   items: vms,
 *   searchKeys: ['name', 'id', 'region'],
 *   defaultFilters: {},
 *   defaultSort: [],
 *   getValue: (item, field) => item[field],
 * });
 * ```
 */

import { useState, useMemo, useCallback } from 'react';
import { usePagination } from './use-pagination';
import { useAdvancedFiltering, type SortConfig, type FilterPreset } from './use-advanced-filtering';
import { DataProcessor, type DataProcessingOptions } from '@/lib/data';
import type { FilterValue } from '@/components/ui/filter-panel';
import { UI } from '@/lib/constants';

export interface UseResourceListPageOptions<T> {
  /**
   * 리소스 이름 (localStorage 키에 사용)
   */
  resourceName: string;
  
  /**
   * 데이터 배열
   */
  items: T[];
  
  /**
   * 검색 키 배열 (Fuse.js에서 사용)
   */
  searchKeys: string[];
  
  /**
   * 기본 필터
   */
  defaultFilters?: Record<string, unknown>;
  
  /**
   * 기본 정렬 설정
   */
  defaultSort?: SortConfig[];
  
  /**
   * 필드 값 가져오기 함수 (정렬에 사용)
   */
  getValue: (item: T, field: string) => unknown;
  
  /**
   * 커스텀 필터 함수 (선택)
   */
  filterFn?: (item: T, filters: FilterValue) => boolean;
  
  /**
   * 검색 옵션 (선택)
   */
  searchOptions?: {
    threshold?: number;
    minMatchCharLength?: number;
  };
  
  /**
   * 초기 페이지 크기
   */
  initialPageSize?: number;
  
  /**
   * 필터 변경 콜백
   */
  onFiltersChange?: (filters: Record<string, unknown>) => void;
  
  /**
   * 정렬 변경 콜백
   */
  onSortChange?: (sort: SortConfig[]) => void;
}

export interface UseResourceListPageReturn<T> {
  // 검색
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  isSearching: boolean;
  clearSearch: () => void;
  
  // 필터
  filters: Record<string, unknown>;
  setAllFilters: (filters: Record<string, unknown>) => void;
  clearFilters: () => void;
  filterValue: FilterValue;
  setFilterValue: (filters: FilterValue) => void;
  
  // 정렬
  sortConfig: SortConfig[];
  toggleSort: (field: string) => void;
  clearSort: () => void;
  getSortIndicator: (field: string) => 'asc' | 'desc' | null;
  
  // 필터 프리셋
  presets: FilterPreset[];
  savePreset: (name: string) => void;
  loadPreset: (presetId: string) => void;
  deletePreset: (presetId: string) => void;
  
  // 페이지네이션
  page: number;
  pageSize: number;
  totalPages: number;
  totalItems: number;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  goToFirstPage: () => void;
  goToLastPage: () => void;
  goToNextPage: () => void;
  goToPreviousPage: () => void;
  canGoNext: boolean;
  canGoPrevious: boolean;
  
  // 처리된 데이터
  processedData: T[];
  paginatedData: T[];
  
  // 전체 데이터 (검색/필터/정렬 적용 전)
  allItems: T[];
}

/**
 * useResourceListPage Hook
 * 
 * 리소스 리스트 페이지의 모든 로직을 통합
 */
export function useResourceListPage<T>({
  resourceName,
  items,
  searchKeys,
  defaultFilters = {},
  defaultSort = [],
  getValue,
  filterFn,
  searchOptions,
  initialPageSize = UI.PAGINATION.DEFAULT_PAGE_SIZE,
  onFiltersChange,
  onSortChange,
}: UseResourceListPageOptions<T>): UseResourceListPageReturn<T> {
  // 검색 상태
  const [searchQuery, setSearchQuery] = useState<string>('');
  const isSearching = searchQuery.trim().length > 0;

  // 고급 필터링 (localStorage 저장 포함)
  const {
    filters,
    sortConfig,
    presets,
    setAllFilters,
    clearFilters,
    savePreset,
    loadPreset,
    deletePreset,
    toggleSort,
    clearSort,
  } = useAdvancedFiltering<T>({
    storageKey: `${resourceName}-page`,
    defaultFilters,
    defaultSort,
    onFiltersChange,
    onSortChange,
  });

  // FilterValue 타입으로 변환 (FilterPanel과 호환)
  const filterValue = useMemo(() => filters as FilterValue, [filters]);
  const setFilterValue = useCallback((newFilters: FilterValue) => {
    setAllFilters(newFilters as Record<string, unknown>);
  }, [setAllFilters]);

  // 검색 초기화
  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  // 데이터 처리 (검색 + 필터 + 정렬)
  const processedData = useMemo(() => {
    const options: DataProcessingOptions<T> = {
      searchQuery: searchQuery.trim() || undefined,
      searchKeys: searchKeys.length > 0 ? searchKeys : undefined,
      filters: filterValue,
      sortConfig: sortConfig.length > 0 ? sortConfig : undefined,
      getValue,
    };

    // 커스텀 필터 함수가 있으면 사용
    if (filterFn) {
      // DataProcessor.combine은 기본 필터링을 사용하므로,
      // 커스텀 필터 함수를 사용하려면 별도로 처리
      let result = items;

      // 1. 검색
      if (options.searchQuery && options.searchKeys && options.searchKeys.length > 0) {
        result = DataProcessor.search(result, options.searchQuery, {
          keys: options.searchKeys,
          threshold: searchOptions?.threshold ?? 0.3,
          minMatchCharLength: searchOptions?.minMatchCharLength ?? 1,
        });
      }

      // 2. 커스텀 필터
      if (options.filters && Object.keys(options.filters).length > 0) {
        result = DataProcessor.filter(result, options.filters, filterFn);
      }

      // 3. 정렬
      if (options.sortConfig && options.sortConfig.length > 0) {
        result = DataProcessor.multiSort(result, options.sortConfig, getValue);
      }

      return result;
    }

    // 기본 필터링 사용
    return DataProcessor.combine(items, options);
  }, [items, searchQuery, searchKeys, filterValue, sortConfig, getValue, filterFn, searchOptions]);

  // 페이지네이션
  const {
    page,
    pageSize,
    totalPages,
    paginatedItems: paginatedData,
    setPage,
    setPageSize,
    goToFirstPage,
    goToLastPage,
    goToNextPage,
    goToPreviousPage,
    canGoNext,
    canGoPrevious,
  } = usePagination(processedData, {
    totalItems: processedData.length,
    initialPage: 1,
    initialPageSize,
  });

  // 정렬 인디케이터 가져오기
  const getSortIndicator = useCallback((field: string): 'asc' | 'desc' | null => {
    return DataProcessor.getSortIndicator(field, sortConfig);
  }, [sortConfig]);

  return {
    // 검색
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    
    // 필터
    filters,
    setAllFilters,
    clearFilters,
    filterValue,
    setFilterValue,
    
    // 정렬
    sortConfig,
    toggleSort,
    clearSort,
    getSortIndicator,
    
    // 필터 프리셋
    presets,
    savePreset,
    loadPreset,
    deletePreset,
    
    // 페이지네이션
    page,
    pageSize,
    totalPages,
    totalItems: processedData.length,
    setPage,
    setPageSize,
    goToFirstPage,
    goToLastPage,
    goToNextPage,
    goToPreviousPage,
    canGoNext,
    canGoPrevious,
    
    // 처리된 데이터
    processedData,
    paginatedData,
    
    // 전체 데이터
    allItems: items,
  };
}


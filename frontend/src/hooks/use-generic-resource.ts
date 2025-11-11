/**
 * Generic Resource Hook
 * 모든 리소스 리스트 페이지의 공통 로직을 통합한 훅
 * 
 * 검색, 필터, 정렬, 페이지네이션을 통합하여 제공합니다.
 * 모든 리소스 리스트 페이지에서 일관된 패턴을 사용할 수 있도록 합니다.
 * 
 * @example
 * ```tsx
 * const {
 *   // 데이터
 *   items,
 *   isLoading,
 *   filteredItems,
 *   paginatedItems,
 *   
 *   // 검색
 *   searchQuery,
 *   setSearchQuery,
 *   isSearching,
 *   clearSearch,
 *   
 *   // 필터
 *   filters,
 *   setFilters,
 *   filterConfigs,
 *   
 *   // 페이지네이션
 *   page,
 *   pageSize,
 *   setPage,
 *   setPageSize,
 *   totalPages,
 *   totalItems,
 *   
 *   // 선택
 *   selectedIds,
 *   setSelectedIds,
 * } = useGenericResource({
 *   resourceName: 'vms',
 *   items: vms,
 *   isLoading,
 *   searchKeys: ['name', 'id', 'region'],
 *   filterConfigs: vmFilterConfigs,
 *   filterFn: (vm, filters) => {
 *     if (filters.status && vm.status !== filters.status) return false;
 *     return true;
 *   },
 * });
 * ```
 */

import { useState, useMemo, useCallback } from 'react';
import { usePagination } from './use-pagination';
import { DataProcessor } from '@/lib/data';
import type { FilterValue } from '@/components/ui/filter-panel';
import type { FilterConfig } from '@/components/ui/filter-panel';
import { UI } from '@/lib/constants';

export interface UseGenericResourceOptions<TItem> {
  /**
   * 리소스 이름 (localStorage 키에 사용)
   */
  resourceName: string;
  
  /**
   * 데이터 배열
   */
  items: TItem[];
  
  /**
   * 로딩 상태
   */
  isLoading?: boolean;
  
  /**
   * 검색 키 배열 (Fuse.js에서 사용)
   */
  searchKeys: string[];
  
  /**
   * 필터 설정
   */
  filterConfigs: FilterConfig[];
  
  /**
   * 커스텀 필터 함수
   */
  filterFn?: (item: TItem, filters: FilterValue) => boolean;
  
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
  onFiltersChange?: (filters: FilterValue) => void;
  
  /**
   * 검색 변경 콜백
   */
  onSearchChange?: (query: string) => void;
}

export interface UseGenericResourceReturn<TItem> {
  /**
   * 원본 데이터
   */
  items: TItem[];
  
  /**
   * 로딩 상태
   */
  isLoading: boolean;
  
  /**
   * 검색 쿼리
   */
  searchQuery: string;
  
  /**
   * 검색 쿼리 설정
   */
  setSearchQuery: (query: string) => void;
  
  /**
   * 검색 중 여부
   */
  isSearching: boolean;
  
  /**
   * 검색 초기화
   */
  clearSearch: () => void;
  
  /**
   * 필터 값
   */
  filters: FilterValue;
  
  /**
   * 필터 설정
   */
  setFilters: (filters: FilterValue | ((prev: FilterValue) => FilterValue)) => void;
  
  /**
   * 필터 설정 (전체 교체)
   */
  setAllFilters: (filters: FilterValue) => void;
  
  /**
   * 필터 초기화
   */
  clearFilters: () => void;
  
  /**
   * 필터 설정 배열
   */
  filterConfigs: FilterConfig[];
  
  /**
   * 필터링된 아이템
   */
  filteredItems: TItem[];
  
  /**
   * 페이지네이션된 아이템
   */
  paginatedItems: TItem[];
  
  /**
   * 현재 페이지
   */
  page: number;
  
  /**
   * 페이지 크기
   */
  pageSize: number;
  
  /**
   * 페이지 설정
   */
  setPage: (page: number) => void;
  
  /**
   * 페이지 크기 설정
   */
  setPageSize: (size: number) => void;
  
  /**
   * 총 페이지 수
   */
  totalPages: number;
  
  /**
   * 총 아이템 수
   */
  totalItems: number;
  
  /**
   * 선택된 ID 배열
   */
  selectedIds: string[];
  
  /**
   * 선택된 ID 설정
   */
  setSelectedIds: (ids: string[] | ((prev: string[]) => string[])) => void;
  
  /**
   * 선택 초기화
   */
  clearSelection: () => void;
}

/**
 * Generic Resource Hook
 */
export function useGenericResource<TItem extends { id: string }>(
  options: UseGenericResourceOptions<TItem>
): UseGenericResourceReturn<TItem> {
  const {
    resourceName: _resourceName, // 사용되지 않지만 옵션으로 유지
    items,
    isLoading = false,
    searchKeys,
    filterConfigs,
    filterFn,
    searchOptions = { threshold: 0.3 },
    initialPageSize = UI.PAGINATION.DEFAULT_PAGE_SIZE,
    onFiltersChange,
    onSearchChange,
  } = options;

  // 검색 상태
  const [searchQuery, setSearchQueryState] = useState('');
  
  // 필터 상태
  const [filters, setFiltersState] = useState<FilterValue>({});
  
  // 선택 상태
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  
  // 페이지 크기 상태
  const [pageSize, setPageSizeState] = useState(initialPageSize);

  // 검색 쿼리 설정 (콜백 호출)
  const setSearchQuery = useCallback((query: string) => {
    setSearchQueryState(query);
    onSearchChange?.(query);
  }, [onSearchChange]);

  // 검색 초기화
  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, [setSearchQuery]);

  // 검색 중 여부
  const isSearching = searchQuery.length > 0;

  // 필터 설정 (콜백 호출)
  const setFilters = useCallback((newFilters: FilterValue | ((prev: FilterValue) => FilterValue)) => {
    setFiltersState((prev) => {
      const updated = typeof newFilters === 'function' ? newFilters(prev) : newFilters;
      onFiltersChange?.(updated);
      return updated;
    });
  }, [onFiltersChange]);

  // 필터 전체 교체
  const setAllFilters = useCallback((newFilters: FilterValue) => {
    setFiltersState(newFilters);
    onFiltersChange?.(newFilters);
  }, [onFiltersChange]);

  // 필터 초기화
  const clearFilters = useCallback(() => {
    setFiltersState({});
    onFiltersChange?.({});
  }, [onFiltersChange]);

  // 선택 초기화
  const clearSelection = useCallback(() => {
    setSelectedIds([]);
  }, []);

  // 필터링된 아이템 (검색 + 필터)
  const filteredItems = useMemo(() => {
    // 1. 검색 적용
    let result = DataProcessor.search(items, searchQuery, {
      keys: searchKeys,
      ...searchOptions,
    });

    // 2. 필터 적용
    if (filterFn) {
      result = DataProcessor.filter(result, filters, filterFn);
    } else {
      // 기본 필터링 (필터 값이 있으면 적용)
      result = DataProcessor.filter(result, filters);
    }

    return result as TItem[];
  }, [items, searchQuery, filters, searchKeys, searchOptions, filterFn]);

  // 페이지네이션
  const {
    page,
    paginatedItems,
    setPage,
    setPageSize: setPaginationPageSize,
    totalPages,
  } = usePagination(filteredItems, {
    totalItems: filteredItems.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // 총 아이템 수 (필터링된 아이템 수)
  const totalItems = filteredItems.length;

  // 페이지 크기 변경 핸들러 (로컬 상태와 페이지네이션 동기화)
  const setPageSize = useCallback((size: number) => {
    setPageSizeState(size);
    setPaginationPageSize(size);
  }, [setPaginationPageSize]);

  return {
    items,
    isLoading,
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filters,
    setFilters,
    setAllFilters,
    clearFilters,
    filterConfigs,
    filteredItems,
    paginatedItems,
    page,
    pageSize,
    setPage,
    setPageSize,
    totalPages,
    totalItems,
    selectedIds,
    setSelectedIds,
    clearSelection,
  };
}


/**
 * useResourceListState Hook
 * 리스트 페이지의 공통 상태를 관리하는 훅
 * 
 * 모든 리스트 페이지에서 반복되는 상태 관리 로직을 통합
 * - 필터 상태
 * - 선택된 항목
 * - 페이지 크기
 * - 삭제 다이얼로그 상태
 * 
 * @example
 * ```tsx
 * const {
 *   filters,
 *   setFilters,
 *   selectedIds,
 *   setSelectedIds,
 *   pageSize,
 *   setPageSize,
 *   deleteDialogState,
 *   setDeleteDialogState,
 * } = useResourceListState({ storageKey: 'vpcs-page' });
 * ```
 */

import { useState, useCallback } from 'react';
import { UI } from '@/lib/constants';
import { FilterValue } from '@/components/ui/filter-panel';

/**
 * 삭제 다이얼로그 상태
 */
export interface DeleteDialogState {
  open: boolean;
  id: string | null;
  region?: string | null;
  name?: string;
}

/**
 * useResourceListState 옵션
 */
export interface UseResourceListStateOptions {
  /**
   * localStorage 저장 키 (선택사항)
   * 페이지 크기 등을 저장하는데 사용
   */
  storageKey?: string;

  /**
   * 기본 페이지 크기
   */
  defaultPageSize?: number;

  /**
   * 기본 필터 값
   */
  defaultFilters?: FilterValue;
}

/**
 * useResourceListState 반환 타입
 */
export interface UseResourceListStateReturn {
  /**
   * 필터 상태
   */
  filters: FilterValue;
  setFilters: (filters: FilterValue | ((prev: FilterValue) => FilterValue)) => void;

  /**
   * 필터 패널 표시 여부
   */
  showFilters: boolean;
  setShowFilters: (show: boolean | ((prev: boolean) => boolean)) => void;

  /**
   * 선택된 항목 ID 배열
   */
  selectedIds: string[];
  setSelectedIds: (ids: string[] | ((prev: string[]) => string[])) => void;

  /**
   * 페이지 크기
   */
  pageSize: number;
  setPageSize: (size: number | ((prev: number) => number)) => void;

  /**
   * 삭제 다이얼로그 상태
   */
  deleteDialogState: DeleteDialogState;
  setDeleteDialogState: (state: DeleteDialogState | ((prev: DeleteDialogState) => DeleteDialogState)) => void;

  /**
   * 필터 초기화
   */
  clearFilters: () => void;

  /**
   * 선택 초기화
   */
  clearSelection: () => void;

  /**
   * 삭제 다이얼로그 열기
   */
  openDeleteDialog: (id: string, region?: string, name?: string) => void;

  /**
   * 삭제 다이얼로그 닫기
   */
  closeDeleteDialog: () => void;
}

/**
 * useResourceListState Hook
 * 리스트 페이지의 공통 상태를 관리합니다.
 */
export function useResourceListState(
  options: UseResourceListStateOptions = {}
): UseResourceListStateReturn {
  const {
    storageKey,
    defaultPageSize = UI.PAGINATION.DEFAULT_PAGE_SIZE,
    defaultFilters = {},
  } = options;

  // 필터 상태
  const [filters, setFilters] = useState<FilterValue>(defaultFilters);

  // 필터 패널 표시 여부
  const [showFilters, setShowFilters] = useState(false);

  // 선택된 항목 ID 배열
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  // 페이지 크기
  const [pageSize, setPageSize] = useState<number>(() => {
    // localStorage에서 페이지 크기 로드
    if (storageKey && typeof window !== 'undefined') {
      try {
        const saved = localStorage.getItem(`${storageKey}-pageSize`);
        if (saved) {
          const parsed = parseInt(saved, 10);
          if (!isNaN(parsed) && parsed > 0) {
            return parsed;
          }
        }
      } catch {
        // 무시
      }
    }
    return defaultPageSize;
  });

  // 삭제 다이얼로그 상태
  const [deleteDialogState, setDeleteDialogState] = useState<DeleteDialogState>({
    open: false,
    id: null,
    region: null,
    name: undefined,
  });

  // 페이지 크기 변경 시 localStorage에 저장
  const handlePageSizeChange = useCallback(
    (size: number | ((prev: number) => number)) => {
      setPageSize((prev) => {
        const newSize = typeof size === 'function' ? size(prev) : size;
        // localStorage에 저장
        if (storageKey && typeof window !== 'undefined') {
          try {
            localStorage.setItem(`${storageKey}-pageSize`, String(newSize));
          } catch {
            // 무시
          }
        }
        return newSize;
      });
    },
    [storageKey]
  );

  // 필터 초기화
  const clearFilters = useCallback(() => {
    setFilters(defaultFilters);
    setShowFilters(false);
  }, [defaultFilters]);

  // 선택 초기화
  const clearSelection = useCallback(() => {
    setSelectedIds([]);
  }, []);

  // 삭제 다이얼로그 열기
  const openDeleteDialog = useCallback((id: string, region?: string, name?: string) => {
    setDeleteDialogState({
      open: true,
      id,
      region: region || null,
      name,
    });
  }, []);

  // 삭제 다이얼로그 닫기
  const closeDeleteDialog = useCallback(() => {
    setDeleteDialogState({
      open: false,
      id: null,
      region: null,
      name: undefined,
    });
  }, []);

  return {
    filters,
    setFilters,
    showFilters,
    setShowFilters,
    selectedIds,
    setSelectedIds,
    pageSize,
    setPageSize: handlePageSizeChange,
    deleteDialogState,
    setDeleteDialogState,
    clearFilters,
    clearSelection,
    openDeleteDialog,
    closeDeleteDialog,
  };
}



/**
 * Notification Pagination Hook
 * 알림 페이지네이션 상태 관리 훅
 * 
 * 알림 목록의 페이지네이션(offset 기반)을 관리합니다.
 * 
 * @example
 * ```tsx
 * const {
 *   offset,
 *   currentPage,
 *   totalPages,
 *   goToNextPage,
 *   goToPreviousPage,
 *   goToPage,
 *   canGoNext,
 *   canGoPrevious,
 * } = useNotificationPagination({ limit: 20, total: 100 });
 * ```
 */

import { useState, useCallback, useMemo } from 'react';

export interface UseNotificationPaginationOptions {
  /** 페이지당 항목 수 */
  limit: number;
  /** 전체 항목 수 */
  total: number;
  /** 초기 offset 값 */
  initialOffset?: number;
}

export interface UseNotificationPaginationReturn {
  /** 현재 offset 값 */
  offset: number;
  /** 현재 페이지 번호 (1부터 시작) */
  currentPage: number;
  /** 전체 페이지 수 */
  totalPages: number;
  /** 다음 페이지로 이동 */
  goToNextPage: () => void;
  /** 이전 페이지로 이동 */
  goToPreviousPage: () => void;
  /** 특정 페이지로 이동 */
  goToPage: (page: number) => void;
  /** 다음 페이지로 이동 가능 여부 */
  canGoNext: boolean;
  /** 이전 페이지로 이동 가능 여부 */
  canGoPrevious: boolean;
  /** 현재 페이지의 시작 인덱스 (1부터 시작) */
  startIndex: number;
  /** 현재 페이지의 끝 인덱스 */
  endIndex: number;
}

/**
 * 알림 페이지네이션 상태 관리 훅
 */
export function useNotificationPagination(
  options: UseNotificationPaginationOptions
): UseNotificationPaginationReturn {
  const { limit, total, initialOffset = 0 } = options;
  const [offset, setOffset] = useState(initialOffset);

  const totalPages = useMemo(() => {
    return Math.ceil(total / limit);
  }, [total, limit]);

  const currentPage = useMemo(() => {
    return Math.floor(offset / limit) + 1;
  }, [offset, limit]);

  const canGoNext = useMemo(() => {
    return offset + limit < total;
  }, [offset, limit, total]);

  const canGoPrevious = useMemo(() => {
    return offset > 0;
  }, [offset]);

  const startIndex = useMemo(() => {
    return offset + 1;
  }, [offset]);

  const endIndex = useMemo(() => {
    return Math.min(offset + limit, total);
  }, [offset, limit, total]);

  const goToNextPage = useCallback(() => {
    if (canGoNext) {
      setOffset(prev => prev + limit);
    }
  }, [canGoNext, limit]);

  const goToPreviousPage = useCallback(() => {
    if (canGoPrevious) {
      setOffset(prev => Math.max(0, prev - limit));
    }
  }, [canGoPrevious, limit]);

  const goToPage = useCallback((page: number) => {
    const newOffset = (page - 1) * limit;
    if (newOffset >= 0 && newOffset < total) {
      setOffset(newOffset);
    }
  }, [limit, total]);

  return {
    offset,
    currentPage,
    totalPages,
    goToNextPage,
    goToPreviousPage,
    goToPage,
    canGoNext,
    canGoPrevious,
    startIndex,
    endIndex,
  };
}


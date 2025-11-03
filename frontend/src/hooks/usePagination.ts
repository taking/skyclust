import { useState, useMemo } from 'react';

interface UsePaginationOptions {
  totalItems: number;
  initialPage?: number;
  initialPageSize?: number;
}

interface UsePaginationReturn {
  page: number;
  pageSize: number;
  totalPages: number;
  paginatedItems: any[];
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  goToFirstPage: () => void;
  goToLastPage: () => void;
  goToNextPage: () => void;
  goToPreviousPage: () => void;
  canGoNext: boolean;
  canGoPrevious: boolean;
}

export function usePagination<T>(
  items: T[],
  options: UsePaginationOptions
): UsePaginationReturn {
  const { totalItems, initialPage = 1, initialPageSize = 20 } = options;

  const [page, setPage] = useState(initialPage);
  const [pageSize, setPageSize] = useState(initialPageSize);

  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));

  // Page should not exceed totalPages
  const currentPage = Math.min(page, totalPages);

  const paginatedItems = useMemo(() => {
    const startIndex = (currentPage - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    return items.slice(startIndex, endIndex);
  }, [items, currentPage, pageSize]);

  const setPageSafe = (newPage: number) => {
    const clampedPage = Math.max(1, Math.min(newPage, totalPages));
    setPage(clampedPage);
  };

  const setPageSizeSafe = (newSize: number) => {
    setPageSize(newSize);
    // Adjust page if current page would be out of bounds
    const newTotalPages = Math.max(1, Math.ceil(totalItems / newSize));
    if (currentPage > newTotalPages) {
      setPage(newTotalPages);
    }
  };

  return {
    page: currentPage,
    pageSize,
    totalPages,
    paginatedItems,
    setPage: setPageSafe,
    setPageSize: setPageSizeSafe,
    goToFirstPage: () => setPageSafe(1),
    goToLastPage: () => setPageSafe(totalPages),
    goToNextPage: () => setPageSafe(currentPage + 1),
    goToPreviousPage: () => setPageSafe(currentPage - 1),
    canGoNext: currentPage < totalPages,
    canGoPrevious: currentPage > 1,
  };
}


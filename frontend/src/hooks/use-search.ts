import { useState, useMemo, useCallback } from 'react';
import { DataProcessor, type SearchOptions } from '@/lib/data';

interface UseSearchResult<T> {
  query: string;
  setQuery: (query: string) => void;
  results: T[];
  isSearching: boolean;
  clearSearch: () => void;
}

/**
 * useSearch Hook
 * DataProcessor를 내부적으로 사용하여 검색 기능을 제공하는 훅
 * 
 * @example
 * ```tsx
 * const { query, setQuery, results, isSearching, clearSearch } = useSearch(items, {
 *   keys: ['name', 'description'],
 *   threshold: 0.3,
 * });
 * ```
 */
export function useSearch<T>(
  data: T[],
  options: SearchOptions
): UseSearchResult<T> {
  const [query, setQuery] = useState('');

  const results = useMemo(() => {
    return DataProcessor.search(data, query, options);
  }, [data, query, options]);

  const isSearching = query.length > 0;

  const clearSearch = useCallback(() => {
    setQuery('');
  }, []);

  return {
    query,
    setQuery,
    results,
    isSearching,
    clearSearch,
  };
}


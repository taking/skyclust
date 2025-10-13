import { useState, useMemo } from 'react';
import Fuse from 'fuse.js';

interface SearchOptions {
  keys: string[];
  threshold?: number;
  includeScore?: boolean;
  includeMatches?: boolean;
  minMatchCharLength?: number;
  shouldSort?: boolean;
  sortFn?: (a: unknown, b: unknown) => number;
}

interface UseSearchResult<T> {
  query: string;
  setQuery: (query: string) => void;
  results: T[];
  isSearching: boolean;
  clearSearch: () => void;
}

export function useSearch<T>(
  data: T[],
  options: SearchOptions
): UseSearchResult<T> {
  const [query, setQuery] = useState('');

  const fuse = useMemo(() => {
    return new Fuse(data, {
      keys: options.keys,
      threshold: options.threshold ?? 0.3,
      includeScore: options.includeScore ?? true,
      includeMatches: options.includeMatches ?? false,
      minMatchCharLength: options.minMatchCharLength ?? 1,
      shouldSort: options.shouldSort ?? true,
      sortFn: options.sortFn,
    });
  }, [data, options]);

  const results = useMemo(() => {
    if (!query.trim()) {
      return data;
    }
    return fuse.search(query).map(result => result.item);
  }, [fuse, query, data]);

  const isSearching = query.length > 0;

  const clearSearch = () => {
    setQuery('');
  };

  return {
    query,
    setQuery,
    results,
    isSearching,
    clearSearch,
  };
}

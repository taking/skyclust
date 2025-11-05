/**
 * DataProcessor
 * 데이터 처리 로직을 중앙화한 클래스
 * Search, Filter, Sort 기능을 통합 제공
 */

import Fuse from 'fuse.js';
import type { SortConfig } from '@/hooks/use-advanced-filtering';
import type { FilterValue } from '@/components/ui/filter-panel';

export interface SearchOptions {
  keys: string[];
  threshold?: number;
  includeScore?: boolean;
  includeMatches?: boolean;
  minMatchCharLength?: number;
  shouldSort?: boolean;
}

export interface DataProcessingOptions<T> {
  searchQuery?: string;
  searchKeys?: string[];
  filters?: FilterValue;
  sortConfig?: SortConfig[];
  getValue?: (item: T, field: string) => unknown;
}

/**
 * DataProcessor 클래스
 * 데이터의 검색, 필터링, 정렬을 통합 처리
 */
export class DataProcessor {
  /**
   * Search: Fuse.js를 사용한 fuzzy search
   */
  static search<T>(
    items: T[],
    query: string,
    options: SearchOptions
  ): T[] {
    if (!query.trim()) {
      return items;
    }

    const fuse = new Fuse(items, {
      keys: options.keys,
      threshold: options.threshold ?? 0.3,
      includeScore: options.includeScore ?? true,
      includeMatches: options.includeMatches ?? false,
      minMatchCharLength: options.minMatchCharLength ?? 1,
      shouldSort: options.shouldSort ?? true,
    });

    return fuse.search(query).map(result => result.item);
  }

  /**
   * Filter: 필터 값에 따라 데이터 필터링
   */
  static filter<T>(
    items: T[],
    filters: FilterValue,
    filterFn?: (item: T, filters: FilterValue) => boolean
  ): T[] {
    if (!filters || Object.keys(filters).length === 0) {
      return items;
    }

    if (filterFn) {
      return items.filter(item => filterFn(item, filters));
    }

    // 기본 필터링 로직
    return items.filter((item: T) => {
      for (const [key, value] of Object.entries(filters)) {
        if (value === null || value === undefined || value === '') {
          continue;
        }

        const itemValue = item[key];

        // 배열 필터 (multiple values)
        if (Array.isArray(value)) {
          if (value.length === 0) continue;
          if (!value.includes(itemValue)) {
            return false;
          }
        }
        // 단일 값 필터
        else if (itemValue !== value) {
          return false;
        }
      }
      return true;
    });
  }

  /**
   * MultiSort: 여러 필드에 대한 정렬
   */
  static multiSort<T>(
    items: T[],
    sortConfig: SortConfig[],
    getValue: (item: T, field: string) => unknown
  ): T[] {
    if (sortConfig.length === 0) {
      return items;
    }

    return [...items].sort((a, b) => {
      for (const sort of sortConfig) {
        const aValue = getValue(a, sort.field);
        const bValue = getValue(b, sort.field);

        // Handle null/undefined values
        if (aValue == null && bValue == null) continue;
        if (aValue == null) return sort.direction === 'asc' ? 1 : -1;
        if (bValue == null) return sort.direction === 'asc' ? -1 : 1;

        // Compare values
        let comparison = 0;
        
        if (typeof aValue === 'string' && typeof bValue === 'string') {
          comparison = aValue.localeCompare(bValue, undefined, { numeric: true, sensitivity: 'base' });
        } else if (typeof aValue === 'number' && typeof bValue === 'number') {
          comparison = aValue - bValue;
        } else if (aValue instanceof Date && bValue instanceof Date) {
          comparison = aValue.getTime() - bValue.getTime();
        } else {
          // Fallback to string comparison
          comparison = String(aValue).localeCompare(String(bValue), undefined, { numeric: true });
        }

        if (comparison !== 0) {
          return sort.direction === 'asc' ? comparison : -comparison;
        }
      }

      return 0;
    });
  }

  /**
   * Combine: Search + Filter + Sort를 한 번에 처리
   */
  static combine<T>(
    items: T[],
    options: DataProcessingOptions<T>
  ): T[] {
    let result = [...items];

    // 1. Search
    if (options.searchQuery && options.searchKeys && options.searchKeys.length > 0) {
      result = this.search(result, options.searchQuery, {
        keys: options.searchKeys,
        threshold: 0.3,
      });
    }

    // 2. Filter
    if (options.filters && Object.keys(options.filters).length > 0) {
      result = this.filter(result, options.filters);
    }

    // 3. Sort
    if (options.sortConfig && options.sortConfig.length > 0 && options.getValue) {
      result = this.multiSort(result, options.sortConfig, options.getValue);
    }

    return result;
  }

  /**
   * Get sort indicator for a table header
   */
  static getSortIndicator(
    field: string,
    sortConfig: SortConfig[]
  ): 'asc' | 'desc' | null {
    const sort = sortConfig.find(s => s.field === field);
    return sort ? sort.direction : null;
  }
}


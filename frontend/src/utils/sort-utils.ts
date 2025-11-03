import { SortConfig } from '@/hooks/use-advanced-filtering';

/**
 * Sort an array of items based on multiple sort configurations
 */
export function multiSort<T>(
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
 * Get sort indicator for a table header
 */
export function getSortIndicator(
  field: string,
  sortConfig: SortConfig[]
): 'asc' | 'desc' | null {
  const sort = sortConfig.find(s => s.field === field);
  return sort ? sort.direction : null;
}


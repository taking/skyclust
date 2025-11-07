import { useMemo } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';

interface UseVirtualListOptions<T> {
  items: T[];
  parentRef: React.RefObject<HTMLElement>;
  estimateSize?: number;
  overscan?: number;
  enabled?: boolean;
  minItems?: number; // Minimum items to enable virtualization
}

export function useVirtualList<T>({
  items,
  parentRef,
  estimateSize = 50,
  overscan = 5,
  enabled = true,
  minItems = 100,
}: UseVirtualListOptions<T>) {
  const shouldVirtualize = enabled && items.length >= minItems;

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => estimateSize,
    overscan,
    enabled: shouldVirtualize,
  });

  const virtualItems = useMemo(() => {
    if (!shouldVirtualize) {
      return items.map((item, index) => ({
        index,
        item,
        start: index * estimateSize,
        size: estimateSize,
      }));
    }
    return virtualizer.getVirtualItems();
  }, [shouldVirtualize, items, virtualizer, estimateSize]);

  return {
    virtualizer,
    virtualItems,
    totalSize: virtualizer.getTotalSize(),
    shouldVirtualize,
  };
}


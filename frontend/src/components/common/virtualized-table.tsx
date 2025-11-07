'use client';

import * as React from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { cn } from '@/lib/utils';

interface VirtualizedTableProps<T> {
  data: T[];
  renderHeader: () => React.ReactNode;
  renderRow: (item: T, index: number) => React.ReactNode;
  estimateSize?: number;
  overscan?: number;
  minItems?: number; // Minimum items to enable virtualization
  containerHeight?: string;
  className?: string;
}

function VirtualizedTableComponent<T extends { id?: string; [key: string]: unknown }>({
  data,
  renderHeader,
  renderRow,
  estimateSize = 60,
  overscan = 5,
  minItems = 50, // Lower threshold since pagination handles chunking
  containerHeight = '600px',
  className,
}: VirtualizedTableProps<T>) {
  const parentRef = React.useRef<HTMLDivElement>(null);
  const shouldVirtualize = data.length >= minItems;

  const virtualizer = useVirtualizer({
    count: data.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => estimateSize,
    overscan,
    enabled: shouldVirtualize,
  });

  if (!shouldVirtualize) {
    // Render normal table if below threshold
    return (
      <div className={cn('overflow-auto', className)}>
        <Table>
          <TableHeader>
            {renderHeader()}
          </TableHeader>
          <TableBody>
            {data.map((item, index) => (
              <React.Fragment key={item.id || index}>
                {renderRow(item, index)}
              </React.Fragment>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  }

  const items = virtualizer.getVirtualItems();

  return (
    <div
      ref={parentRef}
      className={cn('overflow-auto', className)}
      style={{ height: containerHeight, contain: 'strict' }}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        <Table>
          <TableHeader className="sticky top-0 z-10 bg-background shadow-sm">
            {renderHeader()}
          </TableHeader>
          <TableBody>
            <div
              style={{
                position: 'relative',
                height: `${virtualizer.getTotalSize()}px`,
                width: '100%',
              }}
            >
              {items.map((virtualItem) => {
                const item = data[virtualItem.index];
                return (
                  <div
                    key={item.id || virtualItem.key}
                    data-index={virtualItem.index}
                    ref={virtualizer.measureElement}
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      width: '100%',
                      transform: `translateY(${virtualItem.start}px)`,
                    }}
                  >
                    {renderRow(item, virtualItem.index)}
                  </div>
                );
              })}
            </div>
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

// React.memo로 최적화: 데이터가 변경되지 않으면 리렌더링 방지
export const VirtualizedTable = React.memo(VirtualizedTableComponent) as typeof VirtualizedTableComponent;


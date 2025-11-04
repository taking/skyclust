'use client';

import * as React from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useVirtualizer } from '@tanstack/react-virtual';
import { cn } from '@/lib/utils';

interface VirtualTableWrapperProps<T> {
  data: T[];
  renderRow: (item: T, index: number) => React.ReactNode;
  renderHeader?: () => React.ReactNode;
  estimateSize?: number;
  overscan?: number;
  className?: string;
  minItems?: number; // Minimum items to enable virtualization
  containerHeight?: string; // Height of the container (e.g., '500px', '100%')
}

export function VirtualTableWrapper<T extends { id?: string; [key: string]: unknown }>({
  data,
  renderRow,
  renderHeader,
  estimateSize = 50,
  overscan = 5,
  className,
  minItems = 100,
  containerHeight = '600px',
}: VirtualTableWrapperProps<T>) {
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
          {renderHeader && (
            <TableHeader>
              {renderHeader()}
            </TableHeader>
          )}
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
          {renderHeader && (
            <TableHeader className="sticky top-0 z-10 bg-background">
              {renderHeader()}
            </TableHeader>
          )}
          <TableBody>
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
          </TableBody>
        </Table>
      </div>
    </div>
  );
}


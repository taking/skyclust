'use client';

import * as React from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { cn } from '@/lib/utils';

interface VirtualTableProps<T> {
  data: T[];
  columns: Array<{
    key: string;
    header: React.ReactNode;
    cell: (item: T, index: number) => React.ReactNode;
    className?: string;
  }>;
  estimateSize?: number;
  overscan?: number;
  className?: string;
  containerRef?: React.RefObject<HTMLDivElement>;
}

export function VirtualTable<T extends { id?: string; [key: string]: unknown }>({
  data,
  columns,
  estimateSize = 50,
  overscan = 5,
  className,
  containerRef,
}: VirtualTableProps<T>) {
  const parentRef = React.useRef<HTMLDivElement>(null);
  const ref = containerRef || parentRef;

  const virtualizer = useVirtualizer({
    count: data.length,
    getScrollElement: () => ref.current,
    estimateSize: () => estimateSize,
    overscan,
  });

  const items = virtualizer.getVirtualItems();

  return (
    <div
      ref={ref}
      className={cn('h-full overflow-auto', className)}
      style={{ contain: 'strict' }}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        <Table>
          <TableHeader className="sticky top-0 z-10 bg-background">
            <TableRow>
              {columns.map((column) => (
                <TableHead key={column.key} className={column.className}>
                  {column.header}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody
            style={{
              height: `${virtualizer.getTotalSize()}px`,
              width: '100%',
              position: 'relative',
            }}
          >
            {items.map((virtualItem) => {
              const item = data[virtualItem.index];
              return (
                <TableRow
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
                  {columns.map((column) => (
                    <TableCell key={column.key} className={column.className}>
                      {column.cell(item, virtualItem.index)}
                    </TableCell>
                  ))}
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}


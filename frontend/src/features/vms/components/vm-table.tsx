/**
 * VM Table Component
 * Virtual Machine 목록 테이블
 */

'use client';

import * as React from 'react';
import { useMemo, useCallback } from 'react';
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Pagination } from '@/components/ui/pagination';
import { ArrowUp, ArrowDown } from 'lucide-react';
import { VMRow } from './vm-row';
import { DataProcessor } from '@/lib/data-processor';
import type { VM } from '@/lib/types';

interface VMTableProps {
  vms: VM[];
  sortConfig: Array<{ field: string; direction: 'asc' | 'desc' }>;
  onToggleSort: (field: string) => void;
  onStart: (vmId: string) => void;
  onStop: (vmId: string) => void;
  onDelete: (vmId: string) => void;
  isStarting?: boolean;
  isStopping?: boolean;
  isDeleting?: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
}

function VMTableComponent({
  vms,
  sortConfig,
  onToggleSort,
  onStart,
  onStop,
  onDelete,
  isStarting = false,
  isStopping = false,
  isDeleting = false,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
}: VMTableProps) {
  // 정렬 인디케이터 메모이제이션
  const sortIndicators = useMemo(() => ({
    name: DataProcessor.getSortIndicator('name', sortConfig),
    provider: DataProcessor.getSortIndicator('provider', sortConfig),
    instance_type: DataProcessor.getSortIndicator('instance_type', sortConfig),
    region: DataProcessor.getSortIndicator('region', sortConfig),
    status: DataProcessor.getSortIndicator('status', sortConfig),
    created_at: DataProcessor.getSortIndicator('created_at', sortConfig),
  }), [sortConfig]);

  // 정렬 핸들러 메모이제이션
  const handleSortName = useCallback(() => onToggleSort('name'), [onToggleSort]);
  const handleSortProvider = useCallback(() => onToggleSort('provider'), [onToggleSort]);
  const handleSortInstanceType = useCallback(() => onToggleSort('instance_type'), [onToggleSort]);
  const handleSortRegion = useCallback(() => onToggleSort('region'), [onToggleSort]);
  const handleSortStatus = useCallback(() => onToggleSort('status'), [onToggleSort]);
  const handleSortCreatedAt = useCallback(() => onToggleSort('created_at'), [onToggleSort]);

  // VM별 콜백 함수 메모이제이션 (Map 사용)
  const vmCallbacks = useMemo(() => {
    const callbacks = new Map<string, { onStart: () => void; onStop: () => void; onDelete: () => void }>();
    vms.forEach((vm) => {
      callbacks.set(vm.id, {
        onStart: () => onStart(vm.id),
        onStop: () => onStop(vm.id),
        onDelete: () => onDelete(vm.id),
      });
    });
    return callbacks;
  }, [vms, onStart, onStop, onDelete]);

  return (
    <>
      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <Table role="table" aria-label="Virtual machines list">
            <TableHeader>
              <TableRow>
                <TableHead className="min-w-[120px]">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={handleSortName}
                  >
                    Name
                    {sortIndicators.name === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {sortIndicators.name === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="hidden sm:table-cell">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={handleSortProvider}
                  >
                    Provider
                    {sortIndicators.provider === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {sortIndicators.provider === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="hidden md:table-cell">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={handleSortInstanceType}
                  >
                    Type
                    {sortIndicators.instance_type === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {sortIndicators.instance_type === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="hidden lg:table-cell">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={handleSortRegion}
                  >
                    Region
                    {sortIndicators.region === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {sortIndicators.region === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="min-w-[80px]">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={handleSortStatus}
                  >
                    Status
                    {sortIndicators.status === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {sortIndicators.status === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="hidden xl:table-cell">IP Address</TableHead>
                <TableHead className="hidden lg:table-cell">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={handleSortCreatedAt}
                  >
                    Created
                    {sortIndicators.created_at === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {sortIndicators.created_at === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="text-right min-w-[100px]">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {vms.map((vm) => {
                const callbacks = vmCallbacks.get(vm.id);
                if (!callbacks) return null;
                
                return (
                  <TableRow key={vm.id} role="row">
                    <VMRow
                      vm={vm}
                      onStart={callbacks.onStart}
                      onStop={callbacks.onStop}
                      onDelete={callbacks.onDelete}
                      isStarting={isStarting}
                      isStopping={isStopping}
                      isDeleting={isDeleting}
                    />
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      </div>
      
      {/* Pagination */}
      {total > 0 && (
        <div className="border-t mt-4">
          <Pagination
            total={total}
            page={page}
            pageSize={pageSize}
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            pageSizeOptions={[10, 20, 50, 100]}
            showPageSizeSelector={true}
          />
        </div>
      )}
    </>
  );
}

export const VMTable = React.memo(VMTableComponent);


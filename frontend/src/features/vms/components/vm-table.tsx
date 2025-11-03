/**
 * VM Table Component
 * Virtual Machine 목록 테이블
 */

'use client';

import * as React from 'react';
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Pagination } from '@/components/ui/pagination';
import { ArrowUp, ArrowDown } from 'lucide-react';
import { VMRow } from './vm-row';
import { getSortIndicator } from '@/utils/sort-utils';
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
                    onClick={() => onToggleSort('name')}
                  >
                    Name
                    {getSortIndicator('name', sortConfig) === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {getSortIndicator('name', sortConfig) === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="hidden sm:table-cell">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => onToggleSort('provider')}
                  >
                    Provider
                    {getSortIndicator('provider', sortConfig) === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {getSortIndicator('provider', sortConfig) === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="hidden md:table-cell">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => onToggleSort('instance_type')}
                  >
                    Type
                    {getSortIndicator('instance_type', sortConfig) === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {getSortIndicator('instance_type', sortConfig) === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="hidden lg:table-cell">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => onToggleSort('region')}
                  >
                    Region
                    {getSortIndicator('region', sortConfig) === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {getSortIndicator('region', sortConfig) === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="min-w-[80px]">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => onToggleSort('status')}
                  >
                    Status
                    {getSortIndicator('status', sortConfig) === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {getSortIndicator('status', sortConfig) === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="hidden xl:table-cell">IP Address</TableHead>
                <TableHead className="hidden lg:table-cell">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => onToggleSort('created_at')}
                  >
                    Created
                    {getSortIndicator('created_at', sortConfig) === 'asc' && <ArrowUp className="ml-1 h-3 w-3" />}
                    {getSortIndicator('created_at', sortConfig) === 'desc' && <ArrowDown className="ml-1 h-3 w-3" />}
                  </Button>
                </TableHead>
                <TableHead className="text-right min-w-[100px]">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {vms.map((vm) => (
                <TableRow key={vm.id} role="row">
                  <VMRow
                    vm={vm}
                    onStart={() => onStart(vm.id)}
                    onStop={() => onStop(vm.id)}
                    onDelete={() => onDelete(vm.id)}
                    isStarting={isStarting}
                    isStopping={isStopping}
                    isDeleting={isDeleting}
                  />
                </TableRow>
              ))}
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


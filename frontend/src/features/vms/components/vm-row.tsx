/**
 * VM Row Component
 * 개별 VM 행 렌더링
 */

'use client';

import * as React from 'react';
import { TableCell } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Play, Square, Trash2, ExternalLink } from 'lucide-react';
import { ScreenReaderOnly } from '@/components/accessibility/screen-reader-only';
import { getStatusAriaLabel, getActionAriaLabel } from '@/lib/accessibility';
import type { VM } from '@/lib/types';

interface VMRowProps {
  vm: VM;
  onStart: () => void;
  onStop: () => void;
  onDelete: () => void;
  isStarting?: boolean;
  isStopping?: boolean;
  isDeleting?: boolean;
}

function getStatusBadgeVariant(status: string): 'default' | 'secondary' | 'outline' {
  switch (status.toLowerCase()) {
    case 'running':
      return 'default';
    case 'stopped':
      return 'secondary';
    case 'pending':
      return 'outline';
    default:
      return 'outline';
  }
}

function VMRowComponent({
  vm,
  onStart,
  onStop,
  onDelete,
  isStarting = false,
  isStopping = false,
  isDeleting = false,
}: VMRowProps) {
  return (
    <>
      <TableCell className="font-medium" role="cell">
        <div>
          <div className="font-medium">{vm.name}</div>
          <div className="text-sm text-gray-500 sm:hidden">
            {vm.provider} • {vm.instance_type}
            <ScreenReaderOnly>
              Provider: {vm.provider}, Instance type: {vm.instance_type}
            </ScreenReaderOnly>
          </div>
        </div>
      </TableCell>
      <TableCell className="hidden sm:table-cell" role="cell">
        <Badge variant="outline" aria-label={`Provider: ${vm.provider}`}>
          {vm.provider}
        </Badge>
      </TableCell>
      <TableCell className="hidden md:table-cell" role="cell">
        {vm.instance_type}
      </TableCell>
      <TableCell className="hidden lg:table-cell" role="cell">
        {vm.region}
      </TableCell>
      <TableCell role="cell">
        <Badge 
          variant={getStatusBadgeVariant(vm.status)}
          aria-label={`Status: ${getStatusAriaLabel(vm.status)}`}
        >
          {vm.status}
        </Badge>
      </TableCell>
      <TableCell className="hidden xl:table-cell" role="cell">
        {vm.public_ip ? (
          <div className="flex items-center space-x-1">
            <span>{vm.public_ip}</span>
            <ExternalLink className="h-3 w-3 text-gray-400" aria-hidden="true" />
            <ScreenReaderOnly>External link</ScreenReaderOnly>
          </div>
        ) : (
          <span className="text-gray-400" aria-label="No IP address">-</span>
        )}
      </TableCell>
      <TableCell className="hidden lg:table-cell" role="cell">
        {new Date(vm.created_at).toLocaleDateString()}
      </TableCell>
      <TableCell className="text-right" role="cell">
        <div className="flex items-center justify-end space-x-1" role="group" aria-label={`Actions for ${vm.name}`}>
          {vm.status === 'running' ? (
            <Button
              variant="outline"
              size="sm"
              onClick={onStop}
              disabled={isStopping}
              className="h-8 w-8 p-0"
              aria-label={getActionAriaLabel('stop', vm.name)}
            >
              <Square className="h-4 w-4" aria-hidden="true" />
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              onClick={onStart}
              disabled={isStarting}
              className="h-8 w-8 p-0"
              aria-label={getActionAriaLabel('start', vm.name)}
            >
              <Play className="h-4 w-4" aria-hidden="true" />
            </Button>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={onDelete}
            disabled={isDeleting}
            className="h-8 w-8 p-0"
            aria-label={getActionAriaLabel('delete', vm.name)}
          >
            <Trash2 className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>
      </TableCell>
    </>
  );
}

export const VMRow = React.memo(VMRowComponent, (prevProps, nextProps) => {
  // Custom comparison for better memoization
  return (
    prevProps.vm.id === nextProps.vm.id &&
    prevProps.vm.status === nextProps.vm.status &&
    prevProps.isStarting === nextProps.isStarting &&
    prevProps.isStopping === nextProps.isStopping &&
    prevProps.isDeleting === nextProps.isDeleting
  );
});


/**
 * VM Empty State Component
 * VM이 없을 때 표시되는 빈 상태 컴포넌트
 */

'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import { Server, Plus } from 'lucide-react';

interface VMEmptyStateProps {
  isSearching: boolean;
  searchQuery?: string;
  onCreateClick?: () => void;
}

function VMEmptyStateComponent({
  isSearching,
  searchQuery,
  onCreateClick,
}: VMEmptyStateProps) {
  return (
    <div className="text-center py-12">
      <div className="mx-auto h-12 w-12 text-gray-400">
        <Server className="h-12 w-12" />
      </div>
      <h3 className="mt-2 text-sm font-medium text-gray-900">
        {isSearching ? 'No VMs found' : 'No VMs'}
      </h3>
      <p className="mt-1 text-sm text-gray-500">
        {isSearching 
          ? 'Try adjusting your search or filter criteria.'
          : 'Get started by creating your first virtual machine.'
        }
        {searchQuery && ` (${searchQuery})`}
      </p>
      {onCreateClick && (
        <div className="mt-6">
          <Button onClick={onCreateClick}>
            <Plus className="mr-2 h-4 w-4" />
            Create VM
          </Button>
        </div>
      )}
    </div>
  );
}

export const VMEmptyState = React.memo(VMEmptyStateComponent);


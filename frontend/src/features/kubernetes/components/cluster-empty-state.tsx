/**
 * Cluster Empty State Component
 * 클러스터가 없을 때 표시되는 빈 상태 컴포넌트
 */

'use client';

import * as React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Server, Plus } from 'lucide-react';

interface ClusterEmptyStateProps {
  title: string;
  description: string;
  onCreateClick?: () => void;
  showCreateButton?: boolean;
}

function ClusterEmptyStateComponent({
  title,
  description,
  onCreateClick,
  showCreateButton = false,
}: ClusterEmptyStateProps) {
  return (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Server className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">{title}</h3>
        <p className="text-sm text-gray-500 text-center mb-4">{description}</p>
        {showCreateButton && onCreateClick && (
          <Button onClick={onCreateClick}>
            <Plus className="mr-2 h-4 w-4" />
            Create Cluster
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

export const ClusterEmptyState = React.memo(ClusterEmptyStateComponent);


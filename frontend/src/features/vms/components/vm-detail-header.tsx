/**
 * VM Detail Header Component
 * VM 상세 페이지의 헤더 컴포넌트
 */

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, Trash2 } from 'lucide-react';
import type { VM } from '@/lib/types';

interface VMDetailHeaderProps {
  vm: VM;
  onBack: () => void;
  onDelete: () => void;
  isDeleting: boolean;
}

export function VMDetailHeader({ vm, onBack, onDelete, isDeleting }: VMDetailHeaderProps) {
  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
        return 'bg-green-100 text-green-800';
      case 'stopped':
        return 'bg-red-100 text-red-800';
      case 'starting':
        return 'bg-yellow-100 text-yellow-800';
      case 'stopping':
        return 'bg-orange-100 text-orange-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center space-x-4">
        <Button
          variant="outline"
          size="sm"
          onClick={onBack}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back
        </Button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{vm.name}</h1>
          <p className="text-gray-600">VM Details and Management</p>
        </div>
      </div>
      <div className="flex items-center space-x-2">
        <Badge className={getStatusColor(vm.status)}>
          {vm.status}
        </Badge>
        <Button
          variant="outline"
          size="sm"
          onClick={onDelete}
          disabled={isDeleting}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </Button>
      </div>
    </div>
  );
}


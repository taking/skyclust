'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ArrowUpDown, ArrowUp, ArrowDown, X } from 'lucide-react';
import { SortConfig } from '@/hooks/use-advanced-filtering';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuLabel,
} from '@/components/ui/dropdown-menu';

interface MultiSortIndicatorProps {
  sortConfig: SortConfig[];
  availableFields: Array<{ value: string; label: string }>;
  onToggleSort: (field: string) => void;
  onClearSort: () => void;
  className?: string;
}

function MultiSortIndicatorComponent({
  sortConfig,
  availableFields,
  onToggleSort,
  onClearSort,
  className,
}: MultiSortIndicatorProps) {
  const getSortIcon = (direction: 'asc' | 'desc') => {
    return direction === 'asc' ? (
      <ArrowUp className="h-3 w-3" />
    ) : (
      <ArrowDown className="h-3 w-3" />
    );
  };

  const getFieldLabel = (field: string) => {
    return availableFields.find(f => f.value === field)?.label || field;
  };

  return (
    <div className={`flex items-center gap-2 ${className || ''}`}>
      {sortConfig.length > 0 && (
        <>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <ArrowUpDown className="mr-2 h-4 w-4" />
                Sort
                {sortConfig.length > 0 && (
                  <Badge variant="secondary" className="ml-2">
                    {sortConfig.length}
                  </Badge>
                )}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel>Sort Order</DropdownMenuLabel>
              <DropdownMenuSeparator />
              {sortConfig.map((sort, index) => (
                <DropdownMenuItem
                  key={`${sort.field}-${index}`}
                  onClick={() => onToggleSort(sort.field)}
                  className="flex items-center justify-between"
                >
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-gray-500">{index + 1}.</span>
                    <span>{getFieldLabel(sort.field)}</span>
                  </div>
                  {getSortIcon(sort.direction)}
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />
              {availableFields
                .filter(f => !sortConfig.some(s => s.field === f.value))
                .map((field) => (
                  <DropdownMenuItem
                    key={field.value}
                    onClick={() => onToggleSort(field.value)}
                  >
                    Add: {field.label}
                  </DropdownMenuItem>
                ))}
              {sortConfig.length > 0 && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={onClearSort} className="text-red-600">
                    <X className="mr-2 h-4 w-4" />
                    Clear all
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
          
          <div className="flex items-center gap-1 flex-wrap">
            {sortConfig.map((sort, index) => (
              <Badge
                key={`${sort.field}-${index}`}
                variant="secondary"
                className="flex items-center gap-1"
              >
                <span className="text-xs">{index + 1}.</span>
                <span>{getFieldLabel(sort.field)}</span>
                {getSortIcon(sort.direction)}
                <button
                  onClick={() => onToggleSort(sort.field)}
                  className="ml-1 hover:text-destructive"
                  aria-label={`Remove sort by ${getFieldLabel(sort.field)}`}
                >
                  <X className="h-3 w-3" />
                </button>
              </Badge>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

export const MultiSortIndicator = React.memo(MultiSortIndicatorComponent);


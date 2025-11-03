'use client';

import * as React from 'react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Badge } from '@/components/ui/badge';
import { Trash2, Tag, MoreVertical, X } from 'lucide-react';

interface BulkActionsToolbarProps<T extends { id: string }> {
  items: T[];
  selectedIds: string[];
  onSelectionChange: (ids: string[]) => void;
  onBulkDelete?: (ids: string[]) => void;
  onBulkTag?: (ids: string[]) => void;
  getItemDisplayName?: (item: T) => string;
}

function BulkActionsToolbarComponent<T extends { id: string }>({
  items,
  selectedIds,
  onSelectionChange,
  onBulkDelete,
  onBulkTag,
  getItemDisplayName = (item: T) => {
    const itemWithName = item as T & { name?: string };
    return itemWithName.name || item.id;
  },
}: BulkActionsToolbarProps<T>) {
  const [isSelectMode, setIsSelectMode] = useState(false);

  const toggleSelectMode = () => {
    setIsSelectMode(!isSelectMode);
    if (isSelectMode) {
      onSelectionChange([]);
    }
  };

  const toggleSelectAll = () => {
    if (selectedIds.length === items.length) {
      onSelectionChange([]);
    } else {
      onSelectionChange(items.map(item => item.id));
    }
  };

  const toggleSelectItem = (id: string) => {
    if (selectedIds.includes(id)) {
      onSelectionChange(selectedIds.filter(selectedId => selectedId !== id));
    } else {
      onSelectionChange([...selectedIds, id]);
    }
  };

  const handleBulkDelete = () => {
    if (onBulkDelete && selectedIds.length > 0) {
      if (confirm(`Are you sure you want to delete ${selectedIds.length} item(s)? This action cannot be undone.`)) {
        onBulkDelete(selectedIds);
        onSelectionChange([]);
        setIsSelectMode(false);
      }
    }
  };

  const handleBulkTag = () => {
    if (onBulkTag && selectedIds.length > 0) {
      onBulkTag(selectedIds);
    }
  };

  if (!isSelectMode) {
    return (
      <Button variant="outline" onClick={toggleSelectMode}>
        Select Items
      </Button>
    );
  }

  return (
    <div className="flex items-center justify-between p-4 bg-blue-50 border border-blue-200 rounded-lg">
      <div className="flex items-center space-x-4">
        <Button variant="ghost" size="sm" onClick={toggleSelectMode}>
          <X className="h-4 w-4 mr-2" />
          Cancel
        </Button>
        <Checkbox
          checked={selectedIds.length === items.length && items.length > 0}
          onCheckedChange={toggleSelectAll}
          id="select-all"
        />
        <label htmlFor="select-all" className="text-sm font-medium cursor-pointer">
          Select All ({items.length})
        </label>
        <Badge variant="secondary" className="ml-2">
          {selectedIds.length} selected
        </Badge>
      </div>

      <div className="flex items-center space-x-2">
        {selectedIds.length > 0 && (
          <>
            {onBulkDelete && (
              <Button
                variant="destructive"
                size="sm"
                onClick={handleBulkDelete}
                disabled={selectedIds.length === 0}
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete ({selectedIds.length})
              </Button>
            )}
            {onBulkTag && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleBulkTag}
                disabled={selectedIds.length === 0}
              >
                <Tag className="h-4 w-4 mr-2" />
                Tag ({selectedIds.length})
              </Button>
            )}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>More Actions</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleBulkTag} disabled={!onBulkTag || selectedIds.length === 0}>
                  <Tag className="mr-2 h-4 w-4" />
                  Add Tags
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </>
        )}
      </div>
    </div>
  );
}

export const BulkActionsToolbar = React.memo(BulkActionsToolbarComponent) as typeof BulkActionsToolbarComponent;

export function SelectableItem<T extends { id: string }>({
  item,
  isSelected,
  onToggle,
  children,
}: {
  item: T;
  isSelected: boolean;
  onToggle: () => void;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center space-x-2">
      <Checkbox
        checked={isSelected}
        onCheckedChange={onToggle}
        id={`select-${item.id}`}
      />
      <label htmlFor={`select-${item.id}`} className="flex-1 cursor-pointer">
        {children}
      </label>
    </div>
  );
}


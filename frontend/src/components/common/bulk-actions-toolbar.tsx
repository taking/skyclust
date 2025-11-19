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
import { cn } from '@/lib/utils';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';

interface BulkActionsToolbarProps<T extends { id: string }> {
  items: T[];
  selectedIds: string[];
  onSelectionChange: (ids: string[]) => void;
  onBulkDelete?: (ids: string[]) => void;
  onBulkTag?: (ids: string[]) => void;
  getItemDisplayName?: (item: T) => string;
  /**
   * 고정 위치로 표시할지 여부 (기본값: false)
   */
  fixed?: boolean;
  /**
   * 항상 표시할지 여부 (기본값: false, 선택된 항목이 있을 때만 표시)
   */
  alwaysVisible?: boolean;
}

function BulkActionsToolbarComponent<T extends { id: string }>({
  items,
  selectedIds,
  onSelectionChange,
  onBulkDelete,
  onBulkTag,
  fixed = false,
  alwaysVisible = false,
}: BulkActionsToolbarProps<T>) {
  const hasSelection = selectedIds.length > 0;
  const showToolbar = alwaysVisible || hasSelection;
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  const toggleSelectAll = () => {
    if (selectedIds.length === items.length) {
      onSelectionChange([]);
    } else {
      onSelectionChange(items.map(item => item.id));
    }
  };

  const handleClearSelection = () => {
    onSelectionChange([]);
  };

  const handleBulkDelete = () => {
    if (onBulkDelete && selectedIds.length > 0) {
      setIsDeleteDialogOpen(true);
    }
  };

  const handleConfirmBulkDelete = () => {
    if (onBulkDelete && selectedIds.length > 0) {
      onBulkDelete(selectedIds);
      onSelectionChange([]);
      setIsDeleteDialogOpen(false);
    }
  };

  const handleBulkTag = () => {
    if (onBulkTag && selectedIds.length > 0) {
      onBulkTag(selectedIds);
    }
  };

  if (!showToolbar) {
    return null;
  }

  const toolbarContent = (
    <div className={cn(
      "flex items-center justify-between p-3 bg-primary/5 border border-primary/20 rounded-lg shadow-sm",
      fixed && "sticky top-0 z-50 backdrop-blur-sm bg-primary/10"
    )}>
      <div className="flex items-center space-x-4">
        {hasSelection && (
          <>
            <Button 
              variant="ghost" 
              size="sm" 
              onClick={handleClearSelection}
              aria-label="Clear selection"
            >
              <X className="h-4 w-4 mr-2" aria-hidden="true" />
              Clear
            </Button>
            <Checkbox
              checked={selectedIds.length === items.length && items.length > 0}
              onCheckedChange={toggleSelectAll}
              id="select-all"
              aria-label={`Select all ${items.length} items`}
            />
            <label htmlFor="select-all" className="text-sm font-medium cursor-pointer">
              Select All ({items.length})
            </label>
            <Badge variant="secondary" className="ml-2">
              {selectedIds.length} selected
            </Badge>
          </>
        )}
        {!hasSelection && alwaysVisible && (
          <span className="text-sm text-muted-foreground">
            Select items to perform bulk actions
          </span>
        )}
      </div>

      {hasSelection && (
        <div className="flex items-center space-x-2">
          {onBulkDelete && (
            <Button
              variant="destructive"
              size="sm"
              onClick={handleBulkDelete}
              aria-label={`Delete ${selectedIds.length} selected items`}
            >
              <Trash2 className="h-4 w-4 mr-2" aria-hidden="true" />
              Delete ({selectedIds.length})
            </Button>
          )}
          {onBulkTag && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleBulkTag}
              aria-label={`Tag ${selectedIds.length} selected items`}
            >
              <Tag className="h-4 w-4 mr-2" aria-hidden="true" />
              Tag ({selectedIds.length})
            </Button>
          )}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" aria-label="More actions">
                <MoreVertical className="h-4 w-4" aria-hidden="true" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>More Actions</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleBulkTag} disabled={!onBulkTag}>
                <Tag className="mr-2 h-4 w-4" />
                Add Tags
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}
    </div>
  );

  const toolbarWithDialog = (
    <>
      {fixed ? (
        <div className="sticky top-0 z-50 mb-4">
          {toolbarContent}
        </div>
      ) : (
        toolbarContent
      )}
      <DeleteConfirmationDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        onConfirm={handleConfirmBulkDelete}
        title="일괄 삭제 확인"
        description={`선택한 ${selectedIds.length}개 항목을 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다.`}
      />
    </>
  );

  return toolbarWithDialog;
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


/**
 * Resource Group Select Component
 * Azure Resource Group 선택 컴포넌트 (Virtual Scrolling + 검색 기능)
 */

'use client';

import * as React from 'react';
import { useState, useMemo, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Search, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { DataProcessor } from '@/lib/data';
import type { ResourceGroupInfo } from '@/services/resource-group';

interface ResourceGroupSelectProps {
  resourceGroups: ResourceGroupInfo[];
  selectedResourceGroup: string | null;
  onValueChange: (value: string | null) => void;
  isLoading?: boolean;
  error?: Error | null;
}

const ITEM_HEIGHT = 36;
const VISIBLE_ITEMS = 8;
const CONTAINER_HEIGHT = ITEM_HEIGHT * VISIBLE_ITEMS;

export function ResourceGroupSelect({
  resourceGroups,
  selectedResourceGroup,
  onValueChange,
  isLoading = false,
  error,
}: ResourceGroupSelectProps) {
  const [open, setOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const scrollElementRef = useRef<HTMLDivElement>(null);

  // 검색 필터링
  const filteredResourceGroups = useMemo(() => {
    if (!searchQuery.trim()) {
      return resourceGroups;
    }
    return DataProcessor.search(resourceGroups, searchQuery, {
      keys: ['name', 'location'],
      threshold: 0.3,
    });
  }, [resourceGroups, searchQuery]);

  // Virtual Scrolling (50개 이상일 때만 활성화)
  // "All Resource Groups" 옵션을 포함하여 계산
  const shouldVirtualize = filteredResourceGroups.length >= 50;
  const virtualizer = useVirtualizer({
    count: filteredResourceGroups.length,
    getScrollElement: () => scrollElementRef.current,
    estimateSize: () => ITEM_HEIGHT,
    overscan: 5,
    enabled: shouldVirtualize,
  });

  const virtualItems = shouldVirtualize ? virtualizer.getVirtualItems() : null;

  const handleValueChange = (value: string) => {
    if (value === 'all') {
      onValueChange(null);
    } else {
      onValueChange(value);
    }
    setOpen(false);
    setSearchQuery('');
  };

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen);
    if (!newOpen) {
      setSearchQuery('');
    }
  };

  return (
    <Select
      value={selectedResourceGroup || 'all'}
      onValueChange={handleValueChange}
      open={open}
      onOpenChange={handleOpenChange}
    >
      <SelectTrigger className="w-full">
        <SelectValue>
          {selectedResourceGroup ? (
            <span>{selectedResourceGroup}</span>
          ) : (
            'All Resource Groups'
          )}
        </SelectValue>
      </SelectTrigger>
      <SelectContent className="p-0" onOpenAutoFocus={(e) => e.preventDefault()}>
        {/* 검색 입력 필드 */}
        <div className="p-2 border-b sticky top-0 bg-background z-10">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search resource groups..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8 pr-8 h-8"
              onClick={(e) => e.stopPropagation()}
              onKeyDown={(e) => e.stopPropagation()}
            />
            {searchQuery && (
              <Button
                variant="ghost"
                size="sm"
                className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6 p-0"
                onClick={(e) => {
                  e.stopPropagation();
                  setSearchQuery('');
                }}
              >
                <X className="h-3 w-3" />
              </Button>
            )}
          </div>
        </div>

        {/* 로딩 상태 */}
        {isLoading && (
          <div className="p-4 text-center text-sm text-muted-foreground">
            Loading resource groups...
          </div>
        )}

        {/* 에러 상태 */}
        {error && !isLoading && (
          <div className="p-4 text-center text-sm text-destructive">
            Failed to load resource groups
          </div>
        )}

        {/* 빈 상태 */}
        {!isLoading && !error && filteredResourceGroups.length === 0 && (
          <div className="p-4 text-center text-sm text-muted-foreground">
            {searchQuery ? 'No resource groups found' : 'No resource groups available'}
          </div>
        )}

        {/* 리스트 */}
        {!isLoading && !error && filteredResourceGroups.length > 0 && (
          <div
            ref={scrollElementRef}
            className={cn(
              'overflow-auto',
              shouldVirtualize && 'h-[300px]'
            )}
            style={{
              height: shouldVirtualize ? CONTAINER_HEIGHT : 'auto',
              maxHeight: shouldVirtualize ? CONTAINER_HEIGHT : '300px',
            }}
          >
            {shouldVirtualize ? (
              <div
                style={{
                  height: `${virtualizer.getTotalSize() + ITEM_HEIGHT}px`,
                  width: '100%',
                  position: 'relative',
                }}
              >
                {/* "All Resource Groups" 옵션 */}
                <div
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    height: `${ITEM_HEIGHT}px`,
                  }}
                >
                  <SelectItem value="all" className="font-medium">
                    All Resource Groups
                  </SelectItem>
                </div>
                {/* Virtualized items */}
                {virtualItems?.map((virtualItem) => {
                  const rg = filteredResourceGroups[virtualItem.index];
                  const isSelected = selectedResourceGroup === rg.name;

                  return (
                    <div
                      key={rg.name}
                      data-index={virtualItem.index}
                      ref={virtualizer.measureElement}
                      style={{
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        width: '100%',
                        transform: `translateY(${virtualItem.start + ITEM_HEIGHT}px)`,
                      }}
                    >
                      <SelectItem
                        value={rg.name}
                        className={cn(
                          'cursor-pointer',
                          isSelected && 'bg-accent'
                        )}
                      >
                        <div className="flex items-center gap-2 w-full">
                          <span className="truncate flex-1">{rg.name}</span>
                          {rg.location && (
                            <span className="text-xs text-muted-foreground flex-shrink-0">
                              ({rg.location})
                            </span>
                          )}
                        </div>
                      </SelectItem>
                    </div>
                  );
                })}
              </div>
            ) : (
              <>
                <SelectItem value="all" className="font-medium">
                  All Resource Groups
                </SelectItem>
                {filteredResourceGroups.map((rg) => {
                  const isSelected = selectedResourceGroup === rg.name;
                  return (
                    <SelectItem
                      key={rg.name}
                      value={rg.name}
                      className={cn(
                        'cursor-pointer',
                        isSelected && 'bg-accent'
                      )}
                    >
                      <div className="flex items-center gap-2 w-full">
                        <span className="truncate flex-1">{rg.name}</span>
                        {rg.location && (
                          <span className="text-xs text-muted-foreground flex-shrink-0">
                            ({rg.location})
                          </span>
                        )}
                      </div>
                    </SelectItem>
                  );
                })}
              </>
            )}
          </div>
        )}

        {/* 검색 결과 카운트 */}
        {searchQuery && filteredResourceGroups.length > 0 && (
          <div className="p-2 border-t text-xs text-muted-foreground text-center">
            {filteredResourceGroups.length} result{filteredResourceGroups.length !== 1 ? 's' : ''} found
          </div>
        )}
      </SelectContent>
    </Select>
  );
}


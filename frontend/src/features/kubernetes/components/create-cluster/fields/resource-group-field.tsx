/**
 * Resource Group Selection Field Component
 * Azure Resource Group 선택 필드 컴포넌트 (검색 기능 포함)
 */

'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, X } from 'lucide-react';
import type { CreateClusterForm } from '@/lib/types';
import { useResourceGroups } from '@/hooks/use-resource-groups';
import { useTranslation } from '@/hooks/use-translation';
import { DataProcessor } from '@/lib/data';

export interface ResourceGroupFieldProps {
  /** React Hook Form 인스턴스 */
  form: UseFormReturn<CreateClusterForm>;
  /** 필드 변경 핸들러 */
  onFieldChange: (field: keyof CreateClusterForm, value: unknown) => void;
  /** Credential ID */
  credentialId: string;
  /** 로딩 중 여부 */
  isLoading?: boolean;
}

/**
 * Azure Resource Group 선택 필드 컴포넌트
 */
export function ResourceGroupField({
  form,
  onFieldChange,
  credentialId,
  isLoading = false,
}: ResourceGroupFieldProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  // Azure Resource Groups 목록 조회
  const { data: resourceGroups = [], isLoading: isLoadingResourceGroups } = useResourceGroups({
    credentialId,
    limit: 100, // 모든 Resource Group 조회
    enabled: !!credentialId,
  });

  const isDisabled = isLoading || isLoadingResourceGroups || !credentialId;

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

  const handleValueChange = (value: string) => {
    if (value === 'no-resource-groups') {
      return;
    }
    form.setValue('resource_group', value);
    onFieldChange('resource_group', value);
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
    <FormField
      control={form.control}
      name="resource_group"
      render={({ field }) => (
        <FormItem>
          <FormLabel>Resource Group *</FormLabel>
          <Select
            value={field.value || ''}
            onValueChange={handleValueChange}
            disabled={isDisabled}
            open={open}
            onOpenChange={handleOpenChange}
          >
            <FormControl>
              <SelectTrigger>
                <SelectValue 
                  placeholder={
                    isDisabled 
                      ? (isLoadingResourceGroups ? 'Loading resource groups...' : 'Select resource group')
                      : 'Select resource group'
                  } 
                />
              </SelectTrigger>
            </FormControl>
            <SelectContent className="p-0" onCloseAutoFocus={(e) => e.preventDefault()}>
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
              {isLoadingResourceGroups && (
                <div className="p-4 text-center text-sm text-muted-foreground">
                  Loading resource groups...
                </div>
              )}

              {/* 빈 상태 */}
              {!isLoadingResourceGroups && filteredResourceGroups.length === 0 && (
                <div className="p-4 text-center text-sm text-muted-foreground">
                  {searchQuery ? 'No resource groups found' : 'No resource groups available'}
                </div>
              )}

              {/* 리스트 */}
              {!isLoadingResourceGroups && filteredResourceGroups.length > 0 && (
                <div className="max-h-[300px] overflow-y-auto">
                  {filteredResourceGroups.map((rg) => (
                    <SelectItem key={rg.id} value={rg.name}>
                      {rg.name}
                      {rg.location && (
                        <span className="text-muted-foreground ml-2">
                          ({rg.location})
                        </span>
                      )}
                    </SelectItem>
                  ))}
                </div>
              )}
            </SelectContent>
          </Select>
          <FormDescription>
            Azure Resource Group name for the cluster
          </FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}


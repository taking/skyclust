/**
 * Resource Group Table Component
 * Azure Resource Group 목록 테이블 컴포넌트
 */

'use client';

import { useMemo, useCallback } from 'react';
import * as React from 'react';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Pagination } from '@/components/ui/pagination';
import { Trash2 } from 'lucide-react';
import { UI } from '@/lib/constants';
import type { ResourceGroupInfo } from '@/services/resource-group';

interface ResourceGroupTableProps {
  resourceGroups: ResourceGroupInfo[];
  filteredResourceGroups: ResourceGroupInfo[];
  paginatedResourceGroups: ResourceGroupInfo[];
  selectedResourceGroupNames: string[];
  onSelectionChange: (names: string[]) => void;
  onDelete: (name: string) => void;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  isDeleting?: boolean;
}

function ResourceGroupTableComponent({
  resourceGroups,
  filteredResourceGroups,
  paginatedResourceGroups,
  selectedResourceGroupNames,
  onSelectionChange,
  onDelete,
  page,
  pageSize,
  onPageChange,
  onPageSizeChange,
  isDeleting = false,
}: ResourceGroupTableProps) {
  const allSelected = useMemo(
    () => selectedResourceGroupNames.length === filteredResourceGroups.length && filteredResourceGroups.length > 0,
    [selectedResourceGroupNames.length, filteredResourceGroups.length]
  );

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      onSelectionChange(filteredResourceGroups.map(rg => rg.name));
    } else {
      onSelectionChange([]);
    }
  }, [filteredResourceGroups, onSelectionChange]);

  const handleSelectOne = useCallback((name: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedResourceGroupNames, name]);
    } else {
      onSelectionChange(selectedResourceGroupNames.filter(n => n !== name));
    }
  }, [selectedResourceGroupNames, onSelectionChange]);

  const getProvisioningStateBadgeVariant = (state: string) => {
    switch (state.toLowerCase()) {
      case 'succeeded':
        return 'default';
      case 'failed':
        return 'destructive';
      case 'deleting':
        return 'secondary';
      default:
        return 'outline';
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Resource Groups</CardTitle>
        <CardDescription>
          {filteredResourceGroups.length} of {resourceGroups.length} Resource Group{resourceGroups.length !== 1 ? 's' : ''} found
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={allSelected}
                  onCheckedChange={handleSelectAll}
                />
              </TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Location</TableHead>
              <TableHead>Provisioning State</TableHead>
              <TableHead>Tags</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedResourceGroups.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                  No resource groups found
                </TableCell>
              </TableRow>
            ) : (
              paginatedResourceGroups.map((rg) => {
                const isSelected = selectedResourceGroupNames.includes(rg.name);
                
                return (
                  <TableRow key={rg.name}>
                    <TableCell>
                      <Checkbox
                        checked={isSelected}
                        onCheckedChange={(checked) => handleSelectOne(rg.name, checked === true)}
                      />
                    </TableCell>
                    <TableCell className="font-medium">{rg.name}</TableCell>
                    <TableCell>{rg.location}</TableCell>
                    <TableCell>
                      <Badge variant={getProvisioningStateBadgeVariant(rg.provisioning_state)}>
                        {rg.provisioning_state}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {rg.tags && Object.keys(rg.tags).length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {Object.entries(rg.tags).slice(0, 3).map(([key, value]) => (
                            <Badge key={key} variant="outline" className="text-xs">
                              {key}: {value}
                            </Badge>
                          ))}
                          {Object.keys(rg.tags).length > 3 && (
                            <Badge variant="outline" className="text-xs">
                              +{Object.keys(rg.tags).length - 3}
                            </Badge>
                          )}
                        </div>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDelete(rg.name)}
                        disabled={isDeleting}
                      >
                        <Trash2 className="h-4 w-4 text-red-600" />
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
        
        {filteredResourceGroups.length > 0 && (
          <div className="border-t mt-4">
            <Pagination
              total={filteredResourceGroups.length}
              page={page}
              pageSize={pageSize}
              onPageChange={onPageChange}
              onPageSizeChange={onPageSizeChange}
              pageSizeOptions={UI.PAGINATION.PAGE_SIZE_OPTIONS}
              showPageSizeSelector={true}
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export const ResourceGroupTable = React.memo(ResourceGroupTableComponent);


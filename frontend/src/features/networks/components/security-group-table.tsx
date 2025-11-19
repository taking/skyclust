/**
 * Security Group Table Component
 * Security Group 목록 테이블 컴포넌트
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
import { SearchBar } from '@/components/ui/search-bar';
import { Trash2, Edit } from 'lucide-react';
import { UI } from '@/lib/constants';
import { useTranslation } from '@/hooks/use-translation';
import type { SecurityGroup, CloudProvider } from '@/lib/types';

interface SecurityGroupTableProps {
  securityGroups: Array<SecurityGroup & { provider?: CloudProvider; credential_id?: string }>;
  filteredSecurityGroups: Array<SecurityGroup & { provider?: CloudProvider; credential_id?: string }>;
  paginatedSecurityGroups: Array<SecurityGroup & { provider?: CloudProvider; credential_id?: string }>;
  selectedSecurityGroupIds: string[];
  onSelectionChange: (ids: string[]) => void;
  onDelete?: (securityGroupId: string, region: string) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  onSearchClear: () => void;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  isDeleting?: boolean;
  showProviderColumn?: boolean;
}

function SecurityGroupTableComponent({
  securityGroups,
  filteredSecurityGroups,
  paginatedSecurityGroups,
  selectedSecurityGroupIds,
  onSelectionChange,
  onDelete,
  searchQuery,
  onSearchChange,
  onSearchClear,
  page,
  pageSize,
  onPageChange,
  onPageSizeChange,
  isDeleting = false,
  showProviderColumn = false,
}: SecurityGroupTableProps) {
  const { t } = useTranslation();
  const hasMultipleProviders = showProviderColumn || securityGroups.some(sg => sg.provider && sg.provider !== securityGroups[0]?.provider);
  const allSelected = useMemo(
    () => selectedSecurityGroupIds.length === filteredSecurityGroups.length && filteredSecurityGroups.length > 0,
    [selectedSecurityGroupIds.length, filteredSecurityGroups.length]
  );

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      onSelectionChange(filteredSecurityGroups.map(sg => sg.id));
    } else {
      onSelectionChange([]);
    }
  }, [filteredSecurityGroups, onSelectionChange]);

  const handleSelectOne = useCallback((securityGroupId: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedSecurityGroupIds, securityGroupId]);
    } else {
      onSelectionChange(selectedSecurityGroupIds.filter(id => id !== securityGroupId));
    }
  }, [selectedSecurityGroupIds, onSelectionChange]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Security Groups</CardTitle>
        <CardDescription>
          {filteredSecurityGroups.length} of {securityGroups.length} security group{securityGroups.length !== 1 ? 's' : ''} found
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
              {hasMultipleProviders && (
                <TableHead>{t('common.provider') || 'Provider'}</TableHead>
              )}
              <TableHead>{t('common.name') || 'Name'}</TableHead>
              <TableHead>{t('common.description') || 'Description'}</TableHead>
              <TableHead>{t('network.vpc') || 'VPC'}</TableHead>
              <TableHead>{t('common.actions') || 'Actions'}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedSecurityGroups.map((securityGroup) => {
              const isSelected = selectedSecurityGroupIds.includes(securityGroup.id);
              
              return (
                <TableRow key={securityGroup.id}>
                  <TableCell>
                    <Checkbox
                      checked={isSelected}
                      onCheckedChange={(checked) => handleSelectOne(securityGroup.id, checked === true)}
                    />
                  </TableCell>
                  {hasMultipleProviders && (
                    <TableCell>
                      <Badge variant="outline">{securityGroup.provider || '-'}</Badge>
                    </TableCell>
                  )}
                  <TableCell className="font-medium">{securityGroup.name}</TableCell>
                  <TableCell>{securityGroup.description || '-'}</TableCell>
                  <TableCell>{securityGroup.vpc_id || '-'}</TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                      <Button variant="ghost" size="sm">
                        <Edit className="h-4 w-4" />
                      </Button>
                      {onDelete && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => onDelete(securityGroup.id, securityGroup.region)}
                          disabled={isDeleting}
                        >
                          <Trash2 className="h-4 w-4 text-red-600" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
        
        {filteredSecurityGroups.length > 0 && (
          <div className="border-t mt-4">
            <Pagination
              total={filteredSecurityGroups.length}
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

export const SecurityGroupTable = React.memo(SecurityGroupTableComponent);


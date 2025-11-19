/**
 * VPC Table Component
 * VPC 목록 테이블 컴포넌트
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
import { Trash2, Edit } from 'lucide-react';
import { UI } from '@/lib/constants';
import { useTranslation } from '@/hooks/use-translation';
import type { VPC, CloudProvider } from '@/lib/types';

interface VPCTableProps {
  vpcs: Array<VPC & { provider?: CloudProvider; credential_id?: string }>;
  filteredVPCs: Array<VPC & { provider?: CloudProvider; credential_id?: string }>;
  paginatedVPCs: Array<VPC & { provider?: CloudProvider; credential_id?: string }>;
  selectedVPCIds: string[];
  onSelectionChange: (ids: string[]) => void;
  onDelete: (vpcId: string, region?: string) => void;
  selectedRegion?: string;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  isDeleting?: boolean;
  showProviderColumn?: boolean;
}

function VPCTableComponent({
  vpcs,
  filteredVPCs,
  paginatedVPCs,
  selectedVPCIds,
  onSelectionChange,
  onDelete,
  selectedRegion,
  page,
  pageSize,
  onPageChange,
  onPageSizeChange,
  isDeleting = false,
  showProviderColumn = false,
}: VPCTableProps) {
  const { t } = useTranslation();
  const hasMultipleProviders = showProviderColumn || vpcs.some(v => v.provider && v.provider !== vpcs[0]?.provider);
  const allSelected = useMemo(
    () => selectedVPCIds.length === filteredVPCs.length && filteredVPCs.length > 0,
    [selectedVPCIds.length, filteredVPCs.length]
  );

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      onSelectionChange(filteredVPCs.map(v => v.id));
    } else {
      onSelectionChange([]);
    }
  }, [filteredVPCs, onSelectionChange]);

  const handleSelectOne = useCallback((vpcId: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedVPCIds, vpcId]);
    } else {
      onSelectionChange(selectedVPCIds.filter(id => id !== vpcId));
    }
  }, [selectedVPCIds, onSelectionChange]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>VPCs</CardTitle>
        <CardDescription>
          {filteredVPCs.length} of {vpcs.length} VPC{vpcs.length !== 1 ? 's' : ''} found
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
              <TableHead>{t('common.state') || 'State'}</TableHead>
              <TableHead>{t('common.description') || 'Description'}</TableHead>
              <TableHead>{t('common.default') || 'Default'}</TableHead>
              <TableHead>{t('common.actions') || 'Actions'}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedVPCs.map((vpc) => {
              const isSelected = selectedVPCIds.includes(vpc.id);
              
              return (
                <TableRow key={vpc.id}>
                  <TableCell>
                    <Checkbox
                      checked={isSelected}
                      onCheckedChange={(checked) => handleSelectOne(vpc.id, checked === true)}
                    />
                  </TableCell>
                  {hasMultipleProviders && (
                    <TableCell>
                      <Badge variant="outline">{vpc.provider || '-'}</Badge>
                    </TableCell>
                  )}
                  <TableCell className="font-medium">{vpc.name}</TableCell>
                  <TableCell>
                    <Badge variant={vpc.state === 'available' ? 'default' : 'secondary'}>
                      {vpc.state}
                    </Badge>
                  </TableCell>
                  <TableCell>{vpc.description || '-'}</TableCell>
                  <TableCell>
                    {vpc.is_default ? (
                      <Badge variant="outline">{t('common.default') || 'Default'}</Badge>
                    ) : (
                      <span className="text-gray-400">-</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                      <Button variant="ghost" size="sm">
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDelete(vpc.id, selectedRegion || vpc.region)}
                        disabled={isDeleting}
                      >
                        <Trash2 className="h-4 w-4 text-red-600" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
        
        {filteredVPCs.length > 0 && (
          <div className="border-t mt-4">
            <Pagination
              total={filteredVPCs.length}
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

export const VPCTable = React.memo(VPCTableComponent);


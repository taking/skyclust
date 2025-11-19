/**
 * Security Groups Page Content
 * 
 * Security Groups 페이지의 메인 콘텐츠 컴포넌트
 * 테이블만 담당합니다.
 */

'use client';

import * as React from 'react';
import dynamic from 'next/dynamic';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import type { SecurityGroup, CloudProvider } from '@/lib/types';

const SecurityGroupTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.SecurityGroupTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={7} rows={5} showCheckbox={true} />,
  }
);

export interface SecurityGroupsPageContentProps {
  securityGroups: Array<SecurityGroup & { provider?: CloudProvider; credential_id?: string }>;
  filteredSecurityGroups: Array<SecurityGroup & { provider?: CloudProvider; credential_id?: string }>;
  paginatedSecurityGroups: Array<SecurityGroup & { provider?: CloudProvider; credential_id?: string }>;
  selectedProvider: CloudProvider | undefined;
  selectedSecurityGroupIds: string[];
  onSelectionChange: (ids: string[] | ((prev: string[]) => string[])) => void;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  isSearching: boolean;
  searchQuery: string;
  isMultiProviderMode: boolean;
  selectedProviders: CloudProvider[];
}

export function SecurityGroupsPageContent({
  securityGroups,
  filteredSecurityGroups,
  paginatedSecurityGroups,
  selectedProvider,
  selectedSecurityGroupIds,
  onSelectionChange,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  isSearching,
  searchQuery,
  isMultiProviderMode,
  selectedProviders,
}: SecurityGroupsPageContentProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Security Groups</CardTitle>
        <CardDescription>
          {filteredSecurityGroups.length} of {securityGroups.length} security group{securityGroups.length !== 1 ? 's' : ''} 
          {isSearching && ` (${searchQuery})`}
          {isMultiProviderMode && ` • ${selectedProviders.length} provider${selectedProviders.length !== 1 ? 's' : ''}`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <SecurityGroupTable
          securityGroups={securityGroups}
          filteredSecurityGroups={filteredSecurityGroups}
          paginatedSecurityGroups={paginatedSecurityGroups}
          selectedSecurityGroupIds={selectedSecurityGroupIds}
          onSelectionChange={onSelectionChange}
          searchQuery={searchQuery}
          onSearchChange={() => {}}
          onSearchClear={() => {}}
          page={page}
          pageSize={pageSize}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
          showProviderColumn={isMultiProviderMode}
        />
      </CardContent>
    </Card>
  );
}


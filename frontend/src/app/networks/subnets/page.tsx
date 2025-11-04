/**
 * Subnets Page (Refactored)
 * 서브넷 관리 페이지 - 리팩토링된 버전
 */

'use client';

import { useState, useMemo } from 'react';
import * as React from 'react';
import dynamic from 'next/dynamic';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useSearch } from '@/hooks/use-search';
import { usePagination } from '@/hooks/use-pagination';
import { useCreateDialog } from '@/hooks/use-create-dialog';
import { EVENTS, UI } from '@/lib/constants';
import { Layers, Plus } from 'lucide-react';
import { useRequireAuth } from '@/hooks/use-auth';
import { useSSEMonitoring } from '@/hooks/use-sse-monitoring';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { Layout } from '@/components/layout/layout';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import {
  useSubnets,
  useSubnetActions,
  SubnetsPageHeader,
} from '@/features/networks';
import type { CreateSubnetForm } from '@/lib/types';

// Dynamic imports for heavy components
const CreateSubnetDialog = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.CreateSubnetDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

const SubnetTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.SubnetTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={6} rows={5} showCheckbox={true} />,
  }
);

export default function SubnetsPage() {
  const router = useRouter();
  const { isLoading: authLoading } = useRequireAuth();

  const {
    subnets,
    isLoadingSubnets,
    vpcs,
    selectedVPCId,
    setSelectedVPCId,
    credentials,
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
  } = useSubnets();

  const [isCreateDialogOpen, setIsCreateDialogOpen] = useCreateDialog(EVENTS.CREATE_DIALOG.SUBNET);
  const [selectedSubnetIds, setSelectedSubnetIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState<number>(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  useSSEMonitoring();

  const {
    createSubnetMutation,
    deleteSubnetMutation,
    handleBulkDeleteSubnets: handleBulkDelete,
    handleDeleteSubnet,
  } = useSubnetActions({
    selectedProvider,
    selectedCredentialId,
    onSuccess: () => {
      setIsCreateDialogOpen(false);
    },
  });

  // Search functionality
  const {
    query: searchQuery,
    setQuery: setSearchQuery,
    results: searchResults,
    clearSearch: clearSearch,
  } = useSearch(subnets, {
    keys: ['name', 'id', 'cidr_block', 'state'],
    threshold: 0.3,
  });

  // Filtered subnets (memoized for consistency)
  const filteredSubnets = useMemo(() => {
    return searchResults;
  }, [searchResults]);

  // Pagination
  const {
    page,
    paginatedItems: paginatedSubnets,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredSubnets, {
    totalItems: filteredSubnets.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const handleCreateSubnet = (data: CreateSubnetForm) => {
    createSubnetMutation.mutate(data);
  };

  const handleBulkDeleteSubnets = async (subnetIds: string[]) => {
    try {
      await handleBulkDelete(subnetIds, filteredSubnets);
      setSelectedSubnetIds([]);
    } catch (error) {
      // Error already handled in hook
    }
  };

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  };

  const handleVPCChange = (vpcId: string) => {
    setSelectedVPCId(vpcId);
  };

  // Render content with Early Return pattern
  const renderContent = () => {
    // Early Return: No credentials
    if (credentials.length === 0) {
      return <CredentialRequiredState serviceName="Networks (Subnets)" />;
    }

    // Early Return: No provider or credential selected
    if (!selectedProvider || !selectedCredentialId) {
      return (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Layers className="h-12 w-12 text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {!selectedProvider ? 'Select a Provider' : 'Select a Credential'}
            </h3>
            <p className="text-sm text-gray-500 text-center">
              {!selectedProvider
                ? 'Please select a cloud provider to view subnets'
                : 'Please select a credential to view subnets. If you don\'t have any credentials, register one first.'}
            </p>
            {!selectedProvider ? null : (
              <Button
                onClick={() => router.push('/credentials')}
                variant="default"
                className="mt-4"
              >
                <Plus className="mr-2 h-4 w-4" />
                Register Credentials
              </Button>
            )}
          </CardContent>
        </Card>
      );
    }

    // Early Return: No VPC or Region selected
    if (!selectedVPCId || !selectedRegion) {
      return (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Layers className="h-12 w-12 text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              Select VPC and Region
            </h3>
            <p className="text-sm text-gray-500 text-center">
              Please select a VPC and region to view subnets
            </p>
          </CardContent>
        </Card>
      );
    }

    // Early Return: No subnets found
    if (filteredSubnets.length === 0) {
      return (
        <ResourceEmptyState
          resourceName="Subnets"
          icon={Layers}
          onCreateClick={() => setIsCreateDialogOpen(true)}
          description="No subnets found for the selected VPC. Create your first subnet."
          withCard={true}
        />
      );
    }

    // Main content
    return (
      <>
        <BulkActionsToolbar
          items={paginatedSubnets}
          selectedIds={selectedSubnetIds}
          onSelectionChange={setSelectedSubnetIds}
          onBulkDelete={handleBulkDeleteSubnets}
          getItemDisplayName={(subnet) => subnet.name}
        />
        
        <SubnetTable
          subnets={subnets}
          filteredSubnets={filteredSubnets}
          paginatedSubnets={paginatedSubnets}
          selectedSubnetIds={selectedSubnetIds}
          onSelectionChange={setSelectedSubnetIds}
          onDelete={handleDeleteSubnet}
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          onSearchClear={clearSearch}
          page={page}
          pageSize={pageSize}
          onPageChange={setPage}
          onPageSizeChange={handlePageSizeChange}
          isDeleting={deleteSubnetMutation.isPending}
        />

        <CreateSubnetDialog
          open={isCreateDialogOpen}
          onOpenChange={setIsCreateDialogOpen}
          onSubmit={handleCreateSubnet}
          selectedProvider={selectedProvider}
          selectedRegion={selectedRegion}
          selectedVPCId={selectedVPCId}
          vpcs={vpcs}
          onVPCChange={handleVPCChange}
          isPending={createSubnetMutation.isPending}
          disabled={credentials.length === 0 || !selectedVPCId}
        />
      </>
    );
  };

  if (authLoading) {
    return (
      <WorkspaceRequired>
        <Layout>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
              <p className="mt-2 text-gray-600">Loading...</p>
            </div>
          </div>
        </Layout>
      </WorkspaceRequired>
    );
  }

  return (
    <WorkspaceRequired>
      <Layout>
        <div className="space-y-6">
          <SubnetsPageHeader />
          
          {/* Configuration */}
          {selectedProvider && selectedCredentialId && (
            <Card>
              <CardHeader>
                <CardTitle>Configuration</CardTitle>
                <CardDescription>Select region and VPC to view subnets</CardDescription>
              </CardHeader>
              <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Region *</Label>
                  <Input
                    placeholder="e.g., ap-northeast-2"
                    value={selectedRegion || ''}
                    readOnly
                    className="bg-muted"
                  />
                  <p className="text-xs text-muted-foreground">
                    Region selection is now handled in Header
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>VPC *</Label>
                  <Select
                    value={selectedVPCId}
                    onValueChange={handleVPCChange}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select VPC" />
                    </SelectTrigger>
                    <SelectContent>
                      {vpcs.map((vpc) => (
                        <SelectItem key={vpc.id} value={vpc.id}>
                          {vpc.name || vpc.id}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Content */}
          {isLoadingSubnets ? (
            <Card>
              <CardContent className="pt-6">
                <TableSkeleton columns={6} rows={5} showCheckbox={true} />
              </CardContent>
            </Card>
          ) : (
            renderContent()
          )}
        </div>
      </Layout>
    </WorkspaceRequired>
  );
}


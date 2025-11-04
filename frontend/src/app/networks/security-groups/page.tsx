/**
 * Security Groups Page (Refactored)
 * 보안 그룹 관리 페이지 - 리팩토링된 버전
 */

'use client';

import { useState, useMemo } from 'react';
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
import { Shield, Plus } from 'lucide-react';
import { useRequireAuth } from '@/hooks/use-auth';
import { useSSEMonitoring } from '@/hooks/use-sse-monitoring';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { Layout } from '@/components/layout/layout';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import {
  useSecurityGroups,
  useSecurityGroupActions,
  SecurityGroupsPageHeader,
} from '@/features/networks';
import type { CreateSecurityGroupForm } from '@/lib/types';

// Dynamic imports for heavy components
const CreateSecurityGroupDialog = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.CreateSecurityGroupDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

const SecurityGroupTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.SecurityGroupTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={4} rows={5} showCheckbox={true} />,
  }
);

export default function SecurityGroupsPage() {
  const router = useRouter();
  const { isLoading: authLoading } = useRequireAuth();

  const {
    securityGroups,
    isLoadingSecurityGroups,
    vpcs,
    selectedVPCId,
    setSelectedVPCId,
    credentials,
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
  } = useSecurityGroups();

  const [isCreateDialogOpen, setIsCreateDialogOpen] = useCreateDialog(EVENTS.CREATE_DIALOG.SECURITY_GROUP);
  const [selectedSecurityGroupIds, setSelectedSecurityGroupIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  useSSEMonitoring();

  const {
    createSecurityGroupMutation,
    deleteSecurityGroupMutation,
    handleBulkDeleteSecurityGroups: handleBulkDelete,
    handleDeleteSecurityGroup,
  } = useSecurityGroupActions({
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
  } = useSearch(securityGroups, {
    keys: ['name', 'id', 'description'],
    threshold: 0.3,
  });

  // Filtered security groups (memoized for consistency)
  const filteredSecurityGroups = useMemo(() => {
    return searchResults;
  }, [searchResults]);

  // Pagination
  const {
    page,
    paginatedItems: paginatedSecurityGroups,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredSecurityGroups, {
    totalItems: filteredSecurityGroups.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const handleCreateSecurityGroup = (data: CreateSecurityGroupForm) => {
    createSecurityGroupMutation.mutate(data);
  };

  const handleBulkDeleteSecurityGroups = async (securityGroupIds: string[]) => {
    try {
      await handleBulkDelete(securityGroupIds, filteredSecurityGroups);
      setSelectedSecurityGroupIds([]);
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
      return <CredentialRequiredState serviceName="Networks (Security Groups)" />;
    }

    // Early Return: No provider or credential selected
    if (!selectedProvider || !selectedCredentialId) {
      return (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Shield className="h-12 w-12 text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {!selectedProvider ? 'Select a Provider' : 'Select a Credential'}
            </h3>
            <p className="text-sm text-gray-500 text-center">
              {!selectedProvider
                ? 'Please select a cloud provider to view security groups'
                : 'Please select a credential to view security groups. If you don\'t have any credentials, register one first.'}
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
            <Shield className="h-12 w-12 text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              Select VPC and Region
            </h3>
            <p className="text-sm text-gray-500 text-center">
              Please select a VPC and region to view security groups
            </p>
          </CardContent>
        </Card>
      );
    }

    // Early Return: No security groups found
    if (filteredSecurityGroups.length === 0) {
      return (
        <ResourceEmptyState
          resourceName="Security Groups"
          icon={Shield}
          onCreateClick={() => setIsCreateDialogOpen(true)}
          description="No security groups found for the selected VPC. Create your first security group."
          withCard={true}
        />
      );
    }

    // Main content
    return (
      <>
        <BulkActionsToolbar
          items={paginatedSecurityGroups}
          selectedIds={selectedSecurityGroupIds}
          onSelectionChange={setSelectedSecurityGroupIds}
          onBulkDelete={handleBulkDeleteSecurityGroups}
          getItemDisplayName={(sg) => sg.name}
        />
        
        <SecurityGroupTable
          securityGroups={securityGroups}
          filteredSecurityGroups={filteredSecurityGroups}
          paginatedSecurityGroups={paginatedSecurityGroups}
          selectedSecurityGroupIds={selectedSecurityGroupIds}
          onSelectionChange={setSelectedSecurityGroupIds}
          onDelete={handleDeleteSecurityGroup}
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          onSearchClear={clearSearch}
          page={page}
          pageSize={pageSize}
          onPageChange={setPage}
          onPageSizeChange={handlePageSizeChange}
          isDeleting={deleteSecurityGroupMutation.isPending}
        />

        <CreateSecurityGroupDialog
          open={isCreateDialogOpen}
          onOpenChange={setIsCreateDialogOpen}
          onSubmit={handleCreateSecurityGroup}
          selectedProvider={selectedProvider}
          selectedRegion={selectedRegion}
          selectedVPCId={selectedVPCId}
          vpcs={vpcs}
          onVPCChange={handleVPCChange}
          isPending={createSecurityGroupMutation.isPending}
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
          <SecurityGroupsPageHeader />
          
          {/* Configuration */}
          {selectedProvider && selectedCredentialId && (
            <Card>
              <CardHeader>
                <CardTitle>Configuration</CardTitle>
                <CardDescription>Select region and VPC to view security groups</CardDescription>
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
          {isLoadingSecurityGroups ? (
            <Card>
              <CardContent className="pt-6">
                <TableSkeleton columns={4} rows={5} showCheckbox={true} />
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


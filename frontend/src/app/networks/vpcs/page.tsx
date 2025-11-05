/**
 * VPCs Page (Refactored)
 * Virtual Private Cloud 관리 페이지 - 리팩토링된 버전
 */

'use client';

import { useState, useMemo } from 'react';
import dynamic from 'next/dynamic';
import { useRouter } from 'next/navigation';
import { useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { DataProcessor } from '@/lib/data-processor';
import { usePagination } from '@/hooks/use-pagination';
import { useCreateDialog } from '@/hooks/use-create-dialog';
import { EVENTS, UI } from '@/lib/constants';
import { Network, Filter } from 'lucide-react';
import { FilterPanel, FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { SearchBar } from '@/components/ui/search-bar';
import { useRequireAuth } from '@/hooks/use-auth';
import { useSSEMonitoring } from '@/hooks/use-sse-monitoring';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { Layout } from '@/components/layout/layout';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { useTranslation } from '@/hooks/use-translation';
import {
  useVPCs,
  useVPCActions,
  VPCsPageHeader,
} from '@/features/networks';
import type { CreateVPCForm, VPC } from '@/lib/types';

// Dynamic imports for heavy components
const CreateVPCDialog = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.CreateVPCDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

const VPCTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.VPCTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={5} rows={5} showCheckbox={true} />,
  }
);

export default function VPCsPage() {
  const { t } = useTranslation();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { isLoading: authLoading } = useRequireAuth();

  const {
    vpcs,
    isLoadingVPCs,
    credentials,
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
  } = useVPCs();

  const [isCreateDialogOpen, setIsCreateDialogOpen] = useCreateDialog(EVENTS.CREATE_DIALOG.VPC);
  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedVPCIds, setSelectedVPCIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState<number>(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  useSSEMonitoring();

  const {
    createVPCMutation,
    deleteVPCMutation,
    handleBulkDeleteVPCs: handleBulkDelete,
    handleDeleteVPC,
  } = useVPCActions({
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
    onSuccess: () => {
      setIsCreateDialogOpen(false);
    },
  });

  const [searchQuery, setSearchQuery] = useState('');

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'state',
      label: t('filters.status'),
      type: 'select',
      options: [
        { id: 'available', value: 'available', label: t('filters.available') },
        { id: 'pending', value: 'pending', label: t('status.pending') },
        { id: 'deleting', value: 'deleting', label: t('status.deleting') },
      ],
    },
    {
      id: 'is_default',
      label: t('common.type'),
      type: 'select',
      options: [
        { id: 'true', value: 'true', label: t('network.defaultVPC') },
        { id: 'false', value: 'false', label: t('network.customVPC') },
      ],
    },
  ], [t]);

  // Custom filter function for VPC-specific filtering
  const filterFn = (vpc: VPC, filters: FilterValue): boolean => {
    if (filters.state && vpc.state !== filters.state) return false;
    if (filters.is_default !== undefined) {
      const isDefault = filters.is_default === 'true';
      if (vpc.is_default !== isDefault) return false;
    }
    return true;
  };

  // Apply search and filter using DataProcessor (memoized)
  const filteredVPCs = useMemo(() => {
    let result = DataProcessor.search(vpcs, searchQuery, {
      keys: ['name', 'id', 'state'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result as VPC[];
  }, [vpcs, searchQuery, filters]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = () => {
    setSearchQuery('');
  };

  // Pagination
  const {
    page,
    paginatedItems: paginatedVPCs,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredVPCs, {
    totalItems: filteredVPCs.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const handleCreateVPC = (data: CreateVPCForm) => {
    createVPCMutation.mutate(data);
  };

  const handleBulkDeleteVPCs = async (vpcIds: string[]) => {
    try {
      await handleBulkDelete(vpcIds, filteredVPCs);
      setSelectedVPCIds([]);
    } catch (error) {
      // Error already handled in hook
    }
  };

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  };

  // Render content with Early Return pattern
  const renderContent = () => {
    // Early Return: No credentials
    if (credentials.length === 0) {
      return <CredentialRequiredState serviceName={t('network.title')} />;
    }

    // Early Return: No provider or credential selected
    if (!selectedProvider || !selectedCredentialId) {
      return (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Network className="h-12 w-12 text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {!selectedProvider ? t('credential.selectCredential') : t('credential.selectCredential')}
            </h3>
            <p className="text-sm text-gray-500 text-center">
              {!selectedProvider
                ? t('credential.selectCredential')
                : t('credential.selectCredential')}
            </p>
            {!selectedProvider ? null : (
              <Button
                onClick={() => router.push('/credentials')}
                variant="default"
                className="mt-4"
              >
                {t('components.credentialRequired.registerButton')}
              </Button>
            )}
          </CardContent>
        </Card>
      );
    }

    // Early Return: No VPCs found
    if (filteredVPCs.length === 0) {
      return (
        <ResourceEmptyState
          resourceName={t('network.vpcs')}
          icon={Network}
          onCreateClick={() => setIsCreateDialogOpen(true)}
          withCard={true}
        />
      );
    }

    // Main content
    return (
      <>
        <BulkActionsToolbar
          items={paginatedVPCs}
          selectedIds={selectedVPCIds}
          onSelectionChange={setSelectedVPCIds}
          onBulkDelete={handleBulkDeleteVPCs}
          getItemDisplayName={(vpc) => vpc.name}
        />
        
        <VPCTable
          vpcs={vpcs}
          filteredVPCs={filteredVPCs}
          paginatedVPCs={paginatedVPCs}
          selectedVPCIds={selectedVPCIds}
          onSelectionChange={setSelectedVPCIds}
          onDelete={handleDeleteVPC}
          selectedRegion={selectedRegion}
          page={page}
          pageSize={pageSize}
          onPageChange={setPage}
          onPageSizeChange={handlePageSizeChange}
          isDeleting={deleteVPCMutation.isPending}
        />

        <CreateVPCDialog
          open={isCreateDialogOpen}
          onOpenChange={setIsCreateDialogOpen}
          onSubmit={handleCreateVPC}
          selectedProvider={selectedProvider}
          selectedRegion={selectedRegion}
          isPending={createVPCMutation.isPending}
          disabled={credentials.length === 0}
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
              <p className="mt-2 text-gray-600">{t('common.loading')}</p>
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
          <VPCsPageHeader />

          {/* Region Selection */}
          {selectedProvider && selectedCredentialId && (
            <Card>
              <CardHeader>
                <CardTitle>{t('common.configuration')}</CardTitle>
                <CardDescription>{t('network.selectRegionToFilterVPCs')}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <Label>{t('region.select')}</Label>
                  <Input
                    placeholder={t('region.placeholder')}
                    value={selectedRegion || ''}
                    readOnly
                    className="bg-muted"
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('network.regionSelectionHandledInHeader')}
                  </p>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Search and Filter */}
          {selectedProvider && selectedCredentialId && vpcs.length > 0 && (
            <Card>
              <CardContent className="pt-6">
                <div className="flex flex-col md:flex-row gap-4">
                  <div className="flex-1">
                    <SearchBar
                      value={searchQuery}
                      onChange={setSearchQuery}
                      onClear={clearSearch}
                      placeholder={t('network.searchVPCsPlaceholder')}
                    />
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => setShowFilters(!showFilters)}
                    className="flex items-center"
                  >
                    <Filter className="mr-2 h-4 w-4" />
                    {t('common.filter')}
                    {Object.keys(filters).length > 0 && (
                      <Badge variant="secondary" className="ml-2">
                        {Object.keys(filters).length}
                      </Badge>
                    )}
                  </Button>
                </div>
                {showFilters && (
                  <div className="mt-4">
                    <FilterPanel
                      filters={filterConfigs}
                      values={filters}
                      onChange={setFilters}
                      onClear={() => setFilters({})}
                      onApply={() => {}}
                    />
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Content */}
          {isLoadingVPCs ? (
            <Card>
              <CardContent className="pt-6">
                <TableSkeleton columns={5} rows={5} showCheckbox={true} />
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


/**
 * Kubernetes Clusters Page
 * Kubernetes 클러스터 관리 페이지
 * 
 * ResourceListPage 템플릿을 사용한 리팩토링 버전
 */

'use client';

import { useState, useMemo } from 'react';
import * as React from 'react';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { TagFilter } from '@/components/common/tag-filter';
import { BulkOperationProgress } from '@/components/common/bulk-operation-progress';
import { usePagination } from '@/hooks/use-pagination';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { EVENTS, UI } from '@/lib/constants';
import { useWorkspaceStore } from '@/store/workspace';
import { Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FilterPanel, FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCreateDialog } from '@/hooks/use-create-dialog';
import {
  ClusterPageHeader,
  useKubernetesClusters,
  useClusterFilters,
  useClusterBulkActions,
} from '@/features/kubernetes';
import { downloadKubeconfig } from '@/utils/kubeconfig';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import type { CreateClusterForm, CloudProvider } from '@/lib/types';

// Dynamic imports for heavy components
const BulkTagDialog = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.BulkTagDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

const ClusterTable = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.ClusterTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={6} rows={5} showCheckbox={true} />,
  }
);

export default function KubernetesClustersPage() {
  const { success } = useToast();
  const { handleError } = useErrorHandler();

  // Get workspace from store (consistent with other pages)
  const { currentWorkspace } = useWorkspaceStore();

  // Get credential context from global store
  const { selectedCredentialId, selectedRegion } = useCredentialContext();

  // Local state
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useCreateDialog(EVENTS.CREATE_DIALOG.CLUSTER);
  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedClusterIds, setSelectedClusterIds] = useState<string[]>([]);
  const [isTagDialogOpen, setIsTagDialogOpen] = useState(false);
  const [bulkTagKey, setBulkTagKey] = useState('');
  const [bulkTagValue, setBulkTagValue] = useState('');
  const [tagFilters, setTagFilters] = useState<Record<string, string[]>>({});
  const [pageSize, setPageSize] = useState(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  // Kubernetes clusters hook
  const {
    credentials,
    clusters,
    isLoading,
    createClusterMutation,
    deleteClusterMutation,
    downloadKubeconfigMutation,
  } = useKubernetesClusters({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || '',
    selectedRegion: selectedRegion || '',
  });

  // Get selected credential and provider
  const selectedCredential = credentials.find(c => c.id === selectedCredentialId);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  // Cluster filters hook
  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    availableTags,
    filteredClusters,
  } = useClusterFilters({
    clusters,
    filters,
    tagFilters,
  });

  // Bulk actions hook
  const {
    bulkOperationProgress,
    handleBulkDelete,
    handleBulkTag,
    handleCancelOperation,
    clearProgress,
  } = useClusterBulkActions(selectedProvider, selectedCredentialId || undefined);

  // Pagination
  const {
    page,
    totalPages,
    paginatedItems: paginatedClusters,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredClusters, {
    totalItems: filteredClusters.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // Filter configurations (memoized)
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: 'Status',
      type: 'select',
      options: [
        { id: 'active', value: 'ACTIVE', label: 'Active' },
        { id: 'creating', value: 'CREATING', label: 'Creating' },
        { id: 'updating', value: 'UPDATING', label: 'Updating' },
        { id: 'deleting', value: 'DELETING', label: 'Deleting' },
        { id: 'failed', value: 'FAILED', label: 'Failed' },
      ],
    },
    {
      id: 'region',
      label: 'Region',
      type: 'select',
      options: Array.from(new Set(clusters.map(c => c.region)))
        .filter(Boolean)
        .map((r, idx) => ({ 
          id: `region-${idx}`, 
          value: r, 
          label: r 
        })),
    },
  ], [clusters]);

  // Event handlers
  const handleCreateCluster = (data: CreateClusterForm) => {
    if (!selectedProvider) {
      handleError(new Error('Provider not selected'), { operation: 'createCluster', resource: 'Cluster' });
      return;
    }
    createClusterMutation.mutate(
      { provider: selectedProvider, data },
      {
        onSuccess: () => {
          success('Cluster creation initiated');
          setIsCreateDialogOpen(false);
        },
        onError: (error: unknown) => {
          handleError(error, { operation: 'createCluster', resource: 'Cluster' });
        },
      }
    );
  };

  const handleDeleteCluster = (clusterName: string, region: string) => {
    if (!selectedCredentialId || !selectedProvider) return;
    if (confirm(`Are you sure you want to delete cluster ${clusterName}? This action cannot be undone.`)) {
      deleteClusterMutation.mutate(
        {
          provider: selectedProvider,
          clusterName,
          credentialId: selectedCredentialId,
          region,
        },
        {
          onSuccess: () => {
            success('Cluster deletion initiated');
          },
          onError: (error: unknown) => {
            handleError(error, { operation: 'deleteCluster', resource: 'Cluster' });
          },
        }
      );
    }
  };

  const handleDownloadKubeconfig = (clusterName: string, region: string) => {
    if (!selectedCredentialId || !selectedProvider) return;
    downloadKubeconfigMutation.mutate(
      {
        provider: selectedProvider,
        clusterName,
        credentialId: selectedCredentialId,
        region,
      },
      {
        onSuccess: (kubeconfig) => {
          downloadKubeconfig(kubeconfig, clusterName);
          success('Kubeconfig downloaded');
        },
        onError: (error: unknown) => {
          handleError(error, { operation: 'downloadKubeconfig', resource: 'Cluster' });
        },
      }
    );
  };

  const handleBulkDeleteSubmit = (clusterIds: string[]) => {
    handleBulkDelete(clusterIds, filteredClusters, success, (error: unknown) => 
      handleError(error, { operation: 'bulkDeleteClusters', resource: 'Cluster' })
    );
  };

  const handleBulkTagSubmit = () => {
    if (!bulkTagKey.trim() || !bulkTagValue.trim()) return;
    handleBulkTag(
      selectedClusterIds,
      filteredClusters,
      bulkTagKey,
      bulkTagValue,
      success,
      (error: unknown) => handleError(error, { operation: 'bulkTagClusters', resource: 'Cluster' })
    );
    setIsTagDialogOpen(false);
    setBulkTagKey('');
    setBulkTagValue('');
    setSelectedClusterIds([]);
  };

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  };

  // Determine empty state
  const isEmpty = !selectedProvider || !selectedCredentialId || filteredClusters.length === 0;

  // Empty state component
  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName="Kubernetes" />
  ) : !selectedProvider ? (
    <ResourceEmptyState
      resourceName="Clusters"
      title="Select a Provider"
      description="Please select a cloud provider to view Kubernetes clusters"
      withCard={true}
    />
  ) : !selectedCredentialId ? (
    <CredentialRequiredState
      title="Select a Credential"
      description="Please select a credential to view Kubernetes clusters. If you don't have any credentials, register one first."
      serviceName="Kubernetes"
    />
  ) : filteredClusters.length === 0 ? (
    <ResourceEmptyState
      resourceName="Clusters"
      title="No Clusters Found"
      description="No Kubernetes clusters found. Create your first cluster to get started."
      onCreateClick={() => setIsCreateDialogOpen(true)}
      withCard={true}
    />
  ) : null;

  return (
    <ResourceListPage
      title="Kubernetes Clusters"
      resourceName="clusters"
      storageKey="kubernetes-clusters-page"
      header={
        <ClusterPageHeader
          workspaceName={currentWorkspace?.name}
          credentials={credentials}
          selectedCredentialId={selectedCredentialId || ''}
          onCredentialChange={() => {}} // Handled by Header
          selectedRegion={selectedRegion || ''}
          onRegionChange={() => {}} // Handled by Header
          selectedProvider={selectedProvider}
          onCreateCluster={handleCreateCluster}
          isCreatePending={createClusterMutation.isPending}
          isCreateDialogOpen={isCreateDialogOpen}
          onCreateDialogChange={setIsCreateDialogOpen}
        />
      }
      items={filteredClusters}
      isLoading={isLoading}
      isEmpty={isEmpty}
      searchQuery={searchQuery}
      onSearchChange={setSearchQuery}
      onSearchClear={clearSearch}
      isSearching={isSearching}
      searchPlaceholder="Search clusters by name, version, status, or region..."
      filterConfigs={selectedCredentialId && clusters.length > 0 ? filterConfigs : []}
      filters={filters}
      onFiltersChange={setFilters}
      onFiltersClear={() => setFilters({})}
      showFilters={showFilters}
      onToggleFilters={() => setShowFilters(!showFilters)}
      filterCount={Object.keys(filters).length}
      toolbar={
        <>
          {bulkOperationProgress && (
            <BulkOperationProgress
              {...bulkOperationProgress}
              onDismiss={clearProgress}
              onCancel={handleCancelOperation}
            />
          )}
          
          {selectedCredentialId && filteredClusters.length > 0 && (
            <BulkActionsToolbar
              items={filteredClusters}
              selectedIds={selectedClusterIds}
              onSelectionChange={setSelectedClusterIds}
              onBulkDelete={handleBulkDeleteSubmit}
              onBulkTag={() => setIsTagDialogOpen(true)}
              getItemDisplayName={(cluster) => cluster.name}
            />
          )}
        </>
      }
      additionalControls={
        selectedCredentialId && clusters.length > 0 ? (
          <>
            <TagFilter
              availableTags={availableTags}
              selectedTags={tagFilters}
              onTagsChange={setTagFilters}
            />
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              className="flex items-center"
            >
              <Filter className="mr-2 h-4 w-4" />
              Filters
              {Object.keys(filters).length > 0 && (
                <span className="ml-2 px-2 py-1 bg-gray-100 rounded text-sm">
                  {Object.keys(filters).length}
                </span>
              )}
            </Button>
          </>
        ) : null
      }
      emptyState={emptyStateComponent}
      content={
        selectedCredentialId && filteredClusters.length > 0 ? (
          <>
            <Card>
              <CardHeader>
                <CardTitle>Clusters</CardTitle>
                <CardDescription>
                  {filteredClusters.length} of {clusters.length} cluster{clusters.length !== 1 ? 's' : ''} 
                  {isSearching && ` (${searchQuery})`}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ClusterTable
                  clusters={paginatedClusters}
                  selectedIds={selectedClusterIds}
                  onSelectionChange={setSelectedClusterIds}
                  onDelete={handleDeleteCluster}
                  onDownloadKubeconfig={handleDownloadKubeconfig}
                  isDeleting={deleteClusterMutation.isPending}
                  isDownloading={downloadKubeconfigMutation.isPending}
                  page={page}
                  pageSize={pageSize}
                  total={filteredClusters.length}
                  onPageChange={setPage}
                  onPageSizeChange={handlePageSizeChange}
                />
              </CardContent>
            </Card>

            {/* Bulk Tag Dialog */}
            <BulkTagDialog
              open={isTagDialogOpen}
              onOpenChange={setIsTagDialogOpen}
              onSubmit={handleBulkTagSubmit}
              tagKey={bulkTagKey}
              tagValue={bulkTagValue}
              onTagKeyChange={setBulkTagKey}
              onTagValueChange={setBulkTagValue}
              selectedCount={selectedClusterIds.length}
            />
          </>
        ) : emptyStateComponent
      }
      pageSize={pageSize}
      onPageSizeChange={handlePageSizeChange}
      searchResultsCount={filteredClusters.length}
      skeletonColumns={6}
      skeletonRows={5}
      skeletonShowCheckbox={true}
      showFilterButton={false}
      showSearchResultsInfo={false}
    />
  );
}


/**
 * Kubernetes Clusters Page
 * Kubernetes 클러스터 관리 페이지
 * 
 * ResourceListPage 템플릿을 사용한 리팩토링 버전
 */

'use client';

import { useState, useMemo } from 'react';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { TagFilter } from '@/components/common/tag-filter';
import { BulkOperationProgress } from '@/components/common/bulk-operation-progress';
import { usePagination } from '@/hooks/use-pagination';
import { useToast } from '@/hooks/use-toast';
import { useWorkspaceFromUrl } from '@/hooks/use-workspace-from-url';
import { Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FilterPanel, FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import {
  ClusterPageHeader,
  ClusterTable,
  ClusterEmptyState,
  useKubernetesClusters,
  useClusterFilters,
  useClusterBulkActions,
} from '@/features/kubernetes';
import { downloadKubeconfig } from '@/utils/kubeconfig';
import type { CreateClusterForm, CloudProvider } from '@/lib/types';

// Dynamic import for BulkTagDialog
const BulkTagDialog = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.BulkTagDialog })),
  { 
    ssr: false,
    loading: () => null, // Dialog is hidden by default, so no loading state needed
  }
);

export default function KubernetesPage() {
  const { success, error: showError } = useToast();

  // Get workspace from URL (ResourceListPage handles loading)
  const { currentWorkspace } = useWorkspaceFromUrl();

  // Local state
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [selectedRegion, setSelectedRegion] = useState<string>('');
  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedClusterIds, setSelectedClusterIds] = useState<string[]>([]);
  const [isTagDialogOpen, setIsTagDialogOpen] = useState(false);
  const [bulkTagKey, setBulkTagKey] = useState('');
  const [bulkTagValue, setBulkTagValue] = useState('');
  const [tagFilters, setTagFilters] = useState<Record<string, string[]>>({});
  const [selectedCredentialId, setSelectedCredentialId] = useState<string>('');
  const [pageSize, setPageSize] = useState(20);

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
    selectedCredentialId,
    selectedRegion,
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
  } = useClusterBulkActions(selectedProvider, selectedCredentialId);

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
      showError('Provider not selected');
      return;
    }
    createClusterMutation.mutate(
      { provider: selectedProvider, data },
      {
        onSuccess: () => {
          success('Cluster creation initiated');
          setIsCreateDialogOpen(false);
        },
        onError: (error: Error) => {
          showError(`Failed to create cluster: ${error.message}`);
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
          onError: (error: Error) => {
            showError(`Failed to delete cluster: ${error.message}`);
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
        onError: (error: Error) => {
          showError(`Failed to download kubeconfig: ${error.message}`);
        },
      }
    );
  };

  const handleBulkDeleteSubmit = (clusterIds: string[]) => {
    handleBulkDelete(clusterIds, filteredClusters, success, showError);
  };

  const handleBulkTagSubmit = () => {
    if (!bulkTagKey.trim() || !bulkTagValue.trim()) return;
    handleBulkTag(
      selectedClusterIds,
      filteredClusters,
      bulkTagKey,
      bulkTagValue,
      success,
      showError
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
  const emptyStateComponent = !selectedProvider ? (
    <ClusterEmptyState
      title="Select a Provider"
      description="Please select a cloud provider to view Kubernetes clusters"
    />
  ) : !selectedCredentialId ? (
    <ClusterEmptyState
      title="Select a Credential"
      description="Please select a credential to view Kubernetes clusters"
    />
  ) : filteredClusters.length === 0 ? (
    <ClusterEmptyState
      title="No Clusters Found"
      description="No Kubernetes clusters found. Create your first cluster to get started."
      onCreateClick={() => setIsCreateDialogOpen(true)}
      showCreateButton={true}
    />
  ) : null;

  return (
    <ResourceListPage
      title="Kubernetes Clusters"
      resourceName="clusters"
      storageKey="kubernetes-page"
      header={
        <ClusterPageHeader
          workspaceName={currentWorkspace?.name}
          credentials={credentials}
          selectedCredentialId={selectedCredentialId}
          onCredentialChange={setSelectedCredentialId}
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

/**
 * Virtual Machines Page
 * Virtual Machine 관리 페이지
 * 
 * ResourceListPage 템플릿을 사용한 리팩토링 버전
 */

'use client';

import { useState, useEffect, useMemo } from 'react';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { FilterPresetsManager } from '@/components/common/filter-presets-manager';
import { MultiSortIndicator } from '@/components/common/multi-sort-indicator';
import { usePagination } from '@/hooks/use-pagination';
import { useAdvancedFiltering } from '@/hooks/use-advanced-filtering';
import { useKeyboardShortcuts, commonShortcuts } from '@/hooks/use-keyboard-shortcuts';
import { useToast } from '@/hooks/use-toast';
import { useWorkspaceFromUrl } from '@/hooks/use-workspace-from-url';
import { LiveRegion } from '@/components/accessibility/live-region';
import { GlobalKeyboardShortcuts } from '@/components/common/global-keyboard-shortcuts';
import { FilterPanel } from '@/components/ui/filter-panel';
import { VMPageHeader, VMTable, VMEmptyState, useVMs, useVMFilters, useVMActions } from '@/features/vms';
import type { CreateVMForm, VM } from '@/lib/types';
import type { FilterConfig, FilterValue } from '@/components/ui/filter-panel';

export default function VMsPage() {
  const { success: showSuccess, error: showError } = useToast();
  
  // Local state
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [liveMessage, setLiveMessage] = useState('');
  const [selectedCredentialId, setSelectedCredentialId] = useState<string>('');
  const [pageSize, setPageSize] = useState(20);
  const [showFilters, setShowFilters] = useState(false);

  // Advanced filtering with presets
  const {
    filters: advancedFilters,
    sortConfig,
    presets,
    setAllFilters,
    clearFilters,
    savePreset,
    loadPreset,
    deletePreset,
    toggleSort,
    clearSort,
  } = useAdvancedFiltering<VM>({
    storageKey: 'vms-page',
    defaultFilters: {},
    defaultSort: [],
  });

  // Keep local filters state for FilterPanel compatibility
  const [filters, setFilters] = useState<FilterValue>(advancedFilters as FilterValue);

  // Keyboard shortcuts
  useKeyboardShortcuts([
    commonShortcuts.newResource(() => setIsCreateDialogOpen(true)),
    commonShortcuts.escape(() => setIsCreateDialogOpen(false)),
  ]);

  // Get workspace from URL (ResourceListPage handles loading)
  const { currentWorkspace } = useWorkspaceFromUrl();

  // VMs hook
  const {
    credentials,
    vms,
    isLoading,
    createVMMutation,
    deleteVMMutation,
    startVMMutation,
    stopVMMutation,
  } = useVMs({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId,
  });

  // VM filters hook
  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filteredVMs,
  } = useVMFilters({
    vms,
    advancedFilters: advancedFilters as FilterValue,
    sortConfig,
  });

  // VM actions hook
  const {
    handleDeleteVM,
    handleStartVM,
    handleStopVM,
  } = useVMActions({
    workspaceId: currentWorkspace?.id,
    deleteMutation: deleteVMMutation,
    startMutation: startVMMutation,
    stopMutation: stopVMMutation,
    onSuccess: showSuccess,
    onError: showError,
    setLiveMessage,
  });

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: 'Status',
      type: 'multiselect',
      options: [
        { id: 'running', label: 'Running', value: 'running' },
        { id: 'stopped', label: 'Stopped', value: 'stopped' },
        { id: 'starting', label: 'Starting', value: 'starting' },
        { id: 'stopping', label: 'Stopping', value: 'stopping' },
      ],
    },
    {
      id: 'provider',
      label: 'Provider',
      type: 'multiselect',
      options: [
        { id: 'aws', label: 'AWS', value: 'aws' },
        { id: 'gcp', label: 'GCP', value: 'gcp' },
        { id: 'azure', label: 'Azure', value: 'azure' },
      ],
    },
    {
      id: 'region',
      label: 'Region',
      type: 'select',
      options: Array.from(new Set(vms.map(vm => vm.region)))
        .filter(Boolean)
        .map((region, idx) => ({
          id: `region-${idx}`,
          label: region,
          value: region,
        })),
    },
  ], [vms]);

  // Sync local filters with advanced filters
  useEffect(() => {
    setFilters(advancedFilters as FilterValue);
  }, [advancedFilters]);

  // Pagination
  const {
    page,
    totalPages,
    paginatedItems: paginatedVMs,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredVMs, {
    totalItems: filteredVMs.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // Event handlers
  const handleCreateVM = (data: CreateVMForm) => {
    if (!currentWorkspace) return;
    createVMMutation.mutate(
      { workspaceId: currentWorkspace.id, data },
      {
        onSuccess: () => {
          setIsCreateDialogOpen(false);
          setSelectedCredentialId('');
          showSuccess('VM creation initiated');
        },
        onError: (error: Error) => {
          showError(`Failed to create VM: ${error.message}`);
        },
      }
    );
  };

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  };

  const handleFiltersChange = (newFilters: FilterValue) => {
    setFilters(newFilters);
    setAllFilters(newFilters);
  };

  const handleFiltersClear = () => {
    setFilters({});
    clearFilters();
  };

  return (
    <ResourceListPage
      title="Virtual Machines"
      resourceName="VMs"
      storageKey="vms-page"
      header={
        <VMPageHeader
          workspaceName={currentWorkspace?.name}
          credentials={credentials}
          selectedCredentialId={selectedCredentialId}
          onCredentialChange={setSelectedCredentialId}
          onCreateVM={handleCreateVM}
          isCreatePending={createVMMutation.isPending}
          isCreateDialogOpen={isCreateDialogOpen}
          onCreateDialogChange={setIsCreateDialogOpen}
        />
      }
      items={filteredVMs}
      isLoading={isLoading}
      isEmpty={filteredVMs.length === 0 && !isSearching}
      searchQuery={searchQuery}
      onSearchChange={setSearchQuery}
      onSearchClear={clearSearch}
      isSearching={isSearching}
      searchPlaceholder="Search VMs..."
      filterConfigs={filterConfigs}
      filters={filters}
      onFiltersChange={handleFiltersChange}
      onFiltersClear={handleFiltersClear}
      showFilters={showFilters}
      onToggleFilters={() => setShowFilters(!showFilters)}
      filterCount={Object.keys(filters).length}
      sortIndicator={
        <MultiSortIndicator
          sortConfig={sortConfig}
          availableFields={[
            { value: 'name', label: 'Name' },
            { value: 'provider', label: 'Provider' },
            { value: 'status', label: 'Status' },
            { value: 'region', label: 'Region' },
            { value: 'instance_type', label: 'Instance Type' },
            { value: 'created_at', label: 'Created At' },
          ]}
          onToggleSort={toggleSort}
          onClearSort={clearSort}
        />
      }
      additionalControls={
        <>
          <FilterPresetsManager
            presets={presets}
            currentFilters={filters}
            onSavePreset={savePreset}
            onLoadPreset={loadPreset}
            onDeletePreset={deletePreset}
          />
          <FilterPanel
            filters={filterConfigs}
            values={filters}
            onChange={handleFiltersChange}
            onClear={handleFiltersClear}
            onApply={() => {}}
            title="Filter VMs"
            description="Filter VMs by status, provider, and region"
          />
        </>
      }
      emptyState={
        <VMEmptyState
          isSearching={isSearching}
          searchQuery={searchQuery}
          onCreateClick={() => setIsCreateDialogOpen(true)}
        />
      }
      content={
        <VMTable
          vms={paginatedVMs}
          sortConfig={sortConfig}
          onToggleSort={toggleSort}
          onStart={(vmId) => handleStartVM(vmId, vms)}
          onStop={(vmId) => handleStopVM(vmId, vms)}
          onDelete={(vmId) => handleDeleteVM(vmId, vms)}
          isStarting={startVMMutation.isPending}
          isStopping={stopVMMutation.isPending}
          isDeleting={deleteVMMutation.isPending}
          page={page}
          pageSize={pageSize}
          total={filteredVMs.length}
          onPageChange={setPage}
          onPageSizeChange={handlePageSizeChange}
        />
      }
      pageSize={pageSize}
      onPageSizeChange={handlePageSizeChange}
      searchResultsCount={filteredVMs.length}
      skeletonColumns={7}
      skeletonRows={5}
      keyboardShortcuts={<GlobalKeyboardShortcuts />}
      liveRegion={<LiveRegion message={liveMessage} />}
    />
  );
}

/**
 * Compute Snapshots Page
 * Compute Snapshots 관리 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/compute/snapshots
 */

'use client';

import { Suspense } from 'react';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useTranslation } from '@/hooks/use-translation';

function SnapshotsPageContent() {
  const { t } = useTranslation();
  const { workspaceId, credentialId, region } = useRequiredResourceContext();

  const isEmpty = !credentialId;

  const emptyStateComponent = !credentialId ? (
    <CredentialRequiredState serviceName={t('compute.title')} />
  ) : (
    <ResourceEmptyState
      resourceName={t('compute.snapshots') || 'Snapshots'}
      title={t('compute.noSnapshotsFound') || 'No Snapshots found'}
      description={t('compute.snapshotsDescription') || 'Snapshots will be displayed here'}
      withCard={true}
    />
  );

  return (
    <ResourceListPage
      title={t('compute.snapshots') || 'Snapshots'}
      resourceName={t('compute.snapshots') || 'Snapshots'}
      storageKey="snapshots-page"
      header={null}
      items={[]}
      isLoading={false}
      isEmpty={isEmpty}
      searchQuery=""
      onSearchChange={() => {}}
      onSearchClear={() => {}}
      isSearching={false}
      searchPlaceholder={t('compute.searchSnapshots') || 'Search Snapshots...'}
      filterConfigs={[]}
      filters={{}}
      onFiltersChange={() => {}}
      onFiltersClear={() => {}}
      showFilters={false}
      onToggleFilters={() => {}}
      filterCount={0}
      toolbar={null}
      additionalControls={null}
      emptyState={emptyStateComponent}
      content={emptyStateComponent}
      pageSize={20}
      onPageSizeChange={() => {}}
      searchResultsCount={0}
      skeletonColumns={5}
      skeletonRows={5}
      skeletonShowCheckbox={false}
      showFilterButton={false}
      showSearchResultsInfo={false}
    />
  );
}

export default function SnapshotsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <SnapshotsPageContent />
    </Suspense>
  );
}


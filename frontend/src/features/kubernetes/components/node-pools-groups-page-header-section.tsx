/**
 * Node Pools/Groups Page Header Section
 * 
 * Node Pool/Group 페이지의 헤더 섹션 컴포넌트
 * ClustersPageHeaderSection 참고
 */

'use client';

import * as React from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertTriangle } from 'lucide-react';
import { ClusterPageHeader } from './cluster-page-header';
import { UnifiedFilterPanel } from './unified-filter-panel';
import type { Credential, CloudProvider } from '@/lib/types';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';

export interface NodePoolsGroupsPageHeaderSectionProps {
  workspaceId: string;
  workspaceName?: string;
  credentials: Credential[];
  selectedCredentialIds: string[];
  onCredentialSelectionChange: (credentialIds: string[]) => void;
  selectedProvider: CloudProvider | undefined;
  selectedProviders: CloudProvider[];
  selectedRegion: string | null;
  onRegionChange: (region: string | null) => void;
  selectedRegions?: ProviderRegionSelection;
  onRegionSelectionChange?: (selectedRegions: ProviderRegionSelection) => void;
  useProviderRegionFilter?: boolean;
  onRefresh: () => Promise<void>;
  isRefreshing: boolean;
  lastUpdated: Date | null;
  isLoadingCredentials: boolean;
  isLoadingClusters?: boolean;
  errors: Array<{ provider: CloudProvider; credentialId: string; region?: string; error: Error }>;
  onCreateClick?: () => void;
}

export function NodePoolsGroupsPageHeaderSection({
  workspaceId,
  workspaceName,
  credentials,
  selectedCredentialIds,
  onCredentialSelectionChange,
  selectedProvider,
  selectedProviders,
  selectedRegion,
  onRegionChange,
  selectedRegions,
  onRegionSelectionChange,
  useProviderRegionFilter = false,
  onRefresh,
  isRefreshing,
  lastUpdated,
  isLoadingCredentials,
  isLoadingClusters = false,
  errors,
  onCreateClick,
}: NodePoolsGroupsPageHeaderSectionProps) {
  const hasError = errors.length > 0;

  return (
    <div className="space-y-4">
      <ClusterPageHeader
        workspaceName={workspaceName}
        credentials={credentials}
        selectedCredentialId={selectedCredentialIds[0] || ''}
        onCredentialChange={() => {}}
        selectedRegion={selectedRegion || ''}
        onRegionChange={() => {}}
        selectedProvider={selectedProvider}
        onRefresh={onRefresh}
        isRefreshing={isRefreshing}
        lastUpdated={lastUpdated}
        onCreateClick={onCreateClick}
      />
      
      {/* Unified Filter Panel */}
      <UnifiedFilterPanel
        credentials={credentials}
        selectedCredentialIds={selectedCredentialIds}
        onCredentialSelectionChange={onCredentialSelectionChange}
        selectedProviders={selectedProviders}
        selectedRegion={selectedRegion}
        onRegionChange={onRegionChange}
        selectedRegions={selectedRegions}
        onRegionSelectionChange={onRegionSelectionChange}
        disabled={isLoadingCredentials}
        isLoading={isLoadingClusters}
        defaultExpanded={true}
      />
      
      {/* Error Alerts (Non-blocking) */}
      {hasError && errors.length > 0 && (
        <div className="space-y-2">
          {errors.map((error, index) => (
            <Alert key={index} variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                {error.provider} ({error.credentialId}
                {error.region && `, ${error.region}`}): {error.error.message}
              </AlertDescription>
            </Alert>
          ))}
        </div>
      )}
    </div>
  );
}






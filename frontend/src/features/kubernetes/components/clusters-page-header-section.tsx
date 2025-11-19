/**
 * Clusters Page Header Section
 * 
 * 클러스터 페이지의 헤더 섹션 컴포넌트
 * Credential Multi-Select, Region Filter, Error Alerts를 포함합니다.
 */

'use client';

import * as React from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertTriangle } from 'lucide-react';
import { ClusterPageHeader } from './cluster-page-header';
import { UnifiedFilterPanel } from './unified-filter-panel';
import type { Credential, CloudProvider } from '@/lib/types';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';

export interface ClustersPageHeaderSectionProps {
  workspaceId: string;
  workspaceName?: string; // Workspace 이름 추가
  credentials: Credential[];
  selectedCredentialIds: string[];
  onCredentialSelectionChange: (credentialIds: string[]) => void;
  selectedProvider: CloudProvider | undefined;
  selectedProviders: CloudProvider[];
  selectedRegion: string | null;
  onRegionChange: (region: string | null) => void;
  selectedRegions?: ProviderRegionSelection; // Provider별 Region 선택
  onRegionSelectionChange?: (selectedRegions: ProviderRegionSelection) => void;
  onRefresh: () => Promise<void>;
  isRefreshing: boolean;
  lastUpdated: Date | null;
  isLoadingCredentials: boolean;
  isLoadingClusters?: boolean;
  clusterErrors: Array<{ provider: CloudProvider; credentialId: string; region?: string; error: Error }>;
  onCreateClick?: () => void;
}

export function ClustersPageHeaderSection({
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
  onRefresh,
  isRefreshing,
  lastUpdated,
  isLoadingCredentials,
  isLoadingClusters = false,
  clusterErrors,
  onCreateClick,
}: ClustersPageHeaderSectionProps) {
  const hasClusterError = clusterErrors.length > 0;

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
      />
      
      {/* Error Alerts (Non-blocking) */}
      {hasClusterError && clusterErrors.length > 0 && (
        <div className="space-y-2">
          {clusterErrors.map((error, index) => (
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

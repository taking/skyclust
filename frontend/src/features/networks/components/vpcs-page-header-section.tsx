/**
 * VPCs Page Header Section
 * 
 * VPCs 페이지의 헤더 섹션 컴포넌트
 * Credential Multi-Select, Region Filter, Error Alerts를 포함합니다.
 */

'use client';

import * as React from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertTriangle } from 'lucide-react';
import { VPCsPageHeader } from './vpcs-page-header';
import { UnifiedFilterPanel } from '@/features/kubernetes';
import type { Credential, CloudProvider } from '@/lib/types';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';

export interface VPCsPageHeaderSectionProps {
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
  isLoadingVPCs?: boolean;
  vpcErrors: Array<{ provider: CloudProvider; credentialId: string; region?: string; error: Error }>;
  onCreateClick?: () => void;
}

export function VPCsPageHeaderSection({
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
  isLoadingVPCs = false,
  vpcErrors,
  onCreateClick,
}: VPCsPageHeaderSectionProps) {
  const hasVPCError = vpcErrors.length > 0;

  return (
    <div className="space-y-4">
      <VPCsPageHeader
        onRefresh={onRefresh}
        isRefreshing={isRefreshing}
        lastUpdated={lastUpdated}
        onCreateClick={onCreateClick}
        disabled={isLoadingCredentials || selectedCredentialIds.length === 0}
      />
      
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
        isLoading={isLoadingVPCs}
        defaultExpanded={true}
      />
      
      {hasVPCError && vpcErrors.length > 0 && (
        <div className="space-y-2">
          {vpcErrors.map((error, index) => (
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


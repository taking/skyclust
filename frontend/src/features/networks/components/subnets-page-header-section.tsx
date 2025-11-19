/**
 * Subnets Page Header Section
 * 
 * Subnets 페이지의 헤더 섹션 컴포넌트
 * Credential Multi-Select, Region Filter, VPC Selection, Error Alerts를 포함합니다.
 */

'use client';

import * as React from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertTriangle } from 'lucide-react';
import { SubnetsPageHeader } from './subnets-page-header';
import { UnifiedFilterPanel } from '@/features/kubernetes';
import type { Credential, CloudProvider, VPC } from '@/lib/types';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';

export interface SubnetsPageHeaderSectionProps {
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
  vpcs: Array<VPC & { provider?: CloudProvider; credential_id?: string }>;
  selectedVPCId: string;
  onVPCChange: (vpcId: string) => void;
  onRefresh: () => Promise<void>;
  isRefreshing: boolean;
  lastUpdated: Date | null;
  isLoadingCredentials: boolean;
  isLoadingSubnets?: boolean;
  subnetErrors: Array<{ provider: CloudProvider; credentialId: string; region?: string; error: Error }>;
  onCreateClick?: () => void;
}

export function SubnetsPageHeaderSection({
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
  vpcs,
  selectedVPCId,
  onVPCChange,
  onRefresh,
  isRefreshing,
  lastUpdated,
  isLoadingCredentials,
  isLoadingSubnets = false,
  subnetErrors,
  onCreateClick,
}: SubnetsPageHeaderSectionProps) {
  const hasSubnetError = subnetErrors.length > 0;

  return (
    <div className="space-y-4">
      <SubnetsPageHeader
        selectedProvider={selectedProvider}
        selectedCredentialId={selectedCredentialIds[0] || ''}
        selectedVPCId={selectedVPCId}
        vpcs={vpcs}
        onVPCChange={onVPCChange}
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
        isLoading={isLoadingSubnets}
        defaultExpanded={true}
      />
      
      {hasSubnetError && subnetErrors.length > 0 && (
        <div className="space-y-2">
          {subnetErrors.map((error, index) => (
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


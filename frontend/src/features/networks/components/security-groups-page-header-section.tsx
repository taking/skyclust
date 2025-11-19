/**
 * Security Groups Page Header Section
 * 
 * Security Groups 페이지의 헤더 섹션 컴포넌트
 * Credential Multi-Select, Region Filter, VPC Selection, Error Alerts를 포함합니다.
 */

'use client';

import * as React from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertTriangle } from 'lucide-react';
import { SecurityGroupsPageHeader } from './security-groups-page-header';
import { UnifiedFilterPanel } from '@/features/kubernetes';
import type { Credential, CloudProvider, VPC } from '@/lib/types';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';

export interface SecurityGroupsPageHeaderSectionProps {
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
  isLoadingSecurityGroups?: boolean;
  securityGroupErrors: Array<{ provider: CloudProvider; credentialId: string; region?: string; error: Error }>;
  onCreateClick?: () => void;
}

export function SecurityGroupsPageHeaderSection({
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
  isLoadingSecurityGroups = false,
  securityGroupErrors,
  onCreateClick,
}: SecurityGroupsPageHeaderSectionProps) {
  const hasSecurityGroupError = securityGroupErrors.length > 0;

  return (
    <div className="space-y-4">
      <SecurityGroupsPageHeader
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
        isLoading={isLoadingSecurityGroups}
        defaultExpanded={true}
      />
      
      {hasSecurityGroupError && securityGroupErrors.length > 0 && (
        <div className="space-y-2">
          {securityGroupErrors.map((error, index) => (
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


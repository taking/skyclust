/**
 * Basic Cluster Configuration Step
 * Step 1: 클러스터 기본 설정 (Credential, Name, Version, Region, Zone)
 * 
 * 리팩토링: 필드별 컴포넌트와 커스텀 훅으로 분리하여 가독성과 유지보수성 향상
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { Form } from '@/components/ui/form';
import type { CreateClusterForm, Credential, CloudProvider } from '@/lib/types';
import { useClusterMetadata } from '@/features/kubernetes/hooks/use-cluster-metadata';
import { CredentialSelectionField } from './fields/credential-selection-field';
import { ClusterNameField } from './fields/cluster-name-field';
import { VersionSelectionField } from './fields/version-selection-field';
import { RegionZoneFields } from './fields/region-zone-fields';

export interface BasicClusterConfigStepProps {
  form: UseFormReturn<CreateClusterForm>;
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  selectedProvider?: CloudProvider;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
}

/**
 * 클러스터 기본 설정 Step 컴포넌트
 */
export function BasicClusterConfigStep({
  form,
  credentials,
  selectedCredentialId,
  onCredentialChange,
  selectedProvider,
  onDataChange,
}: BasicClusterConfigStepProps) {
  const selectedRegion = form.watch('region');

  // 클러스터 메타데이터 로딩 (버전, 리전, 존)
  const {
    versions,
    regions: awsRegions,
    zones,
    isLoadingVersions,
    isLoadingRegions,
    isLoadingZones,
    versionsError,
    regionsError,
    zonesError,
    canLoadMetadata,
  } = useClusterMetadata({
    provider: selectedProvider,
    credentialId: selectedCredentialId,
    region: selectedRegion || '',
  });

  const handleCredentialChange = (value: string) => {
    onCredentialChange(value);
    form.setValue('credential_id', value);
    // Credential 변경 시 region과 version 초기화 (새 credential의 region 목록을 가져오기 위해)
    if (canLoadMetadata) {
      form.setValue('region', '');
      form.setValue('version', '');
      form.setValue('zone', '');
      onDataChange({ credential_id: value, region: '', version: '', zone: '' });
    } else {
      onDataChange({ credential_id: value });
    }
  };

  const handleFieldChange = (field: keyof CreateClusterForm, value: unknown) => {
    form.setValue(field, value as never);
    onDataChange({ [field]: value });
  };

  const handleRegionChange = (region: string) => {
    // Region 변경 시 zone과 version 초기화는 RegionZoneFields에서 처리됨
    onDataChange({ region, zone: '', version: '' });
  };

  return (
    <Form {...form}>
      <div className="space-y-6">
        {/* Credential Selection */}
        <CredentialSelectionField
          form={form}
          credentials={credentials}
          selectedCredentialId={selectedCredentialId}
          onCredentialChange={handleCredentialChange}
        />

        {/* Cluster Name and Version */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <ClusterNameField
            form={form}
            onFieldChange={handleFieldChange}
          />

          <VersionSelectionField
            form={form}
            onFieldChange={handleFieldChange}
            versions={versions}
            isLoadingVersions={isLoadingVersions}
            versionsError={versionsError}
            canLoadMetadata={canLoadMetadata}
            hasSelectedRegion={!!selectedRegion}
          />
        </div>

        {/* Region and Zone */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <RegionZoneFields
            form={form}
            onFieldChange={handleFieldChange}
            provider={selectedProvider}
            awsRegions={awsRegions}
            isLoadingRegions={isLoadingRegions}
            regionsError={regionsError}
            zones={zones}
            isLoadingZones={isLoadingZones}
            zonesError={zonesError}
            canLoadMetadata={canLoadMetadata}
            onRegionChange={handleRegionChange}
          />
        </div>
      </div>
    </Form>
  );
}

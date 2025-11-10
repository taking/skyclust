/**
 * Basic Cluster Configuration Step
 * Step 1: 클러스터 기본 설정 (Credential, Name, Version, Region, Zone)
 * 
 * 레이아웃:
 * Row 1: Credential (1 column)
 * Row 2: Region (1 column) - Dashboard에서 Region 선택이 안된 경우만 표시
 * Row 3: Cluster Name | Kubernetes Version (2 columns)
 * Row 4: Availability Zone (1 column) - Region 값이 없으면 Disabled, Region 값이 있으면 해당 Region의 Availability Zone select box 표시
 * 
 * 리팩토링: 필드별 컴포넌트와 커스텀 훅으로 분리하여 가독성과 유지보수성 향상
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import type { CreateClusterForm, Credential, CloudProvider } from '@/lib/types';
import { useClusterMetadata } from '@/features/kubernetes/hooks/use-cluster-metadata';
import { CredentialSelectionField } from './fields/credential-selection-field';
import { ClusterNameField } from './fields/cluster-name-field';
import { VersionSelectionField } from './fields/version-selection-field';
import { RegionField } from './fields/region-field';
import { ZoneField } from './fields/zone-field';
import { useCredentialContext } from '@/hooks/use-credential-context';

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
  const { selectedCredentialId: dashboardCredentialId, selectedRegion: dashboardRegion } = useCredentialContext();
  const formRegion = form.watch('region');
  const formCredentialId = form.watch('credential_id');
  const selectedRegion = formRegion || dashboardRegion || '';
  
  // 사이드바에서 선택된 값이 있는지 확인
  const hasDashboardCredential = !!dashboardCredentialId;
  const hasDashboardRegion = !!dashboardRegion;
  
  // 실제로 사용할 credential ID (form > dashboard > prop 순서)
  const currentCredentialId = formCredentialId || dashboardCredentialId || selectedCredentialId || '';

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
    credentialId: currentCredentialId,
    region: selectedRegion || '',
  });

  const handleCredentialChange = (value: string) => {
    onCredentialChange(value);
    form.setValue('credential_id', value);
    // Credential 변경 시 region과 version 초기화 (새 credential의 region 목록을 가져오기 위해)
    // 단, Dashboard에서 선택된 region이 있으면 유지
    if (canLoadMetadata && !dashboardRegion) {
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
    // Region 변경 시 zone과 version 초기화
    onDataChange({ region, zone: '', version: '' });
  };

  return (
    <Form {...form}>
      <div className="space-y-6">
        {/* Row 1: Credential (1 column) - 사이드바에서 선택된 경우 hidden */}
        {!hasDashboardCredential && (
          <CredentialSelectionField
            form={form}
            credentials={credentials}
            selectedCredentialId={selectedCredentialId}
            onCredentialChange={handleCredentialChange}
          />
        )}

        {/* Row 2: Region (1 column) - Dashboard에서 Region 선택이 안된 경우만 표시 */}
        {!hasDashboardRegion && (
          <>
            {/* Credential과 Region 사이 구분선 */}
            {!hasDashboardCredential && (
              <div className="border-t border-gray-200 my-4" />
            )}
            <RegionField
              form={form}
              onFieldChange={handleFieldChange}
              provider={selectedProvider}
              awsRegions={awsRegions}
              isLoadingRegions={isLoadingRegions}
              regionsError={regionsError}
              canLoadMetadata={canLoadMetadata}
              onRegionChange={handleRegionChange}
            />
          </>
        )}

        {/* Region과 Cluster Name 사이 구분선 - Credential 또는 Region이 보일 때만 표시 */}
        {(!hasDashboardCredential || !hasDashboardRegion) && (
          <div className="border-t border-gray-200 my-4" />
        )}

        {/* Row 3: Cluster Name | Kubernetes Version (2 columns) - 좌우 대칭 개선 */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
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

        {/* Row 4: Availability Zone (1 column) - Region 값이 없으면 Disabled, Region 값이 있으면 해당 Region의 Availability Zone select box 표시 */}
        {/* AWS만 표시 */}
        {selectedProvider === 'aws' && (
          <ZoneField
            form={form}
            onFieldChange={handleFieldChange}
            provider={selectedProvider}
            zones={zones}
            isLoadingZones={isLoadingZones}
            zonesError={zonesError}
            canLoadMetadata={canLoadMetadata}
            selectedRegion={selectedRegion}
            disabled={!selectedRegion}
          />
        )}

        {/* Azure Resource Group */}
        {selectedProvider === 'azure' && (
          <FormField
            control={form.control}
            name="resource_group"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Resource Group *</FormLabel>
                <FormControl>
                  <Input
                    {...field}
                    placeholder="my-resource-group"
                    value={field.value || ''}
                  />
                </FormControl>
                <FormDescription>
                  Azure Resource Group name for the cluster
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        )}
      </div>
    </Form>
  );
}

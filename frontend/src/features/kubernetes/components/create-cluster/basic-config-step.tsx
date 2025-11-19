/**
 * Basic Cluster Configuration Step
 * Step 1: 클러스터 기본 설정 (Credential, Name, Version, Region, Zone)
 * 
 * 레이아웃:
 * Row 1: Credential (1 column) - 항상 표시
 * Row 2: Region | Availability Zone (2 columns)
 *   - Availability Zone: Region 값이 없으면 Disabled, Region 값이 있으면 해당 Region의 Availability Zone select box 표시
 * Row 3: Cluster Name | Kubernetes Version (2 columns)
 * 
 * 기능:
 * - Region 추천: Credential 선택 후 첫 번째 Region 추천 표시
 * - Zone 추천: Region 선택 후 첫 번째 Zone 추천 표시 (자동 선택 없음)
 * - Zone 정보 툴팁: 각 Zone에 대한 추가 정보 표시
 * 
 * 리팩토링: 필드별 컴포넌트와 커스텀 훅으로 분리하여 가독성과 유지보수성 향상
 */

'use client';

import * as React from 'react';
import { useRef, useEffect } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { useQueryClient, useQuery } from '@tanstack/react-query';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import type { CreateClusterForm, Credential, CloudProvider } from '@/lib/types';
import { useClusterMetadata } from '@/features/kubernetes/hooks/use-cluster-metadata';
import { queryKeys } from '@/lib/query';
import { credentialService } from '@/services/credential';
import { CredentialSelectionField } from './fields/credential-selection-field';
import { ClusterNameField } from './fields/cluster-name-field';
import { VersionSelectionField } from './fields/version-selection-field';
import { RegionField } from './fields/region-field';
import { ZoneField } from './fields/zone-field';
import { ResourceGroupField } from './fields/resource-group-field';

export interface BasicClusterConfigStepProps {
  form: UseFormReturn<CreateClusterForm>;
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  selectedProvider?: CloudProvider;
  workspaceId?: string;
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
  workspaceId,
  onDataChange,
}: BasicClusterConfigStepProps) {
  const queryClient = useQueryClient();
  const formRegion = form.watch('region');
  const formCredentialId = form.watch('credential_id');
  const selectedRegion = formRegion || '';
  
  // 실제로 사용할 credential ID (form > prop 순서)
  const currentCredentialId = formCredentialId || selectedCredentialId || '';
  
  // Credential 변경 추적 (이전 credentialId, provider 저장)
  const previousCredentialIdRef = useRef<string>(currentCredentialId);
  const previousProviderRef = useRef<CloudProvider | undefined>(selectedProvider);
  const previousRegionRef = useRef<string>(selectedRegion);
  const [isCredentialChanged, setIsCredentialChanged] = React.useState(false);
  
  // Credential 변경 감지 및 zones 캐시 무효화
  useEffect(() => {
    const credentialChanged = previousCredentialIdRef.current && previousCredentialIdRef.current !== currentCredentialId;
    const providerChanged = previousProviderRef.current !== selectedProvider;
    
    if (credentialChanged || providerChanged) {
      setIsCredentialChanged(true);
      
      // Credential 또는 Provider 변경 시 region과 zone 초기화
      form.setValue('region', '');
      form.setValue('zone', '');
      form.setValue('version', '');
      onDataChange({ region: '', zone: '', version: '' });
      
      // 이전 credential/provider의 모든 zones 캐시 무효화
      if (previousProviderRef.current && previousCredentialIdRef.current) {
        // 이전 provider의 모든 region에 대한 zones 캐시 무효화
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesMetadata.availabilityZones(previousProviderRef.current, previousCredentialIdRef.current),
        });
      }
      
      // Region이 선택되면 credential 변경 상태 초기화
      if (selectedRegion) {
        setIsCredentialChanged(false);
      }
    } else if (selectedRegion) {
      setIsCredentialChanged(false);
    }
    
    previousCredentialIdRef.current = currentCredentialId;
    previousProviderRef.current = selectedProvider;
  }, [currentCredentialId, selectedProvider, selectedRegion, form, onDataChange, queryClient]);

  // Region 변경 시 zones 캐시 무효화 및 zone 초기화
  useEffect(() => {
    if (previousRegionRef.current && previousRegionRef.current !== selectedRegion) {
      // Zone 초기화
      form.setValue('zone', '');
      onDataChange({ zone: '' });
      
      // 이전 region의 zones 캐시 무효화 및 제거
      if (selectedProvider && currentCredentialId) {
        // 이전 region의 zones 캐시 무효화
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesMetadata.availabilityZones(selectedProvider, currentCredentialId, previousRegionRef.current),
        });
        
        // 이전 region의 zones 캐시 제거 (강제로 새로운 region의 zones를 로드하기 위해)
        queryClient.removeQueries({
          queryKey: queryKeys.kubernetesMetadata.availabilityZones(selectedProvider, currentCredentialId, previousRegionRef.current),
        });
      }
      
      // 새로운 region의 zones를 강제로 refetch
      if (selectedProvider && currentCredentialId && selectedRegion) {
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesMetadata.availabilityZones(selectedProvider, currentCredentialId, selectedRegion),
        });
      }
    }
    previousRegionRef.current = selectedRegion;
  }, [selectedRegion, selectedProvider, currentCredentialId, form, onDataChange, queryClient]);

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
    workspaceId,
  });

  // GCP credential 선택 시 project_id 추출
  const { data: credentialDetail } = useQuery<Credential>({
    queryKey: queryKeys.credentials.detail(currentCredentialId),
    queryFn: () => credentialService.getCredential(currentCredentialId, workspaceId),
    enabled: !!currentCredentialId && selectedProvider === 'gcp' && !!workspaceId,
  });

  // GCP credential의 project_id를 form에 자동 설정
  useEffect(() => {
    if (selectedProvider === 'gcp' && credentialDetail) {
      // masked_data 필드에서 project_id 추출 (masked_data는 project_id를 포함함)
      const maskedData = credentialDetail.masked_data;
      
      if (maskedData && typeof maskedData === 'object') {
        const projectId = maskedData.project_id as string | undefined;
        
        if (projectId && typeof projectId === 'string' && projectId.trim() !== '') {
          const currentProjectId = form.getValues('project_id');
          // project_id가 없거나 변경된 경우에만 업데이트
          if (currentProjectId !== projectId) {
            form.setValue('project_id', projectId);
            onDataChange({ project_id: projectId });
          }
        } else {
          // project_id가 없거나 유효하지 않은 경우 경고
          console.warn('GCP credential에 project_id가 없거나 유효하지 않습니다.', {
            credentialId: currentCredentialId,
            maskedData,
          });
        }
      } else {
        // masked_data가 없는 경우 경고
        console.warn('GCP credential에 masked_data가 없습니다.', {
          credentialId: currentCredentialId,
          credentialDetail,
        });
      }
    } else if (selectedProvider !== 'gcp') {
      // GCP가 아닌 경우 project_id 초기화
      const currentProjectId = form.getValues('project_id');
      if (currentProjectId) {
        form.setValue('project_id', '');
        onDataChange({ project_id: '' });
      }
    }
  }, [credentialDetail, selectedProvider, form, onDataChange, currentCredentialId]);

  const handleCredentialChange = (value: string) => {
    onCredentialChange(value);
    form.setValue('credential_id', value);
    // Credential 변경 시 항상 region, zone, version 초기화 (canLoadMetadata와 관계없이)
    form.setValue('region', '');
    form.setValue('version', '');
    form.setValue('zone', '');
    onDataChange({ credential_id: value, region: '', version: '', zone: '' });
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
        {/* Row 1: Credential (1 column) - 항상 표시 */}
        <CredentialSelectionField
          form={form}
          credentials={credentials}
          selectedCredentialId={selectedCredentialId}
          onCredentialChange={handleCredentialChange}
        />

        {/* Credential과 Region 사이 구분선 */}
        <div className="border-t border-gray-200 my-4" />

        {/* Row 2: Region | Availability Zone (2 columns) */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <RegionField
            form={form}
            onFieldChange={handleFieldChange}
            provider={selectedProvider}
            awsRegions={awsRegions}
            isLoadingRegions={isLoadingRegions}
            regionsError={regionsError}
            canLoadMetadata={canLoadMetadata}
            isCredentialChanged={isCredentialChanged}
            onRegionChange={handleRegionChange}
          />
          
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
            autoSelectZone={false}
            onRegionChange={handleRegionChange}
          />
        </div>

        {/* Region/Zone과 Cluster Name 사이 구분선 */}
        <div className="border-t border-gray-200 my-4" />

        {/* Row 3: Cluster Name | Kubernetes Version (2 columns) */}
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

        {/* Azure Resource Group */}
        {selectedProvider === 'azure' && (
          <ResourceGroupField
            form={form}
            onFieldChange={handleFieldChange}
            credentialId={currentCredentialId}
            isLoading={isLoadingVersions || isLoadingRegions || isLoadingZones}
          />
        )}

        {/* Azure Deployment Mode Selection */}
        {selectedProvider === 'azure' && (
          <>
            <div className="border-t border-gray-200 my-4" />
            <FormField
              control={form.control}
              name="deployment_mode"
              render={({ field }) => (
                <FormItem className="space-y-3">
                  <FormLabel>Deployment Mode *</FormLabel>
                  <FormControl>
                    <RadioGroup
                      onValueChange={(value: string) => {
                        field.onChange(value);
                        onDataChange({ deployment_mode: value as 'auto' | 'custom' });
                      }}
                      value={field.value || 'auto'}
                      className="flex flex-col space-y-1"
                    >
                      <div className="flex items-center space-x-2">
                        <RadioGroupItem value="auto" id="deployment-mode-auto" />
                        <Label htmlFor="deployment-mode-auto" className="font-normal cursor-pointer">
                          자동 (Automatic)
                        </Label>
                      </div>
                      <div className="flex items-center space-x-2">
                        <RadioGroupItem value="custom" id="deployment-mode-custom" disabled />
                        <Label htmlFor="deployment-mode-custom" className="font-normal cursor-pointer text-muted-foreground">
                          커스텀 (Custom) - 준비 중
                        </Label>
                      </div>
                    </RadioGroup>
                  </FormControl>
                  <FormDescription>
                    자동 모드: Azure가 VNet/Subnet을 자동으로 생성합니다. Network 설정 단계가 건너뜁니다.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </>
        )}
      </div>
    </Form>
  );
}

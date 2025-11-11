/**
 * Region Selection Field Component
 * 리전 선택 필드 컴포넌트 (Dashboard 값 처리 포함)
 */

'use client';

import { useEffect } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { getRegionsByProvider } from '@/lib/regions';
import { useCredentialContext } from '@/hooks/use-credential-context';

export interface RegionFieldProps {
  /** React Hook Form 인스턴스 */
  form: UseFormReturn<CreateClusterForm>;
  /** 필드 변경 핸들러 */
  onFieldChange: (field: keyof CreateClusterForm, value: unknown) => void;
  /** 클라우드 프로바이더 */
  provider?: string;
  /** AWS 리전 목록 (메타데이터에서 로드된 경우) */
  awsRegions?: Array<{ value: string; label: string }>;
  /** 리전 로딩 중 여부 */
  isLoadingRegions?: boolean;
  /** 리전 로딩 에러 */
  regionsError?: Error | null;
  /** 메타데이터 로딩 가능 여부 */
  canLoadMetadata?: boolean;
  /** 리전 변경 시 추가 처리 (예: zone 초기화) */
  onRegionChange?: (region: string) => void;
}

/**
 * 리전 선택 필드 컴포넌트
 * Dashboard에서 선택된 값이 있으면 자동 적용 (비활성화), 없으면 선택 가능
 */
export function RegionField({
  form,
  onFieldChange,
  provider,
  awsRegions = [],
  isLoadingRegions = false,
  regionsError = null,
  canLoadMetadata = false,
  onRegionChange,
}: RegionFieldProps) {
  const { t } = useTranslation();
  const { selectedRegion: dashboardRegion } = useCredentialContext();
  const formRegion = form.watch('region');
  
  // Dashboard에서 Region이 선택되어 있는지 확인
  const hasDashboardRegion = !!dashboardRegion;
  
  // Dashboard에서 선택된 Region이 있으면 form에 자동 설정
  useEffect(() => {
    if (dashboardRegion && !formRegion) {
      form.setValue('region', dashboardRegion);
      onFieldChange('region', dashboardRegion);
      
      // Azure의 경우 location 필드도 업데이트
      if (provider === 'azure') {
        form.setValue('location', dashboardRegion);
        onFieldChange('location', dashboardRegion);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dashboardRegion, formRegion, provider]);
  
  // Dashboard에서 선택된 Region 또는 Form의 Region 사용
  const currentRegion = formRegion || dashboardRegion || '';
  
  // Static regions for non-AWS providers
  const staticRegions = provider ? getRegionsByProvider(provider) : [];
  
  // Use AWS regions if available, otherwise use static regions
  const regions = canLoadMetadata && awsRegions.length > 0 ? awsRegions : staticRegions;

  const handleRegionChange = (value: string) => {
    form.setValue('region', value);
    onFieldChange('region', value);
    
    // Azure의 경우 location 필드도 업데이트
    if (provider === 'azure') {
      form.setValue('location', value);
      onFieldChange('location', value);
    }
    
    // Zone 초기화
    form.setValue('zone', '');
    onFieldChange('zone', '');
    
    // Version 초기화 (AWS의 경우 region이 변경되면 version도 다시 로드해야 함)
    if (canLoadMetadata) {
      form.setValue('version', '');
      onFieldChange('version', '');
    }
    
    onRegionChange?.(value);
  };

  // 현재 선택된 region의 정보 찾기
  const selectedRegionInfo = regions.find(r => r.value === currentRegion);

  return (
    <FormField
      control={form.control}
      name="region"
      render={({ field }) => (
        <FormItem>
          <FormLabel>Region *</FormLabel>
          {canLoadMetadata ? (
            <Select
              value={currentRegion}
              onValueChange={handleRegionChange}
              disabled={isLoadingRegions || hasDashboardRegion}
            >
              <FormControl>
                <SelectTrigger>
                  <SelectValue>
                    {currentRegion ? (
                      <div className="flex items-center gap-2">
                        <span>{currentRegion}</span>
                        {selectedRegionInfo && (
                          <Badge variant="secondary" className="text-xs">
                            {selectedRegionInfo.label}
                          </Badge>
                        )}
                      </div>
                    ) : (
                      isLoadingRegions ? 'Loading regions...' : 'Select region'
                    )}
                  </SelectValue>
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                {regions.length === 0 && !isLoadingRegions ? (
                  <SelectItem value="no-regions" disabled>
                    No regions available
                  </SelectItem>
                ) : (
                  regions.map((region) => (
                    <SelectItem key={region.value} value={region.value}>
                      <div className="flex items-center gap-2">
                        <span>{region.value}</span>
                        <Badge variant="secondary" className="text-xs">
                          {region.label}
                        </Badge>
                      </div>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          ) : (
            <Select
              value={currentRegion}
              onValueChange={(value) => {
                field.onChange(value);
                handleRegionChange(value);
              }}
              disabled={hasDashboardRegion}
            >
              <FormControl>
                <SelectTrigger>
                  <SelectValue>
                    {currentRegion ? (
                      <div className="flex items-center gap-2">
                        <span>{currentRegion}</span>
                        {selectedRegionInfo && (
                          <Badge variant="secondary" className="text-xs">
                            {selectedRegionInfo.label}
                          </Badge>
                        )}
                      </div>
                    ) : (
                      'Select region'
                    )}
                  </SelectValue>
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                {regions.map((region) => (
                  <SelectItem key={region.value} value={region.value}>
                    <div className="flex items-center gap-2">
                      <span>{region.value}</span>
                      <Badge variant="secondary" className="text-xs">
                        {region.label}
                      </Badge>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
          <FormDescription>
            {hasDashboardRegion
              ? `Region selected from Dashboard: ${dashboardRegion}. Change in Sidebar if needed.`
              : t('kubernetes.regionDescription')}
          </FormDescription>
          {canLoadMetadata && regionsError && (
            <FormDescription className="text-destructive">
              Failed to load regions: {regionsError.message}
              {(regionsError.message.includes('IAM permission') || 
                regionsError.message.includes('not authorized') ||
                regionsError.message.includes('UnauthorizedOperation')) && (
                <span className="block mt-1 text-muted-foreground">
                  <strong>Solution:</strong> Add the <code className="px-1 py-0.5 bg-muted rounded">ec2:DescribeRegions</code> permission to your AWS IAM user or role.
                </span>
              )}
            </FormDescription>
          )}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}


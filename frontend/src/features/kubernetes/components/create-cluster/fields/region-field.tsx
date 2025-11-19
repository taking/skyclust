/**
 * Region Selection Field Component
 * 리전 선택 필드 컴포넌트
 */

'use client';

import { useMemo } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Sparkles, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { getRegionsByProvider } from '@/lib/regions';

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
  /** Credential이 변경되었는지 여부 */
  isCredentialChanged?: boolean;
  /** 리전 변경 시 추가 처리 (예: zone 초기화) */
  onRegionChange?: (region: string) => void;
}

/**
 * 리전 선택 필드 컴포넌트
 * 항상 선택 가능한 필드로 표시
 */
export function RegionField({
  form,
  onFieldChange,
  provider,
  awsRegions = [],
  isLoadingRegions = false,
  regionsError = null,
  canLoadMetadata = false,
  isCredentialChanged = false,
  onRegionChange,
}: RegionFieldProps) {
  const { t } = useTranslation();
  const formRegion = form.watch('region');
  
  // Form의 Region 사용
  const currentRegion = formRegion || '';
  
  // Static regions for non-AWS providers
  const staticRegions = provider ? getRegionsByProvider(provider) : [];
  
  // Use AWS regions if available, otherwise use static regions
  const regions = canLoadMetadata && awsRegions.length > 0 ? awsRegions : staticRegions;

  // 추천 Region 계산 (첫 번째 Region)
  const recommendedRegion = useMemo(() => {
    if (regions.length === 0) return null;
    if (!canLoadMetadata && staticRegions.length > 0) {
      return staticRegions[0].value;
    }
    if (canLoadMetadata && awsRegions.length > 0) {
      return awsRegions[0].value;
    }
    return null;
  }, [regions, canLoadMetadata, awsRegions, staticRegions]);

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
              disabled={isLoadingRegions}
            >
              <FormControl>
                <SelectTrigger
                  className={cn(
                    isCredentialChanged && !currentRegion && 'border-amber-300 bg-amber-50/50 dark:bg-amber-950/20 dark:border-amber-800'
                  )}
                >
                  <SelectValue
                    placeholder={
                      isLoadingRegions
                        ? t('kubernetes.loadingRegionsForCredential') || 'Loading regions for new credential...'
                        : isCredentialChanged
                        ? t('kubernetes.selectRegionAfterCredentialChange') || 'Select region (credential changed)'
                        : 'Select region'
                    }
                  >
                    {currentRegion ? (
                      <span className="flex items-center gap-2">
                        <span>
                          {currentRegion}
                          {selectedRegionInfo?.label && (
                            <span className="text-muted-foreground ml-1.5 text-xs">
                              ({selectedRegionInfo.label})
                            </span>
                          )}
                        </span>
                      </span>
                    ) : isCredentialChanged ? (
                      <span className="flex items-center gap-1.5 text-amber-700 dark:text-amber-400">
                        <AlertCircle className="h-3.5 w-3.5" />
                        <span>{t('kubernetes.selectRegionAfterCredentialChange') || 'Select region (credential changed)'}</span>
                      </span>
                    ) : null}
                  </SelectValue>
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                {regions.length === 0 && !isLoadingRegions ? (
                  <SelectItem value="no-regions" disabled>
                    No regions available
                  </SelectItem>
                ) : (
                  regions.map((region) => {
                    const isRecommended = recommendedRegion === region.value;
                    return (
                      <SelectItem key={region.value} value={region.value} title={isRecommended ? `${region.value} (Recommended)` : region.value}>
                        <div className="flex items-center gap-2 w-full">
                          <span className="flex-1">
                            {region.value}
                            {region.label && (
                              <span className="text-muted-foreground ml-1.5 text-xs">
                                ({region.label})
                              </span>
                            )}
                          </span>
                          {isRecommended && (
                            <Badge variant="secondary" className="text-xs py-0 px-1.5 shrink-0">
                              <Sparkles className="h-3 w-3 mr-1" />
                              Recommended
                            </Badge>
                          )}
                        </div>
                      </SelectItem>
                    );
                  })
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
            >
              <FormControl>
                <SelectTrigger
                  className={cn(
                    isCredentialChanged && !currentRegion && 'border-amber-300 bg-amber-50/50 dark:bg-amber-950/20 dark:border-amber-800'
                  )}
                >
                  <SelectValue
                    placeholder={
                      isCredentialChanged
                        ? t('kubernetes.selectRegionAfterCredentialChange') || 'Select region (credential changed)'
                        : 'Select region'
                    }
                  >
                    {currentRegion ? (
                      <span>
                        {currentRegion}
                        {selectedRegionInfo?.label && (
                          <span className="text-muted-foreground ml-1.5 text-xs">
                            ({selectedRegionInfo.label})
                          </span>
                        )}
                      </span>
                    ) : isCredentialChanged ? (
                      <span className="flex items-center gap-1.5 text-amber-700 dark:text-amber-400">
                        <AlertCircle className="h-3.5 w-3.5" />
                        <span>{t('kubernetes.selectRegionAfterCredentialChange') || 'Select region (credential changed)'}</span>
                      </span>
                    ) : null}
                  </SelectValue>
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                {regions.map((region) => {
                  const isRecommended = recommendedRegion === region.value;
                  return (
                    <SelectItem key={region.value} value={region.value} title={isRecommended ? `${region.value} (Recommended)` : region.value}>
                      <div className="flex items-center gap-2 w-full">
                        <span className="flex-1">
                          {region.value}
                          {region.label && (
                            <span className="text-muted-foreground ml-1.5 text-xs">
                              ({region.label})
                            </span>
                          )}
                        </span>
                        {isRecommended && (
                          <Badge variant="secondary" className="text-xs py-0 px-1.5 shrink-0">
                            <Sparkles className="h-3 w-3 mr-1" />
                            Recommended
                          </Badge>
                        )}
                      </div>
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          )}
          <FormDescription>
            {isLoadingRegions && !currentRegion
              ? t('kubernetes.loadingRegionsForCredential') || 'Loading regions for new credential...'
              : !currentRegion && isCredentialChanged
              ? t('kubernetes.selectRegionAfterCredentialChangeDescription') || 'Credential has been changed. Please select a region again.'
              : !currentRegion && recommendedRegion
              ? `Recommended: ${recommendedRegion} (first region for optimal performance)`
              : !currentRegion
              ? t('common.selectRegionFirst') || 'Please select a region to continue.'
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


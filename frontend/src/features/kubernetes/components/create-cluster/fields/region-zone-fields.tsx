/**
 * Region and Zone Selection Fields Component
 * 리전 및 존 선택 필드 컴포넌트
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { getRegionsByProvider } from '@/lib/regions';

export interface RegionZoneFieldsProps {
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
  /** 가용 영역 목록 */
  zones?: string[];
  /** 존 로딩 중 여부 */
  isLoadingZones?: boolean;
  /** 존 로딩 에러 */
  zonesError?: Error | null;
  /** 메타데이터 로딩 가능 여부 */
  canLoadMetadata?: boolean;
  /** 리전 변경 시 추가 처리 (예: zone 초기화) */
  onRegionChange?: (region: string) => void;
}

/**
 * 리전 및 존 선택 필드 컴포넌트
 */
export function RegionZoneFields({
  form,
  onFieldChange,
  provider,
  awsRegions = [],
  isLoadingRegions = false,
  regionsError = null,
  zones = [],
  isLoadingZones = false,
  zonesError = null,
  canLoadMetadata = false,
  onRegionChange,
}: RegionZoneFieldsProps) {
  const { t } = useTranslation();
  const selectedRegion = form.watch('region');

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

  return (
    <>
      <FormField
        control={form.control}
        name="region"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Region *</FormLabel>
            {canLoadMetadata ? (
              <Select
                value={field.value}
                onValueChange={handleRegionChange}
                disabled={isLoadingRegions}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder={isLoadingRegions ? 'Loading regions...' : 'Select region'} />
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
                        {region.label}
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
            ) : (
              <Select
                value={field.value}
                onValueChange={(value) => {
                  field.onChange(value);
                  handleRegionChange(value);
                }}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select region" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {regions.map((region) => (
                    <SelectItem key={region.value} value={region.value}>
                      {region.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            <FormDescription>
              {t('kubernetes.regionDescription')}
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

      {/* Zone Selection (AWS only, required) */}
      {provider === 'aws' && (
        <FormField
          control={form.control}
          name="zone"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Availability Zone *</FormLabel>
              {canLoadMetadata && selectedRegion ? (
                <Select
                  value={field.value || ''}
                  onValueChange={(value) => {
                    field.onChange(value);
                    onFieldChange('zone', value);
                  }}
                  disabled={isLoadingZones}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder={isLoadingZones ? 'Loading zones...' : 'Select zone *'} />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {zones.length === 0 && !isLoadingZones ? (
                      <SelectItem value="no-zones" disabled>
                        No zones available
                      </SelectItem>
                    ) : (
                      zones.map((zone) => (
                        <SelectItem key={zone} value={zone}>
                          {zone}
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
              ) : (
                <FormControl>
                  <Input
                    placeholder="e.g., ap-northeast-3b"
                    {...field}
                    onChange={(e) => {
                      field.onChange(e);
                      onFieldChange('zone', e.target.value);
                    }}
                  />
                </FormControl>
              )}
              <FormDescription>
                {t('kubernetes.zoneDescription')}
              </FormDescription>
              {canLoadMetadata && selectedRegion && zonesError && (
                <FormDescription className="text-destructive">
                  Failed to load availability zones: {zonesError.message}
                  {(zonesError.message.includes('IAM permission') || 
                    zonesError.message.includes('not authorized') ||
                    zonesError.message.includes('UnauthorizedOperation')) && (
                    <span className="block mt-1 text-muted-foreground">
                      <strong>Solution:</strong> Add the <code className="px-1 py-0.5 bg-muted rounded">ec2:DescribeAvailabilityZones</code> permission to your AWS IAM user or role.
                    </span>
                  )}
                </FormDescription>
              )}
              <FormMessage />
            </FormItem>
          )}
        />
      )}
    </>
  );
}


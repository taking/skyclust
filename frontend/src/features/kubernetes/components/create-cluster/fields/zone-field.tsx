/**
 * Zone Selection Field Component
 * 가용 영역 선택 필드 컴포넌트 (AWS만 표시)
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

export interface ZoneFieldProps {
  /** React Hook Form 인스턴스 */
  form: UseFormReturn<CreateClusterForm>;
  /** 필드 변경 핸들러 */
  onFieldChange: (field: keyof CreateClusterForm, value: unknown) => void;
  /** 클라우드 프로바이더 */
  provider?: string;
  /** 가용 영역 목록 */
  zones?: string[];
  /** 존 로딩 중 여부 */
  isLoadingZones?: boolean;
  /** 존 로딩 에러 */
  zonesError?: Error | null;
  /** 메타데이터 로딩 가능 여부 */
  canLoadMetadata?: boolean;
  /** 선택된 리전 */
  selectedRegion?: string;
  /** 필드 비활성화 여부 */
  disabled?: boolean;
}

/**
 * 가용 영역 선택 필드 컴포넌트 (AWS만 표시)
 */
export function ZoneField({
  form,
  onFieldChange,
  provider,
  zones = [],
  isLoadingZones = false,
  zonesError = null,
  canLoadMetadata = false,
  selectedRegion,
  disabled = false,
}: ZoneFieldProps) {
  const { t } = useTranslation();

  // AWS만 표시
  if (provider !== 'aws') {
    return null;
  }

  // Region이 선택되지 않았거나 disabled prop이 true이면 비활성화
  const isDisabled = disabled || !selectedRegion || isLoadingZones;

  return (
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
              disabled={isDisabled}
            >
              <FormControl>
                <SelectTrigger>
                  <SelectValue 
                    placeholder={
                      !selectedRegion
                        ? 'Select region first'
                        : isLoadingZones
                        ? 'Loading zones...'
                        : 'Select zone *'
                    } 
                  />
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
                placeholder={!selectedRegion ? 'Select region first' : 'e.g., ap-northeast-3b'}
                {...field}
                onChange={(e) => {
                  field.onChange(e);
                  onFieldChange('zone', e.target.value);
                }}
                disabled={isDisabled}
              />
            </FormControl>
          )}
          <FormDescription>
            {!selectedRegion
              ? 'Please select a region first to enable availability zone selection.'
              : t('kubernetes.zoneDescription')}
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
  );
}


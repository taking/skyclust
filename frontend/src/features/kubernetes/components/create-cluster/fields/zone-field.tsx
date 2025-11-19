/**
 * Zone Selection Field Component
 * 가용 영역 선택 필드 컴포넌트 (AWS, GCP, Azure 지원)
 * 
 * 기능:
 * - Zone 추천 표시 (첫 번째 Zone 또는 안정적인 Zone)
 * - Zone 정보 툴팁 (가용성 정보)
 * - 자동 선택 없음: 사용자가 직접 선택
 */

'use client';

import { useMemo } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Badge } from '@/components/ui/badge';
import { Info, Sparkles } from 'lucide-react';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';

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
  /** Zone 자동 선택 활성화 여부 */
  autoSelectZone?: boolean;
  /** Region 변경 시 Zone 초기화 핸들러 */
  onRegionChange?: (region: string) => void;
}

/**
 * 가용 영역 선택 필드 컴포넌트 (AWS, GCP, Azure 지원)
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
  autoSelectZone = false,
  onRegionChange,
}: ZoneFieldProps) {
  const { t } = useTranslation();

  // Provider가 없으면 표시하지 않음
  if (!provider) {
    return null;
  }

  // Region이 선택되지 않았거나 disabled prop이 true이면 비활성화
  const isDisabled = disabled || !selectedRegion || isLoadingZones;

  // 추천 Zone 계산 (첫 번째 Zone 또는 안정적인 Zone)
  const recommendedZone = useMemo(() => {
    if (zones.length === 0) return null;
    // 첫 번째 Zone을 추천 Zone으로 설정 (일반적으로 가장 안정적)
    return zones[0];
  }, [zones]);

  return (
    <FormField
      control={form.control}
      name="zone"
      render={({ field }) => (
        <FormItem>
          <div className="flex items-center gap-2">
            <FormLabel>Availability Zone *</FormLabel>
            {selectedRegion && zones.length > 0 && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Info className="h-4 w-4 text-muted-foreground cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p className="text-xs">
                    Select an availability zone for high availability. 
                    The first zone is recommended for optimal performance.
                  </p>
                </TooltipContent>
              </Tooltip>
            )}
          </div>
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
                <SelectTrigger
                  id="zone-select"
                  className={cn(
                    "transition-all duration-200",
                    !selectedRegion && "opacity-50 cursor-not-allowed",
                    selectedRegion && !isDisabled && "opacity-100"
                  )}
                  aria-describedby={selectedRegion ? "zone-description" : "zone-disabled-description"}
                  aria-disabled={isDisabled}
                  aria-label="Availability zone selection"
                >
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
                  zones.map((zone) => {
                    const isRecommended = zone === recommendedZone;
                    return (
                      <SelectItem 
                        key={zone} 
                        value={zone}
                        title={isRecommended ? `${zone} (Recommended)` : zone}
                      >
                        <div className="flex items-center gap-2 w-full">
                          <span className="flex-1">{zone}</span>
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
            <FormControl>
              <Input
                placeholder={!selectedRegion ? 'Select region first' : 'e.g., ap-northeast-3b'}
                {...field}
                onChange={(e) => {
                  field.onChange(e);
                  onFieldChange('zone', e.target.value);
                }}
                disabled={isDisabled}
                aria-describedby={selectedRegion ? "zone-description" : "zone-disabled-description"}
                aria-disabled={isDisabled}
                aria-label="Availability zone input"
              />
            </FormControl>
          )}
          <FormDescription id={selectedRegion ? "zone-description" : "zone-disabled-description"}>
            {!selectedRegion
              ? 'Please select a region first to enable availability zone selection.'
              : isLoadingZones
              ? 'Loading availability zones...'
              : zones.length > 0 && recommendedZone
              ? `Recommended: ${recommendedZone} (first zone for optimal performance)`
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


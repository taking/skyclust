/**
 * Region Filter Component
 * Multi-provider Region 필터 UI
 * 
 * 기능:
 * - 여러 provider의 region을 union으로 표시
 * - Region 선택 드롭다운
 * - Provider별 region 표시
 */

'use client';

import * as React from 'react';
import { useRegionFilter } from '@/hooks/use-region-filter';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { MapPin, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import type { CloudProvider } from '@/lib/types/kubernetes';
import { useTranslation } from '@/hooks/use-translation';

interface RegionFilterProps {
  providers: CloudProvider[];
  selectedRegion?: string | null;
  onRegionChange: (region: string | null) => void;
  disabled?: boolean;
}

// 선택된 region이 어떤 provider에 속하는지 확인하는 헬퍼 함수
function getProvidersForRegion(
  region: string,
  regionsByProvider: Record<CloudProvider, string[]>
): CloudProvider[] {
  const providerList: CloudProvider[] = [];
  (Object.keys(regionsByProvider) as CloudProvider[]).forEach(provider => {
    if (regionsByProvider[provider].includes(region)) {
      providerList.push(provider);
    }
  });
  return providerList;
}

export function RegionFilter({
  providers,
  selectedRegion,
  onRegionChange,
  disabled = false,
}: RegionFilterProps) {
  const { t } = useTranslation();
  
  const {
    availableRegions,
    regionsByProvider,
    setSelectedRegion,
    isRegionValid,
  } = useRegionFilter({
    providers,
    selectedRegion,
    onRegionChange,
  });
  
  const handleClear = React.useCallback(() => {
    setSelectedRegion(null);
  }, [setSelectedRegion]);
  
  // 선택된 region의 provider 목록
  const selectedRegionProviders = React.useMemo(() => {
    if (!selectedRegion) return [];
    return getProvidersForRegion(selectedRegion, regionsByProvider);
  }, [selectedRegion, regionsByProvider]);
  
  return (
    <div className="space-y-2 min-w-0">
      <div className="flex items-center justify-between">
        <Label htmlFor="region-filter" className="flex items-center gap-2">
          <MapPin className="h-4 w-4" />
          <span>Region</span>
        </Label>
        {selectedRegion && (
          <div className="flex items-center gap-1.5">
            {/* 선택된 region의 provider 표시 */}
            {selectedRegionProviders.length > 0 && (
              <div className="flex gap-1">
                {selectedRegionProviders.map(provider => (
                  <Badge
                    key={provider}
                    variant="outline"
                    className="text-xs h-5"
                  >
                    {provider.toUpperCase()}
                  </Badge>
                ))}
              </div>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={handleClear}
              disabled={disabled}
              className="h-6 px-2"
              aria-label={t('common.clearRegionSelection')}
            >
              <X className="h-3 w-3" />
            </Button>
          </div>
        )}
      </div>
      
      <Select
        value={selectedRegion || undefined}
        onValueChange={(value) => {
          setSelectedRegion(value === "__all__" ? null : value);
        }}
        disabled={disabled || availableRegions.length === 0}
      >
        <SelectTrigger id="region-filter" className="w-full h-8">
          <div className="flex items-center justify-between w-full">
            <SelectValue placeholder="All regions" />
            {selectedRegion && selectedRegionProviders.length > 0 && (
              <Badge variant="secondary" className="ml-2 text-xs">
                {selectedRegionProviders.length} provider{selectedRegionProviders.length !== 1 ? 's' : ''}
              </Badge>
            )}
          </div>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="__all__">All regions</SelectItem>
          {availableRegions.map(region => {
            const providerList = getProvidersForRegion(region, regionsByProvider);
            
            return (
              <SelectItem key={region} value={region}>
                <div className="flex items-center justify-between w-full">
                  <span>{region}</span>
                  <div className="flex gap-1 ml-2">
                    {providerList.map(provider => (
                      <Badge
                        key={provider}
                        variant="outline"
                        className="text-xs"
                      >
                        {provider.toUpperCase()}
                      </Badge>
                    ))}
                  </div>
                </div>
              </SelectItem>
            );
          })}
        </SelectContent>
      </Select>
      
      {/* 선택된 region이 유효하지 않은 경우 경고 */}
      {selectedRegion && !isRegionValid && (
        <p className="text-xs text-destructive">
          Selected region may not be available for all providers
        </p>
      )}
      
      {/* Provider별 region 수 표시 (Multi-provider 모드에서만) */}
      {providers.length > 1 && (
        <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
          {providers.map(provider => (
            <span key={provider}>
              {provider.toUpperCase()}: {regionsByProvider[provider]?.length || 0} regions
            </span>
          ))}
        </div>
      )}
    </div>
  );
}


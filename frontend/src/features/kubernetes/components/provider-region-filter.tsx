/**
 * Provider Region Filter Component
 * Provider별 Region 선택 필터 UI
 * 
 * Accordion + Chip 조합 방식
 */

'use client';

import * as React from 'react';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { X, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useProviderRegionFilter, type ProviderRegionSelection } from '@/hooks/use-provider-region-filter';
import type { CloudProvider } from '@/lib/types/kubernetes';
import { useTranslation } from '@/hooks/use-translation';
import { getRegionsByProvider, type RegionOption } from '@/lib/regions';
import { InlineSpinner } from '@/components/ui/spinner';

interface ProviderRegionFilterProps {
  providers: CloudProvider[];
  selectedRegions?: ProviderRegionSelection;
  onRegionSelectionChange?: (selectedRegions: ProviderRegionSelection) => void;
  disabled?: boolean;
  'aria-label'?: string;
  isLoading?: boolean;
}

const providerLabels: Record<CloudProvider, string> = {
  aws: 'AWS',
  gcp: 'GCP',
  azure: 'Azure',
  ncp: 'NCP',
};

const providerColors: Record<CloudProvider, string> = {
  aws: 'bg-orange-100 text-orange-800 border-orange-200',
  gcp: 'bg-blue-100 text-blue-800 border-blue-200',
  azure: 'bg-sky-100 text-sky-800 border-sky-200',
  ncp: 'bg-gray-100 text-gray-800 border-gray-200',
};

export function ProviderRegionFilter({
  providers,
  selectedRegions: externalSelectedRegions,
  onRegionSelectionChange,
  disabled = false,
  'aria-label': ariaLabel,
  isLoading = false,
}: ProviderRegionFilterProps) {
  const { t } = useTranslation();
  
  const {
    selectedRegions,
    regionsByProvider,
    toggleRegion,
    toggleAllRegionsForProvider,
    clearAllRegions,
    hasSelectedRegions,
    selectedCountsByProvider,
  } = useProviderRegionFilter({
    providers,
    initialSelectedRegions: externalSelectedRegions,
    onRegionSelectionChange: (regions) => {
      onRegionSelectionChange?.(regions);
    },
  });
  
  // 기본적으로 닫혀있는 상태 (사용자가 필요할 때 펼칠 수 있음)
  const [openProviders, setOpenProviders] = React.useState<Set<CloudProvider>>(new Set());
  
  const toggleProvider = React.useCallback((provider: CloudProvider) => {
    setOpenProviders(prev => {
      const next = new Set(prev);
      if (next.has(provider)) {
        next.delete(provider);
      } else {
        next.add(provider);
      }
      return next;
    });
  }, []);
  
  // Provider별 Region label 정보 매핑
         const regionMapsByProvider = React.useMemo(() => {
           const maps: Record<CloudProvider, Map<string, RegionOption>> = {
             aws: new Map(),
             gcp: new Map(),
             azure: new Map(),
             ncp: new Map(),
           };
    
    providers.forEach(provider => {
      const regionOptions = getRegionsByProvider(provider);
      const map = new Map<string, RegionOption>();
      regionOptions.forEach(option => {
        map.set(option.value, option);
      });
      maps[provider] = map;
    });
    
    return maps;
  }, [providers]);
  
  if (providers.length === 0) {
    return null;
  }
  
  return (
    <div className="space-y-3 min-w-0" role="group" aria-label={ariaLabel || (t('region.select') || 'Region Filters')}>
      <div className="flex items-center justify-between">
        <Label className="text-sm font-medium" id="region-filters-label">{t('region.title') || 'Regions'}</Label>
        {hasSelectedRegions && !disabled && (
          <Button
            variant="ghost"
            size="sm"
            onClick={clearAllRegions}
            className="h-7 text-xs"
          >
            Clear All
          </Button>
        )}
      </div>
      
      <div className="space-y-2">
        {providers.map(provider => {
          const providerRegions = regionsByProvider[provider] || [];
          const selectedProviderRegions = (selectedRegions as Record<CloudProvider, string[]>)[provider] || [];
          const allSelected = providerRegions.length > 0 && 
            providerRegions.every(r => selectedProviderRegions.includes(r));
          const someSelected = selectedProviderRegions.length > 0 && !allSelected;
          const selectedCount = selectedCountsByProvider[provider];
          const isOpen = openProviders.has(provider);
          const regionMap = regionMapsByProvider[provider];
          
          if (providerRegions.length === 0) {
            return null;
          }
          
          return (
            <Collapsible 
              key={provider} 
              open={isOpen}
              onOpenChange={() => toggleProvider(provider)}
              className="border rounded-lg px-3"
            >
              <div className="flex items-center gap-2 py-3">
                <Checkbox
                  id={`provider-checkbox-${provider}`}
                  checked={allSelected}
                  ref={(el) => {
                    if (el && 'indeterminate' in el) {
                      (el as HTMLInputElement).indeterminate = someSelected;
                    }
                  }}
                  onCheckedChange={() => {
                    if (!disabled) {
                      toggleAllRegionsForProvider(provider);
                    }
                  }}
                  disabled={disabled}
                  className="shrink-0"
                  onClick={(e) => {
                    e.stopPropagation();
                  }}
                />
                <Label
                  htmlFor={`provider-checkbox-${provider}`}
                  className="shrink-0 cursor-pointer flex items-center"
                  onClick={(e) => {
                    e.stopPropagation();
                    if (!disabled) {
                      toggleAllRegionsForProvider(provider);
                    }
                  }}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault();
                      e.stopPropagation();
                      if (!disabled) {
                        toggleAllRegionsForProvider(provider);
                      }
                    }
                  }}
                  role="button"
                  tabIndex={0}
                  aria-label={t('common.selectAll') || 'Select all regions'}
                >
                  <span className="sr-only">{t('common.selectAll') || 'Select all regions'}</span>
                </Label>
                <CollapsibleTrigger asChild className="flex-1">
                  <button
                    type="button"
                    className="flex items-center justify-between w-full pr-4 py-0 text-sm font-medium transition-all hover:underline text-left group"
                  >
                    <span className="font-medium">{providerLabels[provider]}</span>
                    <div className="flex items-center gap-2">
                      <Badge 
                        variant="outline" 
                        className={cn(providerColors[provider], "ml-2")}
                      >
                        {selectedCount > 0 ? `${selectedCount} selected` : `${providerRegions.length} regions`}
                      </Badge>
                      <ChevronDown className={cn(
                        "h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200",
                        isOpen && "rotate-180"
                      )} />
                    </div>
                  </button>
                </CollapsibleTrigger>
              </div>
              <CollapsibleContent className="overflow-hidden data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down">
                <div className="space-y-2 pb-2">
                  {providerRegions.map(region => {
                    const isSelected = selectedProviderRegions.includes(region);
                    const regionInfo = regionMap.get(region);
                    
                    return (
                      <div
                        key={region}
                        className="flex items-center space-x-2 py-1 px-2 rounded hover:bg-accent"
                      >
                        <Checkbox
                          id={`${provider}-${region}`}
                          checked={isSelected}
                          onCheckedChange={() => toggleRegion(provider, region)}
                          disabled={disabled}
                        />
                        <Label
                          htmlFor={`${provider}-${region}`}
                          className="flex-1 cursor-pointer text-sm"
                        >
                          <span className="font-medium">{region}</span>
                          {regionInfo?.label && (
                            <span className="text-muted-foreground ml-1.5 text-xs">
                              ({regionInfo.label})
                            </span>
                          )}
                        </Label>
                        {isLoading && isSelected && (
                          <InlineSpinner className="ml-2" aria-label={t('common.loadingData') || '로딩 중'} />
                        )}
                      </div>
                    );
                  })}
                </div>
              </CollapsibleContent>
            </Collapsible>
          );
        })}
      </div>
      
      {/* 선택된 Region Chip 표시 */}
      {hasSelectedRegions && (
        <div className="flex flex-wrap gap-2 pt-2 border-t">
          {providers.map(provider => {
            const selectedProviderRegions = (selectedRegions as Record<CloudProvider, string[]>)[provider] || [];
            if (selectedProviderRegions.length === 0) return null;
            
            const regionMap = regionMapsByProvider[provider];
            
            return (
              <div key={provider} className="flex flex-wrap gap-1.5">
                {selectedProviderRegions.map((region: string) => {
                  const regionInfo = regionMap?.get(region);
                  
                  return (
                    <Badge
                      key={`${provider}-${region}`}
                      variant="secondary"
                      className={cn(
                        "gap-1 pr-1",
                        providerColors[provider]
                      )}
                    >
                      <span className="text-xs font-medium">
                        {providerLabels[provider]}:
                      </span>
                      <span className="text-xs font-medium">{region}</span>
                      {regionInfo?.label && (
                        <span className="text-xs text-muted-foreground ml-0.5">
                          ({regionInfo.label})
                        </span>
                      )}
                      {isLoading && (
                        <InlineSpinner className="ml-1.5" aria-label={t('common.loadingData') || '로딩 중'} />
                      )}
                      {!disabled && !isLoading && (
                        <button
                          type="button"
                          onClick={() => toggleRegion(provider, region)}
                          className="ml-1 rounded-full hover:bg-black/10 p-0.5"
                          aria-label={t('common.removeRegion', { region, provider: providerLabels[provider] })}
                        >
                          <X className="h-3 w-3" />
                        </button>
                      )}
                    </Badge>
                  );
                })}
              </div>
            );
          })}
        </div>
      )}
      
      {/* 로딩 상태 메시지 */}
      {isLoading && hasSelectedRegions && (
        <div className="flex items-center gap-2 pt-2 border-t text-sm text-muted-foreground">
          <InlineSpinner aria-label={t('common.loadingData') || '로딩 중'} />
          <span>{t('common.loadingClusters') || '클러스터 데이터 로딩 중...'}</span>
        </div>
      )}
    </div>
  );
}


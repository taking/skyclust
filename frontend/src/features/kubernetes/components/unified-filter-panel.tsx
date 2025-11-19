/**
 * Unified Filter Panel Component
 * 통합 필터 패널 - Credential과 Region 선택을 한 곳에 모음
 * 
 * 기능:
 * - Card 기반 통합 필터 패널
 * - 접기/펼치기 기능
 * - 필터 상태 요약 표시
 * - 빠른 액션 버튼 (Clear All, Select All Regions)
 */

'use client';

import * as React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import { ChevronDown, ChevronUp, Filter, X, Info } from 'lucide-react';
import { CredentialListFilter } from './credential-list-filter';
import { ProviderRegionFilter } from './provider-region-filter';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import type { Credential, CloudProvider } from '@/lib/types';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';
import { useTranslation } from '@/hooks/use-translation';

export interface UnifiedFilterPanelProps {
  credentials: Credential[];
  selectedCredentialIds: string[];
  onCredentialSelectionChange: (credentialIds: string[]) => void;
  selectedProviders: CloudProvider[];
  selectedRegion?: string | null;
  onRegionChange?: (region: string | null) => void;
  selectedRegions?: ProviderRegionSelection;
  onRegionSelectionChange?: (selectedRegions: ProviderRegionSelection) => void;
  disabled?: boolean;
  defaultExpanded?: boolean;
  isLoading?: boolean;
}

export function UnifiedFilterPanel({
  credentials,
  selectedCredentialIds,
  onCredentialSelectionChange,
  selectedProviders,
  selectedRegion,
  onRegionChange,
  selectedRegions,
  onRegionSelectionChange,
  disabled = false,
  defaultExpanded = true,
  isLoading = false,
}: UnifiedFilterPanelProps) {
  const { t } = useTranslation();
  const [isExpanded, setIsExpanded] = React.useState(defaultExpanded);
  
  
  // selectedRegion을 ProviderRegionSelection으로 변환 (단일 Credential 지원)
  const normalizedSelectedRegions = React.useMemo(() => {
    if (selectedRegions) {
      return selectedRegions;
    }
    
    if (selectedRegion && selectedProviders.length > 0) {
      const result: ProviderRegionSelection = {
        aws: [],
        gcp: [],
        azure: [],
      } as ProviderRegionSelection;
      
      selectedProviders.forEach(provider => {
        (result as Record<CloudProvider, string[]>)[provider] = [selectedRegion];
      });
      
      return result;
    }
    
    return {
      aws: [],
      gcp: [],
      azure: [],
    } as ProviderRegionSelection;
  }, [selectedRegions, selectedRegion, selectedProviders]);
  
  // Region 선택 변경 핸들러 (ProviderRegionSelection으로 변환)
  const handleRegionSelectionChange = React.useCallback((regions: ProviderRegionSelection) => {
    if (onRegionSelectionChange) {
      onRegionSelectionChange(regions);
    }
    
    if (onRegionChange && selectedProviders.length === 1) {
      const singleProvider = selectedProviders[0];
      const providerRegions = (regions as Record<CloudProvider, string[]>)[singleProvider] || [];
      if (providerRegions.length === 1) {
        onRegionChange(providerRegions[0]);
      } else if (providerRegions.length === 0) {
        onRegionChange(null);
      }
    }
  }, [onRegionSelectionChange, onRegionChange, selectedProviders]);
  
  // 필터 상태 요약 계산
  const filterSummary = React.useMemo(() => {
    const credentialCount = selectedCredentialIds.length;
    const regionCount = Object.values(normalizedSelectedRegions).reduce(
      (sum, regions) => sum + regions.length,
      0
    );
    
    return {
      credentialCount,
      regionCount,
      hasFilters: credentialCount > 0 || regionCount > 0,
    };
  }, [selectedCredentialIds.length, normalizedSelectedRegions]);
  
  // Clear All 핸들러 (debounce 적용, 중복 클릭 방지)
  const clearAllTimeoutRef = React.useRef<NodeJS.Timeout | null>(null);
  const isClearingRef = React.useRef(false);
  
  // 최신 값 참조를 위한 ref
  const selectedCredentialIdsRef = React.useRef(selectedCredentialIds);
  const normalizedSelectedRegionsRef = React.useRef(normalizedSelectedRegions);
  
  // ref 업데이트
  React.useEffect(() => {
    selectedCredentialIdsRef.current = selectedCredentialIds;
  }, [selectedCredentialIds]);
  
  React.useEffect(() => {
    normalizedSelectedRegionsRef.current = normalizedSelectedRegions;
  }, [normalizedSelectedRegions]);
  
  const handleClearAll = React.useCallback(() => {
    if (disabled || isClearingRef.current) return;
    
    // 최신 값으로 확인
    const currentCredentialIds = selectedCredentialIdsRef.current;
    const currentRegions = normalizedSelectedRegionsRef.current;
    
    // 이미 초기화된 상태면 스킵
    const hasCredentials = currentCredentialIds.length > 0;
    const hasRegions = Object.values(currentRegions).some(r => r.length > 0);
    
    if (!hasCredentials && !hasRegions) {
      return;
    }
    
    // Debounce 적용: 기존 timeout 취소
    if (clearAllTimeoutRef.current) {
      clearTimeout(clearAllTimeoutRef.current);
      clearAllTimeoutRef.current = null;
    }
    
    isClearingRef.current = true;
    
    clearAllTimeoutRef.current = setTimeout(() => {
      // 실행 전에 최신 값으로 다시 한번 확인
      const latestCredentialIds = selectedCredentialIdsRef.current;
      const latestRegions = normalizedSelectedRegionsRef.current;
      
      const latestHasCredentials = latestCredentialIds.length > 0;
      const latestHasRegions = Object.values(latestRegions).some(r => r.length > 0);
      
      if (!latestHasCredentials && !latestHasRegions) {
        // 이미 초기화된 상태면 스킵
        isClearingRef.current = false;
        clearAllTimeoutRef.current = null;
        return;
      }
      
      // 초기화 실행
      onCredentialSelectionChange([]);
      
      if (onRegionSelectionChange) {
        const emptyRegions: ProviderRegionSelection = {
          aws: [],
          gcp: [],
          azure: [],
        };
        onRegionSelectionChange(emptyRegions);
      }
      
      if (onRegionChange) {
        onRegionChange(null);
      }
      
      isClearingRef.current = false;
      clearAllTimeoutRef.current = null;
    }, 150);
  }, [disabled, onCredentialSelectionChange, onRegionSelectionChange, onRegionChange]);
  
  // Cleanup
  React.useEffect(() => {
    return () => {
      if (clearAllTimeoutRef.current) {
        clearTimeout(clearAllTimeoutRef.current);
      }
    };
  }, []);
  
  
  return (
    <Card className="w-full">
      <CardHeader className="pb-3">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
          <div className="flex items-center gap-2 flex-wrap">
            <CardTitle className="flex items-center gap-2 text-base">
              <Filter className="h-4 w-4" aria-hidden="true" />
              <span>{t('common.filter') || 'Filters'}</span>
            </CardTitle>
            {filterSummary.hasFilters && (
              <Badge variant="secondary" className="text-xs">
                {filterSummary.credentialCount > 0 && `${filterSummary.credentialCount} ${t('credential.title') || 'credentials'}`}
                {filterSummary.credentialCount > 0 && filterSummary.regionCount > 0 && ', '}
                {filterSummary.regionCount > 0 && `${filterSummary.regionCount} ${t('region.select') || 'regions'}`}
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-2 flex-shrink-0">
            {filterSummary.hasFilters && !disabled && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleClearAll}
                className="h-7 text-xs"
                aria-label={t('common.clearAllFilters') || 'Clear all filters'}
                disabled={isClearingRef.current}
              >
                <X className="h-3 w-3 mr-1" aria-hidden="true" />
                <span className="sr-only sm:not-sr-only">{t('common.clearAllFilters') || 'Clear All Filters'}</span>
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
              className="h-7 w-7 p-0"
              aria-label={isExpanded ? (t('common.collapse') || 'Collapse filters') : (t('common.expand') || 'Expand filters')}
              aria-expanded={isExpanded}
            >
              {isExpanded ? (
                <ChevronUp className="h-4 w-4" aria-hidden="true" />
              ) : (
                <ChevronDown className="h-4 w-4" aria-hidden="true" />
              )}
            </Button>
          </div>
        </div>
        {filterSummary.hasFilters && (
          <CardDescription className="text-xs mt-1">
            {filterSummary.credentialCount} {t('credential.title') || 'credential(s)'}, {filterSummary.regionCount} {t('region.select') || 'region(s)'} selected
          </CardDescription>
        )}
      </CardHeader>
      {isExpanded && (
        <CardContent className="pt-0">
          <div className="flex flex-col md:flex-row gap-4 md:gap-6">
            {/* Credentials Section */}
            <div className="flex-1 min-w-0">
              <div id="credential-filter" className="min-w-0">
                <CredentialListFilter
                  credentials={credentials}
                  selectedCredentialIds={selectedCredentialIds}
                  onSelectionChange={onCredentialSelectionChange}
                  disabled={disabled}
                />
              </div>
            </div>
            
            {/* Divider - Desktop only */}
            <div className="hidden md:block w-px bg-border self-stretch flex-shrink-0" aria-hidden="true" />
            
            {/* Regions Section - 항상 표시 */}
            <div className="flex-1 min-w-0">
              <div id="region-filter" className="min-w-0">
                {selectedCredentialIds.length === 0 ? (
                  <div className="space-y-3">
                    <Alert 
                      variant="default" 
                      className="bg-blue-50 dark:bg-blue-950 border-blue-200 dark:border-blue-800"
                      role="alert"
                      aria-live="polite"
                    >
                      <Info className="h-4 w-4 text-blue-600 dark:text-blue-400" aria-hidden="true" />
                      <AlertTitle className="text-blue-900 dark:text-blue-100">
                        {t('common.selectCredentialFirst') || '자격증명을 먼저 선택해주세요'}
                      </AlertTitle>
                      <AlertDescription className="text-blue-800 dark:text-blue-200">
                        {t('common.selectCredentialFirstDescription') || '리전을 선택하려면 먼저 자격증명을 선택해주세요.'}
                      </AlertDescription>
                    </Alert>
                    <ProviderRegionFilter
                      providers={[]}
                      selectedRegions={normalizedSelectedRegions}
                      onRegionSelectionChange={handleRegionSelectionChange}
                      disabled={true}
                      aria-label={t('common.selectCredentialFirstDescription') || '자격증명을 선택하면 사용 가능한 리전이 표시됩니다.'}
                    />
                  </div>
                ) : (
                  <ProviderRegionFilter
                    providers={selectedProviders}
                    selectedRegions={normalizedSelectedRegions}
                    onRegionSelectionChange={handleRegionSelectionChange}
                    disabled={disabled || isLoading}
                    isLoading={isLoading}
                  />
                )}
              </div>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
}


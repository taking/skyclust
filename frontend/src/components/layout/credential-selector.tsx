/**
 * CredentialSelector 컴포넌트
 * 
 * 사이드바에 표시되는 자격 증명 선택 컴포넌트입니다.
 * Workspace Selector 아래에 위치하여 Workspace → Credential → Region 선택 흐름을 제공합니다.
 * 
 * @example
 * ```tsx
 * // 사이드바에서 자동으로 사용됨
 * <Sidebar>
 *   <WorkspaceSelector />
 *   <CredentialSelector />  // 자동으로 표시됨
 * </Sidebar>
 * ```
 * 
 * 기능:
 * - 자격 증명 선택 및 변경
 * - 프로바이더별 리전 선택 (GCP, AWS, Azure)
 * - 자격 증명 변경 시 기본 리전 자동 설정
 * - URL 쿼리 파라미터와 동기화
 */

'use client';

import * as React from 'react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Key, Plus, ChevronDown, Check } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentials } from '@/hooks/use-credentials';
import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { usePathname, useSearchParams } from 'next/navigation';
import { getRegionsForProvider, supportsRegionSelection, getDefaultRegionForProvider } from '@/lib/regions';
import type { CloudProvider } from '@/lib/types/kubernetes';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';

export function CredentialSelector() {
  const { t } = useTranslation();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion } = useCredentialContext();
  
  // 자격 증명 선택기를 표시할 경로인지 확인
  const shouldShow = React.useMemo(() => {
    return pathname.startsWith('/compute') || 
           pathname.startsWith('/kubernetes') || 
           pathname.startsWith('/networks') ||
           pathname.startsWith('/dashboard');
  }, [pathname]);

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: !!currentWorkspace && shouldShow,
  });
  
  const regions = React.useMemo(() => getRegionsForProvider(selectedProvider), [selectedProvider]);
  const showRegionSelector = supportsRegionSelection(selectedProvider);
  
  /**
   * 자격 증명 변경 핸들러
   * - 자격 증명 변경 시 리전 초기화
   * - 프로바이더가 리전을 지원하면 기본 리전으로 자동 설정
   * - URL 쿼리 파라미터 업데이트
   */
  const handleCredentialChange = (credentialId: string) => {
    // 1. 전역 상태에 선택된 자격 증명 저장
    setSelectedCredential(credentialId);
    
    // 2. URL 쿼리 파라미터 업데이트 준비
    const params = new URLSearchParams(searchParams.toString());
    if (credentialId) {
      params.set('credentialId', credentialId);
    } else {
      params.delete('credentialId');
    }
    
    // 3. 변경된 자격 증명 정보 조회
    const newCredential = credentials.find((c) => c.id === credentialId);
    
    // 4. 자격 증명이 없거나 리전을 지원하지 않는 경우 리전 제거
    if (!newCredential) {
      // 자격 증명을 찾을 수 없는 경우 리전 초기화
      params.delete('region');
      setSelectedRegion(null);
    } else if (!supportsRegionSelection(newCredential.provider as CloudProvider)) {
      // 프로바이더가 리전 선택을 지원하지 않는 경우 리전 초기화
      params.delete('region');
      setSelectedRegion(null);
    } else {
      // 5. 프로바이더가 리전을 지원하는 경우 기본 리전으로 설정
      const defaultRegion = getDefaultRegionForProvider(newCredential.provider);
      if (defaultRegion) {
        // 기본 리전이 있으면 URL과 상태 모두 업데이트
        params.set('region', defaultRegion);
        setSelectedRegion(defaultRegion);
      } else {
        // 기본 리전이 없으면 리전 초기화
        params.delete('region');
        setSelectedRegion(null);
      }
    }
    
    // 6. URL 업데이트 (스크롤 없이)
    router.replace(`${pathname}?${params.toString()}`, { scroll: false });
  };
  
  /**
   * 리전 변경 핸들러
   * - 리전 선택 시 URL 쿼리 파라미터 업데이트
   */
  const handleRegionChange = (region: string) => {
    // 1. 'all' 선택 시 빈 문자열로 변환 (전체 리전 선택 의미)
    const regionValue = region === 'all' ? '' : region;
    
    // 2. 전역 상태에 선택된 리전 저장 (빈 문자열이면 null로 변환)
    setSelectedRegion(regionValue || null);
    
    // 3. URL 쿼리 파라미터 업데이트
    const params = new URLSearchParams(searchParams.toString());
    if (regionValue) {
      // 리전이 선택된 경우 URL에 추가
      params.set('region', regionValue);
    } else {
      // 전체 리전 선택('all')인 경우 URL에서 제거
      params.delete('region');
    }
    
    // 4. URL 업데이트 (스크롤 없이)
    router.replace(`${pathname}?${params.toString()}`, { scroll: false });
  };

  if (!shouldShow || !currentWorkspace) {
    return null;
  }

  // 자격 증명이 없을 때 안내 메시지 표시
  if (credentials.length === 0) {
    return (
      <div className="mb-4 space-y-2">
        <label className="text-sm font-medium text-muted-foreground">
          {t('credential.title')}
        </label>
        <div className="flex items-center gap-2 p-3 rounded-lg border border-yellow-200/50 bg-yellow-50/50">
          <Key className="h-4 w-4 text-yellow-600 flex-shrink-0" />
          <div className="flex-1 min-w-0">
            {/* <p className="text-xs text-yellow-800 truncate">
              {t('components.credentialIndicator.noCredentials')}
            </p> */}
          </div>
          <Button
            variant="outline"
            size="sm"
            className="h-7 px-2 text-xs flex-shrink-0"
            onClick={() => router.push('/credentials')}
          >
            <Plus className="h-3 w-3 mr-1" />
            {t('credential.create')}
          </Button>
        </div>
      </div>
    );
  }
  
  // 자격 증명이 있지만 선택되지 않았을 때
  if (!selectedCredentialId) {
    return (
      <div className="mb-4 space-y-2">
        <label className="text-sm font-medium text-muted-foreground">
          {t('credential.title')}
        </label>
        <Select value="" onValueChange={handleCredentialChange}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder={t('credential.select')} />
          </SelectTrigger>
          <SelectContent>
            {credentials.map((credential) => (
              <SelectItem key={credential.id} value={credential.id} className="truncate">
                {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
              </SelectItem>
            ))}
            <div className="h-px bg-border my-1" />
            <div className="px-2 py-1.5">
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start h-8"
                onClick={() => router.push('/credentials')}
              >
                <Plus className="mr-2 h-4 w-4" />
                {t('credential.create')}
              </Button>
            </div>
          </SelectContent>
        </Select>
      </div>
    );
  }
  
  // 자격 증명이 선택된 경우
  return (
    <div className="mb-4 space-y-2">
      <label className="text-sm font-medium text-muted-foreground">
        {t('credential.title')}
      </label>
      <Select value={selectedCredentialId} onValueChange={handleCredentialChange}>
        <SelectTrigger className="w-full">
          <SelectValue>
            <div className="flex items-center gap-2 min-w-0">
              <span className="truncate">
                {selectedCredential?.name || 
                 `${selectedCredential?.provider.toUpperCase()} (${selectedCredentialId.slice(0, 8)})`}
              </span>
              <Badge variant="outline" className="text-xs flex-shrink-0">
                {selectedCredential?.provider.toUpperCase()}
              </Badge>
            </div>
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          {credentials.map((credential) => (
            <SelectItem key={credential.id} value={credential.id} className="truncate">
              <div className="flex items-center gap-2 w-full">
                <span className="truncate">
                  {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
                </span>
                {credential.id === selectedCredentialId && (
                  <Check className="h-3 w-3 ml-auto flex-shrink-0" />
                )}
              </div>
            </SelectItem>
          ))}
          <div className="h-px bg-border my-1" />
          <div className="px-2 py-1.5">
            <Button
              variant="ghost"
              size="sm"
              className="w-full justify-start h-8"
              onClick={() => router.push('/credentials')}
            >
              <Plus className="mr-2 h-4 w-4" />
              {t('credential.create')}
            </Button>
          </div>
        </SelectContent>
      </Select>
      
      {/* 리전 선택기 (프로바이더가 리전을 지원하는 경우) */}
      {showRegionSelector && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-muted-foreground">
            {t('region.select')}
          </label>
          <Select 
            value={selectedRegion || 'all'} 
            onValueChange={handleRegionChange}
          >
            <SelectTrigger className="w-full">
              <SelectValue>
                {selectedRegion ? (
                  <div className="flex items-center gap-2">
                    <span>{selectedRegion}</span>
                    {regions.find(r => r.value === selectedRegion) && (
                      <Badge variant="secondary" className="text-xs">
                        {regions.find(r => r.value === selectedRegion)?.label}
                      </Badge>
                    )}
                  </div>
                ) : (
                  t('region.all')
                )}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">
                {t('region.all')}
              </SelectItem>
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
        </div>
      )}
    </div>
  );
}


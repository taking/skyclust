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
import { getRegionsByProvider, supportsRegionSelection, getDefaultRegionForProvider } from '@/lib/regions';
import type { CloudProvider } from '@/lib/types/kubernetes';
import { useResourceGroups } from '@/hooks/use-resource-groups';
import { ResourceGroupSelect } from './resource-group-select';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';
import { useResourceContext } from '@/hooks/use-resource-context';
import { buildResourcePath, buildManagementPath, updatePathFilters } from '@/lib/routing/helpers';

export function CredentialSelector() {
  const { t } = useTranslation();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();
  const { 
    selectedCredentialId, 
    selectedRegion, 
    selectedResourceGroup,
    setSelectedCredential, 
    setSelectedRegion,
    setSelectedResourceGroup,
  } = useCredentialContext();
  
  // 자격 증명 선택기를 표시할 경로인지 확인
  const shouldShow = React.useMemo(() => {
    return pathname.startsWith('/compute') || 
           pathname.startsWith('/kubernetes') || 
           pathname.startsWith('/networks') ||
           pathname.startsWith('/dashboard') ||
           pathname.startsWith('/azure/iam');
  }, [pathname]);

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: !!currentWorkspace && shouldShow,
  });
  
  const regions = React.useMemo(() => getRegionsByProvider(selectedProvider), [selectedProvider]);
  const showRegionSelector = supportsRegionSelection(selectedProvider);
  const isAzure = selectedProvider === 'azure';
  
  // Azure Resource Groups 조회 (Azure credential이 선택된 경우에만)
  // Sidebar용이므로 limit=100으로 설정하여 충분한 데이터 조회
  const { data: resourceGroups = [], isLoading: isLoadingResourceGroups, error: resourceGroupsError } = useResourceGroups({
    credentialId: isAzure && selectedCredentialId ? selectedCredentialId : undefined,
    enabled: isAzure && !!selectedCredentialId,
    limit: 100, // Sidebar용: 충분한 데이터 조회
  });
  
  // 디버깅: 데이터 확인
  React.useEffect(() => {
    if (isAzure && selectedCredentialId && process.env.NODE_ENV === 'development') {
      import('@/lib/logging').then(({ log }) => {
        log.debug('Azure Resource Groups', {
          isAzure,
          selectedCredentialId,
          resourceGroupsCount: resourceGroups.length,
          resourceGroups,
          isLoading: isLoadingResourceGroups,
          error: resourceGroupsError,
        });
      });
    }
  }, [isAzure, selectedCredentialId, resourceGroups, isLoadingResourceGroups, resourceGroupsError]);
  
  // Path Parameter에서 컨텍스트 추출 (새로운 라우팅 구조)
  const resourceContext = useResourceContext();
  
  /**
   * 자격 증명 변경 핸들러
   * - 자격 증명 변경 시 리전 초기화
   * - 프로바이더가 리전을 지원하면 기본 리전으로 자동 설정
   * - 새로운 라우팅 구조에 맞게 URL 업데이트
   */
  const handleCredentialChange = (credentialId: string) => {
    // 1. 전역 상태에 선택된 자격 증명 저장
    setSelectedCredential(credentialId);
    
    // 2. 변경된 자격 증명 정보 조회
    const newCredential = credentials.find((c) => c.id === credentialId);
    
    // 3. 새로운 라우팅 구조에 맞게 경로 업데이트
    if (resourceContext.isResourcePage && resourceContext.workspaceId) {
      // 리소스 페이지: /{workspaceId}/{credentialId}/resource-type/...
      // credentialId를 path parameter로 업데이트
      const currentPathParts = pathname.split('/');
      const resourceTypeIndex = currentPathParts.findIndex(part => 
        ['kubernetes', 'networks', 'compute', 'azure'].includes(part)
      );
      
      if (resourceTypeIndex > 0) {
        // 기존 credentialId 위치에 새 credentialId로 교체
        const newPathParts = [...currentPathParts];
        newPathParts[resourceTypeIndex - 1] = credentialId;
        const newPath = newPathParts.join('/');
        
        // 리전 처리
        let region = resourceContext.region;
        if (newCredential) {
          if (supportsRegionSelection(newCredential.provider as CloudProvider)) {
            const defaultRegion = getDefaultRegionForProvider(newCredential.provider);
            if (defaultRegion && !region) {
              region = defaultRegion;
              setSelectedRegion(defaultRegion);
            }
          } else {
            region = undefined;
            setSelectedRegion(null);
          }
        }
        
        // Azure의 경우 Resource Group 초기화
        if (newCredential?.provider === 'azure') {
          setSelectedResourceGroup(null);
        }
        
        const filters: Record<string, string | undefined> = {};
        if (region) filters.region = region;
        
        const finalPath = buildResourcePath(
          resourceContext.workspaceId,
          credentialId,
          currentPathParts[resourceTypeIndex] as 'kubernetes' | 'networks' | 'compute' | 'azure',
          '/' + currentPathParts.slice(resourceTypeIndex + 1).join('/'),
          filters
        );
        
        router.replace(finalPath, { scroll: false });
      }
    } else if (resourceContext.isManagementPage && resourceContext.workspaceId) {
      // 관리 페이지: /{workspaceId}/management-type?credentialId=xxx
      // credentialId를 query parameter로 업데이트
      const filters: Record<string, string | undefined> = {};
      if (credentialId) filters.credentialId = credentialId;
      
      // 리전 처리
      let region = resourceContext.region;
      if (newCredential) {
        if (supportsRegionSelection(newCredential.provider as CloudProvider)) {
          const defaultRegion = getDefaultRegionForProvider(newCredential.provider);
          if (defaultRegion && !region) {
            region = defaultRegion;
            setSelectedRegion(defaultRegion);
          }
          if (region) filters.region = region;
        } else {
          setSelectedRegion(null);
        }
      }
      
      // Azure의 경우 Resource Group 초기화
      if (newCredential?.provider === 'azure') {
        setSelectedResourceGroup(null);
      }
      
      const managementType = pathname.split('/').pop() || 'dashboard';
      const finalPath = buildManagementPath(resourceContext.workspaceId, managementType, filters);
      router.replace(finalPath, { scroll: false });
    } else {
      // 기존 방식 (레거시 라우트)
      const params = new URLSearchParams(searchParams.toString());
      if (credentialId) {
        params.set('credentialId', credentialId);
      } else {
        params.delete('credentialId');
      }
      
      if (newCredential) {
        if (supportsRegionSelection(newCredential.provider as CloudProvider)) {
          const defaultRegion = getDefaultRegionForProvider(newCredential.provider);
          if (defaultRegion) {
            params.set('region', defaultRegion);
            setSelectedRegion(defaultRegion);
          }
        } else {
          params.delete('region');
          setSelectedRegion(null);
        }
      }
      
      router.replace(`${pathname}?${params.toString()}`, { scroll: false });
    }
  };
  
  /**
   * Region 변경 핸들러
   * - 새로운 라우팅 구조에 맞게 URL 업데이트
   */
  const handleRegionChange = (region: string) => {
    const regionValue = region === 'all' ? '' : region;
    setSelectedRegion(regionValue || null);
    
    // 새로운 라우팅 구조에 맞게 경로 업데이트
    if (resourceContext.isResourcePage && resourceContext.workspaceId && resourceContext.credentialId) {
      // 리소스 페이지: 필터만 업데이트 (query parameter)
      const newPath = updatePathFilters(pathname, { region: regionValue || undefined });
      router.replace(newPath, { scroll: false });
    } else if (resourceContext.isManagementPage && resourceContext.workspaceId) {
      // 관리 페이지: query parameter로 업데이트
      const filters: Record<string, string | undefined> = {};
      if (resourceContext.credentialId) filters.credentialId = resourceContext.credentialId;
      if (regionValue) filters.region = regionValue;
      
      const managementType = pathname.split('/').pop() || 'dashboard';
      const finalPath = buildManagementPath(resourceContext.workspaceId, managementType, filters);
      router.replace(finalPath, { scroll: false });
    } else {
      // 기존 방식 (레거시 라우트)
      const params = new URLSearchParams(searchParams.toString());
      if (regionValue) {
        params.set('region', regionValue);
      } else {
        params.delete('region');
      }
      router.replace(`${pathname}?${params.toString()}`, { scroll: false });
    }
  };

  /**
   * Resource Group 변경 핸들러 (Azure 전용)
   * - Resource Group 선택 시 URL 쿼리 파라미터 업데이트
   */
  const handleResourceGroupChange = (resourceGroup: string) => {
    // 1. 'all' 선택 시 빈 문자열로 변환 (전체 Resource Group 선택 의미)
    const rgValue = resourceGroup === 'all' ? '' : resourceGroup;
    
    // 2. 전역 상태에 선택된 Resource Group 저장 (빈 문자열이면 null로 변환)
    setSelectedResourceGroup(rgValue || null);
    
    // 3. URL 쿼리 파라미터 업데이트
    const params = new URLSearchParams(searchParams.toString());
    if (rgValue) {
      // Resource Group이 선택된 경우 URL에 추가
      params.set('resourceGroup', rgValue);
    } else {
      // 전체 Resource Group 선택('all')인 경우 URL에서 제거
      params.delete('resourceGroup');
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
            <p className="text-xs text-yellow-800 truncate">
              {t('components.credentialIndicator.noCredentials')}
            </p>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="h-7 px-2 text-xs flex-shrink-0"
            onClick={() => {
              if (currentWorkspace?.id) {
                router.push(buildManagementPath(currentWorkspace.id, 'credentials'));
              } else {
                router.push('/credentials');
              }
            }}
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
                onClick={() => {
              if (currentWorkspace?.id) {
                router.push(buildManagementPath(currentWorkspace.id, 'credentials'));
              } else {
                router.push('/credentials');
              }
            }}
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
              onClick={() => {
              if (currentWorkspace?.id) {
                router.push(buildManagementPath(currentWorkspace.id, 'credentials'));
              } else {
                router.push('/credentials');
              }
            }}
            >
              <Plus className="mr-2 h-4 w-4" />
              {t('credential.create')}
            </Button>
          </div>
        </SelectContent>
      </Select>
      
      {/* Azure의 경우 Resource Group 선택기 (Region보다 먼저 표시) */}
      {isAzure && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-muted-foreground">
            Resource Group
          </label>
          <ResourceGroupSelect
            resourceGroups={resourceGroups}
            selectedResourceGroup={selectedResourceGroup}
            onValueChange={(value) => handleResourceGroupChange(value || 'all')}
            isLoading={isLoadingResourceGroups}
            error={resourceGroupsError}
          />
        </div>
      )}
      
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


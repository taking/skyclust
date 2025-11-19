/**
 * Credential Indicator Component
 * 페이지 상단에 현재 선택된 자격증명을 표시하는 인디케이터 컴포넌트
 * 
 * Card 형태로 현재 자격증명을 표시하고, "Change" 버튼으로 변경할 수 있는 드롭다운 제공
 */

'use client';

import * as React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Key, ChevronDown, Check, Plus } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentials } from '@/hooks/use-credentials';
import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import type { Credential } from '@/lib/types/credential';
import type { CloudProvider } from '@/lib/types/kubernetes';
import { getRegionsByProvider, supportsRegionSelection } from '@/lib/regions';
import { buildManagementPath } from '@/lib/routing/helpers';

interface CredentialIndicatorProps {
  /**
   * Region 선택 표시 여부
   */
  showRegion?: boolean;
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

export function CredentialIndicator({
  showRegion = true,
  className,
}: CredentialIndicatorProps) {
  const { t } = useTranslation();
  const router = useRouter();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion } = useCredentialContext();
  
  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: !!currentWorkspace,
  });
  
  const regions = React.useMemo(() => getRegionsByProvider(selectedProvider), [selectedProvider]);
  const showRegionSelector = showRegion && supportsRegionSelection(selectedProvider);
  
  const handleCredentialChange = (credentialId: string) => {
    setSelectedCredential(credentialId);
    
    // URL 업데이트
    const params = new URLSearchParams(window.location.search);
    if (credentialId) {
      params.set('credentialId', credentialId);
    } else {
      params.delete('credentialId');
    }
    
    // Provider 변경 시 Region 초기화
    const newCredential = credentials.find((c) => c.id === credentialId);
    if (newCredential && !supportsRegionSelection(newCredential.provider as CloudProvider)) {
      params.delete('region');
      setSelectedRegion(null);
    }
    
    router.replace(`${window.location.pathname}?${params.toString()}`, { scroll: false });
  };
  
  const handleRegionChange = (region: string) => {
    const regionValue = region === 'all' ? '' : region;
    setSelectedRegion(regionValue || null);
    
    // URL 업데이트
    const params = new URLSearchParams(window.location.search);
    if (regionValue) {
      params.set('region', regionValue);
    } else {
      params.delete('region');
    }
    
    router.replace(`${window.location.pathname}?${params.toString()}`, { scroll: false });
  };
  
  // 자격증명이 없을 때
  if (credentials.length === 0) {
    return (
      <Card className={cn('border-yellow-200/50 bg-yellow-50/50 shadow-sm', className)}>
        <CardContent className="flex items-center justify-between py-2.5 px-4">
          <div className="flex items-center gap-2">
            <Key className="h-4 w-4 text-yellow-600" />
            <span className="text-sm text-yellow-800">
              {t('components.credentialIndicator.noCredentials')}
            </span>
          </div>
          <Button
            variant="outline"
            size="sm"
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
        </CardContent>
      </Card>
    );
  }
  
  // 자격증명이 있지만 선택되지 않았을 때
  if (!selectedCredentialId) {
    return (
      <Card className={cn('border-blue-200/50 bg-blue-50/50 shadow-sm', className)}>
        <CardContent className="flex items-center justify-between py-2.5 px-4">
          <div className="flex items-center gap-2">
            <Key className="h-4 w-4 text-blue-600" />
            <span className="text-sm text-blue-800">
              {t('components.credentialIndicator.selectCredential')}
            </span>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                {t('credential.select')}
                <ChevronDown className="ml-2 h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {credentials.map((credential) => (
                <DropdownMenuItem
                  key={credential.id}
                  onClick={() => handleCredentialChange(credential.id)}
                >
                  {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
                </DropdownMenuItem>
              ))}
              <div className="h-px bg-border my-1" />
              <DropdownMenuItem onClick={() => {
                if (currentWorkspace?.id) {
                  router.push(buildManagementPath(currentWorkspace.id, 'credentials'));
                } else {
                  router.push('/credentials');
                }
              }}>
                <Plus className="mr-2 h-4 w-4" />
                {t('credential.create')}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </CardContent>
      </Card>
    );
  }
  
  // 자격증명이 선택된 경우 - Card 형태로 표시
  return (
    <Card className={cn('border-border/50 bg-card shadow-sm', className)}>
      <CardContent className="flex items-center justify-between py-2.5 px-4">
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-shrink-0">
            <Key className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {t('components.credentialIndicator.using')}
            </span>
          </div>
          <div className="flex items-center gap-2 min-w-0">
            <span className="text-sm font-medium truncate">
              {selectedCredential?.name || 
               `${selectedCredential?.provider.toUpperCase()} (${selectedCredentialId.slice(0, 8)})`}
            </span>
            <Badge variant="outline" className="text-xs flex-shrink-0">
              {selectedCredential?.provider.toUpperCase()}
            </Badge>
            {selectedRegion && showRegionSelector && (
              <Badge variant="secondary" className="text-xs flex-shrink-0">
                {selectedRegion}
              </Badge>
            )}
          </div>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm">
              {t('common.change')}
              <ChevronDown className="ml-2 h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {credentials.map((credential) => (
              <DropdownMenuItem
                key={credential.id}
                onClick={() => handleCredentialChange(credential.id)}
              >
                <div className="flex items-center gap-2 w-full">
                  {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
                  {credential.id === selectedCredentialId && (
                    <Check className="h-3 w-3 ml-auto" />
                  )}
                </div>
              </DropdownMenuItem>
            ))}
            {showRegionSelector && (
              <>
                <div className="h-px bg-border my-1" />
                <div className="px-2 py-1.5">
                  <div className="text-xs font-medium mb-1">{t('region.select')}</div>
                  <div className="space-y-1">
                    <DropdownMenuItem
                      onClick={() => handleRegionChange('all')}
                      className={!selectedRegion ? 'bg-accent' : ''}
                    >
                      {t('region.all')}
                    </DropdownMenuItem>
                    {regions.map((region) => (
                      <DropdownMenuItem
                        key={region.value}
                        onClick={() => handleRegionChange(region.value)}
                        className={selectedRegion === region.value ? 'bg-accent' : ''}
                      >
                        {region.value} - {region.label}
                      </DropdownMenuItem>
                    ))}
                  </div>
                </div>
              </>
            )}
            <div className="h-px bg-border my-1" />
            <DropdownMenuItem onClick={() => {
              if (currentWorkspace?.id) {
                router.push(buildManagementPath(currentWorkspace.id, 'credentials'));
              } else {
                router.push('/credentials');
              }
            }}>
              <Plus className="mr-2 h-4 w-4" />
              {t('credential.create')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardContent>
    </Card>
  );
}


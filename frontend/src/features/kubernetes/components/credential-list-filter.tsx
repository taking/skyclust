/**
 * Credential List Filter Component
 * 직접 선택 리스트 방식의 Credential 필터 UI
 * 
 * 기능:
 * - Provider별 그룹화
 * - 체크박스 기반 직접 선택
 * - Collapsible로 공간 효율성
 * - Provider 단위 전체 선택/해제
 */

'use client';

import * as React from 'react';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { ChevronDown, X } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { Credential } from '@/lib/types';
import type { CloudProvider } from '@/lib/types/kubernetes';
import { useTranslation } from '@/hooks/use-translation';

interface CredentialListFilterProps {
  credentials: Credential[];
  selectedCredentialIds: string[];
  onSelectionChange: (credentialIds: string[]) => void;
  disabled?: boolean;
  useCollapsible?: boolean;
  maxHeight?: string;
}

const providerLabels: Record<CloudProvider, string> = {
  aws: 'AWS',
  gcp: 'GCP',
  azure: 'Azure',
};

const providerColors: Record<CloudProvider, string> = {
  aws: 'bg-orange-100 text-orange-800 border-orange-200',
  gcp: 'bg-blue-100 text-blue-800 border-blue-200',
  azure: 'bg-sky-100 text-sky-800 border-sky-200',
};

const COLLAPSIBLE_THRESHOLD = 6;

export function CredentialListFilter({
  credentials,
  selectedCredentialIds,
  onSelectionChange,
  disabled = false,
  useCollapsible,
  maxHeight = 'max-h-64',
}: CredentialListFilterProps) {
  const { t } = useTranslation();
  
  // Provider별로 그룹화
  const credentialsByProvider = React.useMemo(() => {
    const grouped: Record<CloudProvider, Credential[]> = {
      aws: [],
      gcp: [],
      azure: [],
    };
    
    credentials.forEach(cred => {
      if (cred.provider in grouped) {
        grouped[cred.provider as CloudProvider].push(cred);
      }
    });
    
    return grouped;
  }, [credentials]);
  
  // 총 credential 수 계산
  const totalCredentials = React.useMemo(() => 
    credentials.length,
    [credentials.length]
  );
  
  // Collapsible 사용 여부 결정 (credential 수가 많으면 사용)
  const shouldUseCollapsible = React.useMemo(() => 
    useCollapsible !== undefined 
      ? useCollapsible 
      : totalCredentials > COLLAPSIBLE_THRESHOLD,
    [useCollapsible, totalCredentials]
  );
  
  // Provider별 열림 상태 관리 (초기값: 모든 provider 열림)
  const [openProviders, setOpenProviders] = React.useState<Set<CloudProvider>>(() => {
    const providersWithCreds = (Object.keys(credentialsByProvider) as CloudProvider[]).filter(
      provider => credentialsByProvider[provider].length > 0
    );
    return new Set(providersWithCreds);
  });
  
  // Credentials 변경 시 열림 상태 업데이트
  React.useEffect(() => {
    const providersWithCreds = (Object.keys(credentialsByProvider) as CloudProvider[]).filter(
      provider => credentialsByProvider[provider].length > 0
    );
    
    // 새로운 provider가 추가되면 자동으로 열기
    setOpenProviders(prev => {
      const next = new Set(prev);
      providersWithCreds.forEach(provider => {
        if (!next.has(provider)) {
          next.add(provider);
        }
      });
      return next;
    });
  }, [credentialsByProvider]);
  
  // 초기화는 상위 컴포넌트(clusters/page.tsx)에서 처리하므로 여기서는 제거
  // CredentialListFilter는 순수하게 UI만 담당
  
  // Credential 토글 (완전히 controlled)
  const handleToggleCredential = React.useCallback((credentialId: string, checked: boolean) => {
    if (disabled) return;
    
    if (checked) {
      // 선택: credential 추가 (중복 방지)
      if (!selectedCredentialIds.includes(credentialId)) {
        onSelectionChange([...selectedCredentialIds, credentialId]);
      }
    } else {
      // 해제: credential 제거
      onSelectionChange(selectedCredentialIds.filter(id => id !== credentialId));
    }
  }, [disabled, selectedCredentialIds, onSelectionChange]);
  
  // Provider별 전체 선택/해제 (완전히 controlled)
  const handleToggleProvider = React.useCallback((provider: CloudProvider, checked: boolean) => {
    if (disabled) return;
    
    const providerCreds = credentialsByProvider[provider];
    if (!providerCreds || providerCreds.length === 0) return;
    
    const providerCredIds = providerCreds.map(c => c.id);
    
    if (checked) {
      // 전체 선택: 해당 provider의 credential 추가 (중복 방지)
      const newSelection = [...selectedCredentialIds];
      providerCredIds.forEach(id => {
        if (!newSelection.includes(id)) {
          newSelection.push(id);
        }
      });
      onSelectionChange(newSelection);
    } else {
      // 전체 해제: 해당 provider의 credential 제거
      const newSelection = selectedCredentialIds.filter(id => !providerCredIds.includes(id));
      onSelectionChange(newSelection);
    }
  }, [disabled, credentialsByProvider, selectedCredentialIds, onSelectionChange]);
  
  // 모든 선택 해제 (debounce 적용, 중복 클릭 방지)
  const clearAllTimeoutRef = React.useRef<NodeJS.Timeout | null>(null);
  const isClearingRef = React.useRef(false);
  
  const handleClearAll = React.useCallback(() => {
    if (disabled || isClearingRef.current) return;
    
    // 이미 초기화된 상태면 스킵
    if (selectedCredentialIds.length === 0) {
      return;
    }
    
    // Debounce 적용
    if (clearAllTimeoutRef.current) {
      clearTimeout(clearAllTimeoutRef.current);
    }
    
    isClearingRef.current = true;
    
    clearAllTimeoutRef.current = setTimeout(() => {
      onSelectionChange([]);
      isClearingRef.current = false;
    }, 150);
  }, [disabled, selectedCredentialIds.length, onSelectionChange]);
  
  // Cleanup
  React.useEffect(() => {
    return () => {
      if (clearAllTimeoutRef.current) {
        clearTimeout(clearAllTimeoutRef.current);
      }
    };
  }, []);
  
  // Provider 토글
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
  
  const hasSelection = selectedCredentialIds.length > 0;
  
  // Provider 목록 (credential이 있는 것만)
  const providersWithCredentials = React.useMemo(() => 
    (Object.keys(credentialsByProvider) as CloudProvider[]).filter(
      provider => credentialsByProvider[provider].length > 0
    ),
    [credentialsByProvider]
  );
  
  if (credentials.length === 0) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        <p>{t('credential.noCredentials') || 'No credentials available.'}</p>
        <p className="text-xs mt-1">{t('credential.addCredentialsFirst') || 'Please add credentials first.'}</p>
      </div>
    );
  }
  
  return (
    <div className={cn("space-y-2", shouldUseCollapsible && maxHeight && "overflow-y-auto", maxHeight)}>
      <Label className="text-sm font-medium" id="credential-filters-label">{t('credential.title') || 'Credentials'}</Label> 
      {providersWithCredentials.map(provider => {
        const providerCreds = credentialsByProvider[provider];
        const providerCredIds = providerCreds.map(c => c.id);
        const allSelected = providerCredIds.length > 0 && 
          providerCredIds.every(id => selectedCredentialIds.includes(id));
        const someSelected = providerCredIds.some(id => selectedCredentialIds.includes(id));
        const selectedCount = providerCredIds.filter(id => selectedCredentialIds.includes(id)).length;
        const isOpen = openProviders.has(provider);
        
        const content = (
          <div className="space-y-1.5 pl-6">
            {providerCreds.map(credential => {
              const isSelected = selectedCredentialIds.includes(credential.id);
              
              return (
                <div
                  key={credential.id}
                  className="flex items-center space-x-2 py-1 px-2 rounded hover:bg-accent"
                >
                  <Checkbox
                    id={`credential-${credential.id}`}
                    checked={isSelected}
                    onCheckedChange={(checked) => handleToggleCredential(credential.id, checked === true)}
                    disabled={disabled}
                  />
                  <Label
                    htmlFor={`credential-${credential.id}`}
                    className="flex-1 cursor-pointer text-sm"
                  >
                    {credential.name}
                  </Label>
                  {(credential as any).description && (
                    <span className="text-xs text-muted-foreground truncate max-w-[200px]">
                      {(credential as any).description}
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        );
        
        return (
          <div key={provider} className="border rounded-lg px-3 py-2">
            {shouldUseCollapsible ? (
              <Collapsible 
                open={isOpen}
                onOpenChange={() => toggleProvider(provider)}
              >
                <div className="flex items-center gap-2">
                  <Checkbox
                    id={`provider-${provider}`}
                    checked={allSelected}
                    ref={(el) => {
                      if (el && 'indeterminate' in el) {
                        (el as HTMLInputElement).indeterminate = someSelected && !allSelected;
                      }
                    }}
                    onCheckedChange={(checked) => {
                      if (!disabled) {
                        handleToggleProvider(provider, checked === true);
                      }
                    }}
                    disabled={disabled}
                    className="shrink-0"
                    onClick={(e) => {
                      e.stopPropagation();
                    }}
                  />
                  <CollapsibleTrigger asChild className="flex-1">
                    <button
                      type="button"
                      className="flex items-center justify-between w-full py-1 text-sm font-medium transition-all hover:underline text-left group"
                      onClick={(e) => {
                        e.stopPropagation();
                      }}
                    >
                      <div className="flex items-center gap-2">
                        <span>{providerLabels[provider]}</span>
                        <Badge 
                          variant="outline" 
                          className={cn(providerColors[provider], "text-xs")}
                        >
                          {selectedCount > 0 ? `${selectedCount}/${providerCreds.length}` : providerCreds.length}
                        </Badge>
                      </div>
                      <ChevronDown className={cn(
                        "h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200",
                        isOpen && "rotate-180"
                      )} />
                    </button>
                  </CollapsibleTrigger>
                </div>
                <CollapsibleContent className="overflow-hidden data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down">
                  {content}
                </CollapsibleContent>
              </Collapsible>
            ) : (
              <>
                <div className="flex items-center gap-2 pb-2">
                  <Checkbox
                    id={`provider-${provider}`}
                    checked={allSelected}
                    ref={(el) => {
                      if (el && 'indeterminate' in el) {
                        (el as HTMLInputElement).indeterminate = someSelected && !allSelected;
                      }
                    }}
                    onCheckedChange={(checked) => {
                      if (!disabled) {
                        handleToggleProvider(provider, checked === true);
                      }
                    }}
                    disabled={disabled}
                    className="shrink-0"
                  />
                  <Label
                    htmlFor={`provider-${provider}`}
                    className="flex-1 cursor-pointer text-sm font-medium"
                    onClick={(e) => {
                      e.stopPropagation();
                      if (!disabled) {
                        // 현재 상태를 기반으로 토글 (전체 선택이면 해제, 아니면 선택)
                        handleToggleProvider(provider, !allSelected);
                      }
                    }}
                  >
                    {providerLabels[provider]}
                  </Label>
                  <Badge 
                    variant="outline" 
                    className={cn(providerColors[provider], "text-xs")}
                  >
                    {selectedCount > 0 ? `${selectedCount}/${providerCreds.length}` : providerCreds.length}
                  </Badge>
                </div>
                {content}
              </>
            )}
          </div>
        );
      })}
      
      {/* Clear Credentials 버튼 */}
      {hasSelection && !disabled && (
        // <div className="pt-2 border-t">
        //   <Button
        //     variant="ghost"
        //     size="sm"
        //     onClick={handleClearAll}
        //     className="h-8 text-xs w-full"
        //     disabled={isClearingRef.current}
        //     aria-label={t('credential.clearCredentials') || 'Clear credentials'}
        //   >
        //     <X className="h-3 w-3 mr-1" />
        //     {t('credential.clearCredentials') || 'Clear Credentials'}
        //   </Button>
        // </div>
        <></>
      )}
    </div>
  );
}


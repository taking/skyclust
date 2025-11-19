/**
 * Credential Multi-Select Compact Component
 * 컴팩트한 Multi-provider credential 선택 UI
 * 
 * 개선 사항:
 * - Chip 기반 컴팩트 UI
 * - Popover 내부에 상세 목록
 * - 선택된 credential이 항상 보임
 * - 빠른 제거 가능
 */

'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Key, X, CheckCircle2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { Credential } from '@/lib/types';
import type { CloudProvider } from '@/lib/types/kubernetes';
import { useTranslation } from '@/hooks/use-translation';

interface CredentialMultiSelectCompactProps {
  credentials: Credential[];
  selectedCredentialIds: string[];
  onSelectionChange: (credentialIds: string[]) => void;
  disabled?: boolean;
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

export function CredentialMultiSelectCompact({
  credentials,
  selectedCredentialIds,
  onSelectionChange,
  disabled = false,
}: CredentialMultiSelectCompactProps) {
  const { t } = useTranslation();
  const [open, setOpen] = React.useState(false);
  
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
  
  // Credential 토글
  const handleToggleCredential = React.useCallback((credentialId: string) => {
    if (disabled) return;
    
    const isSelected = selectedCredentialIds.includes(credentialId);
    if (isSelected) {
      onSelectionChange(selectedCredentialIds.filter(id => id !== credentialId));
    } else {
      onSelectionChange([...selectedCredentialIds, credentialId]);
    }
  }, [disabled, selectedCredentialIds, onSelectionChange]);
  
  // Provider별 전체 선택/해제
  const handleToggleProvider = React.useCallback((provider: CloudProvider) => {
    if (disabled) return;
    
    const providerCreds = credentialsByProvider[provider];
    const providerCredIds = providerCreds.map(c => c.id);
    const allSelected = providerCredIds.every(id => selectedCredentialIds.includes(id));
    
    if (allSelected) {
      onSelectionChange(selectedCredentialIds.filter(id => !providerCredIds.includes(id)));
    } else {
      const newSelection = [...selectedCredentialIds];
      providerCredIds.forEach(id => {
        if (!newSelection.includes(id)) {
          newSelection.push(id);
        }
      });
      onSelectionChange(newSelection);
    }
  }, [disabled, credentialsByProvider, selectedCredentialIds, onSelectionChange]);
  
  // 모든 선택 해제
  const handleClearAll = React.useCallback(() => {
    if (disabled) return;
    onSelectionChange([]);
  }, [disabled, onSelectionChange]);
  
  // Credential 제거
  const handleRemove = React.useCallback((credentialId: string) => {
    if (disabled) return;
    onSelectionChange(selectedCredentialIds.filter(id => id !== credentialId));
  }, [disabled, selectedCredentialIds, onSelectionChange]);
  
  const hasSelection = selectedCredentialIds.length > 0;
  const selectedCredentials = credentials.filter(c => selectedCredentialIds.includes(c.id));
  
  return (
    <div className="flex items-center gap-2 flex-wrap">
      {/* 선택 버튼 */}
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            disabled={disabled}
            className="h-8"
          >
            <Key className="mr-2 h-4 w-4" />
            <span>Credentials</span>
            {hasSelection && (
              <Badge variant="secondary" className="ml-2 h-5 px-1.5">
                {selectedCredentialIds.length}
              </Badge>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80 p-0" align="start">
          <Command>
            <CommandInput placeholder="Search credentials..." />
            <CommandList>
              <CommandEmpty>No credentials found.</CommandEmpty>
              
              {/* Provider별 그룹 */}
              {(Object.keys(credentialsByProvider) as CloudProvider[]).map(provider => {
                const providerCreds = credentialsByProvider[provider];
                if (providerCreds.length === 0) return null;
                
                const providerCredIds = providerCreds.map(c => c.id);
                const allSelected = providerCredIds.every(id => selectedCredentialIds.includes(id));
                const someSelected = providerCredIds.some(id => selectedCredentialIds.includes(id));
                
                return (
                  <CommandGroup key={provider} heading={providerLabels[provider]}>
                    {/* Provider 전체 선택 */}
                    <div className="flex items-center space-x-2 px-2 py-1.5 border-b">
                      <Checkbox
                        id={`provider-${provider}`}
                        checked={allSelected}
                        onCheckedChange={() => handleToggleProvider(provider)}
                        disabled={disabled}
                      />
                      <Label
                        htmlFor={`provider-${provider}`}
                        className="flex-1 cursor-pointer text-sm font-medium"
                      >
                        Select All {providerLabels[provider]}
                      </Label>
                      <Badge variant="outline" className={cn(providerColors[provider])}>
                        {providerCreds.length}
                      </Badge>
                    </div>
                    
                    {/* Credential 목록 */}
                    {providerCreds.map(credential => {
                      const isSelected = selectedCredentialIds.includes(credential.id);
                      
                      return (
                        <CommandItem
                          key={credential.id}
                          value={credential.id}
                          onSelect={() => handleToggleCredential(credential.id)}
                          className="flex items-center space-x-2"
                        >
                          <Checkbox
                            checked={isSelected}
                            onCheckedChange={() => handleToggleCredential(credential.id)}
                            disabled={disabled}
                          />
                          <div className="flex-1">
                            <div className="flex items-center justify-between">
                              <span className="font-medium">{credential.name}</span>
                              {isSelected && (
                                <CheckCircle2 className="h-4 w-4 text-primary" />
                              )}
                            </div>
                            {credential.description && (
                              <p className="text-xs text-muted-foreground truncate">
                                {credential.description}
                              </p>
                            )}
                          </div>
                        </CommandItem>
                      );
                    })}
                  </CommandGroup>
                );
              })}
              
              {/* 빈 상태 */}
              {credentials.length === 0 && (
                <div className="p-4 text-center text-sm text-muted-foreground">
                  <Key className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                  <p>No credentials available.</p>
                  <p className="text-xs mt-1">Please add credentials first.</p>
                </div>
              )}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
      
      {/* 선택된 credential Chip 표시 */}
      {hasSelection && (
        <>
          <div className="flex items-center gap-1.5 flex-wrap">
            {selectedCredentials.map(credential => (
              <Badge
                key={credential.id}
                variant="secondary"
                className={cn(
                  "gap-1 pr-1",
                  providerColors[credential.provider as CloudProvider]
                )}
              >
                <Key className="h-3 w-3" />
                <span className="text-xs">{credential.name}</span>
                <span className="text-xs opacity-70">
                  ({providerLabels[credential.provider as CloudProvider]})
                </span>
                {!disabled && (
                  <button
                    type="button"
                    onClick={() => handleRemove(credential.id)}
                    className="ml-1 rounded-full hover:bg-black/10 p-0.5"
                    aria-label={t('common.removeCredential', { name: credential.name })}
                  >
                    <X className="h-3 w-3" />
                  </button>
                )}
              </Badge>
            ))}
          </div>
          
          {/* 전체 해제 버튼 */}
          {!disabled && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleClearAll}
              className="h-8 text-xs"
            >
              Clear
            </Button>
          )}
        </>
      )}
    </div>
  );
}


/**
 * Credential Multi-Select Component
 * Multi-provider credential 선택 UI
 * 
 * 기능:
 * - 여러 credential 동시 선택
 * - Provider별 그룹화 표시
 * - 체크박스 UI
 */

'use client';

import * as React from 'react';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Key, CheckCircle2 } from 'lucide-react';
import type { Credential } from '@/lib/types';
import type { CloudProvider } from '@/lib/types/kubernetes';
import { cn } from '@/lib/utils';

interface CredentialMultiSelectProps {
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
  aws: 'bg-orange-100 text-orange-800',
  gcp: 'bg-blue-100 text-blue-800',
  azure: 'bg-sky-100 text-sky-800',
};

export function CredentialMultiSelect({
  credentials,
  selectedCredentialIds,
  onSelectionChange,
  disabled = false,
}: CredentialMultiSelectProps) {
  
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
  const handleToggleCredential = (credentialId: string) => {
    if (disabled) return;
    
    const isSelected = selectedCredentialIds.includes(credentialId);
    if (isSelected) {
      onSelectionChange(selectedCredentialIds.filter(id => id !== credentialId));
    } else {
      onSelectionChange([...selectedCredentialIds, credentialId]);
    }
  };
  
  // Provider별 전체 선택/해제
  const handleToggleProvider = (provider: CloudProvider) => {
    if (disabled) return;
    
    const providerCreds = credentialsByProvider[provider];
    const providerCredIds = providerCreds.map(c => c.id);
    const allSelected = providerCredIds.every(id => selectedCredentialIds.includes(id));
    
    if (allSelected) {
      // 전체 해제
      onSelectionChange(selectedCredentialIds.filter(id => !providerCredIds.includes(id)));
    } else {
      // 전체 선택
      const newSelection = [...selectedCredentialIds];
      providerCredIds.forEach(id => {
        if (!newSelection.includes(id)) {
          newSelection.push(id);
        }
      });
      onSelectionChange(newSelection);
    }
  };
  
  const hasSelection = selectedCredentialIds.length > 0;
  
  return (
    <div className="space-y-4">
      {/* 선택된 credential 요약 */}
      {hasSelection && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Selected Credentials</CardTitle>
            <CardDescription>
              {selectedCredentialIds.length} credential{selectedCredentialIds.length !== 1 ? 's' : ''} selected
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {selectedCredentialIds.map(credentialId => {
                const credential = credentials.find(c => c.id === credentialId);
                if (!credential) return null;
                
                return (
                  <Badge
                    key={credentialId}
                    variant="secondary"
                    className={cn(providerColors[credential.provider as CloudProvider])}
                  >
                    <Key className="mr-1 h-3 w-3" />
                    {credential.name} ({providerLabels[credential.provider as CloudProvider]})
                  </Badge>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}
      
      {/* Provider별 credential 목록 */}
      <div className="space-y-4">
        {(Object.keys(credentialsByProvider) as CloudProvider[]).map(provider => {
          const providerCreds = credentialsByProvider[provider];
          if (providerCreds.length === 0) return null;
          
          const providerCredIds = providerCreds.map(c => c.id);
          const allSelected = providerCredIds.every(id => selectedCredentialIds.includes(id));
          const someSelected = providerCredIds.some(id => selectedCredentialIds.includes(id));
          
          return (
            <Card key={provider}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <CardTitle className="text-base">
                      {providerLabels[provider]}
                    </CardTitle>
                    <Badge variant="outline" className={cn(providerColors[provider])}>
                      {providerCreds.length}
                    </Badge>
                  </div>
                  <Checkbox
                    checked={allSelected}
                    onCheckedChange={() => handleToggleProvider(provider)}
                    disabled={disabled}
                  />
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {providerCreds.map(credential => {
                    const isSelected = selectedCredentialIds.includes(credential.id);
                    
                    return (
                      <div
                        key={credential.id}
                        className={cn(
                          "flex items-center space-x-3 p-2 rounded-md border transition-colors",
                          isSelected && "bg-muted",
                          !disabled && "cursor-pointer hover:bg-muted/50"
                        )}
                        onClick={() => handleToggleCredential(credential.id)}
                      >
                        <Checkbox
                          checked={isSelected}
                          onCheckedChange={() => handleToggleCredential(credential.id)}
                          disabled={disabled}
                        />
                        <Label
                          htmlFor={`credential-${credential.id}`}
                          className="flex-1 cursor-pointer"
                        >
                          <div className="flex items-center justify-between">
                            <span className="font-medium">{credential.name}</span>
                            {isSelected && (
                              <CheckCircle2 className="h-4 w-4 text-primary" />
                            )}
                          </div>
                          {credential.description && (
                            <p className="text-sm text-muted-foreground">
                              {credential.description}
                            </p>
                          )}
                        </Label>
                      </div>
                    );
                  })}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
      
      {/* 빈 상태 */}
      {credentials.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Key className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-sm text-muted-foreground">
              No credentials available. Please add credentials first.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}


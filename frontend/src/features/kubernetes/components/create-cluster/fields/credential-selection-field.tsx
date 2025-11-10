/**
 * Credential Selection Field Component
 * Credential 선택 필드 컴포넌트 (Dashboard 값 처리 포함)
 */

'use client';

import { useEffect } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateClusterForm, Credential } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { useCredentialContext } from '@/hooks/use-credential-context';

export interface CredentialSelectionFieldProps {
  /** React Hook Form 인스턴스 */
  form: UseFormReturn<CreateClusterForm>;
  /** Credential 목록 */
  credentials: Credential[];
  /** 선택된 Credential ID */
  selectedCredentialId: string;
  /** Credential 변경 핸들러 */
  onCredentialChange: (credentialId: string) => void;
}

/**
 * Credential 선택 필드 컴포넌트
 * Dashboard에서 선택된 값이 있으면 자동 적용 (비활성화), 없으면 선택 가능
 */
export function CredentialSelectionField({
  form,
  credentials,
  selectedCredentialId,
  onCredentialChange,
}: CredentialSelectionFieldProps) {
  const { t } = useTranslation();
  const { selectedCredentialId: dashboardCredentialId } = useCredentialContext();
  const formCredentialId = form.watch('credential_id');
  
  // Dashboard에서 Credential이 선택되어 있는지 확인
  const hasDashboardCredential = !!dashboardCredentialId;
  
  // Dashboard에서 선택된 Credential이 있으면 form에 자동 설정
  useEffect(() => {
    if (dashboardCredentialId && !formCredentialId) {
      form.setValue('credential_id', dashboardCredentialId);
      onCredentialChange(dashboardCredentialId);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dashboardCredentialId, formCredentialId]);
  
  // Dashboard에서 선택된 Credential 또는 Form의 Credential 사용
  const currentCredentialId = formCredentialId || dashboardCredentialId || selectedCredentialId || '';
  
  // 선택된 Credential 정보 찾기
  const selectedCredential = credentials.find(c => c.id === currentCredentialId);

  return (
    <FormField
      control={form.control}
      name="credential_id"
      render={({ field }) => (
        <FormItem>
          <FormLabel>{t('kubernetes.credential')} *</FormLabel>
          <Select
            value={currentCredentialId}
            onValueChange={(value) => {
              field.onChange(value);
              onCredentialChange(value);
            }}
            disabled={hasDashboardCredential}
          >
            <FormControl>
              <SelectTrigger className="w-full">
                <SelectValue 
                  placeholder={
                    hasDashboardCredential && selectedCredential
                      ? `Selected: ${selectedCredential.provider} - ${selectedCredential.id.substring(0, 8)}...`
                      : t('kubernetes.selectCredential')
                  } 
                />
              </SelectTrigger>
            </FormControl>
            <SelectContent>
              {credentials.map((cred) => (
                <SelectItem key={cred.id} value={cred.id}>
                  <div className="flex items-center gap-2 min-w-0">
                    <span className="truncate">{cred.provider} - {cred.id.substring(0, 8)}...</span>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FormDescription>
            {hasDashboardCredential
              ? `Credential selected from Dashboard. Change in Sidebar if needed.`
              : t('kubernetes.credentialDescription')}
          </FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}


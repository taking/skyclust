/**
 * Credential Selection Field Component
 * Credential 선택 필드 컴포넌트
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateClusterForm, Credential } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

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
 * 항상 선택 가능한 필드로 표시
 */
export function CredentialSelectionField({
  form,
  credentials,
  selectedCredentialId,
  onCredentialChange,
}: CredentialSelectionFieldProps) {
  const { t } = useTranslation();
  const formCredentialId = form.watch('credential_id');
  
  // Form의 Credential 또는 prop으로 전달된 Credential 사용
  const currentCredentialId = formCredentialId || selectedCredentialId || '';
  
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
          >
            <FormControl>
              <SelectTrigger className="w-full">
                <SelectValue 
                  placeholder={t('kubernetes.selectCredential')}
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
            {t('kubernetes.credentialDescription')}
          </FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}


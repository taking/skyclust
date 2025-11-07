/**
 * Credential Selection Field Component
 * Credential 선택 필드 컴포넌트
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
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
 */
export function CredentialSelectionField({
  form,
  credentials,
  selectedCredentialId,
  onCredentialChange,
}: CredentialSelectionFieldProps) {
  const { t } = useTranslation();

  return (
    <FormField
      control={form.control}
      name="credential_id"
      render={({ field }) => (
        <FormItem>
          <FormLabel>{t('kubernetes.credential')} *</FormLabel>
          <Select
            value={field.value}
            onValueChange={(value) => {
              field.onChange(value);
              onCredentialChange(value);
            }}
          >
            <FormControl>
              <SelectTrigger className="w-full">
                <SelectValue placeholder={t('kubernetes.selectCredential')} />
              </SelectTrigger>
            </FormControl>
            <SelectContent>
              {credentials.map((cred) => (
                <SelectItem key={cred.id} value={cred.id}>
                  {cred.provider} - {cred.id.substring(0, 8)}...
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}


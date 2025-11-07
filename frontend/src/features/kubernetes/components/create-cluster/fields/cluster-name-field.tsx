/**
 * Cluster Name Field Component
 * 클러스터 이름 입력 필드 컴포넌트
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

export interface ClusterNameFieldProps {
  /** React Hook Form 인스턴스 */
  form: UseFormReturn<CreateClusterForm>;
  /** 필드 변경 핸들러 */
  onFieldChange: (field: keyof CreateClusterForm, value: unknown) => void;
}

/**
 * 클러스터 이름 입력 필드 컴포넌트
 */
export function ClusterNameField({
  form,
  onFieldChange,
}: ClusterNameFieldProps) {
  const { t } = useTranslation();

  return (
    <FormField
      control={form.control}
      name="name"
      render={({ field }) => (
        <FormItem>
          <FormLabel>{t('kubernetes.clusterName')} *</FormLabel>
          <FormControl>
            <Input
              placeholder={t('kubernetes.enterClusterName')}
              {...field}
              onChange={(e) => {
                field.onChange(e);
                onFieldChange('name', e.target.value);
              }}
            />
          </FormControl>
          <FormDescription>
            {t('kubernetes.clusterNameDescription')}
          </FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}


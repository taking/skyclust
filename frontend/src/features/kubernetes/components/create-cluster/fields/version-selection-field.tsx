/**
 * Version Selection Field Component
 * Kubernetes 버전 선택 필드 컴포넌트
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

export interface VersionSelectionFieldProps {
  /** React Hook Form 인스턴스 */
  form: UseFormReturn<CreateClusterForm>;
  /** 필드 변경 핸들러 */
  onFieldChange: (field: keyof CreateClusterForm, value: unknown) => void;
  /** 버전 목록 (AWS metadata에서 로드된 경우) */
  versions?: string[];
  /** 버전 로딩 중 여부 */
  isLoadingVersions?: boolean;
  /** 버전 로딩 에러 */
  versionsError?: Error | null;
  /** 메타데이터 로딩 가능 여부 */
  canLoadMetadata?: boolean;
  /** 리전이 선택되었는지 여부 */
  hasSelectedRegion?: boolean;
}

/**
 * Kubernetes 버전 선택 필드 컴포넌트
 */
export function VersionSelectionField({
  form,
  onFieldChange,
  versions = [],
  isLoadingVersions = false,
  versionsError = null,
  canLoadMetadata = false,
  hasSelectedRegion = false,
}: VersionSelectionFieldProps) {
  const { t } = useTranslation();

  return (
    <FormField
      control={form.control}
      name="version"
      render={({ field }) => (
        <FormItem className="flex flex-col h-full min-h-[100px]">
          <FormLabel className="mb-2">Kubernetes Version *</FormLabel>
          {canLoadMetadata && hasSelectedRegion ? (
            <Select
              value={field.value}
              onValueChange={(value) => {
                field.onChange(value);
                onFieldChange('version', value);
              }}
              disabled={isLoadingVersions}
            >
              <FormControl>
                <SelectTrigger>
                  <SelectValue placeholder={isLoadingVersions ? 'Loading versions...' : 'Select Kubernetes version'} />
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                {versions.length === 0 && !isLoadingVersions ? (
                  <SelectItem value="no-versions" disabled>
                    No versions available
                  </SelectItem>
                ) : (
                  versions.map((version) => (
                    <SelectItem key={version} value={version}>
                      {version}
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          ) : (
            <FormControl>
              <Input
                placeholder="e.g., 1.34"
                {...field}
                onChange={(e) => {
                  field.onChange(e);
                  onFieldChange('version', e.target.value);
                }}
              />
            </FormControl>
          )}
          <FormDescription className="mt-1">
            {t('kubernetes.versionDescription')}
          </FormDescription>
          {canLoadMetadata && hasSelectedRegion && versionsError && (
            <FormDescription className="text-destructive mt-1">
              Failed to load Kubernetes versions: {versionsError.message}
              {(versionsError.message.includes('IAM permission') || 
                versionsError.message.includes('not authorized') ||
                versionsError.message.includes('UnauthorizedOperation')) && (
                <span className="block mt-1 text-muted-foreground">
                  <strong>Solution:</strong> Add the required EKS permissions to your AWS IAM user or role.
                </span>
              )}
            </FormDescription>
          )}
          <FormMessage className="mt-1" />
        </FormItem>
      )}
    />
  );
}


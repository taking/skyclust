/**
 * Create VPC Dialog Component
 * VPC 생성 다이얼로그 컴포넌트
 */

'use client';

import { useEffect } from 'react';
import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { useToast } from '@/hooks/use-toast';
import { createValidationSchemas } from '@/lib/validations';
import type { CreateVPCForm, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { useVPCActions } from '@/features/networks/hooks/use-vpc-actions';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query-keys';

interface CreateVPCDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedProvider?: string;
  selectedCredentialId: string;
  selectedRegion?: string;
  onSuccess?: (vpcId: string) => void;
  disabled?: boolean;
}

export function CreateVPCDialog({
  open,
  onOpenChange,
  selectedProvider,
  selectedCredentialId,
  selectedRegion,
  onSuccess,
  disabled = false,
}: CreateVPCDialogProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const schemas = createValidationSchemas(t);

  // VPC 생성 mutation
  const { createVPCMutation } = useVPCActions({
    selectedProvider,
    selectedCredentialId,
    selectedRegion: selectedRegion || '',
    onSuccess: () => {
      // VPC 목록 갱신
      if (selectedProvider && selectedCredentialId) {
        queryClient.invalidateQueries({
          queryKey: queryKeys.vpcs.list(selectedProvider as CloudProvider, selectedCredentialId, selectedRegion),
        });
      }
    },
  });

  const {
    form,
    handleSubmit,
    isLoading,
    reset,
    setValue,
  } = useFormWithValidation<CreateVPCForm>({
    schema: schemas.createVPCSchema,
    defaultValues: {
      name: '',
      description: '',
      cidr_block: '',
      region: selectedRegion || '',
      credential_id: selectedCredentialId, // credential_id를 defaultValues에 포함
    },
    onSubmit: async (data) => {
      if (!selectedProvider) {
        throw new Error('Provider not selected');
      }
      
      // selectedRegion이 있으면 항상 그 값을 사용
      const finalRegion = selectedRegion || data.region;
      
      // credential_id는 이미 form에 포함되어 있지만, 확실하게 설정
      const vpcData: CreateVPCForm = {
        ...data,
        credential_id: data.credential_id || selectedCredentialId,
        region: finalRegion,
      };
      
      // Provider별 필드 정리
      if (selectedProvider === 'azure') {
        // Azure는 location과 resource_group, address_space 사용
        vpcData.location = finalRegion;
        // cidr_block은 사용하지 않음 (address_space 사용)
        delete vpcData.cidr_block;
      } else if (selectedProvider === 'gcp') {
        // GCP는 cidr_block이 optional (auto-mode VPC인 경우)
        // project_id, auto_create_subnets, routing_mode, mtu는 form에서 입력받음
      } else if (selectedProvider === 'aws') {
        // AWS는 cidr_block과 region 필수
        if (!vpcData.cidr_block) {
          throw new Error('CIDR block is required for AWS VPC');
        }
      }
      
      // Validation: Provider별 필수 필드 확인
      if (!vpcData.name) {
        throw new Error('Name is required');
      }
      
      if (selectedProvider === 'aws' && !vpcData.cidr_block) {
        throw new Error('CIDR block is required for AWS VPC');
      }
      
      if (selectedProvider === 'azure') {
        if (!vpcData.location || !vpcData.resource_group || !vpcData.address_space || vpcData.address_space.length === 0) {
          throw new Error('Location, resource group, and address space are required for Azure VPC');
        }
      }
      
      const result = await createVPCMutation.mutateAsync(vpcData);
      
      // 성공 시 콜백 호출
      if (result?.id && onSuccess) {
        onSuccess(result.id);
      }
      
      // 다이얼로그 닫기
      onOpenChange(false);
    },
    onSuccess: () => {
      reset();
    },
    onError: (error) => {
      // 에러는 mutation에서 자동 처리됨
    },
    resetOnSuccess: true,
  });

  // Reset form when dialog opens/closes or region changes
  useEffect(() => {
    if (open) {
      const resetValues = {
        name: '',
        description: '',
        cidr_block: '',
        region: selectedRegion || '',
        credential_id: selectedCredentialId, // credential_id를 reset values에 포함
        location: selectedProvider === 'azure' ? selectedRegion || '' : undefined,
        resource_group: undefined,
        address_space: undefined,
      };
      reset(resetValues);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, selectedRegion, selectedProvider, selectedCredentialId]);

  // Update region/location when selectedRegion changes (only when dialog is open)
  useEffect(() => {
    if (open && selectedRegion) {
      form.setValue('region', selectedRegion, { shouldValidate: true });
      // credential_id도 함께 설정 (validation을 위해)
      if (selectedCredentialId) {
        form.setValue('credential_id', selectedCredentialId, { shouldValidate: true });
      }
      if (selectedProvider === 'azure') {
        form.setValue('location', selectedRegion, { shouldValidate: true });
      }
      // Validation 상태 확인
      form.trigger(['region', 'credential_id']);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, selectedRegion, selectedProvider, selectedCredentialId]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('network.createVPCTitle')}</DialogTitle>
          <DialogDescription>
            {selectedRegion 
              ? t('network.createVPCDescription', { 
                  region: selectedRegion, 
                  provider: selectedProvider?.toUpperCase() || '' 
                })
              : t('network.createVPCDescriptionNoRegion', { 
                  provider: selectedProvider?.toUpperCase() || '' 
                })
            }
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form 
            onSubmit={handleSubmit} 
            className="space-y-4"
          >
            <EnhancedField
              name="name"
              label={t('network.vpcName') + ' *'}
              placeholder="my-vpc"
            />
            <EnhancedField
              name="description"
              label={t('network.vpcDescription')}
              placeholder={t('network.vpcDescriptionPlaceholder')}
              type="textarea"
            />
            <EnhancedField
              name="cidr_block"
              label={t('network.cidrBlock') + ' *'}
              placeholder="10.0.0.0/16"
            />
            <FormField
              control={form.control}
              name="region"
              render={({ field }) => {
                // selectedRegion이 있으면 항상 그 값을 사용하고 form에 등록
                const regionValue = selectedRegion || field.value || '';
                
                return (
                  <FormItem>
                    <FormLabel>{t('region.select')} *</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder="ap-northeast-3"
                        value={regionValue}
                        disabled={true}
                        readOnly
                        className="bg-muted cursor-not-allowed"
                      />
                    </FormControl>
                    {selectedRegion && (
                      <p className="text-sm text-muted-foreground">
                        Region is fixed from Basic Configuration step
                      </p>
                    )}
                    <FormMessage />
                  </FormItem>
                );
              }}
            />

            {/* Azure specific fields */}
            {selectedProvider === 'azure' && (
              <div className="space-y-4 pt-4 border-t">
                <div className="grid grid-cols-2 gap-4">
                  <EnhancedField
                    name="location"
                    label="Location *"
                    placeholder="eastus"
                    required
                  />
                  <EnhancedField
                    name="resource_group"
                    label="Resource Group *"
                    placeholder="my-resource-group"
                    required
                  />
                </div>
                <EnhancedField
                  name="address_space"
                  label="Address Space (CIDR blocks, comma-separated)"
                  placeholder="10.0.0.0/16"
                  description="Azure Virtual Network address spaces (comma-separated CIDR blocks)"
                />
              </div>
            )}

            <div className="flex justify-end space-x-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading || createVPCMutation.isPending}>
                {t('common.cancel')}
              </Button>
              <Button 
                type="submit" 
                disabled={isLoading || createVPCMutation.isPending || disabled || !selectedProvider || !selectedCredentialId}
              >
                {isLoading || createVPCMutation.isPending ? t('actions.creating') : t('network.createVPC')}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}


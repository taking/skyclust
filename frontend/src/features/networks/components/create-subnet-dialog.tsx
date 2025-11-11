/**
 * Create Subnet Dialog Component
 * Subnet 생성 다이얼로그 컴포넌트
 */

'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useValidation } from '@/lib/validation';
import type { CreateSubnetForm, VPC, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { useSubnetActions } from '@/features/networks/hooks/use-subnet-actions';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import { useAvailabilityZones } from '@/features/kubernetes/hooks/use-kubernetes-metadata';

interface CreateSubnetDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedProvider?: string;
  selectedCredentialId: string;
  selectedRegion?: string;
  selectedZone?: string;
  selectedVPCId: string;
  vpcs: VPC[];
  onVPCChange: (vpcId: string) => void;
  onSuccess?: (subnetId: string) => void;
  disabled?: boolean;
}

export function CreateSubnetDialog({
  open,
  onOpenChange,
  selectedProvider,
  selectedCredentialId,
  selectedRegion,
  selectedZone,
  selectedVPCId,
  vpcs,
  onVPCChange,
  onSuccess,
  disabled: _disabled = false,
}: CreateSubnetDialogProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { schemas } = useValidation();

  // Availability Zones 조회 (AWS의 경우)
  const {
    data: availabilityZones = [],
    isLoading: isLoadingZones,
  } = useAvailabilityZones({
    provider: selectedProvider as CloudProvider,
    credentialId: selectedCredentialId,
    region: selectedRegion || '',
  });

  const canSelectZone = selectedProvider === 'aws' && !!selectedCredentialId && !!selectedRegion;

  // Subnet 생성 mutation
  const { createSubnetMutation } = useSubnetActions({
    selectedProvider,
    selectedCredentialId,
    onSuccess: () => {
      // Subnet 목록 갱신
      if (selectedProvider && selectedCredentialId && selectedVPCId) {
        queryClient.invalidateQueries({
          queryKey: queryKeys.subnets.list(
            selectedProvider as CloudProvider,
            selectedCredentialId,
            selectedVPCId,
            selectedRegion || ''
          ),
        });
      }
    },
  });

  const form = useForm<CreateSubnetForm>({
    resolver: zodResolver(schemas.createSubnetSchema),
    defaultValues: {
      name: '',
      cidr_block: '',
      region: selectedRegion || '',
      availability_zone: selectedZone || '',
      vpc_id: selectedVPCId,
      credential_id: selectedCredentialId,
    },
  });

  // Reset form when dialog opens/closes
  React.useEffect(() => {
    if (open) {
      form.reset({
        name: '',
        cidr_block: '',
        region: selectedRegion || '',
        availability_zone: selectedZone || '',
        vpc_id: selectedVPCId,
        credential_id: selectedCredentialId,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, selectedRegion, selectedZone, selectedVPCId, selectedCredentialId]);

  // Update form when VPC changes (only when dialog is open)
  React.useEffect(() => {
    if (open && selectedVPCId) {
      form.setValue('vpc_id', selectedVPCId);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, selectedVPCId]);

  // Update form when region changes (only when dialog is open)
  React.useEffect(() => {
    if (open && selectedRegion) {
      form.setValue('region', selectedRegion, { shouldValidate: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, selectedRegion]);

  // Update form when zone changes (only when dialog opens, not when selectedZone changes while dialog is open)
  // 사용자가 dialog에서 직접 선택한 값을 보존하기 위해 dialog가 열릴 때만 초기값 설정
  React.useEffect(() => {
    if (open && selectedZone && !form.getValues('availability_zone')) {
      // form에 값이 없을 때만 selectedZone을 초기값으로 설정
      // 사용자가 이미 선택한 값이 있으면 덮어쓰지 않음
      form.setValue('availability_zone', selectedZone, { shouldValidate: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]); // selectedZone을 dependency에서 제거하여 사용자 선택을 보존

  // Update form when credential_id changes (only when dialog is open)
  React.useEffect(() => {
    if (open && selectedCredentialId) {
      form.setValue('credential_id', selectedCredentialId, { shouldValidate: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, selectedCredentialId]);

  const handleSubmit = form.handleSubmit(async (data) => {
    if (!selectedProvider) {
      throw new Error('Provider not selected');
    }
    
    // Region: selectedRegion이 있으면 우선 사용, 없으면 form의 값 사용
    const finalRegion = selectedRegion || data.region;
    
    // Zone: form에서 사용자가 선택한 값을 우선 사용, 없으면 selectedZone 사용
    // 사용자가 dialog에서 직접 선택한 값이 있으면 그것을 사용
    const finalZone = data.availability_zone || data.zone || selectedZone;
    
    // credential_id 자동 추가
    const subnetData: CreateSubnetForm = {
      ...data,
      credential_id: selectedCredentialId,
      vpc_id: selectedVPCId,
      region: finalRegion,
    };
    
    // Provider별 필드 설정
    if (selectedProvider === 'gcp') {
      // GCP는 zone 사용 (availability_zone 대신)
      subnetData.zone = finalZone;
      // availability_zone은 사용하지 않음
      delete subnetData.availability_zone;
    } else {
      // AWS/Azure는 availability_zone 사용
      subnetData.availability_zone = finalZone;
      // zone은 사용하지 않음
      delete subnetData.zone;
    }
    
    // Validation: Provider별 필수 필드 확인
    if (!subnetData.name || !subnetData.cidr_block || !subnetData.region) {
      throw new Error('Name, CIDR block, and region are required');
    }
    
    if (selectedProvider === 'gcp') {
      if (!subnetData.zone) {
        throw new Error('Zone is required for GCP subnet');
      }
    } else {
      if (!subnetData.availability_zone) {
        throw new Error('Availability zone is required');
      }
    }
    
    const result = await createSubnetMutation.mutateAsync(subnetData);
    
    // 성공 시 콜백 호출
    if (result?.id && onSuccess) {
      onSuccess(result.id);
    }
    
    // 다이얼로그 닫기
    onOpenChange(false);
    
    // Form reset
    form.reset({
      name: '',
      cidr_block: '',
      region: selectedRegion || '',
      availability_zone: selectedZone || '',
      vpc_id: selectedVPCId,
      credential_id: selectedCredentialId,
    });
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('network.createSubnetTitle')}</DialogTitle>
          <DialogDescription>
            {selectedRegion && selectedZone
              ? t('network.createSubnetDescription', {
                  region: selectedRegion,
                  zone: selectedZone,
                  provider: selectedProvider?.toUpperCase() || ''
                })
              : selectedRegion
              ? t('network.createSubnetDescriptionRegionOnly', {
                  region: selectedRegion,
                  provider: selectedProvider?.toUpperCase() || ''
                })
              : t('network.createSubnetDescriptionNoRegion', {
                  provider: selectedProvider?.toUpperCase() || ''
                })
            }
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('network.subnetName')} *</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder="my-subnet" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            <FormField
              control={form.control}
              name="vpc_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('network.vpcId')} *</FormLabel>
                  <FormControl>
                    <Select
                      value={selectedVPCId}
                      onValueChange={(value) => {
                        onVPCChange(value);
                        field.onChange(value);
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={t('network.selectVPC')} />
                      </SelectTrigger>
                      <SelectContent>
                        {vpcs.map((vpc) => (
                          <SelectItem key={vpc.id} value={vpc.id}>
                            {vpc.name || vpc.id}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            <FormField
              control={form.control}
              name="cidr_block"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('network.cidrBlock')} *</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder="10.0.1.0/24" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="region"
                render={({ field }) => {
                  // selectedRegion이 있으면 항상 그 값을 사용하고 form에 등록
                  const regionValue = selectedRegion || field.value || '';
                  
                  return (
                    <FormItem>
                      <FormLabel>Region *</FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder="ap-northeast-2"
                          value={regionValue}
                          disabled={true}
                          readOnly
                          className="bg-muted cursor-not-allowed"
                        />
                      </FormControl>
                      {selectedRegion && (
                        <p className="text-sm text-muted-foreground">
                          Fixed from Basic Configuration
                        </p>
                      )}
                      <FormMessage />
                    </FormItem>
                  );
                }}
              />
              
              <FormField
                control={form.control}
                name="availability_zone"
                render={({ field }) => {
                  // AWS의 경우 Select로 표시, 그 외는 Input으로 표시
                  if (canSelectZone) {
                    // AWS: Select로 표시 (선택 가능)
                    const zoneValue = field.value || selectedZone || '';
                    
                    return (
                      <FormItem>
                        <FormLabel>Availability Zone *</FormLabel>
                        <FormControl>
                          <Select
                            value={zoneValue}
                            onValueChange={(value) => {
                              field.onChange(value);
                            }}
                            disabled={isLoadingZones || !selectedRegion}
                          >
                            <SelectTrigger>
                              <SelectValue placeholder={isLoadingZones ? 'Loading zones...' : 'Select availability zone'} />
                            </SelectTrigger>
                            <SelectContent>
                              {availabilityZones.length === 0 && !isLoadingZones ? (
                                <div className="p-2 text-sm text-muted-foreground">No zones available</div>
                              ) : (
                                availabilityZones.map((zone) => (
                                  <SelectItem key={zone} value={zone}>
                                    {zone}
                                  </SelectItem>
                                ))
                              )}
                            </SelectContent>
                          </Select>
                        </FormControl>
                        <FormDescription>
                          Select an availability zone for the subnet
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    );
                  } else {
                    // AWS가 아닌 경우 또는 zone을 선택할 수 없는 경우: Input으로 표시
                    const zoneValue = selectedZone || field.value || '';
                    const isReadOnly = !!selectedZone;
                    
                    return (
                      <FormItem>
                        <FormLabel>Availability Zone *</FormLabel>
                        <FormControl>
                          <Input
                            {...field}
                            placeholder="ap-northeast-2a"
                            value={zoneValue}
                            disabled={isReadOnly}
                            readOnly={isReadOnly}
                            className={isReadOnly ? 'bg-muted cursor-not-allowed' : ''}
                            onChange={(e) => {
                              if (!isReadOnly) {
                                field.onChange(e.target.value);
                              }
                            }}
                          />
                        </FormControl>
                        {selectedZone && (
                          <p className="text-sm text-muted-foreground">
                            Fixed from Basic Configuration
                          </p>
                        )}
                        <FormMessage />
                      </FormItem>
                    );
                  }
                }}
              />
            </div>
            <div className="flex justify-end space-x-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={createSubnetMutation.isPending}>
                {t('common.cancel')}
              </Button>
              <Button 
                type="submit" 
                disabled={createSubnetMutation.isPending || !selectedProvider || !selectedCredentialId || !selectedVPCId}
              >
                {createSubnetMutation.isPending ? t('actions.creating') : t('network.createSubnet')}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}


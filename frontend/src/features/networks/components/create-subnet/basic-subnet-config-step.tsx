/**
 * Basic Subnet Configuration Step
 * Step 1: 기본 Subnet 설정
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateSubnetForm, CloudProvider, VPC } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useNetworkResources } from '@/features/networks/hooks/use-network-resources';
import { useAvailabilityZones } from '@/features/kubernetes/hooks/use-kubernetes-metadata';

interface BasicSubnetConfigStepProps {
  form: UseFormReturn<CreateSubnetForm>;
  selectedProvider?: CloudProvider;
  onDataChange: (data: Partial<CreateSubnetForm>) => void;
}

export function BasicSubnetConfigStep({
  form,
  selectedProvider,
  onDataChange,
}: BasicSubnetConfigStepProps) {
  const { t } = useTranslation();
  const { selectedRegion, selectedCredentialId } = useCredentialContext();
  const { vpcs } = useNetworkResources({ resourceType: 'vpcs' });
  const selectedVPCId = form.watch('vpc_id');
  const formRegion = form.watch('region');
  // Dashboard에서 선택된 Region이 있으면 우선 사용, 없으면 form의 region 사용
  const activeRegion = selectedRegion || formRegion || '';

  // Fetch availability zones for AWS/Azure when region is selected
  // AWS와 Azure 모두 Availability Zone을 지원하므로 둘 다 처리
  const {
    data: availabilityZones = [],
    isLoading: isLoadingZones,
    isError: isZonesError,
    error: zonesError,
  } = useAvailabilityZones({
    provider: selectedProvider,
    credentialId: selectedCredentialId || '',
    region: activeRegion,
  });

  // AWS만 useAvailabilityZones hook을 통해 select box로 표시
  // Azure와 GCP는 현재 API가 없으므로 Input 필드로 표시
  const canLoadZones = selectedProvider === 'aws' && !!selectedCredentialId && !!activeRegion;

  return (
    <Form {...form}>
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Subnet Name */}
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem className="flex flex-col h-full">
                  <FormLabel className="mb-2">{t('network.subnetName')} *</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder="my-subnet" />
                  </FormControl>
                  <FormDescription className="mt-1">
                    Name for your subnet
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              )}
            />
          </div>

          {/* VPC ID */}
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="vpc_id"
              render={({ field }) => (
                <FormItem className="flex flex-col h-full">
                  <FormLabel className="mb-2">{t('network.vpcId')} *</FormLabel>
                  <FormControl>
                    <Select
                      value={field.value || ''}
                      onValueChange={(value) => {
                        field.onChange(value);
                        onDataChange({ vpc_id: value });
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={t('network.selectVPC')} />
                      </SelectTrigger>
                      <SelectContent>
                        {vpcs.map((vpc) => (
                          <SelectItem key={vpc.id} value={vpc.id}>
                            {vpc.name || vpc.id} {vpc.cidr_block && `(${vpc.cidr_block})`}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormControl>
                  <FormDescription className="mt-1">
                    Select the VPC for this subnet
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              )}
            />
          </div>
        </div>

        {/* CIDR Block */}
        <div className="space-y-1 min-h-[100px] flex flex-col h-full">
          <FormField
            control={form.control}
            name="cidr_block"
            render={({ field }) => (
              <FormItem className="flex flex-col h-full">
                <FormLabel className="mb-2">{t('network.cidrBlock')} *</FormLabel>
                <FormControl>
                  <Input {...field} placeholder="10.0.1.0/24" />
                </FormControl>
                <FormDescription className="mt-1">
                  CIDR block for the subnet (e.g., 10.0.1.0/24)
                </FormDescription>
                <FormMessage className="mt-1" />
              </FormItem>
            )}
          />
        </div>

        {/* Region and Availability Zone */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Region */}
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="region"
              render={({ field }) => {
                const regionValue = selectedRegion || field.value || '';
                return (
                  <FormItem className="flex flex-col h-full">
                    <FormLabel className="mb-2">{t('region.select')} *</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        value={regionValue}
                        placeholder="ap-northeast-3"
                        disabled={!!selectedRegion}
                        readOnly={!!selectedRegion}
                        className={selectedRegion ? 'bg-muted cursor-not-allowed' : ''}
                      />
                    </FormControl>
                    <FormDescription className="mt-1">
                      {selectedRegion 
                        ? 'Region is selected from dashboard'
                        : 'Select the region for the subnet'}
                    </FormDescription>
                    <FormMessage className="mt-1" />
                  </FormItem>
                );
              }}
            />
          </div>

          {/* Availability Zone / Zone */}
          {selectedProvider === 'gcp' ? (
            <div className="space-y-1 min-h-[100px] flex flex-col h-full">
              <FormField
                control={form.control}
                name="zone"
                render={({ field }) => (
                  <FormItem className="flex flex-col h-full">
                    <FormLabel className="mb-2">Zone *</FormLabel>
                    <FormControl>
                      <Input {...field} placeholder="asia-northeast3-a" />
                    </FormControl>
                    <FormDescription className="mt-1">
                      GCP zone for the subnet
                    </FormDescription>
                    <FormMessage className="mt-1" />
                  </FormItem>
                )}
              />
            </div>
          ) : (
            <div className="space-y-1 min-h-[100px] flex flex-col h-full">
              <FormField
                control={form.control}
                name="availability_zone"
                render={({ field }) => (
                  <FormItem className="flex flex-col h-full">
                    <FormLabel className="mb-2">Availability Zone *</FormLabel>
                    <FormControl>
                      {canLoadZones && activeRegion ? (
                        <Select
                          value={field.value || ''}
                          onValueChange={(value) => {
                            field.onChange(value);
                            onDataChange({ availability_zone: value });
                          }}
                          disabled={isLoadingZones || !activeRegion}
                        >
                          <SelectTrigger>
                            <SelectValue
                              placeholder={
                                !activeRegion
                                  ? 'Select region first'
                                  : isLoadingZones
                                  ? 'Loading zones...'
                                  : 'Select availability zone *'
                              }
                            />
                          </SelectTrigger>
                          <SelectContent>
                            {availabilityZones.length === 0 && !isLoadingZones ? (
                              <SelectItem value="no-zones" disabled>
                                No zones available
                              </SelectItem>
                            ) : (
                              availabilityZones.map((zone) => (
                                <SelectItem key={zone} value={zone}>
                                  {zone}
                                </SelectItem>
                              ))
                            )}
                          </SelectContent>
                        </Select>
                      ) : (
                        <Input
                          {...field}
                          placeholder={!activeRegion ? 'Select region first' : 'ap-northeast-3a'}
                          disabled={!activeRegion}
                        />
                      )}
                    </FormControl>
                    <FormDescription className="mt-1">
                      {!activeRegion
                        ? 'Please select a region first to enable availability zone selection.'
                        : canLoadZones && isLoadingZones
                        ? 'Loading availability zones...'
                        : 'Select availability zone for the subnet'}
                    </FormDescription>
                    {canLoadZones && activeRegion && isZonesError && zonesError && (
                      <FormDescription className="mt-1 text-destructive">
                        Failed to load availability zones: {zonesError.message}
                        {(zonesError.message.includes('IAM permission') ||
                          zonesError.message.includes('not authorized') ||
                          zonesError.message.includes('UnauthorizedOperation')) && (
                          <span className="block mt-1 text-muted-foreground">
                            <strong>Solution:</strong> Add the{' '}
                            <code className="px-1 py-0.5 bg-muted rounded">
                              ec2:DescribeAvailabilityZones
                            </code>{' '}
                            permission to your AWS IAM user or role.
                          </span>
                        )}
                      </FormDescription>
                    )}
                    <FormMessage className="mt-1" />
                  </FormItem>
                )}
              />
            </div>
          )}
        </div>
      </div>
    </Form>
  );
}


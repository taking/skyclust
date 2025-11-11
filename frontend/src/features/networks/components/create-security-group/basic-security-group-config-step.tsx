/**
 * Basic Security Group Configuration Step
 * Step 1: 기본 Security Group 설정
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateSecurityGroupForm, CloudProvider, VPC } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useNetworkResources } from '@/features/networks/hooks/use-network-resources';

interface BasicSecurityGroupConfigStepProps {
  form: UseFormReturn<CreateSecurityGroupForm>;
  selectedProvider?: CloudProvider;
  onDataChange: (data: Partial<CreateSecurityGroupForm>) => void;
}

export function BasicSecurityGroupConfigStep({
  form,
  selectedProvider,
  onDataChange,
}: BasicSecurityGroupConfigStepProps) {
  const { t } = useTranslation();
  const { selectedRegion } = useCredentialContext();
  const { vpcs } = useNetworkResources({ resourceType: 'vpcs' });
  const selectedVPCId = form.watch('vpc_id');
  const formRegion = form.watch('region');
  const activeRegion = selectedRegion || formRegion || '';

  return (
    <Form {...form}>
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Security Group Name */}
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem className="flex flex-col h-full">
                  <FormLabel className="mb-2">Security Group Name *</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder="my-security-group" />
                  </FormControl>
                  <FormDescription className="mt-1">
                    Name for your security group
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
                    Select the VPC for this security group
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              )}
            />
          </div>
        </div>

        {/* Description */}
        <div className="space-y-1 min-h-[100px] flex flex-col h-full">
          <FormField
            control={form.control}
            name="description"
            render={({ field }) => (
              <FormItem className="flex flex-col h-full">
                <FormLabel className="mb-2">Description *</FormLabel>
                <FormControl>
                  <Input {...field} placeholder="Security group description" />
                </FormControl>
                <FormDescription className="mt-1">
                  Description for your security group
                </FormDescription>
                <FormMessage className="mt-1" />
              </FormItem>
            )}
          />
        </div>

        {/* Region */}
        <div className="space-y-1 min-h-[100px] flex flex-col h-full">
          <FormField
            control={form.control}
            name="region"
            render={({ field }) => {
              const regionValue = activeRegion || field.value || '';
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
                      : 'Select the region for the security group'}
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              );
            }}
          />
        </div>

        {/* GCP specific fields */}
        {selectedProvider === 'gcp' && (
          <div className="space-y-4 pt-4 border-t">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-1 min-h-[100px] flex flex-col h-full">
                <FormField
                  control={form.control}
                  name="project_id"
                  render={({ field }) => (
                    <FormItem className="flex flex-col h-full">
                      <FormLabel className="mb-2">Project ID</FormLabel>
                      <FormControl>
                        <Input {...field} placeholder="my-gcp-project" />
                      </FormControl>
                      <FormDescription className="mt-1">
                        GCP Project ID (optional)
                      </FormDescription>
                      <FormMessage className="mt-1" />
                    </FormItem>
                  )}
                />
              </div>

              <div className="space-y-1 min-h-[100px] flex flex-col h-full">
                <FormField
                  control={form.control}
                  name="priority"
                  render={({ field }) => (
                    <FormItem className="flex flex-col h-full">
                      <FormLabel className="mb-2">Priority</FormLabel>
                      <FormControl>
                        <Input 
                          {...field} 
                          type="number"
                          placeholder="1000"
                          value={field.value || ''}
                          onChange={(e) => field.onChange(e.target.value ? parseInt(e.target.value) : undefined)}
                        />
                      </FormControl>
                      <FormDescription className="mt-1">
                        Firewall rule priority (0-65535)
                      </FormDescription>
                      <FormMessage className="mt-1" />
                    </FormItem>
                  )}
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-1 min-h-[100px] flex flex-col h-full">
                <FormField
                  control={form.control}
                  name="direction"
                  render={({ field }) => (
                    <FormItem className="flex flex-col h-full">
                      <FormLabel className="mb-2">Direction</FormLabel>
                      <FormControl>
                        <Select
                          value={field.value || 'INGRESS'}
                          onValueChange={(value) => {
                            field.onChange(value);
                            onDataChange({ direction: value as 'INGRESS' | 'EGRESS' });
                          }}
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="INGRESS">INGRESS</SelectItem>
                            <SelectItem value="EGRESS">EGRESS</SelectItem>
                          </SelectContent>
                        </Select>
                      </FormControl>
                      <FormDescription className="mt-1">
                        Traffic direction (INGRESS or EGRESS)
                      </FormDescription>
                      <FormMessage className="mt-1" />
                    </FormItem>
                  )}
                />
              </div>

              <div className="space-y-1 min-h-[100px] flex flex-col h-full">
                <FormField
                  control={form.control}
                  name="action"
                  render={({ field }) => (
                    <FormItem className="flex flex-col h-full">
                      <FormLabel className="mb-2">Action</FormLabel>
                      <FormControl>
                        <Select
                          value={field.value || 'ALLOW'}
                          onValueChange={(value) => {
                            field.onChange(value);
                            onDataChange({ action: value as 'ALLOW' | 'DENY' });
                          }}
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="ALLOW">ALLOW</SelectItem>
                            <SelectItem value="DENY">DENY</SelectItem>
                          </SelectContent>
                        </Select>
                      </FormControl>
                      <FormDescription className="mt-1">
                        Action to take (ALLOW or DENY)
                      </FormDescription>
                      <FormMessage className="mt-1" />
                    </FormItem>
                  )}
                />
              </div>
            </div>
          </div>
        )}

        {/* Azure specific fields */}
        {selectedProvider === 'azure' && (
          <div className="space-y-4 pt-4 border-t">
            <div className="space-y-1 min-h-[100px] flex flex-col h-full">
              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem className="flex flex-col h-full">
                    <FormLabel className="mb-2">Location</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        value={activeRegion || field.value || ''}
                        placeholder="eastus"
                        disabled={!!selectedRegion}
                        readOnly={!!selectedRegion}
                        className={selectedRegion ? 'bg-muted cursor-not-allowed' : ''}
                      />
                    </FormControl>
                    <FormDescription className="mt-1">
                      Azure location for the Network Security Group
                    </FormDescription>
                    <FormMessage className="mt-1" />
                  </FormItem>
                )}
              />
            </div>
          </div>
        )}
      </div>
    </Form>
  );
}


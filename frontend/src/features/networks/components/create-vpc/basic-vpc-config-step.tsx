/**
 * Basic VPC Configuration Step
 * Step 1: 기본 VPC 설정
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { EnhancedField } from '@/hooks/use-form-with-validation';
import type { CreateVPCForm, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { useCredentialContext } from '@/hooks/use-credential-context';

interface BasicVPCConfigStepProps {
  form: UseFormReturn<CreateVPCForm>;
  selectedProvider?: CloudProvider;
  onDataChange: (data: Partial<CreateVPCForm>) => void;
}

export function BasicVPCConfigStep({
  form,
  selectedProvider,
  onDataChange,
}: BasicVPCConfigStepProps) {
  const { t } = useTranslation();
  const { selectedRegion } = useCredentialContext();

  return (
    <Form {...form}>
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* VPC Name */}
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem className="flex flex-col h-full">
                  <FormLabel className="mb-2">{t('network.vpcName')} *</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder="my-vpc" />
                  </FormControl>
                  <FormDescription className="mt-1">
                    {t('network.vpcNameDescription')}
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              )}
            />
          </div>

          {/* Description */}
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem className="flex flex-col h-full">
                  <FormLabel className="mb-2">{t('network.vpcDescription')}</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder={t('network.vpcDescriptionPlaceholder')} />
                  </FormControl>
                  <FormDescription className="mt-1">
                    Optional description for your VPC
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              )}
            />
          </div>
        </div>

        {/* CIDR Block (AWS required, GCP optional) */}
        {selectedProvider !== 'azure' && (
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="cidr_block"
              render={({ field }) => (
                <FormItem className="flex flex-col h-full">
                  <FormLabel className="mb-2">
                    {t('network.cidrBlock')} {selectedProvider === 'aws' ? '*' : ''}
                  </FormLabel>
                  <FormControl>
                    <Input {...field} placeholder="10.0.0.0/16" />
                  </FormControl>
                  <FormDescription className="mt-1">
                    {selectedProvider === 'aws' 
                      ? 'CIDR block is required for AWS VPC (e.g., 10.0.0.0/16)'
                      : 'CIDR block is optional for GCP (auto-mode VPC can be created without CIDR)'}
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              )}
            />
          </div>
        )}

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
                      : 'Select the region where the VPC will be created'}
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              );
            }}
          />
        </div>

        {/* Azure specific fields */}
        {selectedProvider === 'azure' && (
          <div className="space-y-4 pt-4 border-t">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-1 min-h-[100px] flex flex-col h-full">
                <FormField
                  control={form.control}
                  name="location"
                  render={({ field }) => {
                    const locationValue = selectedRegion || field.value || '';
                    return (
                      <FormItem className="flex flex-col h-full">
                        <FormLabel className="mb-2">Location *</FormLabel>
                        <FormControl>
                          <Input
                            {...field}
                            value={locationValue}
                            placeholder="eastus"
                            disabled={!!selectedRegion}
                            readOnly={!!selectedRegion}
                            className={selectedRegion ? 'bg-muted cursor-not-allowed' : ''}
                          />
                        </FormControl>
                        <FormDescription className="mt-1">
                          Azure region for the Virtual Network
                        </FormDescription>
                        <FormMessage className="mt-1" />
                      </FormItem>
                    );
                  }}
                />
              </div>

              <div className="space-y-1 min-h-[100px] flex flex-col h-full">
                <FormField
                  control={form.control}
                  name="resource_group"
                  render={({ field }) => (
                    <FormItem className="flex flex-col h-full">
                      <FormLabel className="mb-2">Resource Group *</FormLabel>
                      <FormControl>
                        <Input {...field} placeholder="my-resource-group" />
                      </FormControl>
                      <FormDescription className="mt-1">
                        Azure Resource Group name
                      </FormDescription>
                      <FormMessage className="mt-1" />
                    </FormItem>
                  )}
                />
              </div>
            </div>

            <div className="space-y-1 min-h-[100px] flex flex-col h-full">
              <FormField
                control={form.control}
                name="address_space"
                render={({ field }) => (
                  <FormItem className="flex flex-col h-full">
                    <FormLabel className="mb-2">Address Space *</FormLabel>
                    <FormControl>
                      <Input 
                        {...field} 
                        placeholder="10.0.0.0/16" 
                        value={Array.isArray(field.value) ? field.value.join(', ') : field.value || ''}
                        onChange={(e) => {
                          const value = e.target.value;
                          const addresses = value.split(',').map(addr => addr.trim()).filter(Boolean);
                          field.onChange(addresses.length > 0 ? addresses : undefined);
                        }}
                      />
                    </FormControl>
                    <FormDescription className="mt-1">
                      Comma-separated CIDR blocks (e.g., &quot;10.0.0.0/16, 10.1.0.0/16&quot;)
                    </FormDescription>
                    <FormMessage className="mt-1" />
                  </FormItem>
                )}
              />
            </div>
          </div>
        )}

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
            </div>
          </div>
        )}
      </div>
    </Form>
  );
}


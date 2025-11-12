/**
 * Basic Resource Group Configuration Step
 * Step 1: 기본 Resource Group 설정
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateResourceGroupForm } from '@/features/resource-groups/hooks/use-resource-group-actions';
import { useTranslation } from '@/hooks/use-translation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { getRegionsByProvider } from '@/lib/regions';

interface BasicResourceGroupConfigStepProps {
  form: UseFormReturn<CreateResourceGroupForm>;
  onDataChange: (data: Partial<CreateResourceGroupForm>) => void;
}

export function BasicResourceGroupConfigStep({
  form,
  onDataChange,
}: BasicResourceGroupConfigStepProps) {
  const { t } = useTranslation();
  const { selectedRegion } = useCredentialContext();
  const regions = getRegionsByProvider('azure');

  return (
    <Form {...form}>
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Resource Group Name */}
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem className="flex flex-col h-full">
                  <FormLabel className="mb-2">Resource Group Name *</FormLabel>
                  <FormControl>
                    <Input 
                      {...field} 
                      placeholder="my-resource-group"
                      onChange={(e) => {
                        field.onChange(e);
                        onDataChange({ name: e.target.value });
                      }}
                    />
                  </FormControl>
                  <FormDescription className="mt-1">
                    A unique name for the resource group (alphanumeric, hyphens, underscores, and periods allowed)
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              )}
            />
          </div>

          {/* Location */}
          <div className="space-y-1 min-h-[100px] flex flex-col h-full">
            <FormField
              control={form.control}
              name="location"
              render={({ field }) => (
                <FormItem className="flex flex-col h-full">
                  <FormLabel className="mb-2">Location *</FormLabel>
                  <Select
                    value={field.value || selectedRegion || ''}
                    onValueChange={(value) => {
                      field.onChange(value);
                      onDataChange({ location: value });
                    }}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a location" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {regions.map((region) => (
                        <SelectItem key={region.value} value={region.value}>
                          {region.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormDescription className="mt-1">
                    The Azure region where the resource group will be created
                  </FormDescription>
                  <FormMessage className="mt-1" />
                </FormItem>
              )}
            />
          </div>
        </div>
      </div>
    </Form>
  );
}


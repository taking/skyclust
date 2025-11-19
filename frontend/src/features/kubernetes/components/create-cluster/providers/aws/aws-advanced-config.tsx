/**
 * AWS Advanced Configuration Component
 * AWS EKS 고급 설정: Access Config
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormField, FormItem, FormLabel, FormControl, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import type { CreateClusterForm } from '@/lib/types';

interface AWSAdvancedConfigProps {
  form: UseFormReturn<CreateClusterForm>;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
}

export function AWSAdvancedConfig({
  form,
  onDataChange,
}: AWSAdvancedConfigProps) {
  const accessConfig = form.watch('access_config') || {
    authentication_mode: 'API',
    bootstrap_cluster_creator_admin_permissions: true,
  };

  const handleAccessConfigChange = (field: string, value: unknown) => {
    const newAccessConfig = { ...accessConfig, [field]: value };
    form.setValue('access_config', newAccessConfig);
    onDataChange({ access_config: newAccessConfig });
  };

  return (
    <div className="space-y-6 mt-6 pt-6 border-t">
      <h3 className="text-lg font-semibold">AWS EKS Configuration</h3>
      
      {/* Access Config */}
      <div className="space-y-4">
        <div>
          <Label>Access Configuration</Label>
          <FormDescription className="mb-2">
            Configure cluster access settings
          </FormDescription>
        </div>

        <div className="space-y-4 p-4 border rounded-md">
          <FormField
            control={form.control}
            name="access_config.authentication_mode"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Authentication Mode</FormLabel>
                <FormControl>
                  <Input
                    placeholder="API"
                    value={field.value || 'API'}
                    onChange={(e) => {
                      field.onChange(e);
                      handleAccessConfigChange('authentication_mode', e.target.value);
                    }}
                  />
                </FormControl>
                <FormDescription>
                  Cluster authentication mode (default: API)
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <div className="flex items-center space-x-2">
            <Checkbox
              id="bootstrap-admin"
              checked={accessConfig.bootstrap_cluster_creator_admin_permissions ?? true}
              onCheckedChange={(checked) => {
                handleAccessConfigChange('bootstrap_cluster_creator_admin_permissions', checked);
              }}
            />
            <Label
              htmlFor="bootstrap-admin"
              className="text-sm font-normal cursor-pointer"
            >
              Bootstrap cluster creator admin permissions
            </Label>
          </div>
        </div>
      </div>
    </div>
  );
}


/**
 * Advanced Configuration Step
 * Step 3: 고급 설정 (Role ARN, Tags, Access Config)
 */

'use client';

import { useState } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Plus, X } from 'lucide-react';
import type { CreateClusterForm } from '@/lib/types';

interface AdvancedConfigStepProps {
  form: UseFormReturn<CreateClusterForm>;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
}

export function AdvancedConfigStep({
  form,
  onDataChange,
}: AdvancedConfigStepProps) {
  const [tagKey, setTagKey] = useState('');
  const [tagValue, setTagValue] = useState('');

  const tags = form.watch('tags') || {};
  const accessConfig = form.watch('access_config') || {
    authentication_mode: 'API',
    bootstrap_cluster_creator_admin_permissions: true,
  };

  const handleRoleARNChange = (value: string) => {
    form.setValue('role_arn', value);
    onDataChange({ role_arn: value });
  };

  const handleAddTag = () => {
    if (!tagKey.trim()) return;
    
    const currentTags = tags || {};
    const newTags = { ...currentTags, [tagKey.trim()]: tagValue.trim() };
    form.setValue('tags', newTags);
    onDataChange({ tags: newTags });
    setTagKey('');
    setTagValue('');
  };

  const handleRemoveTag = (key: string) => {
    const currentTags = tags || {};
    const newTags = { ...currentTags };
    delete newTags[key];
    form.setValue('tags', newTags);
    onDataChange({ tags: newTags });
  };

  const handleAccessConfigChange = (field: string, value: unknown) => {
    const newAccessConfig = { ...accessConfig, [field]: value };
    form.setValue('access_config', newAccessConfig);
    onDataChange({ access_config: newAccessConfig });
  };

  return (
    <Form {...form}>
      <div className="space-y-6">
        {/* Role ARN */}
        <FormField
          control={form.control}
          name="role_arn"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Role ARN (Optional)</FormLabel>
              <FormControl>
                <Input
                  placeholder="e.g., arn:aws:iam::123456789012:role/EKSServiceRole"
                  {...field}
                  onChange={(e) => {
                    field.onChange(e);
                    handleRoleARNChange(e.target.value);
                  }}
                />
              </FormControl>
              <FormDescription>
                IAM role ARN for the EKS cluster service role
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Tags */}
        <div className="space-y-4">
          <div>
            <Label>Tags (Optional)</Label>
            <FormDescription className="mb-2">
              Add key-value pairs to tag your cluster
            </FormDescription>
          </div>

          {/* Add Tag Input */}
          <div className="flex gap-2">
            <Input
              placeholder="Tag key"
              value={tagKey}
              onChange={(e) => setTagKey(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  handleAddTag();
                }
              }}
            />
            <Input
              placeholder="Tag value"
              value={tagValue}
              onChange={(e) => setTagValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  handleAddTag();
                }
              }}
            />
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={handleAddTag}
              disabled={!tagKey.trim()}
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>

          {/* Tags Display */}
          {Object.keys(tags).length > 0 && (
            <div className="space-y-2">
              {Object.entries(tags).map(([key, value]) => (
                <div key={key} className="flex items-center gap-2 p-2 bg-muted rounded-md">
                  <span className="text-sm font-medium">{key}:</span>
                  <span className="text-sm text-muted-foreground flex-1">{value}</span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveTag(key)}
                    className="h-6 w-6"
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>

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
    </Form>
  );
}


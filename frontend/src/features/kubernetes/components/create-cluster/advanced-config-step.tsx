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
import type { CreateClusterForm, CloudProvider } from '@/lib/types';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface AdvancedConfigStepProps {
  form: UseFormReturn<CreateClusterForm>;
  selectedProvider?: CloudProvider;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
}

export function AdvancedConfigStep({
  form,
  selectedProvider,
  onDataChange,
}: AdvancedConfigStepProps) {
  const [tagKey, setTagKey] = useState('');
  const [tagValue, setTagValue] = useState('');

  const tags = form.watch('tags') || {};
  const accessConfig = form.watch('access_config') || {
    authentication_mode: 'API',
    bootstrap_cluster_creator_admin_permissions: true,
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

        {/* Azure specific: Node Pool Configuration */}
        {selectedProvider === 'azure' && (
          <div className="space-y-6 mt-6 pt-6 border-t">
            <h3 className="text-lg font-semibold">Azure Node Pool Configuration</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Node Pool Name */}
              <FormField
                control={form.control}
                name="node_pool.name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Node Pool Name *</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="nodepool1"
                        value={field.value || ''}
                        onChange={(e) => {
                          field.onChange(e);
                          const currentNodePool = form.getValues('node_pool') || {};
                          form.setValue('node_pool', { ...currentNodePool, name: e.target.value });
                          onDataChange({ node_pool: { ...currentNodePool, name: e.target.value } });
                        }}
                      />
                    </FormControl>
                    <FormDescription>
                      Name for the node pool
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* VM Size */}
              <FormField
                control={form.control}
                name="node_pool.vm_size"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>VM Size *</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Standard_D2s_v3"
                        value={field.value || ''}
                        onChange={(e) => {
                          field.onChange(e);
                          const currentNodePool = form.getValues('node_pool') || {};
                          form.setValue('node_pool', { ...currentNodePool, vm_size: e.target.value });
                          onDataChange({ node_pool: { ...currentNodePool, vm_size: e.target.value } });
                        }}
                      />
                    </FormControl>
                    <FormDescription>
                      Azure VM size (e.g., Standard_D2s_v3)
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {/* Node Count */}
              <FormField
                control={form.control}
                name="node_pool.node_count"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Node Count *</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min="1"
                        placeholder="3"
                        value={field.value || ''}
                        onChange={(e) => {
                          const value = parseInt(e.target.value, 10) || 0;
                          field.onChange(value);
                          const currentNodePool = form.getValues('node_pool') || {};
                          form.setValue('node_pool', { ...currentNodePool, node_count: value });
                          onDataChange({ node_pool: { ...currentNodePool, node_count: value } });
                        }}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Min Count (for auto-scaling) */}
              <FormField
                control={form.control}
                name="node_pool.min_count"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Min Count</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min="0"
                        placeholder="1"
                        value={field.value || ''}
                        onChange={(e) => {
                          const value = parseInt(e.target.value, 10) || 0;
                          field.onChange(value);
                          const currentNodePool = form.getValues('node_pool') || {};
                          form.setValue('node_pool', { ...currentNodePool, min_count: value });
                          onDataChange({ node_pool: { ...currentNodePool, min_count: value } });
                        }}
                      />
                    </FormControl>
                    <FormDescription>Minimum nodes (for auto-scaling)</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Max Count (for auto-scaling) */}
              <FormField
                control={form.control}
                name="node_pool.max_count"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Max Count</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min="1"
                        placeholder="10"
                        value={field.value || ''}
                        onChange={(e) => {
                          const value = parseInt(e.target.value, 10) || 0;
                          field.onChange(value);
                          const currentNodePool = form.getValues('node_pool') || {};
                          form.setValue('node_pool', { ...currentNodePool, max_count: value });
                          onDataChange({ node_pool: { ...currentNodePool, max_count: value } });
                        }}
                      />
                    </FormControl>
                    <FormDescription>Maximum nodes (for auto-scaling)</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* Enable Auto Scaling */}
            <FormField
              control={form.control}
              name="node_pool.enable_auto_scaling"
              render={({ field }) => (
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="enable-auto-scaling"
                    checked={field.value || false}
                    onCheckedChange={(checked) => {
                      field.onChange(checked);
                      const currentNodePool = form.getValues('node_pool') || {};
                      form.setValue('node_pool', { ...currentNodePool, enable_auto_scaling: checked });
                      onDataChange({ node_pool: { ...currentNodePool, enable_auto_scaling: checked } });
                    }}
                  />
                  <Label
                    htmlFor="enable-auto-scaling"
                    className="text-sm font-normal cursor-pointer"
                  >
                    Enable Auto Scaling
                  </Label>
                </div>
              )}
            />
          </div>
        )}
      </div>
    </Form>
  );
}


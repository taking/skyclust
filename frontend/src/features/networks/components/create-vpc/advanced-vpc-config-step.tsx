/**
 * Advanced VPC Configuration Step
 * Step 2: 고급 VPC 설정 (Optional)
 */

'use client';

import { useState } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Plus, X } from 'lucide-react';
import type { CreateVPCForm, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface AdvancedVPCConfigStepProps {
  form: UseFormReturn<CreateVPCForm>;
  selectedProvider?: CloudProvider;
  onDataChange: (data: Partial<CreateVPCForm>) => void;
}

export function AdvancedVPCConfigStep({
  form,
  selectedProvider,
  onDataChange,
}: AdvancedVPCConfigStepProps) {
  const { t } = useTranslation();
  const [tagKey, setTagKey] = useState('');
  const [tagValue, setTagValue] = useState('');

  const tags = form.watch('tags') || {};

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

  return (
    <Form {...form}>
      <div className="space-y-6">
        {/* Tags */}
        <div className="space-y-4">
          <div>
            <Label>Tags (Optional)</Label>
            <FormDescription className="mb-2">
              Add key-value pairs to tag your VPC
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
            <div className="flex flex-wrap gap-2 pt-2">
              {Object.entries(tags).map(([key, value]) => (
                <Badge key={key} variant="secondary" className="flex items-center gap-2">
                  <span>{key}: {value}</span>
                  <button
                    type="button"
                    onClick={() => handleRemoveTag(key)}
                    className="ml-1 hover:bg-destructive/20 rounded-full p-0.5"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              ))}
            </div>
          )}
        </div>

        {/* GCP specific advanced options */}
        {selectedProvider === 'gcp' && (
          <div className="space-y-4 pt-4 border-t">
            <div className="space-y-1 min-h-[100px] flex flex-col h-full">
              <FormField
                control={form.control}
                name="auto_create_subnets"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center space-x-3 space-y-0">
                    <FormControl>
                      <Checkbox
                        checked={field.value || false}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <div className="space-y-1 leading-none">
                      <FormLabel>Auto-create Subnets</FormLabel>
                      <FormDescription>
                        Automatically create subnets in all regions
                      </FormDescription>
                    </div>
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-1 min-h-[100px] flex flex-col h-full">
                <FormField
                  control={form.control}
                  name="routing_mode"
                  render={({ field }) => (
                    <FormItem className="flex flex-col h-full">
                      <FormLabel className="mb-2">Routing Mode</FormLabel>
                      <FormControl>
                        <Input {...field} placeholder="REGIONAL" />
                      </FormControl>
                      <FormDescription className="mt-1">
                        GCP routing mode (REGIONAL or GLOBAL)
                      </FormDescription>
                      <FormMessage className="mt-1" />
                    </FormItem>
                  )}
                />
              </div>

              <div className="space-y-1 min-h-[100px] flex flex-col h-full">
                <FormField
                  control={form.control}
                  name="mtu"
                  render={({ field }) => (
                    <FormItem className="flex flex-col h-full">
                      <FormLabel className="mb-2">MTU</FormLabel>
                      <FormControl>
                        <Input 
                          {...field} 
                          type="number"
                          placeholder="1460"
                          value={field.value || ''}
                          onChange={(e) => field.onChange(e.target.value ? parseInt(e.target.value) : undefined)}
                        />
                      </FormControl>
                      <FormDescription className="mt-1">
                        Maximum Transmission Unit (1280-8896)
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


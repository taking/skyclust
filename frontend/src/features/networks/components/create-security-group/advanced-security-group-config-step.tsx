/**
 * Advanced Security Group Configuration Step
 * Step 2: 고급 Security Group 설정 (Optional)
 */

'use client';

import { useState } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import { Plus, X } from 'lucide-react';
import type { CreateSecurityGroupForm, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface AdvancedSecurityGroupConfigStepProps {
  form: UseFormReturn<CreateSecurityGroupForm>;
  selectedProvider?: CloudProvider;
  onDataChange: (data: Partial<CreateSecurityGroupForm>) => void;
}

export function AdvancedSecurityGroupConfigStep({
  form,
  selectedProvider,
  onDataChange,
}: AdvancedSecurityGroupConfigStepProps) {
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
              Add key-value pairs to tag your security group
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
                name="source_ranges"
                render={({ field }) => (
                  <FormItem className="flex flex-col h-full">
                    <FormLabel className="mb-2">Source Ranges</FormLabel>
                    <FormControl>
                      <Input 
                        {...field} 
                        placeholder="0.0.0.0/0"
                        value={Array.isArray(field.value) ? field.value.join(', ') : field.value || ''}
                        onChange={(e) => {
                          const value = e.target.value;
                          const ranges = value.split(',').map(range => range.trim()).filter(Boolean);
                          field.onChange(ranges.length > 0 ? ranges : undefined);
                        }}
                      />
                    </FormControl>
                    <FormDescription className="mt-1">
                      Comma-separated CIDR blocks (e.g., &quot;0.0.0.0/0, 10.0.0.0/8&quot;)
                    </FormDescription>
                    <FormMessage className="mt-1" />
                  </FormItem>
                )}
              />
            </div>

            <div className="space-y-1 min-h-[100px] flex flex-col h-full">
              <FormField
                control={form.control}
                name="target_tags"
                render={({ field }) => (
                  <FormItem className="flex flex-col h-full">
                    <FormLabel className="mb-2">Target Tags</FormLabel>
                    <FormControl>
                      <Input 
                        {...field} 
                        placeholder="gke-node"
                        value={Array.isArray(field.value) ? field.value.join(', ') : field.value || ''}
                        onChange={(e) => {
                          const value = e.target.value;
                          const tags = value.split(',').map(tag => tag.trim()).filter(Boolean);
                          field.onChange(tags.length > 0 ? tags : undefined);
                        }}
                      />
                    </FormControl>
                    <FormDescription className="mt-1">
                      Comma-separated target tags (e.g., &quot;gke-node, web-server&quot;)
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


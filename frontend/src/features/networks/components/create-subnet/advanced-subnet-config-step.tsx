/**
 * Advanced Subnet Configuration Step
 * Step 2: 고급 Subnet 설정 (Optional)
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
import type { CreateSubnetForm, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface AdvancedSubnetConfigStepProps {
  form: UseFormReturn<CreateSubnetForm>;
  selectedProvider?: CloudProvider;
  onDataChange: (data: Partial<CreateSubnetForm>) => void;
}

export function AdvancedSubnetConfigStep({
  form,
  selectedProvider,
  onDataChange,
}: AdvancedSubnetConfigStepProps) {
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
              Add key-value pairs to tag your subnet
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
                name="private_ip_google_access"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center space-x-3 space-y-0">
                    <FormControl>
                      <Checkbox
                        checked={field.value || false}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <div className="space-y-1 leading-none">
                      <FormLabel>Private IP Google Access</FormLabel>
                      <FormDescription>
                        Enable private Google access for instances without external IP
                      </FormDescription>
                    </div>
                  </FormItem>
                )}
              />
            </div>

            <div className="space-y-1 min-h-[100px] flex flex-col h-full">
              <FormField
                control={form.control}
                name="flow_logs"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center space-x-3 space-y-0">
                    <FormControl>
                      <Checkbox
                        checked={field.value || false}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <div className="space-y-1 leading-none">
                      <FormLabel>Flow Logs</FormLabel>
                      <FormDescription>
                        Enable VPC flow logs for this subnet
                      </FormDescription>
                    </div>
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


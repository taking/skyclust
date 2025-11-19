/**
 * Advanced Configuration Step
 * Step 3: 고급 설정 (Tags)
 */

'use client';

import { useState } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { Form, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Plus, X } from 'lucide-react';
import type { CreateClusterForm, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { AWSAdvancedConfig } from './providers/aws/aws-advanced-config';
import { GCPAdvancedConfig } from './providers/gcp/gcp-advanced-config';
import { AzureAdvancedConfig } from './providers/azure/azure-advanced-config';

interface AdvancedConfigStepProps {
  form: UseFormReturn<CreateClusterForm>;
  selectedProvider?: CloudProvider;
  selectedCredentialId?: string;
  selectedProjectId?: string;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
}

export function AdvancedConfigStep({
  form,
  selectedProvider,
  selectedCredentialId,
  selectedProjectId,
  onDataChange,
}: AdvancedConfigStepProps) {
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

        {/* Provider-specific Advanced Configuration */}
        {selectedProvider === 'aws' && (
          <AWSAdvancedConfig
            form={form}
            onDataChange={onDataChange}
          />
        )}
        {selectedProvider === 'gcp' && (
          <GCPAdvancedConfig
            form={form}
            selectedCredentialId={selectedCredentialId}
            selectedProjectId={selectedProjectId}
            onDataChange={onDataChange}
          />
        )}
        {selectedProvider === 'azure' && (
          <AzureAdvancedConfig
            form={form}
            onDataChange={onDataChange}
            deploymentMode={form.watch('deployment_mode')}
          />
        )}
      </div>
    </Form>
  );
}


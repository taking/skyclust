'use client';

import * as React from 'react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { X, Plus, Tag } from 'lucide-react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';

interface TagManagerProps {
  tags: Record<string, string>;
  onTagsChange: (tags: Record<string, string>) => void;
  readonly?: boolean;
}

function TagManagerComponent({ tags, onTagsChange, readonly = false }: TagManagerProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newTagKey, setNewTagKey] = useState('');
  const [newTagValue, setNewTagValue] = useState('');

  const handleAddTag = () => {
    if (newTagKey.trim() && newTagValue.trim()) {
      const updatedTags = { ...tags, [newTagKey.trim()]: newTagValue.trim() };
      onTagsChange(updatedTags);
      setNewTagKey('');
      setNewTagValue('');
      setIsDialogOpen(false);
    }
  };

  const handleRemoveTag = (key: string) => {
    const updatedTags = { ...tags };
    delete updatedTags[key];
    onTagsChange(updatedTags);
  };

  const tagEntries = Object.entries(tags);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <Label>Tags</Label>
        {!readonly && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => setIsDialogOpen(true)}
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Tag
          </Button>
        )}
      </div>
      
      {tagEntries.length === 0 ? (
        <p className="text-sm text-gray-500">No tags assigned</p>
      ) : (
        <div className="flex flex-wrap gap-2">
          {tagEntries.map(([key, value]) => (
            <Badge key={key} variant="secondary" className="flex items-center space-x-1">
              <Tag className="h-3 w-3" />
              <span>{key}: {value}</span>
              {!readonly && (
                <button
                  onClick={() => handleRemoveTag(key)}
                  className="ml-1 hover:bg-gray-300 rounded-full p-0.5"
                  aria-label={`Remove tag ${key}`}
                >
                  <X className="h-3 w-3" />
                </button>
              )}
            </Badge>
          ))}
        </div>
      )}

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Tag</DialogTitle>
            <DialogDescription>
              Add a key-value pair tag to this resource
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="tag-key">Key *</Label>
              <Input
                id="tag-key"
                value={newTagKey}
                onChange={(e) => setNewTagKey(e.target.value)}
                placeholder="e.g., Environment"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    document.getElementById('tag-value')?.focus();
                  }
                }}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="tag-value">Value *</Label>
              <Input
                id="tag-value"
                value={newTagValue}
                onChange={(e) => setNewTagValue(e.target.value)}
                placeholder="e.g., Production"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && newTagKey && newTagValue) {
                    e.preventDefault();
                    handleAddTag();
                  }
                }}
              />
            </div>
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleAddTag} disabled={!newTagKey.trim() || !newTagValue.trim()}>
                Add Tag
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export const TagManager = React.memo(TagManagerComponent);


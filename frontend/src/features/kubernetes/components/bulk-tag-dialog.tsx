/**
 * Bulk Tag Dialog Component
 * 여러 클러스터에 일괄 태그 추가 다이얼로그
 */

'use client';

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface BulkTagDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: () => void;
  tagKey: string;
  tagValue: string;
  onTagKeyChange: (key: string) => void;
  onTagValueChange: (value: string) => void;
  selectedCount: number;
  isPending?: boolean;
}

export function BulkTagDialog({
  open,
  onOpenChange,
  onSubmit,
  tagKey,
  tagValue,
  onTagKeyChange,
  onTagValueChange,
  selectedCount,
  isPending = false,
}: BulkTagDialogProps) {
  const handleSubmit = () => {
    if (tagKey.trim() && tagValue.trim()) {
      onSubmit();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && tagKey.trim() && tagValue.trim()) {
      handleSubmit();
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add Tags to Selected Clusters</DialogTitle>
          <DialogDescription>
            Add the same tag to {selectedCount} selected cluster(s)
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="bulk-tag-key">Tag Key *</Label>
            <Input
              id="bulk-tag-key"
              value={tagKey}
              onChange={(e) => onTagKeyChange(e.target.value)}
              placeholder="e.g., Environment"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="bulk-tag-value">Tag Value *</Label>
            <Input
              id="bulk-tag-value"
              value={tagValue}
              onChange={(e) => onTagValueChange(e.target.value)}
              placeholder="e.g., Production"
              onKeyDown={handleKeyDown}
            />
          </div>
          <div className="flex justify-end space-x-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={!tagKey.trim() || !tagValue.trim() || isPending}>
              {isPending ? 'Adding...' : 'Add Tag'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}


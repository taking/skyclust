/**
 * Create Node Pool Dialog Component
 * 노드 풀 생성 다이얼로그 컴포넌트
 */

'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createNodePoolSchema } from '@/lib/validations';
import type { CreateNodePoolForm } from '@/lib/types';

interface CreateNodePoolDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clusterName: string;
  defaultRegion: string;
  defaultCredentialId: string;
  onSubmit: (data: CreateNodePoolForm) => void;
  onCredentialIdChange: (credentialId: string) => void;
  onRegionChange: (region: string) => void;
  isPending: boolean;
}

export function CreateNodePoolDialog({
  open,
  onOpenChange,
  clusterName,
  defaultRegion,
  defaultCredentialId,
  onSubmit,
  onCredentialIdChange,
  onRegionChange,
  isPending,
}: CreateNodePoolDialogProps) {
  const form = useForm<CreateNodePoolForm>({
    resolver: zodResolver(createNodePoolSchema),
    defaultValues: {
      cluster_name: clusterName,
      region: defaultRegion,
      credential_id: defaultCredentialId,
    },
  });

  const handleSubmit = form.handleSubmit((data) => {
    onSubmit(data);
    form.reset({
      cluster_name: clusterName,
      region: defaultRegion,
      credential_id: defaultCredentialId,
    });
  });

  // Sync form values when props change
  React.useEffect(() => {
    if (defaultCredentialId) {
      form.setValue('credential_id', defaultCredentialId);
      onCredentialIdChange(defaultCredentialId);
    }
  }, [defaultCredentialId]); // eslint-disable-line react-hooks/exhaustive-deps

  React.useEffect(() => {
    if (defaultRegion) {
      form.setValue('region', defaultRegion);
      onRegionChange(defaultRegion);
    }
  }, [defaultRegion]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Node Pool</DialogTitle>
          <DialogDescription>
            Create a new node pool for this cluster
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="np-name">Name *</Label>
              <Input id="np-name" {...form.register('name')} />
              {form.formState.errors.name && (
                <p className="text-sm text-red-600">{form.formState.errors.name.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="np-instance-type">Instance Type *</Label>
              <Input id="np-instance-type" {...form.register('instance_type')} placeholder="n1-standard-2" />
              {form.formState.errors.instance_type && (
                <p className="text-sm text-red-600">{form.formState.errors.instance_type.message}</p>
              )}
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="np-min-nodes">Min Nodes</Label>
              <Input
                id="np-min-nodes"
                type="number"
                {...form.register('min_nodes', { valueAsNumber: true })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="np-max-nodes">Max Nodes</Label>
              <Input
                id="np-max-nodes"
                type="number"
                {...form.register('max_nodes', { valueAsNumber: true })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="np-node-count">Node Count</Label>
              <Input
                id="np-node-count"
                type="number"
                {...form.register('node_count', { valueAsNumber: true })}
              />
            </div>
          </div>
          <div className="flex justify-end space-x-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating...' : 'Create Node Pool'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}


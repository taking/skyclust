/**
 * Create Node Group Dialog Component
 * 노드 그룹 생성 다이얼로그 컴포넌트 (EKS용)
 */

'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import type { CreateNodeGroupForm } from '@/lib/types';

const createNodeGroupSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required'),
  cluster_name: z.string().min(1, 'Cluster name is required'),
  instance_type: z.string().min(1, 'Instance type is required'),
  disk_size_gb: z.number().min(10).optional(),
  min_size: z.number().min(0),
  max_size: z.number().min(1),
  desired_size: z.number().min(0),
  region: z.string().min(1, 'Region is required'),
});

interface CreateNodeGroupDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clusterName: string;
  defaultRegion: string;
  defaultCredentialId: string;
  onSubmit: (data: CreateNodeGroupForm) => void;
  onCredentialIdChange: (credentialId: string) => void;
  onRegionChange: (region: string) => void;
  isPending: boolean;
}

export function CreateNodeGroupDialog({
  open,
  onOpenChange,
  clusterName,
  defaultRegion,
  defaultCredentialId,
  onSubmit,
  onCredentialIdChange,
  onRegionChange,
  isPending,
}: CreateNodeGroupDialogProps) {
  const form = useForm<CreateNodeGroupForm>({
    resolver: zodResolver(createNodeGroupSchema),
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
          <DialogTitle>Create Node Group</DialogTitle>
          <DialogDescription>
            Create a new node group for this cluster
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="ng-name">Name *</Label>
              <Input id="ng-name" {...form.register('name')} />
              {form.formState.errors.name && (
                <p className="text-sm text-red-600">{form.formState.errors.name.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="ng-instance-type">Instance Type *</Label>
              <Input id="ng-instance-type" {...form.register('instance_type')} placeholder="t3.medium" />
              {form.formState.errors.instance_type && (
                <p className="text-sm text-red-600">{form.formState.errors.instance_type.message}</p>
              )}
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="ng-min-size">Min Size</Label>
              <Input
                id="ng-min-size"
                type="number"
                {...form.register('min_size', { valueAsNumber: true })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ng-max-size">Max Size</Label>
              <Input
                id="ng-max-size"
                type="number"
                {...form.register('max_size', { valueAsNumber: true })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ng-desired-size">Desired Size</Label>
              <Input
                id="ng-desired-size"
                type="number"
                {...form.register('desired_size', { valueAsNumber: true })}
              />
            </div>
          </div>
          <div className="flex justify-end space-x-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating...' : 'Create Node Group'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}


/**
 * Create VPC Dialog Component
 * VPC 생성 다이얼로그 컴포넌트
 */

'use client';

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createVPCSchema } from '@/lib/validations';
import { Plus } from 'lucide-react';
import type { CreateVPCForm } from '@/lib/types';

interface CreateVPCDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateVPCForm) => void;
  selectedProvider?: string;
  selectedRegion?: string;
  isPending: boolean;
  disabled?: boolean;
}

export function CreateVPCDialog({
  open,
  onOpenChange,
  onSubmit,
  selectedProvider,
  selectedRegion,
  isPending,
  disabled = false,
}: CreateVPCDialogProps) {
  const form = useForm<CreateVPCForm>({
    resolver: zodResolver(createVPCSchema),
    defaultValues: {
      region: selectedRegion || '',
    },
  });

  const handleSubmit = form.handleSubmit((data) => {
    onSubmit(data);
    form.reset({
      region: selectedRegion || '',
    });
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        <Button disabled={disabled}>
          <Plus className="mr-2 h-4 w-4" />
          Create VPC
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create VPC</DialogTitle>
          <DialogDescription>
            Create a new VPC on {selectedProvider?.toUpperCase()}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="vpc-name">Name *</Label>
            <Input id="vpc-name" {...form.register('name')} placeholder="my-vpc" />
            {form.formState.errors.name && (
              <p className="text-sm text-red-600">{form.formState.errors.name.message}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="vpc-description">Description</Label>
            <Input id="vpc-description" {...form.register('description')} placeholder="VPC description" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="vpc-cidr">CIDR Block</Label>
              <Input id="vpc-cidr" {...form.register('cidr_block')} placeholder="10.0.0.0/16" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="vpc-region">Region</Label>
              <Input
                id="vpc-region"
                {...form.register('region')}
                placeholder="ap-northeast-2"
                defaultValue={selectedRegion || ''}
                onChange={(e) => {
                  form.setValue('region', e.target.value);
                }}
              />
            </div>
          </div>
          <div className="flex justify-end space-x-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating...' : 'Create VPC'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}


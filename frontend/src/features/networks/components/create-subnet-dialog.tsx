/**
 * Create Subnet Dialog Component
 * Subnet 생성 다이얼로그 컴포넌트
 */

'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createSubnetSchema } from '@/lib/validations';
import { Plus } from 'lucide-react';
import type { CreateSubnetForm, VPC } from '@/lib/types';

interface CreateSubnetDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateSubnetForm) => void;
  selectedProvider?: string;
  selectedRegion?: string;
  selectedVPCId: string;
  vpcs: VPC[];
  onVPCChange: (vpcId: string) => void;
  isPending: boolean;
  disabled?: boolean;
}

export function CreateSubnetDialog({
  open,
  onOpenChange,
  onSubmit,
  selectedProvider,
  selectedRegion,
  selectedVPCId,
  vpcs,
  onVPCChange,
  isPending,
  disabled = false,
}: CreateSubnetDialogProps) {
  const form = useForm<CreateSubnetForm>({
    resolver: zodResolver(createSubnetSchema),
    defaultValues: {
      region: selectedRegion || '',
      vpc_id: selectedVPCId,
    },
  });

  // Update form when VPC changes
  React.useEffect(() => {
    if (selectedVPCId) {
      form.setValue('vpc_id', selectedVPCId);
    }
  }, [selectedVPCId, form]);

  // Update form when region changes
  React.useEffect(() => {
    if (selectedRegion) {
      form.setValue('region', selectedRegion);
    }
  }, [selectedRegion, form]);

  const handleSubmit = form.handleSubmit((data) => {
    onSubmit(data);
    form.reset({
      region: selectedRegion || '',
      vpc_id: selectedVPCId,
    });
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        <Button disabled={disabled}>
          <Plus className="mr-2 h-4 w-4" />
          Create Subnet
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Subnet</DialogTitle>
          <DialogDescription>
            Create a new subnet on {selectedProvider?.toUpperCase()}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="subnet-name">Name *</Label>
            <Input id="subnet-name" {...form.register('name')} placeholder="my-subnet" />
            {form.formState.errors.name && (
              <p className="text-sm text-red-600">{form.formState.errors.name.message}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="subnet-vpc">VPC ID *</Label>
            <Select
              value={selectedVPCId}
              onValueChange={(value) => {
                onVPCChange(value);
                form.setValue('vpc_id', value);
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select VPC" />
              </SelectTrigger>
              <SelectContent>
                {vpcs.map((vpc) => (
                  <SelectItem key={vpc.id} value={vpc.id}>
                    {vpc.name || vpc.id}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {form.formState.errors.vpc_id && (
              <p className="text-sm text-red-600">{form.formState.errors.vpc_id.message}</p>
            )}
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="subnet-cidr">CIDR Block *</Label>
              <Input id="subnet-cidr" {...form.register('cidr_block')} placeholder="10.0.1.0/24" />
              {form.formState.errors.cidr_block && (
                <p className="text-sm text-red-600">{form.formState.errors.cidr_block.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="subnet-az">Availability Zone *</Label>
              <Input id="subnet-az" {...form.register('availability_zone')} placeholder="ap-northeast-2a" />
              {form.formState.errors.availability_zone && (
                <p className="text-sm text-red-600">{form.formState.errors.availability_zone.message}</p>
              )}
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="subnet-region">Region *</Label>
            <Input
              id="subnet-region"
              {...form.register('region')}
              placeholder="ap-northeast-2"
              defaultValue={selectedRegion || ''}
              onChange={(e) => {
                form.setValue('region', e.target.value);
              }}
            />
            {form.formState.errors.region && (
              <p className="text-sm text-red-600">{form.formState.errors.region.message}</p>
            )}
          </div>
          <div className="flex justify-end space-x-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating...' : 'Create Subnet'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}


/**
 * Create Security Group Dialog Component
 * Security Group 생성 다이얼로그 컴포넌트
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
import { createValidationSchemas } from '@/lib/validations';
import { Plus } from 'lucide-react';
import type { CreateSecurityGroupForm, VPC } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface CreateSecurityGroupDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateSecurityGroupForm) => void;
  selectedProvider?: string;
  selectedRegion?: string;
  selectedVPCId: string;
  vpcs: VPC[];
  onVPCChange: (vpcId: string) => void;
  isPending: boolean;
  disabled?: boolean;
}

export function CreateSecurityGroupDialog({
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
}: CreateSecurityGroupDialogProps) {
  const { t } = useTranslation();
  const schemas = createValidationSchemas(t);
  const form = useForm<CreateSecurityGroupForm>({
    resolver: zodResolver(schemas.createSecurityGroupSchema),
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
          Create Security Group
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Security Group</DialogTitle>
          <DialogDescription>
            Create a new security group on {selectedProvider?.toUpperCase()}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="sg-name">Name *</Label>
            <Input id="sg-name" {...form.register('name')} placeholder="my-security-group" />
            {form.formState.errors.name && (
              <p className="text-sm text-red-600">{form.formState.errors.name.message}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="sg-description">Description</Label>
            <Input id="sg-description" {...form.register('description')} placeholder="Security group description" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="sg-vpc">VPC ID *</Label>
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
          <div className="space-y-2">
            <Label htmlFor="sg-region">Region *</Label>
            <Input
              id="sg-region"
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
              {isPending ? 'Creating...' : 'Create Security Group'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}


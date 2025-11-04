/**
 * Create Cluster Dialog Component
 * Kubernetes 클러스터 생성 폼 다이얼로그
 */

'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateClusterForm, Credential, CloudProvider } from '@/lib/types';

const createClusterSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  version: z.string().min(1, 'Version is required'),
  region: z.string().min(1, 'Region is required'),
  zone: z.string().optional(),
  subnet_ids: z.array(z.string()).min(1, 'At least one subnet is required'),
  vpc_id: z.string().optional(),
  role_arn: z.string().optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

interface CreateClusterDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateClusterForm) => void;
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  selectedProvider?: CloudProvider;
  isPending?: boolean;
}

export function CreateClusterDialog({
  open,
  onOpenChange,
  onSubmit,
  credentials,
  selectedCredentialId,
  onCredentialChange,
  selectedProvider,
  isPending = false,
}: CreateClusterDialogProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
  } = useForm<CreateClusterForm>({
    resolver: zodResolver(createClusterSchema),
    defaultValues: {
      credential_id: selectedCredentialId || '',
    },
  });

  const handleFormSubmit = (data: CreateClusterForm) => {
    onSubmit(data);
    reset();
  };

  const handleCancel = () => {
    reset();
    onOpenChange(false);
  };

  return (
    <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Create Kubernetes Cluster</DialogTitle>
        <DialogDescription>
          Create a new Kubernetes cluster on {selectedProvider?.toUpperCase() || 'your cloud provider'}
        </DialogDescription>
      </DialogHeader>
      <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="credential_id">Credential *</Label>
          <Select
            value={selectedCredentialId || ''}
            onValueChange={(value) => {
              onCredentialChange(value);
              setValue('credential_id', value);
            }}
          >
            <SelectTrigger>
              <SelectValue placeholder="Select credential" />
            </SelectTrigger>
            <SelectContent>
              {credentials.map((cred) => (
                <SelectItem key={cred.id} value={cred.id}>
                  {cred.provider} - {cred.id.substring(0, 8)}...
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.credential_id && (
            <p className="text-sm text-red-600">{errors.credential_id.message}</p>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="name">Cluster Name *</Label>
            <Input
              id="name"
              {...register('name')}
              placeholder="my-cluster"
            />
            {errors.name && (
              <p className="text-sm text-red-600">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="version">Kubernetes Version *</Label>
            <Input
              id="version"
              {...register('version')}
              placeholder="1.28"
            />
            {errors.version && (
              <p className="text-sm text-red-600">{errors.version.message}</p>
            )}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="region">Region *</Label>
            <Input
              id="region"
              {...register('region')}
              placeholder="ap-northeast-2"
            />
            {errors.region && (
              <p className="text-sm text-red-600">{errors.region.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="zone">Zone (Optional)</Label>
            <Input
              id="zone"
              {...register('zone')}
              placeholder="ap-northeast-2a"
            />
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="subnet_ids">Subnet IDs *</Label>
          <Input
            id="subnet_ids"
            placeholder="subnet-12345,subnet-67890"
            onChange={(e) => {
              const subnets = e.target.value.split(',').map(s => s.trim()).filter(Boolean);
              setValue('subnet_ids', subnets);
            }}
          />
          <p className="text-sm text-gray-500">Comma-separated list of subnet IDs</p>
          {errors.subnet_ids && (
            <p className="text-sm text-red-600">{errors.subnet_ids.message}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="vpc_id">VPC ID (Optional)</Label>
          <Input
            id="vpc_id"
            {...register('vpc_id')}
            placeholder="vpc-12345"
          />
        </div>

        <div className="flex justify-end space-x-2">
          <Button
            type="button"
            variant="outline"
            onClick={handleCancel}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={isPending}
          >
            {isPending ? 'Creating...' : 'Create Cluster'}
          </Button>
        </div>
      </form>
    </DialogContent>
  );
}


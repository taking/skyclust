/**
 * Create VM Dialog Component
 * Virtual Machine 생성 폼 다이얼로그
 */

'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { CheckCircle2, XCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { createValidationSchemas } from '@/lib/validations';
import type { CreateVMForm, Credential, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface CreateVMDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateVMForm) => void;
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  isPending?: boolean;
}

export function CreateVMDialog({
  open: _open,
  onOpenChange,
  onSubmit,
  credentials,
  selectedCredentialId,
  onCredentialChange,
  isPending = false,
}: CreateVMDialogProps) {
  const { t } = useTranslation();
  const schemas = createValidationSchemas(t);
  const {
    register,
    handleSubmit,
    formState: { errors, touchedFields },
    reset,
    setValue,
  } = useForm<CreateVMForm>({
    resolver: zodResolver(schemas.createVMSchema),
    mode: 'onChange',
  });

  const handleFormSubmit = (data: CreateVMForm) => {
    onSubmit(data);
    reset();
  };

  const handleCancel = () => {
    reset();
    onOpenChange(false);
  };

  return (
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <DialogTitle>Create New VM</DialogTitle>
        <DialogDescription>
          Create a new virtual machine in your workspace.
        </DialogDescription>
      </DialogHeader>
      <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2 relative">
            <Label htmlFor="name" className="flex items-center gap-1">
              VM Name
              <span className="text-red-500">*</span>
            </Label>
            <div className="relative">
              <Input
                id="name"
                placeholder="Enter VM name"
                {...register('name')}
                className={cn(
                  errors.name && 'border-red-500 focus-visible:ring-red-500',
                  !errors.name && touchedFields.name && 'border-green-500 focus-visible:ring-green-500',
                  'pr-10'
                )}
              />
              {errors.name && (
                <XCircle className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-red-500 pointer-events-none" aria-hidden="true" />
              )}
              {!errors.name && touchedFields.name && (
                <CheckCircle2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-green-500 pointer-events-none" aria-hidden="true" />
              )}
            </div>
            {errors.name && (
              <p className="text-sm text-red-600 flex items-center gap-1">
                <XCircle className="h-3 w-3" aria-hidden="true" />
                {errors.name.message}
              </p>
            )}
            {!errors.name && touchedFields.name && (
              <p className="text-sm text-green-600 flex items-center gap-1">
                <CheckCircle2 className="h-3 w-3" aria-hidden="true" />
                Looks good!
              </p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="credential">Credential (Provider)</Label>
            <Select 
              value={selectedCredentialId}
              onValueChange={(value) => {
                onCredentialChange(value);
                const credential = credentials.find(c => c.id === value);
                if (credential) {
                  setValue('provider', credential.provider as CloudProvider);
                }
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select credential" />
              </SelectTrigger>
              <SelectContent>
                {credentials.map((credential) => (
                  <SelectItem key={credential.id} value={credential.id}>
                    {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.provider && (
              <p className="text-sm text-red-600">{errors.provider.message}</p>
            )}
            {credentials.length === 0 && (
              <p className="text-sm text-yellow-600">No credentials available. Please create a credential first.</p>
            )}
          </div>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="instance_type">Instance Type</Label>
            <Input
              id="instance_type"
              placeholder="e.g., t3.micro"
              {...register('instance_type')}
            />
            {errors.instance_type && (
              <p className="text-sm text-red-600">{errors.instance_type.message}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="region">Region</Label>
            <Input
              id="region"
              placeholder="e.g., us-east-1"
              {...register('region')}
            />
            {errors.region && (
              <p className="text-sm text-red-600">{errors.region.message}</p>
            )}
          </div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="image_id">Image ID</Label>
          <Input
            id="image_id"
            placeholder="e.g., ami-12345678"
            {...register('image_id')}
          />
          {errors.image_id && (
            <p className="text-sm text-red-600">{errors.image_id.message}</p>
          )}
        </div>
        <div className="flex justify-end space-x-2">
          <Button
            type="button"
            variant="outline"
            onClick={handleCancel}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? 'Creating...' : 'Create VM'}
          </Button>
        </div>
      </form>
    </DialogContent>
  );
}


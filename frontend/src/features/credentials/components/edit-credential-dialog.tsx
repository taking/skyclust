/**
 * Edit Credential Dialog Component
 * Credential 수정 다이얼로그 컴포넌트
 */

'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createValidationSchemas } from '@/lib/validations';
import { ProviderFormFields } from './provider-form-fields';
import type { CreateCredentialForm, Credential } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface EditCredentialDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  credential: Credential | null;
  onSubmit: (data: CreateCredentialForm) => void;
  onClose: () => void;
  isPending: boolean;
}

export function EditCredentialDialog({
  open,
  onOpenChange,
  credential,
  onSubmit,
  onClose,
  isPending,
}: EditCredentialDialogProps) {
  const { t } = useTranslation();
  const schemas = createValidationSchemas(t);
  const {
    register,
    handleSubmit,
    reset,
    setValue,
  } = useForm<CreateCredentialForm>({
    resolver: zodResolver(schemas.createCredentialSchema),
  });

  // Initialize form when credential changes
  React.useEffect(() => {
    if (credential) {
      setValue('provider', credential.provider);
    }
  }, [credential, setValue]);

  const handleFormSubmit = handleSubmit((data) => {
    onSubmit(data);
    reset();
    onClose();
  });

  const handleClose = () => {
    onOpenChange(false);
    reset();
    onClose();
  };

  if (!credential) {
    return null;
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit Credentials</DialogTitle>
          <DialogDescription>
            Update your cloud provider credentials.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleFormSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-provider">Provider</Label>
            <Input
              id="edit-provider"
              value={credential.provider || ''}
              disabled
            />
          </div>
          
          <div className="space-y-4">
            <div className="text-sm text-gray-600">
              Update your {credential.provider?.toUpperCase()} credentials:
            </div>
            
            {credential.provider === 'aws' && (
              <ProviderFormFields
                provider="aws"
                register={register}
                setValue={setValue}
              />
            )}
          </div>
          
          <div className="flex justify-end space-x-2">
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Updating...' : 'Update Credentials'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}


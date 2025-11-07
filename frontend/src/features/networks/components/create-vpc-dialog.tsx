/**
 * Create VPC Dialog Component
 * VPC 생성 다이얼로그 컴포넌트
 */

'use client';

import { useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { useToast } from '@/hooks/use-toast';
import { createValidationSchemas } from '@/lib/validations';
import type { CreateVPCForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

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
  const { t } = useTranslation();
  const { success: showSuccess, error: showError } = useToast();
  const schemas = createValidationSchemas(t);

  const {
    form,
    handleSubmit,
    isLoading,
    reset,
    setValue,
  } = useFormWithValidation<CreateVPCForm>({
    schema: schemas.createVPCSchema,
    defaultValues: {
      name: '',
      description: '',
      cidr_block: '',
      region: selectedRegion || '',
    },
    onSubmit: async (data) => {
      await onSubmit(data);
    },
    onSuccess: () => {
      showSuccess(t('messages.created', { resource: t('network.vpcs') }));
      reset();
      onOpenChange(false);
    },
    onError: (error) => {
      showError(error.message || t('messages.operationFailed'));
    },
    resetOnSuccess: true,
  });

  // Reset form when dialog opens/closes or region changes
  useEffect(() => {
    if (open) {
      reset({
        name: '',
        description: '',
        cidr_block: '',
        region: selectedRegion || '',
      });
    }
  }, [open, selectedRegion, reset]);

  // Update region when selectedRegion changes
  useEffect(() => {
    if (selectedRegion) {
      setValue('region', selectedRegion);
    }
  }, [selectedRegion, setValue]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create VPC</DialogTitle>
          <DialogDescription>
            Create a new VPC on {selectedProvider?.toUpperCase()}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <EnhancedField
              name="name"
              label={t('network.vpcs') + ' ' + t('form.validation.nameRequired')}
              placeholder="my-vpc"
              required
            />
            <EnhancedField
              name="description"
              label={t('form.validation.descriptionRequired')}
              placeholder="VPC description"
              type="textarea"
            />
            <div className="grid grid-cols-2 gap-4">
              <EnhancedField
                name="cidr_block"
                label="CIDR Block"
                placeholder="10.0.0.0/16"
              />
              <EnhancedField
                name="region"
                label={t('region.select')}
                placeholder="ap-northeast-2"
              />
            </div>
            <div className="flex justify-end space-x-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading || isPending}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" disabled={isLoading || isPending || disabled}>
                {isLoading || isPending ? t('actions.creating') : t('network.createVPC')}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}


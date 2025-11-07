/**
 * Create Credential Dialog Component
 * Credential 생성 다이얼로그 컴포넌트
 */

'use client';

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createValidationSchemas } from '@/lib/validations';
import { ProviderFormFields } from './provider-form-fields';
import type { CreateCredentialForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface CreateCredentialDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateCredentialForm) => void;
  isPending: boolean;
  gcpInputMode: 'json' | 'file';
  onGcpInputModeChange: (mode: 'json' | 'file') => void;
}

export function CreateCredentialDialog({
  open,
  onOpenChange,
  onSubmit,
  isPending,
  gcpInputMode,
  onGcpInputModeChange,
}: CreateCredentialDialogProps) {
  const { t } = useTranslation();
  const schemas = createValidationSchemas(t);
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
    watch,
  } = useForm<CreateCredentialForm>({
    resolver: zodResolver(schemas.createCredentialSchema),
  });

  const selectedProvider = watch('provider');

  const handleFormSubmit = handleSubmit((data) => {
    onSubmit(data);
    reset();
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t('credential.addNewCredentials')}</DialogTitle>
          <DialogDescription>
            {t('credential.addCredentialsDialogDescription')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleFormSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">{t('common.name')}</Label>
            <Input
              id="name"
              placeholder={t('credential.namePlaceholder')}
              {...register('name')}
            />
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="provider">{t('credential.provider')}</Label>
            <Select onValueChange={(value) => setValue('provider', value)}>
              <SelectTrigger>
                <SelectValue placeholder={t('credential.selectProvider')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="aws">AWS</SelectItem>
                <SelectItem value="gcp">Google Cloud</SelectItem>
                <SelectItem value="azure">Azure</SelectItem>
              </SelectContent>
            </Select>
            {errors.provider && (
              <p className="text-sm text-red-600">{errors.provider.message}</p>
            )}
          </div>
          
          {selectedProvider && (
            <div className="space-y-4">
              <div className="text-sm text-gray-600">
                {t('credential.enterCredentials', { provider: selectedProvider.toUpperCase() })}
              </div>
              
              <ProviderFormFields
                provider={selectedProvider}
                gcpInputMode={gcpInputMode}
                onGcpInputModeChange={onGcpInputModeChange}
                register={register}
                setValue={setValue}
              />
            </div>
          )}
          
          <div className="flex justify-end space-x-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                onOpenChange(false);
                reset();
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? t('credential.adding') : t('credential.add')}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}


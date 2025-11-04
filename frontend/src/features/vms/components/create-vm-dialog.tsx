/**
 * Create VM Dialog Component
 * Virtual Machine 생성 폼 다이얼로그
 * 
 * use-form-with-validation 훅을 사용한 리팩토링 버전
 */

'use client';

import { Form } from '@/components/ui/form';
import { DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { useToast } from '@/hooks/use-toast';
import { createVMSchema } from '@/lib/validations';
import type { CreateVMForm, Credential } from '@/lib/types';

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
  open,
  onOpenChange,
  onSubmit,
  credentials,
  selectedCredentialId,
  onCredentialChange,
  isPending = false,
}: CreateVMDialogProps) {
  const { success: showSuccess, error: showError } = useToast();

  const {
    form,
    handleSubmit,
    isLoading,
    error,
    reset,
    getFieldError,
    getFieldValidationState,
    setValue,
  } = useFormWithValidation<CreateVMForm>({
    schema: createVMSchema,
    defaultValues: {
      name: '',
      provider: '',
      instance_type: '',
      region: '',
      image_id: '',
    },
    onSubmit: async (data) => {
      onSubmit(data);
    },
    onSuccess: () => {
      showSuccess('VM created successfully');
      reset();
      onOpenChange(false);
    },
    onError: (error) => {
      showError(`Failed to create VM: ${error.message}`);
    },
    resetOnSuccess: true,
  });

  const handleCancel = () => {
    reset();
    onOpenChange(false);
  };

  const handleCredentialChange = (value: string) => {
    onCredentialChange(value);
    const credential = credentials.find(c => c.id === value);
    if (credential) {
      setValue('provider', credential.provider);
    }
  };

  return (
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <DialogTitle>Create New VM</DialogTitle>
        <DialogDescription>
          Create a new virtual machine in your workspace.
        </DialogDescription>
      </DialogHeader>
      <Form {...form}>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <EnhancedField
              name="name"
              label="VM Name"
              type="text"
              placeholder="Enter VM name"
              required
              getFieldError={getFieldError}
              getFieldValidationState={getFieldValidationState}
            />
            <div className="space-y-2">
              <label htmlFor="credential" className="flex items-center gap-1">
                Credential (Provider)
              </label>
              <Select 
                value={selectedCredentialId}
                onValueChange={handleCredentialChange}
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
              {getFieldError('provider') && (
                <p className="text-sm text-red-600">{getFieldError('provider')}</p>
              )}
              {credentials.length === 0 && (
                <p className="text-sm text-yellow-600">No credentials available. Please create a credential first.</p>
              )}
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <EnhancedField
              name="instance_type"
              label="Instance Type"
              type="text"
              placeholder="e.g., t3.micro"
              required
              getFieldError={getFieldError}
              getFieldValidationState={getFieldValidationState}
            />
            <EnhancedField
              name="region"
              label="Region"
              type="text"
              placeholder="e.g., us-east-1"
              required
              getFieldError={getFieldError}
              getFieldValidationState={getFieldValidationState}
            />
          </div>
          <EnhancedField
            name="image_id"
            label="Image ID"
            type="text"
            placeholder="e.g., ami-12345678"
            required
            getFieldError={getFieldError}
            getFieldValidationState={getFieldValidationState}
          />
          
          {error && (
            <div className="text-sm text-red-600 text-center" role="alert">
              {error}
            </div>
          )}

          <div className="flex justify-end space-x-2">
            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading || isPending}>
              {isLoading || isPending ? 'Creating...' : 'Create VM'}
            </Button>
          </div>
        </form>
      </Form>
    </DialogContent>
  );
}


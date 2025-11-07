/**
 * Create Cluster Dialog Component
 * Kubernetes 클러스터 생성 폼 다이얼로그
 * 
 * use-form-with-validation 훅을 사용한 리팩토링 버전
 */

'use client';

import { Form } from '@/components/ui/form';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { useToast } from '@/hooks/use-toast';
import { createValidationSchemas } from '@/lib/validations';
import type { CreateClusterForm, Credential, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

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
  open: _open,
  onOpenChange,
  onSubmit,
  credentials,
  selectedCredentialId,
  onCredentialChange,
  selectedProvider,
  isPending = false,
}: CreateClusterDialogProps) {
  const { t } = useTranslation();
  const { success: showSuccess, error: showError } = useToast();
  const schemas = createValidationSchemas(t);

  const {
    form,
    handleSubmit,
    isLoading,
    error,
    reset,
    getFieldError,
    getFieldValidationState,
    setValue,
  } = useFormWithValidation<CreateClusterForm>({
    schema: schemas.createClusterSchema,
    defaultValues: {
      credential_id: selectedCredentialId || '',
      name: '',
      version: '',
      region: '',
      zone: '',
      subnet_ids: [],
      vpc_id: '',
      role_arn: '',
      tags: {},
    },
    onSubmit: async (data) => {
      onSubmit(data);
    },
    onSuccess: () => {
      showSuccess('Cluster created successfully');
      reset();
      onOpenChange(false);
    },
    onError: (error) => {
      showError(`Failed to create cluster: ${error.message}`);
    },
    resetOnSuccess: true,
  });

  const handleCancel = () => {
    reset();
    onOpenChange(false);
  };

  const _handleCredentialChange = (value: string) => {
    onCredentialChange(value);
    setValue('credential_id', value);
  };

  return (
    <Dialog open={_open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Kubernetes Cluster</DialogTitle>
          <DialogDescription>
            Create a new Kubernetes cluster on {selectedProvider?.toUpperCase() || 'your cloud provider'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={handleSubmit} className="space-y-4">
            <EnhancedField
              name="credential_id"
              label="Credential"
              type="select"
              required
              placeholder="Select credential"
              options={credentials.map(cred => ({
                value: cred.id,
                label: `${cred.provider} - ${cred.id.substring(0, 8)}...`,
              }))}
              getFieldError={getFieldError}
              getFieldValidationState={getFieldValidationState}
            />
            
            <div className="grid grid-cols-2 gap-4">
              <EnhancedField
                name="name"
                label="Cluster Name"
                type="text"
                placeholder="Enter cluster name"
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
              <EnhancedField
                name="version"
                label="Kubernetes Version"
                type="text"
                placeholder="e.g., 1.28"
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <EnhancedField
                name="region"
                label="Region"
                type="text"
                placeholder="e.g., us-east-1"
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
              <EnhancedField
                name="zone"
                label="Zone (Optional)"
                type="text"
                placeholder="e.g., us-east-1a"
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
            </div>

            <EnhancedField
              name="vpc_id"
              label="VPC ID (Optional)"
              type="text"
              placeholder="Enter VPC ID"
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
                {isLoading || isPending ? 'Creating...' : 'Create Cluster'}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}


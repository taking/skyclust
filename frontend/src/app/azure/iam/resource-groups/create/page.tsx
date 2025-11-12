/**
 * Create Resource Group Page
 * Azure Resource Group 생성 페이지 (Stepper 방식)
 */

'use client';

import { useEffect, Suspense } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useValidation } from '@/lib/validation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentials } from '@/hooks/use-credentials';
import { useResourceGroupActions, type CreateResourceGroupForm } from '@/features/resource-groups';
import { useStepperForm } from '@/hooks/use-stepper-form';
import { CreateResourceStepperLayout } from '@/components/common/create-resource-stepper-layout';
import { TableSkeleton } from '@/components/ui/table-skeleton';

// Dynamic imports for step components (lazy loading)
const BasicResourceGroupConfigStep = dynamic(
  () => import('@/features/resource-groups/components/create-resource-group/basic-resource-group-config-step').then(mod => ({ default: mod.BasicResourceGroupConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const ReviewResourceGroupStep = dynamic(
  () => import('@/features/resource-groups/components/create-resource-group/review-resource-group-step').then(mod => ({ default: mod.ReviewResourceGroupStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const STEPS = [
  {
    label: 'Basic Configuration',
    description: 'Configure basic Resource Group settings',
  },
  {
    label: 'Review & Create',
    description: 'Review and create Resource Group',
  },
];

function CreateResourceGroupPageContent() {
  const router = useRouter();
  const { handleError } = useErrorHandler();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const { currentWorkspace } = useWorkspaceStore();
  const { schemas } = useValidation();
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    provider: 'azure',
    updateUrl: true,
  });

  const { credentials } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  const selectedCredential = credentials.find(c => c.id === selectedCredentialId);
  const selectedProvider = selectedCredential?.provider;

  const defaultFormValues: Partial<CreateResourceGroupForm> = {
    credential_id: selectedCredentialId || '',
    name: '',
    location: selectedRegion || '',
    tags: {},
  };

  const { createResourceGroupMutation } = useResourceGroupActions({
    selectedCredentialId: selectedCredentialId || '',
    onSuccess: () => {
      router.push('/azure/iam/resource-groups');
    },
  });

  // Stepper Form 훅 사용
  const {
    currentStep,
    form,
    formData,
    updateFormData,
    handleNext,
    handlePrevious,
  } = useStepperForm<CreateResourceGroupForm>({
    totalSteps: STEPS.length,
    schema: schemas.createResourceGroupSchema,
    defaultValues: defaultFormValues,
    formOptions: {
      mode: 'onChange',
    },
  });

  // Update form when credential/region changes
  useEffect(() => {
    if (selectedCredentialId) {
      form.setValue('credential_id', selectedCredentialId);
      updateFormData({ credential_id: selectedCredentialId });
    }
    if (selectedRegion) {
      form.setValue('location', selectedRegion);
      updateFormData({ location: selectedRegion });
    }
  }, [selectedCredentialId, selectedRegion, form, updateFormData]);

  const handleCreateResourceGroup = async () => {
    if (selectedProvider !== 'azure') {
      handleError(new Error('Azure provider is required'), { operation: 'createResourceGroup' });
      return;
    }

    const isValid = await form.trigger();
    if (!isValid) {
      return;
    }

    const data = formData as CreateResourceGroupForm;
    createResourceGroupMutation.mutate(data);
  };

  const handleCancel = () => {
    router.push('/azure/iam/resource-groups');
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <BasicResourceGroupConfigStep
            form={form}
            onDataChange={updateFormData}
          />
        );
      case 2:
        return (
          <ReviewResourceGroupStep
            formData={formData as CreateResourceGroupForm}
          />
        );
      default:
        return null;
    }
  };

  return (
    <CreateResourceStepperLayout
      backPath="/azure/iam/resource-groups"
      title="Create Resource Group"
      description="Create a new Azure Resource Group"
      steps={STEPS}
      currentStep={currentStep}
      renderStepContent={renderStepContent}
      navigationProps={{
        onPrevious: handlePrevious,
        onNext: handleNext,
        onCancel: handleCancel,
        onSubmit: handleCreateResourceGroup,
        isLoading: createResourceGroupMutation.isPending,
        submitButtonText: 'Create Resource Group',
        submittingButtonText: 'Creating...',
      }}
      onCancel={handleCancel}
    />
  );
}

export default function CreateResourceGroupPage() {
  return (
    <Suspense fallback={<TableSkeleton columns={2} rows={5} />}>
      <CreateResourceGroupPageContent />
    </Suspense>
  );
}


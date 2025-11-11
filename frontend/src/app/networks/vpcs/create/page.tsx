/**
 * Create VPC Page
 * VPC 생성 페이지 (Stepper 방식)
 */

'use client';

import { useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import dynamic from 'next/dynamic';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useTranslation } from '@/hooks/use-translation';
import { useValidation } from '@/lib/validation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentials } from '@/hooks/use-credentials';
import type { CreateVPCForm, CloudProvider } from '@/lib/types';
import { useVPCActions } from '@/features/networks/hooks/use-vpc-actions';
import { useStepperForm } from '@/hooks/use-stepper-form';
import { CreateResourceStepperLayout } from '@/components/common/create-resource-stepper-layout';
import { UI_MESSAGES } from '@/lib/constants';
import { TableSkeleton } from '@/components/ui/table-skeleton';

// Dynamic imports for step components (lazy loading)
const BasicVPCConfigStep = dynamic(
  () => import('@/features/networks/components/create-vpc/basic-vpc-config-step').then(mod => ({ default: mod.BasicVPCConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const AdvancedVPCConfigStep = dynamic(
  () => import('@/features/networks/components/create-vpc/advanced-vpc-config-step').then(mod => ({ default: mod.AdvancedVPCConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const ReviewVPCStep = dynamic(
  () => import('@/features/networks/components/create-vpc/review-vpc-step').then(mod => ({ default: mod.ReviewVPCStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const STEPS = [
  {
    label: 'Basic Configuration',
    description: 'Configure basic VPC settings',
  },
  {
    label: 'Advanced Configuration',
    description: 'Optional advanced settings',
  },
  {
    label: 'Review & Create',
    description: 'Review and create VPC',
  },
];

function CreateVPCPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { t } = useTranslation();
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const { currentWorkspace } = useWorkspaceStore();
  const { schemas } = useValidation();
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'network',
    updateUrl: true,
  });

  const { credentials } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  const selectedCredential = credentials.find(c => c.id === selectedCredentialId);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  const defaultFormValues: Partial<CreateVPCForm> = {
    credential_id: selectedCredentialId || '',
    name: '',
    description: '',
    cidr_block: '',
    region: selectedRegion || '',
    tags: {},
  };

  const { createVPCMutation } = useVPCActions({
    selectedProvider,
    selectedCredentialId,
    selectedRegion: selectedRegion || '',
  });

  // Stepper Form 훅 사용
  const {
    currentStep,
    form,
    formData,
    updateFormData,
    handleNext,
    handlePrevious,
    handleSkipAdvanced,
    isAdvancedStep,
    isLastStep,
  } = useStepperForm<CreateVPCForm>({
    totalSteps: STEPS.length,
    schema: schemas.createVPCSchema,
    defaultValues: defaultFormValues,
    advancedStepNumber: 2, // Advanced Configuration step
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
      form.setValue('region', selectedRegion);
      const updateData: Partial<CreateVPCForm> = { region: selectedRegion };
      if (selectedProvider === 'azure') {
        form.setValue('location', selectedRegion);
        updateData.location = selectedRegion;
      }
      updateFormData(updateData);
    }
  }, [selectedCredentialId, selectedRegion, selectedProvider, form, updateFormData]);

  const handleCreateVPC = async () => {
    if (!selectedProvider) {
      handleError(new Error('Provider not selected'), { operation: 'createVPC' });
      return;
    }

    const validatedData = await form.trigger();
    if (!validatedData) {
      handleError(new Error('Please fix validation errors'), { operation: 'createVPC' });
      return;
    }

    const finalData = form.getValues();
    
    // Provider별 필드 정리
    const vpcData: CreateVPCForm = {
      ...finalData,
      credential_id: finalData.credential_id || selectedCredentialId,
      region: selectedRegion || finalData.region,
    };

    if (selectedProvider === 'azure') {
      vpcData.location = selectedRegion || finalData.region;
      delete vpcData.cidr_block;
    } else if (selectedProvider === 'aws' && !vpcData.cidr_block) {
      handleError(new Error('CIDR block is required for AWS VPC'), { operation: 'createVPC' });
      return;
    }

    try {
      await createVPCMutation.mutateAsync(vpcData);
      success('VPC creation initiated');
      router.push('/networks/vpcs');
    } catch (error) {
      handleError(error, { operation: 'createVPC', resource: 'VPC' });
    }
  };

  const handleCancel = () => {
    if (confirm(UI_MESSAGES.CONFIRM_CANCEL)) {
      router.push('/networks/vpcs');
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <BasicVPCConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={updateFormData}
          />
        );
      case 2:
        return (
          <AdvancedVPCConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={updateFormData}
          />
        );
      case 3:
        return (
          <ReviewVPCStep
            formData={formData as CreateVPCForm}
            selectedProvider={selectedProvider}
          />
        );
      default:
        return null;
    }
  };

  return (
    <CreateResourceStepperLayout
      backPath="/networks/vpcs"
      title="network.createVPCTitle"
      description="network.createVPCDescriptionNoRegion"
      descriptionParams={{ provider: selectedProvider?.toUpperCase() || '' }}
      steps={STEPS}
      currentStep={currentStep}
      renderStepContent={renderStepContent}
      navigationProps={{
        onPrevious: handlePrevious,
        onNext: handleNext,
        onCancel: handleCancel,
        onSkipAdvanced: handleSkipAdvanced,
        onSubmit: handleCreateVPC,
        isLoading: createVPCMutation.isPending,
        submitButtonText: 'network.createVPC',
        submittingButtonText: 'actions.creating',
      }}
      onCancel={handleCancel}
      advancedStepNumber={2}
    />
  );
}

export default function CreateVPCPage() {
  return (
    <Suspense fallback={<div>{UI_MESSAGES.LOADING}</div>}>
      <CreateVPCPageContent />
    </Suspense>
  );
}


/**
 * Create Subnet Page
 * Subnet 생성 페이지 (Stepper 방식)
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
import type { CreateSubnetForm, CloudProvider } from '@/lib/types';
import { useSubnetActions } from '@/features/networks/hooks/use-subnet-actions';
import { useStepperForm } from '@/hooks/use-stepper-form';
import { CreateResourceStepperLayout } from '@/components/common/create-resource-stepper-layout';
import { UI_MESSAGES } from '@/lib/constants';
import { TableSkeleton } from '@/components/ui/table-skeleton';

// Dynamic imports for step components (lazy loading)
const BasicSubnetConfigStep = dynamic(
  () => import('@/features/networks/components/create-subnet/basic-subnet-config-step').then(mod => ({ default: mod.BasicSubnetConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const AdvancedSubnetConfigStep = dynamic(
  () => import('@/features/networks/components/create-subnet/advanced-subnet-config-step').then(mod => ({ default: mod.AdvancedSubnetConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const ReviewSubnetStep = dynamic(
  () => import('@/features/networks/components/create-subnet/review-subnet-step').then(mod => ({ default: mod.ReviewSubnetStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const STEPS = [
  {
    label: 'network.Basic Configuration',
    description: 'network.Configure basic subnet settings',
  },
  {
    label: 'network.Advanced Configuration',
    description: 'network.Optional advanced settings',
  },
  {
    label: 'network.Review & Create',
    description: 'network.Review and create subnet',
  },
];

function CreateSubnetPageContent() {
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

  // Get VPC ID from URL params if available
  const vpcIdFromUrl = searchParams?.get('vpc_id') || '';

  const defaultFormValues: Partial<CreateSubnetForm> = {
    credential_id: selectedCredentialId || '',
    name: '',
    vpc_id: vpcIdFromUrl,
    cidr_block: '',
    region: selectedRegion || '',
    availability_zone: '',
    tags: {},
  };

  const { createSubnetMutation } = useSubnetActions({
    selectedProvider,
    selectedCredentialId,
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
  } = useStepperForm<CreateSubnetForm>({
    totalSteps: STEPS.length,
    schema: schemas.createSubnetSchema,
    defaultValues: defaultFormValues,
    advancedStepNumber: 2, // Advanced Configuration step
    formOptions: {
      mode: 'onChange',
    },
  });

  // Update form when credential/region/vpc changes
  useEffect(() => {
    if (selectedCredentialId) {
      form.setValue('credential_id', selectedCredentialId);
      updateFormData({ credential_id: selectedCredentialId });
    }
    if (selectedRegion) {
      form.setValue('region', selectedRegion);
      updateFormData({ region: selectedRegion });
    }
    if (vpcIdFromUrl) {
      form.setValue('vpc_id', vpcIdFromUrl);
      updateFormData({ vpc_id: vpcIdFromUrl });
    }
  }, [selectedCredentialId, selectedRegion, vpcIdFromUrl, form, updateFormData]);

  const handleCreateSubnet = async () => {
    if (!selectedProvider) {
      handleError(new Error('Provider not selected'), { operation: 'createSubnet' });
      return;
    }

    const validatedData = await form.trigger();
    if (!validatedData) {
      handleError(new Error('Please fix validation errors'), { operation: 'createSubnet' });
      return;
    }

    const finalData = form.getValues();
    
    // Provider별 필드 설정
    const subnetData: CreateSubnetForm = {
      ...finalData,
      credential_id: finalData.credential_id || selectedCredentialId,
      region: selectedRegion || finalData.region,
    };

    if (selectedProvider === 'gcp') {
      subnetData.zone = finalData.zone || finalData.availability_zone;
      delete subnetData.availability_zone;
    } else {
      subnetData.availability_zone = finalData.availability_zone || finalData.zone;
      delete subnetData.zone;
    }

    if (!subnetData.vpc_id) {
      handleError(new Error('VPC ID is required'), { operation: 'createSubnet' });
      return;
    }

    try {
      const result = await createSubnetMutation.mutateAsync(subnetData);
      success('Subnet creation initiated');
      
      // Subnet 생성 후 리스트 페이지로 이동할 때 VPC ID를 URL 파라미터로 전달
      const params = new URLSearchParams();
      if (subnetData.vpc_id) {
        params.set('vpc_id', subnetData.vpc_id);
      }
      router.push(`/networks/subnets${params.toString() ? `?${params.toString()}` : ''}`);
    } catch (error) {
      handleError(error, { operation: 'createSubnet', resource: 'Subnet' });
    }
  };

  const handleCancel = () => {
    if (confirm(UI_MESSAGES.CONFIRM_CANCEL)) {
      router.push('/networks/subnets');
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <BasicSubnetConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={updateFormData}
          />
        );
      case 2:
        return (
          <AdvancedSubnetConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={updateFormData}
          />
        );
      case 3:
        return (
          <ReviewSubnetStep
            formData={formData as CreateSubnetForm}
            selectedProvider={selectedProvider}
          />
        );
      default:
        return null;
    }
  };

  return (
    <CreateResourceStepperLayout
      backPath="/networks/subnets"
      title="network.createSubnetTitle"
      description="network.createSubnetDescriptionNoRegion"
      descriptionParams={{ provider: selectedProvider?.toUpperCase() || '' }}
      steps={STEPS}
      currentStep={currentStep}
      renderStepContent={renderStepContent}
      navigationProps={{
        onPrevious: handlePrevious,
        onNext: handleNext,
        onCancel: handleCancel,
        onSkipAdvanced: handleSkipAdvanced,
        onSubmit: handleCreateSubnet,
        isLoading: createSubnetMutation.isPending,
        submitButtonText: 'network.createSubnet',
        submittingButtonText: 'actions.creating',
      }}
      onCancel={handleCancel}
      advancedStepNumber={2}
    />
  );
}

export default function CreateSubnetPage() {
  return (
    <Suspense fallback={<div>{UI_MESSAGES.LOADING}</div>}>
      <CreateSubnetPageContent />
    </Suspense>
  );
}


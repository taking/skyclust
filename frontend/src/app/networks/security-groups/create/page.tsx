/**
 * Create Security Group Page
 * Security Group 생성 페이지 (Stepper 방식)
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
import type { CreateSecurityGroupForm, CloudProvider } from '@/lib/types';
import { useSecurityGroupActions } from '@/features/networks/hooks/use-security-group-actions';
import { useStepperForm } from '@/hooks/use-stepper-form';
import { CreateResourceStepperLayout } from '@/components/common/create-resource-stepper-layout';
import { UI_MESSAGES } from '@/lib/constants';
import { TableSkeleton } from '@/components/ui/table-skeleton';

// Dynamic imports for step components (lazy loading)
const BasicSecurityGroupConfigStep = dynamic(
  () => import('@/features/networks/components/create-security-group/basic-security-group-config-step').then(mod => ({ default: mod.BasicSecurityGroupConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const AdvancedSecurityGroupConfigStep = dynamic(
  () => import('@/features/networks/components/create-security-group/advanced-security-group-config-step').then(mod => ({ default: mod.AdvancedSecurityGroupConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const ReviewSecurityGroupStep = dynamic(
  () => import('@/features/networks/components/create-security-group/review-security-group-step').then(mod => ({ default: mod.ReviewSecurityGroupStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const STEPS = [
  {
    label: 'network.Basic Configuration',
    description: 'network.Configure basic security group settings',
  },
  {
    label: 'network.Advanced Configuration',
    description: 'network.Optional advanced settings',
  },
  {
    label: 'network.Review & Create',
    description: 'network.Review and create security group',
  },
];

function CreateSecurityGroupPageContent() {
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

  const defaultFormValues: Partial<CreateSecurityGroupForm> = {
    credential_id: selectedCredentialId || '',
    name: '',
    description: '',
    vpc_id: vpcIdFromUrl,
    region: selectedRegion || '',
    tags: {},
  };

  const { createSecurityGroupMutation } = useSecurityGroupActions({
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
  } = useStepperForm<CreateSecurityGroupForm>({
    totalSteps: STEPS.length,
    schema: schemas.createSecurityGroupSchema,
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

  const handleCreateSecurityGroup = async () => {
    if (!selectedProvider) {
      handleError(new Error('Provider not selected'), { operation: 'createSecurityGroup' });
      return;
    }

    const validatedData = await form.trigger();
    if (!validatedData) {
      handleError(new Error('Please fix validation errors'), { operation: 'createSecurityGroup' });
      return;
    }

    const finalData = form.getValues();
    
    // Provider별 필드 설정
    const securityGroupData: CreateSecurityGroupForm = {
      ...finalData,
      credential_id: finalData.credential_id || selectedCredentialId,
      region: selectedRegion || finalData.region,
    };

    if (!securityGroupData.vpc_id) {
      handleError(new Error('VPC ID is required'), { operation: 'createSecurityGroup' });
      return;
    }

    try {
      await createSecurityGroupMutation.mutateAsync(securityGroupData);
      success('Security group creation initiated');
      
      // Security Group 생성 후 리스트 페이지로 이동할 때 VPC ID를 URL 파라미터로 전달
      const params = new URLSearchParams();
      if (securityGroupData.vpc_id) {
        params.set('vpc_id', securityGroupData.vpc_id);
      }
      router.push(`/networks/security-groups${params.toString() ? `?${params.toString()}` : ''}`);
    } catch (error) {
      handleError(error, { operation: 'createSecurityGroup', resource: 'SecurityGroup' });
    }
  };

  const handleCancel = () => {
    if (confirm(UI_MESSAGES.CONFIRM_CANCEL)) {
      router.push('/networks/security-groups');
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <BasicSecurityGroupConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={updateFormData}
          />
        );
      case 2:
        return (
          <AdvancedSecurityGroupConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={updateFormData}
          />
        );
      case 3:
        return (
          <ReviewSecurityGroupStep
            formData={formData as CreateSecurityGroupForm}
            selectedProvider={selectedProvider}
          />
        );
      default:
        return null;
    }
  };

  return (
    <CreateResourceStepperLayout
      backPath="/networks/security-groups"
      title="network.createSecurityGroup"
      description="network.createSecurityGroupDescription"
      descriptionParams={{ provider: selectedProvider?.toUpperCase() || '' }}
      steps={STEPS}
      currentStep={currentStep}
      renderStepContent={renderStepContent}
      navigationProps={{
        onPrevious: handlePrevious,
        onNext: handleNext,
        onCancel: handleCancel,
        onSkipAdvanced: handleSkipAdvanced,
        onSubmit: handleCreateSecurityGroup,
        isLoading: createSecurityGroupMutation.isPending,
        submitButtonText: 'network.createSecurityGroup',
        submittingButtonText: 'actions.creating',
      }}
      onCancel={handleCancel}
      advancedStepNumber={2}
    />
  );
}

export default function CreateSecurityGroupPage() {
  return (
    <Suspense fallback={<div>{UI_MESSAGES.LOADING}</div>}>
      <CreateSecurityGroupPageContent />
    </Suspense>
  );
}


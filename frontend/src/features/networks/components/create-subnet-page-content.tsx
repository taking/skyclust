/**
 * Create Subnet Page Content Component
 * Subnet 생성 페이지 컴포넌트 (Stepper 형태)
 * 
 * 기존 Step 컴포넌트를 재사용하여 Stepper 형태로 구현
 */

'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { useValidation } from '@/lib/validation';
import { useTranslation } from '@/hooks/use-translation';
import { useSubnetActions } from '@/features/networks/hooks/use-subnet-actions';
import { useCredentials } from '@/hooks/use-credentials';
import { buildCredentialResourcePath } from '@/lib/routing/helpers';
import { CreateResourceStepperLayout, type StepConfig } from '@/components/common/create-resource-stepper-layout';
import { BasicSubnetConfigStep } from './create-subnet/basic-subnet-config-step';
import { AdvancedSubnetConfigStep } from './create-subnet/advanced-subnet-config-step';
import { ReviewSubnetStep } from './create-subnet/review-subnet-step';
import type { CreateSubnetForm, CloudProvider } from '@/lib/types';

interface CreateSubnetPageContentProps {
  workspaceId: string;
  credentialId: string;
  region?: string | null;
  onCancel?: () => void;
}

export function CreateSubnetPageContent({
  workspaceId,
  credentialId,
  region,
  onCancel,
}: CreateSubnetPageContentProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const { success, error: showError } = useToast();
  const { schemas } = useValidation();

  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<Partial<CreateSubnetForm>>({});

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId,
    selectedCredentialId: credentialId,
    enabled: !!workspaceId && !!credentialId,
  });

  const { createSubnetMutation } = useSubnetActions({
    selectedProvider,
    selectedCredentialId: credentialId,
    onSuccess: () => {
      const path = buildCredentialResourcePath(
        workspaceId,
        credentialId,
        'networks',
        '/subnets',
        { region: region || undefined }
      );
      router.push(path);
    },
  });

  const form = useForm<CreateSubnetForm>({
    resolver: zodResolver(schemas.createSubnetSchema),
    defaultValues: {
      name: '',
      cidr_block: '',
      region: region || '',
      availability_zone: '',
      vpc_id: '',
      credential_id: credentialId,
      description: '',
      tags: {},
    },
    mode: 'onChange',
  });

  useEffect(() => {
    if (region) {
      form.setValue('region', region);
      setFormData(prev => ({ ...prev, region }));
    }
  }, [region, form]);

  useEffect(() => {
    if (credentialId) {
      form.setValue('credential_id', credentialId);
      setFormData(prev => ({ ...prev, credential_id: credentialId }));
    }
  }, [credentialId, form]);

  const handleDataChange = useCallback((data: Partial<CreateSubnetForm>) => {
    setFormData(prev => ({ ...prev, ...data }));
    Object.entries(data).forEach(([key, value]) => {
      form.setValue(key as keyof CreateSubnetForm, value as never);
    });
  }, [form]);

  const canProceedToNextStep = useCallback((): boolean => {
    const values = form.getValues();
    
    switch (currentStep) {
      case 1:
        return !!(values.name && values.vpc_id && values.cidr_block && values.region);
      case 2:
        return true;
      case 3:
        return true;
      default:
        return false;
    }
  }, [currentStep, form]);

  const handleNext = useCallback(() => {
    if (!canProceedToNextStep()) {
      form.trigger();
      return;
    }
    
    if (currentStep < 3) {
      setCurrentStep(currentStep + 1);
    }
  }, [currentStep, canProceedToNextStep, form]);

  const handlePrevious = useCallback(() => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
    }
  }, [currentStep]);

  const handleSkipAdvanced = useCallback(() => {
    setCurrentStep(3);
  }, []);

  const handleSubmit = useCallback(async () => {
    const isValid = await form.trigger();
    if (!isValid) {
      return;
    }

    const data = form.getValues();
    
    if (!selectedProvider) {
      showError(t('network.providerNotSelected') || 'Provider not selected');
      return;
    }

    const subnetData: CreateSubnetForm = {
      ...data,
      credential_id: data.credential_id || credentialId,
      region: region || data.region || '',
    };

    try {
      await createSubnetMutation.mutateAsync(subnetData);
      success(t('network.subnetCreated') || 'Subnet creation initiated successfully');
    } catch (error) {
      ErrorHandler.logError(error, { operation: 'createSubnet', source: 'create-subnet-page' });
      showError(t('network.subnetCreateFailed') || 'Failed to create subnet');
    }
  }, [form, selectedProvider, createSubnetMutation, success, showError, t, credentialId, region]);

  const handleCancel = useCallback(() => {
    if (onCancel) {
      onCancel();
    } else {
      const path = buildCredentialResourcePath(
        workspaceId,
        credentialId,
        'networks',
        '/subnets',
        { region: region || undefined }
      );
      router.push(path);
    }
  }, [onCancel, workspaceId, credentialId, region, router]);

  const steps: StepConfig[] = [
    {
      label: 'network.basicConfig',
      description: 'network.basicConfigDescription',
    },
    {
      label: 'network.advancedConfig',
      description: 'network.advancedConfigDescription',
    },
    {
      label: 'common.review',
      description: 'common.reviewDescription',
    },
  ];

  const renderStepContent = () => {
    const formValues = form.getValues();

    switch (currentStep) {
      case 1:
        return (
          <BasicSubnetConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={handleDataChange}
          />
        );
      case 2:
        return (
          <AdvancedSubnetConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={handleDataChange}
          />
        );
      case 3:
        return (
          <ReviewSubnetStep
            formData={formValues}
            selectedProvider={selectedProvider}
          />
        );
      default:
        return null;
    }
  };

  const backPath = buildCredentialResourcePath(
    workspaceId,
    credentialId,
    'networks',
    '/subnets',
    { region: region || undefined }
  );

  return (
    <CreateResourceStepperLayout
      backPath={backPath}
      title="network.createSubnet"
      description="network.createSubnetDescription"
      descriptionParams={selectedProvider ? { provider: selectedProvider.toUpperCase() } : undefined}
      steps={steps}
      currentStep={currentStep}
      renderStepContent={renderStepContent}
      navigationProps={{
        onNext: handleNext,
        onPrevious: handlePrevious,
        onCancel: handleCancel,
        onSkipAdvanced: handleSkipAdvanced,
        onSubmit: handleSubmit,
        isLoading: createSubnetMutation.isPending,
        submitButtonText: 'common.create',
        submittingButtonText: 'actions.creating',
      }}
      onCancel={handleCancel}
      advancedStepNumber={2}
    />
  );
}


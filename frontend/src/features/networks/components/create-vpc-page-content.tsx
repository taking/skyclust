/**
 * Create VPC Page Content Component
 * VPC 생성 페이지 컴포넌트 (Stepper 형태)
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
import { useVPCActions } from '@/features/networks/hooks/use-vpc-actions';
import { useCredentials } from '@/hooks/use-credentials';
import { buildCredentialResourcePath } from '@/lib/routing/helpers';
import { CreateResourceStepperLayout, type StepConfig } from '@/components/common/create-resource-stepper-layout';
import { BasicVPCConfigStep } from './create-vpc/basic-vpc-config-step';
import { AdvancedVPCConfigStep } from './create-vpc/advanced-vpc-config-step';
import { ReviewVPCStep } from './create-vpc/review-vpc-step';
import type { CreateVPCForm, CloudProvider } from '@/lib/types';

interface CreateVPCPageContentProps {
  workspaceId: string;
  credentialId: string;
  region?: string | null;
  onCancel?: () => void;
}

export function CreateVPCPageContent({
  workspaceId,
  credentialId,
  region,
  onCancel,
}: CreateVPCPageContentProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const { success, error: showError } = useToast();
  const { schemas } = useValidation();

  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<Partial<CreateVPCForm>>({});

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId,
    selectedCredentialId: credentialId,
    enabled: !!workspaceId && !!credentialId,
  });

  const { createVPCMutation } = useVPCActions({
    selectedProvider,
    selectedCredentialId: credentialId,
    selectedRegion: region || '',
    onSuccess: () => {
      const path = buildCredentialResourcePath(
        workspaceId,
        credentialId,
        'networks',
        '/vpcs',
        { region: region || undefined }
      );
      router.push(path);
    },
  });

  const form = useForm<CreateVPCForm>({
    resolver: zodResolver(schemas.createVPCSchema),
    defaultValues: {
      name: '',
      description: '',
      cidr_block: '',
      region: region || '',
      credential_id: credentialId,
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

  const handleDataChange = useCallback((data: Partial<CreateVPCForm>) => {
    setFormData(prev => ({ ...prev, ...data }));
    Object.entries(data).forEach(([key, value]) => {
      form.setValue(key as keyof CreateVPCForm, value as never);
    });
  }, [form]);

  const canProceedToNextStep = useCallback((): boolean => {
    const values = form.getValues();
    
    switch (currentStep) {
      case 1:
        if (!values.name) return false;
        if (selectedProvider === 'aws' && !values.cidr_block) return false;
        if (selectedProvider === 'azure') {
          if (!values.location || !values.resource_group || !values.address_space || values.address_space.length === 0) {
            return false;
          }
        }
        return true;
      case 2:
        return true;
      case 3:
        return true;
      default:
        return false;
    }
  }, [currentStep, form, selectedProvider]);

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

    const vpcData: CreateVPCForm = {
      ...data,
      credential_id: data.credential_id || credentialId,
      region: region || data.region || '',
    };

    if (selectedProvider === 'azure') {
      vpcData.location = region || data.location || '';
      delete vpcData.cidr_block;
    } else if (selectedProvider === 'aws') {
      if (!vpcData.cidr_block) {
        showError(t('network.cidrBlockRequired') || 'CIDR block is required for AWS VPC');
        return;
      }
    }

    try {
      await createVPCMutation.mutateAsync(vpcData);
      success(t('network.vpcCreated') || 'VPC creation initiated successfully');
    } catch (error) {
      ErrorHandler.logError(error, { operation: 'createVPC', source: 'create-vpc-page' });
      showError(t('network.vpcCreateFailed') || 'Failed to create VPC');
    }
  }, [form, selectedProvider, createVPCMutation, success, showError, t, credentialId, region]);

  const handleCancel = useCallback(() => {
    if (onCancel) {
      onCancel();
    } else {
      const path = buildCredentialResourcePath(
        workspaceId,
        credentialId,
        'networks',
        '/vpcs',
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
          <BasicVPCConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={handleDataChange}
          />
        );
      case 2:
        return (
          <AdvancedVPCConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={handleDataChange}
          />
        );
      case 3:
        return (
          <ReviewVPCStep
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
    '/vpcs',
    { region: region || undefined }
  );

  return (
    <CreateResourceStepperLayout
      backPath={backPath}
      title="network.createVPC"
      description="network.createVPCDescription"
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
        isLoading: createVPCMutation.isPending,
        submitButtonText: 'common.create',
        submittingButtonText: 'actions.creating',
      }}
      onCancel={handleCancel}
      advancedStepNumber={2}
    />
  );
}


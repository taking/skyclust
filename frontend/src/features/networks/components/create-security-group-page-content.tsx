/**
 * Create Security Group Page Content Component
 * Security Group 생성 페이지 컴포넌트 (Stepper 형태)
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
import { useSecurityGroupActions } from '@/features/networks/hooks/use-security-group-actions';
import { useCredentials } from '@/hooks/use-credentials';
import { buildCredentialResourcePath } from '@/lib/routing/helpers';
import { CreateResourceStepperLayout, type StepConfig } from '@/components/common/create-resource-stepper-layout';
import { BasicSecurityGroupConfigStep } from './create-security-group/basic-security-group-config-step';
import { AdvancedSecurityGroupConfigStep } from './create-security-group/advanced-security-group-config-step';
import { ReviewSecurityGroupStep } from './create-security-group/review-security-group-step';
import type { CreateSecurityGroupForm, CloudProvider } from '@/lib/types';

interface CreateSecurityGroupPageContentProps {
  workspaceId: string;
  credentialId: string;
  region?: string | null;
  onCancel?: () => void;
}

export function CreateSecurityGroupPageContent({
  workspaceId,
  credentialId,
  region,
  onCancel,
}: CreateSecurityGroupPageContentProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const { success, error: showError } = useToast();
  const { schemas } = useValidation();

  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<Partial<CreateSecurityGroupForm>>({});

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId,
    selectedCredentialId: credentialId,
    enabled: !!workspaceId && !!credentialId,
  });

  const { createSecurityGroupMutation } = useSecurityGroupActions({
    selectedProvider,
    selectedCredentialId: credentialId,
    onSuccess: () => {
      const path = buildCredentialResourcePath(
        workspaceId,
        credentialId,
        'networks',
        '/security-groups',
        { region: region || undefined }
      );
      router.push(path);
    },
  });

  const form = useForm<CreateSecurityGroupForm>({
    resolver: zodResolver(schemas.createSecurityGroupSchema),
    defaultValues: {
      name: '',
      description: '',
      vpc_id: '',
      region: region || '',
      credential_id: credentialId,
      rules: [],
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

  const handleDataChange = useCallback((data: Partial<CreateSecurityGroupForm>) => {
    setFormData(prev => ({ ...prev, ...data }));
    Object.entries(data).forEach(([key, value]) => {
      form.setValue(key as keyof CreateSecurityGroupForm, value as never);
    });
  }, [form]);

  const canProceedToNextStep = useCallback((): boolean => {
    const values = form.getValues();
    
    switch (currentStep) {
      case 1:
        return !!(values.name && values.vpc_id && values.region);
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

    const securityGroupData: CreateSecurityGroupForm = {
      ...data,
      credential_id: data.credential_id || credentialId,
      region: region || data.region || '',
    };

    try {
      await createSecurityGroupMutation.mutateAsync(securityGroupData);
      success(t('network.securityGroupCreated') || 'Security Group creation initiated successfully');
    } catch (error) {
      ErrorHandler.logError(error, { operation: 'createSecurityGroup', source: 'create-security-group-page' });
      showError(t('network.securityGroupCreateFailed') || 'Failed to create security group');
    }
  }, [form, selectedProvider, createSecurityGroupMutation, success, showError, t, credentialId, region]);

  const handleCancel = useCallback(() => {
    if (onCancel) {
      onCancel();
    } else {
      const path = buildCredentialResourcePath(
        workspaceId,
        credentialId,
        'networks',
        '/security-groups',
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
          <BasicSecurityGroupConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={handleDataChange}
          />
        );
      case 2:
        return (
          <AdvancedSecurityGroupConfigStep
            form={form}
            selectedProvider={selectedProvider}
            onDataChange={handleDataChange}
          />
        );
      case 3:
        return (
          <ReviewSecurityGroupStep
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
    '/security-groups',
    { region: region || undefined }
  );

  return (
    <CreateResourceStepperLayout
      backPath={backPath}
      title="network.createSecurityGroup"
      description="network.createSecurityGroupDescription"
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
        isLoading: createSecurityGroupMutation.isPending,
        submitButtonText: 'common.create',
        submittingButtonText: 'actions.creating',
      }}
      onCancel={handleCancel}
      advancedStepNumber={2}
    />
  );
}


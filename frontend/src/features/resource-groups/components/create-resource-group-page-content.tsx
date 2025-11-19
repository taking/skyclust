/**
 * Create Resource Group Page Content Component
 * Azure Resource Group 생성 페이지 컴포넌트 (Stepper 형태)
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
import { useResourceGroupActions, type CreateResourceGroupForm } from '@/features/resource-groups/hooks/use-resource-group-actions';
import { useCredentials } from '@/hooks/use-credentials';
import { buildCredentialResourcePath } from '@/lib/routing/helpers';
import { CreateResourceStepperLayout, type StepConfig } from '@/components/common/create-resource-stepper-layout';
import { BasicResourceGroupConfigStep } from './create-resource-group/basic-resource-group-config-step';
import { ReviewResourceGroupStep } from './create-resource-group/review-resource-group-step';

interface CreateResourceGroupPageContentProps {
  workspaceId: string;
  credentialId: string;
  onCancel?: () => void;
}

export function CreateResourceGroupPageContent({
  workspaceId,
  credentialId,
  onCancel,
}: CreateResourceGroupPageContentProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const { success, error: showError } = useToast();
  const { schemas } = useValidation();

  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<Partial<CreateResourceGroupForm>>({});

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId,
    selectedCredentialId: credentialId,
    enabled: !!workspaceId && !!credentialId,
  });

  const { createResourceGroupMutation } = useResourceGroupActions({
    selectedCredentialId: credentialId,
    onSuccess: () => {
      const path = buildCredentialResourcePath(
        workspaceId,
        credentialId,
        'azure',
        '/iam/resource-groups'
      );
      router.push(path);
    },
  });

  const form = useForm<CreateResourceGroupForm>({
    resolver: zodResolver(schemas.createResourceGroupSchema),
    defaultValues: {
      name: '',
      location: '',
      credential_id: credentialId,
      tags: {},
    },
    mode: 'onChange',
  });

  useEffect(() => {
    if (credentialId) {
      form.setValue('credential_id', credentialId);
      setFormData(prev => ({ ...prev, credential_id: credentialId }));
    }
  }, [credentialId, form]);

  const handleDataChange = useCallback((data: Partial<CreateResourceGroupForm>) => {
    setFormData(prev => ({ ...prev, ...data }));
    Object.entries(data).forEach(([key, value]) => {
      form.setValue(key as keyof CreateResourceGroupForm, value as never);
    });
  }, [form]);

  const canProceedToNextStep = useCallback((): boolean => {
    const values = form.getValues();
    
    switch (currentStep) {
      case 1:
        return !!(values.name && values.location);
      case 2:
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
    
    if (currentStep < 2) {
      setCurrentStep(currentStep + 1);
    }
  }, [currentStep, canProceedToNextStep, form]);

  const handlePrevious = useCallback(() => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
    }
  }, [currentStep]);

  const handleSubmit = useCallback(async () => {
    const isValid = await form.trigger();
    if (!isValid) {
      return;
    }

    const data = form.getValues();
    
    const resourceGroupData: CreateResourceGroupForm = {
      ...data,
      credential_id: data.credential_id || credentialId,
    };

    try {
      await createResourceGroupMutation.mutateAsync(resourceGroupData);
      success(t('resourceGroup.created') || 'Resource Group creation initiated successfully');
    } catch (error) {
      ErrorHandler.logError(error, { operation: 'createResourceGroup', source: 'create-resource-group-page' });
      showError(t('resourceGroup.createFailed') || 'Failed to create resource group');
    }
  }, [form, createResourceGroupMutation, success, showError, t, credentialId]);

  const handleCancel = useCallback(() => {
    if (onCancel) {
      onCancel();
    } else {
      const path = buildCredentialResourcePath(
        workspaceId,
        credentialId,
        'azure',
        '/iam/resource-groups'
      );
      router.push(path);
    }
  }, [onCancel, workspaceId, credentialId, router]);

  const steps: StepConfig[] = [
    {
      label: 'resourceGroup.basicConfig',
      description: 'resourceGroup.basicConfigDescription',
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
          <BasicResourceGroupConfigStep
            form={form}
            onDataChange={handleDataChange}
          />
        );
      case 2:
        return (
          <ReviewResourceGroupStep
            formData={formValues}
          />
        );
      default:
        return null;
    }
  };

  const backPath = buildCredentialResourcePath(
    workspaceId,
    credentialId,
    'azure',
    '/iam/resource-groups'
  );

  return (
    <CreateResourceStepperLayout
      backPath={backPath}
      title="resourceGroup.createResourceGroup"
      description="resourceGroup.createResourceGroupDescription"
      steps={steps}
      currentStep={currentStep}
      renderStepContent={renderStepContent}
      navigationProps={{
        onNext: handleNext,
        onPrevious: handlePrevious,
        onCancel: handleCancel,
        onSubmit: handleSubmit,
        isLoading: createResourceGroupMutation.isPending,
        submitButtonText: 'common.create',
        submittingButtonText: 'actions.creating',
      }}
      onCancel={handleCancel}
    />
  );
}


/**
 * Create Subnet Page
 * Subnet 생성 페이지 (Stepper 방식)
 */

'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Stepper } from '@/components/ui/stepper';
import { StepContent } from '@/components/ui/stepper';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useTranslation } from '@/hooks/use-translation';
import { createValidationSchemas } from '@/lib/validations';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { ArrowLeft } from 'lucide-react';
import { useCredentials } from '@/hooks/use-credentials';
import type { CreateSubnetForm, CloudProvider } from '@/lib/types';
import { BasicSubnetConfigStep } from '@/features/networks/components/create-subnet/basic-subnet-config-step';
import { AdvancedSubnetConfigStep } from '@/features/networks/components/create-subnet/advanced-subnet-config-step';
import { ReviewSubnetStep } from '@/features/networks/components/create-subnet/review-subnet-step';
import { useSubnetActions } from '@/features/networks/hooks/use-subnet-actions';

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
  const schemas = createValidationSchemas(t);
  
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

  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<Partial<CreateSubnetForm>>({
    credential_id: selectedCredentialId || '',
    name: '',
    vpc_id: vpcIdFromUrl,
    cidr_block: '',
    region: selectedRegion || '',
    availability_zone: '',
    tags: {},
  });

  const { createSubnetMutation } = useSubnetActions({
    selectedProvider,
    selectedCredentialId,
  });

  const form = useForm<CreateSubnetForm>({
    resolver: zodResolver(schemas.createSubnetSchema),
    defaultValues: {
      credential_id: selectedCredentialId || '',
      name: '',
      vpc_id: vpcIdFromUrl,
      cidr_block: '',
      region: selectedRegion || '',
      availability_zone: '',
      tags: {},
    },
  });

  // Update form when credential/region/vpc changes
  useEffect(() => {
    if (selectedCredentialId) {
      form.setValue('credential_id', selectedCredentialId);
      setFormData(prev => ({ ...prev, credential_id: selectedCredentialId }));
    }
    if (selectedRegion) {
      form.setValue('region', selectedRegion);
      setFormData(prev => ({ ...prev, region: selectedRegion }));
    }
    if (vpcIdFromUrl) {
      form.setValue('vpc_id', vpcIdFromUrl);
      setFormData(prev => ({ ...prev, vpc_id: vpcIdFromUrl }));
    }
  }, [selectedCredentialId, selectedRegion, vpcIdFromUrl, form]);

  const updateFormData = (data: Partial<CreateSubnetForm>) => {
    setFormData(prev => ({ ...prev, ...data }));
    Object.entries(data).forEach(([key, value]) => {
      form.setValue(key as keyof CreateSubnetForm, value as any);
    });
  };

  const handleNext = async () => {
    const isValid = await form.trigger();
    if (isValid && currentStep < STEPS.length) {
      setCurrentStep(prev => prev + 1);
    }
  };

  const handlePrevious = () => {
    if (currentStep > 1) {
      setCurrentStep(prev => prev - 1);
    }
  };

  const handleSkipAdvanced = () => {
    setCurrentStep(3); // Skip to Review step
  };

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
    if (confirm('Are you sure you want to cancel? All entered data will be lost.')) {
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
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <Button
            variant="ghost"
            onClick={handleCancel}
            className="mb-4"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            {t('common.back')}
          </Button>
          <h1 className="text-3xl font-bold text-gray-900">{t('network.createSubnetTitle')}</h1>
          <p className="text-gray-600 mt-2">
            {t('network.createSubnetDescriptionNoRegion', { provider: selectedProvider?.toUpperCase() || '' })}
          </p>
        </div>

        {/* Stepper */}
        <Card className="mb-6">
          <CardContent className="pt-6">
            <Stepper
              currentStep={currentStep}
              totalSteps={STEPS.length}
              steps={STEPS.map(step => ({
                label: t(step.label) || step.label.replace('network.', ''),
                description: t(step.description) || step.description.replace('network.', ''),
              }))}
            />
          </CardContent>
        </Card>

        {/* Step Content */}
        <Card>
          <CardHeader className="pb-6">
            <CardTitle>{t(STEPS[currentStep - 1].label) || STEPS[currentStep - 1].label.replace('network.', '')}</CardTitle>
            <CardDescription>{t(STEPS[currentStep - 1].description) || STEPS[currentStep - 1].description.replace('network.', '')}</CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
            <StepContent>{renderStepContent()}</StepContent>

            {/* Navigation Buttons */}
            <div className="flex justify-between mt-8 pt-6 border-t">
              <Button
                type="button"
                variant="outline"
                onClick={currentStep === 1 ? handleCancel : handlePrevious}
                disabled={createSubnetMutation.isPending}
              >
                {currentStep === 1 ? t('common.cancel') : t('common.back')}
              </Button>
              <div className="flex gap-2">
                {currentStep === 2 && (
                  <Button variant="outline" onClick={handleSkipAdvanced}>
                    {t('network.skipAdvancedOptions')}
                  </Button>
                )}
                {currentStep < STEPS.length ? (
                  <Button
                    type="button"
                    onClick={handleNext}
                    disabled={createSubnetMutation.isPending}
                  >
                    {t('common.next')}
                  </Button>
                ) : (
                  <Button
                    type="button"
                    onClick={handleCreateSubnet}
                    disabled={createSubnetMutation.isPending}
                  >
                    {createSubnetMutation.isPending ? t('actions.creating') : t('network.createSubnet')}
                  </Button>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default function CreateSubnetPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <CreateSubnetPageContent />
    </Suspense>
  );
}


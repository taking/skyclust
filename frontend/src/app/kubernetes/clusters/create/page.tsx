/**
 * Create Kubernetes Cluster Page
 * Step 방식 클러스터 생성 페이지
 */

'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Stepper, StepContent } from '@/components/ui/stepper';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useTranslation } from '@/hooks/use-translation';
import { createValidationSchemas } from '@/lib/validations';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { ArrowLeft } from 'lucide-react';
import { useKubernetesClusters } from '@/features/kubernetes';
import type { CreateClusterForm, CloudProvider } from '@/lib/types';
import { BasicClusterConfigStep } from '@/features/kubernetes/components/create-cluster/basic-config-step';
import { NetworkConfigStep } from '@/features/kubernetes/components/create-cluster/network-config-step';
import { AdvancedConfigStep } from '@/features/kubernetes/components/create-cluster/advanced-config-step';
import { ReviewStep } from '@/features/kubernetes/components/create-cluster/review-step';

const STEPS = [
  {
    label: 'basicConfiguration',
    description: 'basicConfigurationDescription',
  },
  {
    label: 'networkConfiguration',
    description: 'networkConfigurationDescription',
  },
  {
    label: 'advancedSettings',
    description: 'advancedSettingsDescription',
  },
  {
    label: 'reviewAndCreate',
    description: 'reviewAndCreateDescription',
  },
];

function CreateClusterPageContent() {
  const router = useRouter();
  const { t } = useTranslation();
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const { currentWorkspace } = useWorkspaceStore();
  const { createClusterSchema } = createValidationSchemas(t);
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'kubernetes',
    updateUrl: true,
  });

  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<Partial<CreateClusterForm>>({
    credential_id: selectedCredentialId || '',
    name: '',
    version: '',
    region: selectedRegion || '',
    subnet_ids: [],
    vpc_id: '',
    role_arn: '',
    tags: {},
    access_config: {
      authentication_mode: 'API',
      bootstrap_cluster_creator_admin_permissions: true,
    },
  });

  const {
    credentials,
    createClusterMutation,
  } = useKubernetesClusters({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || '',
    selectedRegion: selectedRegion || '',
  });

  const selectedCredential = credentials.find(c => c.id === selectedCredentialId);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  const form = useForm<CreateClusterForm>({
    resolver: zodResolver(createClusterSchema),
    defaultValues: formData as CreateClusterForm,
    mode: 'onChange',
  });

  // Form data 동기화 - selectedCredentialId가 변경되면 form 업데이트
  useEffect(() => {
    if (selectedCredentialId) {
      form.setValue('credential_id', selectedCredentialId);
      setFormData(prev => ({ ...prev, credential_id: selectedCredentialId }));
    }
  }, [selectedCredentialId, form]);
  
  useEffect(() => {
    if (selectedRegion) {
      form.setValue('region', selectedRegion);
      setFormData(prev => ({ ...prev, region: selectedRegion }));
    }
  }, [selectedRegion, form]);
  
  // 초기 마운트 시 form 초기화
  useEffect(() => {
    form.reset(formData as CreateClusterForm);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // 빈 의존성 배열로 초기 마운트 시에만 실행

  const updateFormData = (data: Partial<CreateClusterForm>) => {
    setFormData(prev => ({ ...prev, ...data }));
  };

  const handleNext = async () => {
    // Step별 필수 필드 검증
    let fieldsToValidate: (keyof CreateClusterForm)[] = [];
    
    switch (currentStep) {
      case 1:
        fieldsToValidate = ['credential_id', 'name', 'version', 'region'];
        break;
      case 2:
        fieldsToValidate = ['subnet_ids'];
        // VPC 선택도 확인 (UI에서만 사용, API에는 전송 안함)
        if (!formData.vpc_id) {
          handleError(new Error('Please select a VPC first'), { operation: 'validateStep2' });
          return;
        }
        break;
      case 3:
        // Step 3는 모두 optional이므로 검증 불필요
        break;
      default:
        break;
    }

    if (fieldsToValidate.length > 0) {
      const isValid = await form.trigger(fieldsToValidate);
      if (!isValid) {
        return;
      }
    }

    if (currentStep < STEPS.length) {
      setCurrentStep(prev => prev + 1);
    }
  };

  const handleBack = () => {
    if (currentStep > 1) {
      setCurrentStep(prev => prev - 1);
    }
  };

  const handleCreateCluster = async () => {
    if (!selectedProvider) {
      handleError(new Error('Provider not selected'), { operation: 'createCluster' });
      return;
    }

    const validatedData = await form.trigger();
    if (!validatedData) {
      handleError(new Error('Please fix validation errors'), { operation: 'createCluster' });
      return;
    }

    const finalData = form.getValues();
    
    // VPC ID는 API에 전송하지 않음 (subnet_ids만 전송)
    const { vpc_id, ...apiData } = finalData;
    
    createClusterMutation.mutate(
      { provider: selectedProvider, data: apiData },
      {
        onSuccess: () => {
          success('Cluster creation initiated');
          router.push('/kubernetes/clusters');
        },
        onError: (error: unknown) => {
          handleError(error, { operation: 'createCluster', resource: 'Cluster' });
        },
      }
    );
  };

  const handleCancel = () => {
    if (confirm('Are you sure you want to cancel? All entered data will be lost.')) {
      router.push('/kubernetes/clusters');
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <BasicClusterConfigStep
            form={form}
            credentials={credentials}
            selectedCredentialId={selectedCredentialId || ''}
            onCredentialChange={(id) => updateFormData({ credential_id: id })}
            selectedProvider={selectedProvider}
            onDataChange={updateFormData}
          />
        );
      case 2:
        return (
          <NetworkConfigStep
            form={form}
            selectedProvider={selectedProvider}
            selectedCredentialId={selectedCredentialId || ''}
            selectedRegion={selectedRegion || ''}
            onDataChange={updateFormData}
          />
        );
      case 3:
        return (
          <AdvancedConfigStep
            form={form}
            onDataChange={updateFormData}
          />
        );
      case 4:
        return (
          <ReviewStep
            formData={formData as CreateClusterForm}
            selectedProvider={selectedProvider}
            onCreate={handleCreateCluster}
            isPending={createClusterMutation.isPending}
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
            {t('kubernetes.backToClusters')}
          </Button>
          <h1 className="text-3xl font-bold text-gray-900">{t('kubernetes.createKubernetesCluster')}</h1>
          <p className="text-gray-600 mt-2">
            {t('kubernetes.configureClusterStepByStep')}
          </p>
        </div>

        {/* Stepper */}
        <Card className="mb-6">
          <CardContent className="pt-6">
            <Stepper
              currentStep={currentStep}
              totalSteps={STEPS.length}
              steps={STEPS}
            />
          </CardContent>
        </Card>

        {/* Step Content */}
        <Card>
          <CardHeader>
            <CardTitle>{t(`kubernetes.${STEPS[currentStep - 1].label}`)}</CardTitle>
            <CardDescription>{t(`kubernetes.${STEPS[currentStep - 1].description}`)}</CardDescription>
          </CardHeader>
          <CardContent>
            <StepContent>{renderStepContent()}</StepContent>

            {/* Navigation Buttons */}
            <div className="flex justify-between mt-8 pt-6 border-t">
              <Button
                type="button"
                variant="outline"
                onClick={currentStep === 1 ? handleCancel : handleBack}
                disabled={createClusterMutation.isPending}
              >
                {currentStep === 1 ? t('common.cancel') : t('common.back')}
              </Button>
              <div className="flex gap-2">
                {currentStep < STEPS.length ? (
                  <Button
                    type="button"
                    onClick={handleNext}
                    disabled={createClusterMutation.isPending}
                  >
                    {t('common.next')}
                  </Button>
                ) : (
                  <Button
                    type="button"
                    onClick={handleCreateCluster}
                    disabled={createClusterMutation.isPending}
                  >
                    {createClusterMutation.isPending ? t('kubernetes.creating') : t('kubernetes.createCluster')}
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

export default function CreateClusterPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-gray-50 py-8 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <CreateClusterPageContent />
    </Suspense>
  );
}


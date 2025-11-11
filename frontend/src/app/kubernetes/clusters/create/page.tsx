/**
 * Create Kubernetes Cluster Page
 * Step 방식 클러스터 생성 페이지
 */

'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Stepper, StepContent } from '@/components/ui/stepper';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useTranslation } from '@/hooks/use-translation';
import { useValidation } from '@/lib/validation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { ArrowLeft } from 'lucide-react';
import { useKubernetesClusters } from '@/features/kubernetes';
import type { CreateClusterForm, CloudProvider } from '@/lib/types';
import { useVPCs } from '@/features/networks/hooks/use-vpcs';
import { useSubnets } from '@/features/networks/hooks/use-subnets';
import { UI_MESSAGES, VALIDATION } from '@/lib/constants';
import { TableSkeleton } from '@/components/ui/table-skeleton';

// Dynamic imports for step components (lazy loading)
const BasicClusterConfigStep = dynamic(
  () => import('@/features/kubernetes/components/create-cluster/basic-config-step').then(mod => ({ default: mod.BasicClusterConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const NetworkConfigStep = dynamic(
  () => import('@/features/kubernetes/components/create-cluster/network-config-step').then(mod => ({ default: mod.NetworkConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const AdvancedConfigStep = dynamic(
  () => import('@/features/kubernetes/components/create-cluster/advanced-config-step').then(mod => ({ default: mod.AdvancedConfigStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

const ReviewStep = dynamic(
  () => import('@/features/kubernetes/components/create-cluster/review-step').then(mod => ({ default: mod.ReviewStep })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={2} rows={5} />,
  }
);

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
  const { schemas } = useValidation();
  const createClusterSchema = schemas.createClusterSchema;
  
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

  // Step 2에서 VPC와 Subnet 정보를 가져오기 위해 사용
  const { vpcs } = useVPCs();
  
  // Subnets는 vpc_id가 있을 때만 가져오기
  const { subnets, setSelectedVPCId: setSubnetVPCId } = useSubnets();
  
  // formData.vpc_id가 변경되면 subnets hook의 selectedVPCId 업데이트
  useEffect(() => {
    if (formData.vpc_id && setSubnetVPCId) {
      setSubnetVPCId(formData.vpc_id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData.vpc_id]);

  const form = useForm<CreateClusterForm>({
    resolver: zodResolver(createClusterSchema) as any, // Type compatibility workaround
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
        fieldsToValidate = ['credential_id', 'name', 'version', 'region', 'zone'];
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
    
    // AWS EKS: 서브넷이 최소 2개의 다른 AZ에 있는지 검증
    if (selectedProvider === 'aws' && finalData.subnet_ids && finalData.subnet_ids.length > 0) {
      const selectedSubnets = finalData.subnet_ids
        .map(id => subnets.find(s => s.id === id))
        .filter(Boolean);
      
      const uniqueAZs = new Set(
        selectedSubnets
          .map(s => s?.availability_zone)
          .filter(Boolean)
      );
      
      if (uniqueAZs.size < VALIDATION.ARRAY.MIN_AZ_COUNT_FOR_AWS_EKS) {
        handleError(
          new Error('AWS EKS requires subnets from at least two different availability zones for high availability.'),
          { operation: 'createCluster', resource: 'Cluster' }
        );
        return;
      }
    }
    
    // Azure의 경우 특별한 데이터 변환 필요
    if (selectedProvider === 'azure') {
      const azureData: CreateClusterForm = {
        credential_id: finalData.credential_id,
        name: finalData.name,
        version: finalData.version,
        region: finalData.region || finalData.location || '', // Required field
        location: finalData.location || finalData.region, // Azure uses 'location'
        resource_group: finalData.resource_group || '',
        subnet_ids: finalData.subnet_ids || [finalData.network?.subnet_id || ''].filter(Boolean), // Required field
        network: finalData.network || {
          virtual_network_id: finalData.vpc_id || '',
          subnet_id: finalData.subnet_ids?.[0] || '',
          network_plugin: 'azure',
        },
        node_pool: finalData.node_pool || {
          name: 'nodepool1',
          vm_size: 'Standard_D2s_v3',
          node_count: 3,
        },
        security: finalData.security,
        tags: finalData.tags,
      };
      
      createClusterMutation.mutate(
        { provider: selectedProvider, data: azureData },
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
    } else {
      // AWS/GCP의 경우 기존 로직 사용
      // VPC ID와 role_arn은 API에 전송하지 않음 (role_arn은 백엔드에서 자동 생성)
      const { vpc_id, role_arn, ...apiData } = finalData;
      
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
    }
  };

  const handleCancel = () => {
    if (confirm(UI_MESSAGES.CONFIRM_CANCEL)) {
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
            selectedProvider={selectedProvider}
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
              steps={STEPS.map(step => ({
                label: t(`kubernetes.${step.label}`),
                description: t(`kubernetes.${step.description}`),
              }))}
            />
          </CardContent>
        </Card>

        {/* Step Content */}
        <Card>
          <CardHeader className="pb-6">
            <CardTitle>{t(`kubernetes.${STEPS[currentStep - 1].label}`)}</CardTitle>
            <CardDescription>{t(`kubernetes.${STEPS[currentStep - 1].description}`)}</CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
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

            {/* Selected Info - Option C: CardContent 하단, 전체 너비, 우측 정렬 */}
            <div className="mt-4 pt-4 border-t flex justify-end items-center">
              <span className="text-sm text-muted-foreground">
                {(() => {
                  // Step 1: Credential | Region | Cluster
                  if (currentStep === 1) {
                    return `${selectedCredential?.name || selectedCredential?.provider?.toUpperCase() || '-'} | ${formData.region || selectedRegion || '-'} | ${formData.name || '-'}`;
                  }
                  
                  // Step 2: Credential | Region | Cluster | VPC | Subnets
                  if (currentStep === 2) {
                    const vpcId = formData.vpc_id || '';
                    const selectedVPC = vpcs.find(v => v.id === vpcId);
                    const vpcName = selectedVPC?.name || selectedVPC?.id || '-';
                    
                    const subnetIds = formData.subnet_ids || [];
                    let subnetsDisplay = '-';
                    if (subnetIds.length > 0) {
                      const selectedSubnets = subnetIds
                        .map(id => subnets.find(s => s.id === id))
                        .filter(Boolean);
                      
                      if (selectedSubnets.length > 2) {
                        // 3개 이상: 첫 번째 + 개수
                        const firstName = selectedSubnets[0]?.name || selectedSubnets[0]?.id || subnetIds[0];
                        subnetsDisplay = `${firstName}, +${selectedSubnets.length - 1} more`;
                      } else if (selectedSubnets.length > 0) {
                        // 1-2개: 이름 나열
                        subnetsDisplay = selectedSubnets
                          .map(s => s?.name || s?.id)
                          .join(', ');
                      } else {
                        // 이름을 찾을 수 없으면 ID 표시
                        subnetsDisplay = subnetIds.length > 2
                          ? `${subnetIds[0]}, +${subnetIds.length - 1} more`
                          : subnetIds.join(', ');
                      }
                    }
                    
                    return `${selectedCredential?.name || selectedCredential?.provider?.toUpperCase() || '-'} | ${formData.region || selectedRegion || '-'} | ${formData.name || '-'} | ${vpcName} | ${subnetsDisplay}`;
                  }
                  
                  // Step 3, 4: Step 1과 동일
                  return `${selectedCredential?.name || selectedCredential?.provider?.toUpperCase() || '-'} | ${formData.region || selectedRegion || '-'} | ${formData.name || '-'}`;
                })()}
              </span>
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


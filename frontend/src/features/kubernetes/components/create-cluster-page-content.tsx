/**
 * Create Cluster Page Content Component
 * Kubernetes 클러스터 생성 페이지 컴포넌트 (Stepper 형태)
 * 
 * 기존 Step 컴포넌트를 재사용하여 Stepper 형태로 구현
 */

'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { createValidationSchemas } from '@/lib/validation';
import { useTranslation } from '@/hooks/use-translation';
import { useKubernetesClusters } from '@/features/kubernetes';
import { useCredentials } from '@/hooks/use-credentials';
import { buildCredentialResourcePath, buildCredentialResourceDetailPath } from '@/lib/routing/helpers';
import { getRegionLabel } from '@/lib/regions/list';
import { CreateResourceStepperLayout, type StepConfig } from '@/components/common/create-resource-stepper-layout';
import type { SelectedValue } from '@/components/common/stepper-selected-values';
import { BasicClusterConfigStep } from './create-cluster/basic-config-step';
import { NetworkConfigStep } from './create-cluster/network-config-step';
import { AdvancedConfigStep } from './create-cluster/advanced-config-step';
import { ReviewStep } from './create-cluster/review-step';
import type { CreateClusterForm, CloudProvider } from '@/lib/types';

interface CreateClusterPageContentProps {
  workspaceId: string;
  credentialId: string;
  region?: string | null;
  onCancel?: () => void;
}

export function CreateClusterPageContent({
  workspaceId,
  credentialId,
  region,
  onCancel,
}: CreateClusterPageContentProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const { success, error: showError } = useToast();
  const schemas = createValidationSchemas(t);

  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<Partial<CreateClusterForm>>({});

  const { credentials, selectedCredential, selectedProvider: initialSelectedProvider } = useCredentials({
    workspaceId,
    selectedCredentialId: credentialId,
    enabled: !!workspaceId,
  });

  const form = useForm<CreateClusterForm>({
    resolver: zodResolver(schemas.createClusterSchema),
    defaultValues: {
      credential_id: credentialId || '',
      name: '',
      version: '',
      region: region || '',
      zone: '',
      subnet_ids: [],
      vpc_id: '',
      role_arn: '',
      tags: {},
      access_config: {
        authentication_mode: 'API',
        bootstrap_cluster_creator_admin_permissions: true,
      },
      deployment_mode: 'auto', // Azure 기본값: 자동 모드
    },
    mode: 'onChange',
  });
  
  // Form에서 선택된 credential을 기반으로 provider 결정
  const formCredentialId = form.watch('credential_id');
  const selectedProvider = useMemo(() => {
    const currentCredentialId = formCredentialId || credentialId;
    const currentCredential = credentials.find(c => c.id === currentCredentialId);
    return currentCredential?.provider as CloudProvider | undefined;
  }, [formCredentialId, credentialId, credentials]);

  const { createClusterMutation } = useKubernetesClusters({
    workspaceId,
    selectedCredentialId: credentialId,
    selectedRegion: region || undefined,
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

  const handleDataChange = useCallback((data: Partial<CreateClusterForm>) => {
    setFormData(prev => ({ ...prev, ...data }));
    Object.entries(data).forEach(([key, value]) => {
      form.setValue(key as keyof CreateClusterForm, value as never);
    });
  }, [form]);

  // GCP provider 선택 시 기본값 설정
  useEffect(() => {
    if (selectedProvider === 'gcp') {
      const currentValues = form.getValues();
      
      // Network 기본값 설정 (VPC/Subnet이 선택된 경우에만)
      if (currentValues.vpc_id && currentValues.subnet_ids && currentValues.subnet_ids.length > 0) {
        const currentNetwork = currentValues.network || {};
        if (!currentNetwork.pod_cidr && !currentNetwork.service_cidr) {
          const updatedNetwork = {
            ...currentNetwork,
            subnet_id: currentValues.subnet_ids[0] || '',
            pod_cidr: currentNetwork.pod_cidr || '10.0.0.0/16',
            service_cidr: currentNetwork.service_cidr || '10.1.0.0/16',
            master_authorized_networks: currentNetwork.master_authorized_networks || [],
            private_endpoint: currentNetwork.private_endpoint || false,
            private_nodes: currentNetwork.private_nodes || false,
          };
          form.setValue('network', updatedNetwork);
          handleDataChange({ network: updatedNetwork });
        }
      }

      // Node Pool 기본값 설정
      const currentNodePool = currentValues.node_pool;
      if (!currentNodePool || !currentNodePool.node_count || !currentNodePool.disk_size_gb) {
        const updatedNodePool = {
          ...currentNodePool,
          name: currentNodePool?.name || 'default-pool',
          machine_type: currentNodePool?.machine_type || '',
          node_count: currentNodePool?.node_count || 2,
          disk_size_gb: currentNodePool?.disk_size_gb || 30,
          disk_type: currentNodePool?.disk_type || 'pd-standard',
          auto_scaling: currentNodePool?.auto_scaling || {
            enabled: true,
            min_node_count: 2,
            max_node_count: 3,
          },
          preemptible: currentNodePool?.preemptible || false,
          spot: currentNodePool?.spot || false,
        };
        form.setValue('node_pool', updatedNodePool);
        handleDataChange({ node_pool: updatedNodePool });
      }

      // Cluster Mode 기본값 설정
      if (!currentValues.cluster_mode) {
        const clusterMode = {
          type: 'standard',
          remove_default_node_pool: false,
        };
        form.setValue('cluster_mode', clusterMode);
        handleDataChange({ cluster_mode: clusterMode });
      }

      // Security 기본값 설정
      if (!currentValues.security) {
        const security = {
          binary_authorization: false,
          network_policy: true,
          pod_security_policy: false,
          enable_workload_identity: true,
        };
        form.setValue('security', security);
        handleDataChange({ security });
      }
    }
  }, [selectedProvider, form, handleDataChange]);

  // Azure provider 선택 시 기본값 설정 (자동 모드일 때만)
  useEffect(() => {
    if (selectedProvider === 'azure' && form.watch('deployment_mode') === 'auto') {
      const currentValues = form.getValues();
      const currentNodePool = currentValues.node_pool;
      
      // Node Pool 기본값 설정 (이미 값이 있으면 건너뜀)
      if (!currentNodePool || !currentNodePool.name) {
        const updatedNodePool = {
          ...currentNodePool,
          name: currentNodePool?.name || 'nodepool1',
          vm_size: currentNodePool?.vm_size || '',
          node_count: currentNodePool?.node_count ?? 3,
          min_count: currentNodePool?.min_count ?? 1,
          max_count: currentNodePool?.max_count ?? 10,
          enable_auto_scaling: currentNodePool?.enable_auto_scaling ?? false,
          os_disk_size_gb: currentNodePool?.os_disk_size_gb ?? 128,
          os_disk_type: currentNodePool?.os_disk_type || 'Managed',
          os_type: currentNodePool?.os_type || 'Linux',
          os_sku: currentNodePool?.os_sku || 'Ubuntu',
          max_pods: currentNodePool?.max_pods ?? 30,
          mode: currentNodePool?.mode || 'System',
        };
        form.setValue('node_pool', updatedNodePool);
        handleDataChange({ node_pool: updatedNodePool });
      }
      
      // Security 기본값 설정
      const currentSecurity = currentValues.security;
      if (!currentSecurity || Object.keys(currentSecurity).length === 0) {
        const updatedSecurity = {
          enable_rbac: true,
          enable_pod_security_policy: false,
          enable_private_cluster: false,
          enable_azure_policy: false,
          enable_workload_identity: false,
        };
        form.setValue('security', updatedSecurity);
        handleDataChange({ security: updatedSecurity });
      }
    }
  }, [selectedProvider, form, handleDataChange]);

  const handleCredentialChange = useCallback((newCredentialId: string) => {
    handleDataChange({ credential_id: newCredentialId });
  }, [handleDataChange]);

  const canProceedToNextStep = useCallback((): boolean => {
    const values = form.getValues();
    const errors = form.formState.errors;
    const isAzureAutoMode = selectedProvider === 'azure' && values.deployment_mode === 'auto';
    
    switch (currentStep) {
      case 1:
        // Azure일 때는 resource_group과 deployment_mode도 필수
        if (selectedProvider === 'azure') {
          return !!(
            values.name && 
            values.credential_id && 
            values.version && 
            values.region &&
            values.resource_group &&
            values.deployment_mode &&
            !errors.name &&
            !errors.credential_id &&
            !errors.version &&
            !errors.region &&
            !errors.resource_group &&
            !errors.deployment_mode
          );
        }
        return !!(
          values.name && 
          values.credential_id && 
          values.version && 
          values.region &&
          !errors.name &&
          !errors.credential_id &&
          !errors.version &&
          !errors.region
        );
      case 2:
        // Network step (Azure 자동 모드일 때는 스킵되므로 여기 도달하지 않음)
        // 일반 모드일 때만 검증
        if (!isAzureAutoMode) {
          return !!(
            values.vpc_id && 
            values.subnet_ids && 
            values.subnet_ids.length > 0 &&
            !errors.vpc_id &&
            !errors.subnet_ids
          );
        }
        return true;
      case 3:
        // Advanced step
        // Azure 자동 모드일 때는 node_pool 필수
        if (isAzureAutoMode && selectedProvider === 'azure') {
          return !!(
            values.node_pool &&
            values.node_pool.name &&
            values.node_pool.vm_size &&
            values.node_pool.node_count &&
            !errors.node_pool
          );
        }
        // 일반 모드일 때는 role_arn과 tags 검증
        return !errors.role_arn && !errors.tags;
      case 4:
        // Review step - 항상 진행 가능
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
    
    // Azure 자동 모드일 때 network step 스킵
    const formValues = form.getValues();
    const isAzureAutoMode = selectedProvider === 'azure' && formValues.deployment_mode === 'auto';
    
    // steps 배열은 항상 4개이므로 totalSteps는 4
    const totalSteps = 4;
    
    if (currentStep === 1 && isAzureAutoMode) {
      // Step 1 (Basic) -> Step 3 (Advanced)로 건너뛰기
      setCurrentStep(3);
    } else if (currentStep < totalSteps) {
      setCurrentStep(currentStep + 1);
    }
  }, [currentStep, canProceedToNextStep, form, selectedProvider]);

  const handlePrevious = useCallback(() => {
    if (currentStep > 1) {
      // Azure 자동 모드일 때 network step 스킵
      const formValues = form.getValues();
      const isAzureAutoMode = selectedProvider === 'azure' && formValues.deployment_mode === 'auto';
      
      if (currentStep === 3 && isAzureAutoMode) {
        // Step 3 (Advanced) -> Step 1 (Basic)로 이동
        setCurrentStep(1);
      } else {
        setCurrentStep(currentStep - 1);
      }
    }
  }, [currentStep, form, selectedProvider]);

  const handleSkipAdvanced = useCallback(() => {
    setCurrentStep(4);
  }, []);

  const handleSubmit = useCallback(async () => {
    console.log('[CreateCluster] handleSubmit called');
    const data = form.getValues();
    console.log('[CreateCluster] Form data:', data);
    
    if (!selectedProvider) {
      console.log('[CreateCluster] Provider not selected');
      showError(t('kubernetes.providerNotSelected') || 'Provider not selected');
      return;
    }

    // Azure 자동 모드일 때는 validation 스킵
    const isAzureAutoMode = selectedProvider === 'azure' && data.deployment_mode === 'auto';
    
    // Azure Auto 모드가 아닌 경우에만 validation 수행
    if (!isAzureAutoMode) {
      const isValid = await form.trigger();
      if (!isValid) {
        const errors = form.formState.errors;
        console.log('[CreateCluster] Form validation failed. Errors:', errors);
        console.log('[CreateCluster] Form values:', form.getValues());
        
        // 첫 번째 에러 메시지 표시
        const firstError = Object.values(errors)[0];
        if (firstError) {
          const errorMessage = firstError.message || 'Form validation failed';
          showError(errorMessage);
        } else {
          showError(t('form.validation.genericError') || 'Please check all required fields');
        }
        return;
      }
    }
    
    // 서브넷이 선택된 경우, credential_id와 region이 올바른지 검증 (Azure 자동 모드 제외)
    if (!isAzureAutoMode && data.subnet_ids && data.subnet_ids.length > 0) {
      if (!data.credential_id || !data.region) {
        console.log('[CreateCluster] Missing credential or region when subnets are selected');
        showError(t('kubernetes.missingCredentialOrRegion') || 'Credential and region must be selected when subnets are selected');
        return;
      }
      
      // AWS의 경우 최소 2개의 서브넷이 필요
      if (selectedProvider === 'aws' && data.subnet_ids.length < 2) {
        console.log('[CreateCluster] AWS requires at least 2 subnets');
        showError(t('kubernetes.awsRequiresTwoSubnets') || 'AWS EKS requires at least 2 subnets from different availability zones');
        return;
      }
    }

    try {
      console.log('[CreateCluster] Starting mutation with provider:', selectedProvider, 'data:', data);
      const createdCluster = await createClusterMutation.mutateAsync({
        provider: selectedProvider,
        data,
      });
      console.log('[CreateCluster] Mutation successful, created cluster:', createdCluster);
      
      success(t('kubernetes.clusterCreated') || 'Cluster creation initiated successfully');
      
      if (createdCluster?.name) {
        const detailPath = buildCredentialResourceDetailPath(
          workspaceId,
          credentialId,
          'k8s',
          'clusters',
          createdCluster.name,
          { region: createdCluster.region || region || undefined }
        );
        router.push(detailPath);
      } else {
        const path = buildCredentialResourcePath(
          workspaceId,
          credentialId,
          'k8s',
          '/clusters',
          { region: region || undefined }
        );
        router.push(path);
      }
    } catch (error) {
      ErrorHandler.logError(error, { operation: 'createCluster', source: 'create-cluster-page' });
      showError(t('kubernetes.clusterCreateFailed') || 'Failed to create cluster');
    }
  }, [form, selectedProvider, createClusterMutation, success, showError, t, workspaceId, credentialId, region, router]);

  const backPath = useMemo(() => {
    return buildCredentialResourcePath(
      workspaceId,
      credentialId,
      'k8s',
      '/clusters',
      { region: region || undefined }
    );
  }, [workspaceId, credentialId, region]);

  const handleCancel = useCallback(() => {
    router.push(backPath);
  }, [router, backPath]);

  // Azure 자동 모드일 때 network step 숨김
  const formValues = form.getValues();
  const isAzureAutoMode = selectedProvider === 'azure' && formValues.deployment_mode === 'auto';
  
  // steps 배열은 항상 4개로 유지 (Step 2는 Azure 자동 모드일 때 스킵되지만 표시는 됨)
  const steps: StepConfig[] = [
    {
      label: 'kubernetes.basicConfig.label',
      description: 'kubernetes.basicConfig.description',
    },
    {
      label: 'kubernetes.networkConfig',
      description: 'kubernetes.networkConfigDescription',
    },
    {
      label: 'kubernetes.advancedConfig',
      description: 'kubernetes.advancedConfigDescription',
    },
    {
      label: 'common.review',
      description: 'common.reviewDescription',
    },
  ];

  // 선택된 값들 추출 (실시간 업데이트)
  const selectedValues = useMemo<SelectedValue[]>(() => {
    const formValues = form.getValues();
    const currentFormCredentialId = formValues.credential_id || credentialId;
    const currentCredential = credentials.find(c => c.id === currentFormCredentialId);
    const credentialName = currentCredential?.name || '';
    const selectedRegion = formValues.region || region || '';
    const regionLabel = getRegionLabel(selectedProvider, selectedRegion);
    const selectedZone = formValues.zone || '';
    const clusterName = formValues.name || '';
    const kubernetesVersion = formValues.version || '';

    return [
      {
        label: t('credential.title') || '자격증명',
        value: credentialName,
        placeholder: t('common.notSelected') || '미선택',
      },
      {
        label: t('common.region') || '리전',
        value: regionLabel || selectedRegion,
        placeholder: t('common.notSelected') || '미선택',
      },
      {
        label: t('common.zone') || '영역',
        value: selectedZone,
        placeholder: t('common.notSelected') || '미선택',
      },
      {
        label: t('kubernetes.basicConfig.clusterName') || t('kubernetes.clusterName') || '클러스터명',
        value: clusterName,
        placeholder: t('common.notSelected') || '미선택',
      },
      {
        label: t('kubernetes.basicConfig.kubernetesVersion') || t('kubernetes.version') || 'Kubernetes 버전',
        value: kubernetesVersion,
        placeholder: t('common.notSelected') || '미선택',
      },
    ];
  }, [
    form.watch('credential_id'),
    form.watch('region'),
    form.watch('zone'),
    form.watch('name'),
    form.watch('version'),
    credentials,
    selectedProvider,
    region,
    credentialId,
    t,
  ]);

  // Review step에서 사용할 최신 form 값들 (watch를 사용하여 실시간 업데이트)
  const watchedFormValues = {
    credential_id: form.watch('credential_id'),
    name: form.watch('name'),
    version: form.watch('version'),
    region: form.watch('region'),
    zone: form.watch('zone'),
    subnet_ids: form.watch('subnet_ids'),
    vpc_id: form.watch('vpc_id'),
    role_arn: form.watch('role_arn'),
    tags: form.watch('tags'),
    access_config: form.watch('access_config'),
    project_id: form.watch('project_id'),
    cluster_mode: form.watch('cluster_mode'),
    location: form.watch('location'),
    resource_group: form.watch('resource_group'),
    network: form.watch('network'),
    node_pool: form.watch('node_pool'),
    security: form.watch('security'),
  } as CreateClusterForm;

  // selectedProjectId 계산 (GCP일 때만 사용)
  const selectedProjectId = form.watch('project_id');

  const renderStepContent = () => {
    const formValues = form.getValues();
    const selectedRegion = formValues.region || region || '';
    const currentFormCredentialId = formValues.credential_id || credentialId;

    switch (currentStep) {
      case 1:
        return (
          <BasicClusterConfigStep
            form={form}
            credentials={credentials}
            selectedCredentialId={currentFormCredentialId}
            onCredentialChange={handleCredentialChange}
            selectedProvider={selectedProvider}
            workspaceId={workspaceId}
            onDataChange={handleDataChange}
          />
        );
      case 2:
        // Azure 자동 모드일 때 network step 스킵
        const formValues = form.getValues();
        const isAzureAutoMode = selectedProvider === 'azure' && formValues.deployment_mode === 'auto';
        
        if (isAzureAutoMode) {
          // 자동 모드일 때는 Advanced step으로 이동 (Step 2는 사용하지 않음)
          return (
            <AdvancedConfigStep
              form={form}
              selectedProvider={selectedProvider}
              selectedCredentialId={formCredentialId}
              selectedProjectId={selectedProjectId}
              onDataChange={handleDataChange}
            />
          );
        }
        
        return (
          <NetworkConfigStep
            form={form}
            selectedProvider={selectedProvider}
            selectedCredentialId={formCredentialId}
            selectedRegion={selectedRegion}
            selectedZone={form.watch('zone')}
            selectedProjectId={selectedProjectId}
            onDataChange={handleDataChange}
          />
        );
      case 3:
        // Azure 자동 모드일 때는 Step 3이 Advanced, 일반 모드일 때도 Step 3이 Advanced
        return (
          <AdvancedConfigStep
            form={form}
            selectedProvider={selectedProvider}
            selectedCredentialId={formCredentialId}
            selectedProjectId={selectedProjectId}
            onDataChange={handleDataChange}
          />
        );
      case 4:
        // Review step에서는 watch를 사용한 최신 값 사용
        return (
          <ReviewStep
            formData={watchedFormValues}
            selectedProvider={selectedProvider}
            selectedProjectId={selectedProjectId}
            onCreate={handleSubmit}
            isPending={createClusterMutation.isPending}
          />
        );
      default:
        return null;
    }
  };

  return (
    <CreateResourceStepperLayout
      backPath={backPath}
      title="kubernetes.createCluster"
      description="kubernetes.createClusterDescription"
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
        isLoading: createClusterMutation.isPending,
        submitButtonText: 'common.create',
        submittingButtonText: 'actions.creating',
      }}
      onCancel={handleCancel}
      advancedStepNumber={3}
      selectedValues={selectedValues}
    />
  );
}

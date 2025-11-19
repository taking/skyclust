/**
 * Create Node Group Dialog Component
 * 노드 그룹 생성 다이얼로그 컴포넌트 (EKS용)
 * GPU 필터링 및 추천 사양 자동 선택 기능 포함
 */

'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import type { CreateNodeGroupForm } from '@/lib/types';
import { useInstanceTypes, useEKSAmitTypes, useAvailabilityZones } from '../hooks/use-kubernetes-metadata';
import { useNetworkResources } from '@/features/networks/hooks/use-network-resources';
import { Sparkles, AlertCircle } from 'lucide-react';
import { createNodeGroupSchema } from '@/lib/validation/schemas';
import type { InstanceTypeInfo, GPUQuotaAvailability } from '../services/kubernetes';
import type { AWSCluster, ProviderCluster } from '@/lib/types';
import { kubernetesService } from '../services/kubernetes';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useQuery } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES } from '@/lib/query';

interface CreateNodeGroupDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clusterName: string;
  cluster?: ProviderCluster | null; // 클러스터 정보 (VPC ID, 기본 서브넷 정보 포함)
  defaultRegion: string;
  defaultCredentialId: string;
  onSubmit: (data: CreateNodeGroupForm) => void;
  onCredentialIdChange: (credentialId: string) => void;
  onRegionChange: (region: string) => void;
  isPending: boolean;
  initialData?: Partial<CreateNodeGroupForm>; // 편집 모드용 초기 데이터
}

// Helper function to check if AMI type supports GPU
function isGPUAMIType(amiType: string): boolean {
  return amiType.includes('NVIDIA') || amiType.includes('NEURON');
}

// Helper function to get compatible AMI types
function getCompatibleAMITypes(
  instanceType: string | null,
  instanceTypes: InstanceTypeInfo[],
  amiTypes: string[],
  useGPU: boolean
): string[] {
  if (!instanceType || instanceTypes.length === 0 || amiTypes.length === 0) {
    return amiTypes;
  }

  const selectedInstance = instanceTypes.find(it => it.instance_type === instanceType);
  if (!selectedInstance) {
    return amiTypes;
  }

  const architecture = selectedInstance.architecture;
  const hasGPU = selectedInstance.has_gpu;

  // GPU 매칭 확인
  if (useGPU !== hasGPU) {
    return [];
  }

  return amiTypes.filter(amiType => {
    const amiHasGPU = isGPUAMIType(amiType);
    if (hasGPU !== amiHasGPU) {
      return false;
    }

    // 아키텍처 매칭
    const isX86 = amiType.includes('x86_64');
    const isARM = amiType.includes('ARM_64');
    if (architecture === 'x86_64' && !isX86) {
      return false;
    }
    if (architecture === 'arm64' && !isARM) {
      return false;
    }

    return true;
  });
}

// Helper function to get recommended instance type
function getRecommendedInstanceType(
  instanceTypes: InstanceTypeInfo[],
  useGPU: boolean,
  architecture: string = 'x86_64'
): string | null {
  if (instanceTypes.length === 0) {
    return useGPU ? 'g5.xlarge' : 't3.medium';
  }

  const candidates = instanceTypes.filter(it => {
    if (it.has_gpu !== useGPU) return false;
    if (it.architecture !== architecture) return false;
    return true;
  });

  if (candidates.length === 0) {
    // Fallback: GPU requirement only
    const fallbackCandidates = instanceTypes.filter(it => it.has_gpu === useGPU);
    if (fallbackCandidates.length === 0) {
      return useGPU ? 'g5.xlarge' : 't3.medium';
    }
    return fallbackCandidates[0].instance_type;
  }

  if (useGPU) {
    // GPU 인스턴스 중 g5.xlarge 우선
    const g5xlarge = candidates.find(c => c.instance_type === 'g5.xlarge');
    if (g5xlarge) return g5xlarge.instance_type;
    return candidates[0].instance_type;
  } else {
    // t3.medium 우선
    const t3medium = candidates.find(c => c.instance_type === 't3.medium');
    if (t3medium) return t3medium.instance_type;
    // 가장 작은 VCPU
    const smallest = candidates.reduce((min, curr) => 
      curr.vcpu < min.vcpu ? curr : min
    );
    return smallest.instance_type;
  }
}

// Helper function to get recommended AMI type
function getRecommendedAMIType(
  compatibleAMITypes: string[],
  hasGPU: boolean,
  architecture: string
): string | null {
  if (compatibleAMITypes.length === 0) {
    if (hasGPU) {
      return architecture === 'arm64' ? 'AL2023_ARM_64_NVIDIA' : 'AL2023_x86_64_NVIDIA';
    } else {
      return architecture === 'arm64' ? 'AL2023_ARM_64_STANDARD' : 'AL2023_x86_64_STANDARD';
    }
  }

  // AL2023 우선
  const al2023 = compatibleAMITypes.find(ami => ami.startsWith('AL2023'));
  if (al2023) return al2023;

  return compatibleAMITypes[0];
}

export function CreateNodeGroupDialog({
  open,
  onOpenChange,
  clusterName,
  cluster,
  defaultRegion,
  defaultCredentialId,
  onSubmit,
  onCredentialIdChange,
  onRegionChange,
  isPending,
  initialData,
}: CreateNodeGroupDialogProps) {
  const [useGPU, setUseGPU] = React.useState(false);
  const isEditMode = !!initialData;
  
  // 클러스터 정보에서 VPC ID와 기본 서브넷 ID 추출
  const clusterVPCId = React.useMemo(() => {
    if (!cluster) return '';
    // AWS 클러스터의 경우 network_config.vpc_id 또는 resources_vpc_config.vpc_id 사용
    if ('resources_vpc_config' in cluster && cluster.resources_vpc_config?.vpc_id) {
      return cluster.resources_vpc_config.vpc_id;
    }
    if (cluster.network_config && 'vpc_id' in cluster.network_config) {
      return cluster.network_config.vpc_id;
    }
    return '';
  }, [cluster]);

  const defaultSubnetIds = React.useMemo(() => {
    // 편집 모드에서는 initialData의 subnet_ids 우선 사용
    if (initialData?.subnet_ids && initialData.subnet_ids.length > 0) {
      return initialData.subnet_ids;
    }
    if (!cluster) return [];
    // AWS 클러스터의 경우 resources_vpc_config.subnet_ids 사용
    if ('resources_vpc_config' in cluster && cluster.resources_vpc_config?.subnet_ids) {
      return cluster.resources_vpc_config.subnet_ids;
    }
    return [];
  }, [cluster, initialData]);

  const [selectedVPCId, setSelectedVPCId] = React.useState<string>(clusterVPCId || '');

  const form = useForm<CreateNodeGroupForm>({
    resolver: zodResolver(createNodeGroupSchema),
    defaultValues: {
      cluster_name: initialData?.cluster_name || clusterName,
      region: initialData?.region || defaultRegion,
      credential_id: initialData?.credential_id || defaultCredentialId,
      name: initialData?.name || '',
      instance_types: initialData?.instance_types || [],
      availability_zone: initialData?.availability_zone || undefined,
      ami_type: initialData?.ami_type,
      disk_size: initialData?.disk_size,
      min_size: initialData?.min_size ?? 1,
      max_size: initialData?.max_size ?? 10,
      desired_size: initialData?.desired_size ?? 1,
      capacity_type: initialData?.capacity_type || 'ON_DEMAND',
      subnet_ids: defaultSubnetIds,
    },
  });

  // initialData가 변경되면 폼 업데이트
  React.useEffect(() => {
    if (initialData && open) {
      form.reset({
        cluster_name: initialData.cluster_name || clusterName,
        region: initialData.region || defaultRegion,
        credential_id: initialData.credential_id || defaultCredentialId,
        name: initialData.name || '',
        instance_types: initialData.instance_types || [],
        availability_zone: initialData.availability_zone || undefined,
        ami_type: initialData.ami_type,
        disk_size: initialData.disk_size,
        min_size: initialData.min_size ?? 1,
        max_size: initialData.max_size ?? 10,
        desired_size: initialData.desired_size ?? 1,
        capacity_type: initialData.capacity_type || 'ON_DEMAND',
        subnet_ids: initialData.subnet_ids || defaultSubnetIds,
      });
    }
  }, [initialData, open, clusterName, defaultRegion, defaultCredentialId, defaultSubnetIds, form]);

  const selectedRegion = form.watch('region');
  const selectedCredentialId = form.watch('credential_id');
  const selectedInstanceType = form.watch('instance_types')?.[0] || null;
  const desiredSize = form.watch('desired_size') || 1;

  // Fetch instance types and AMI types
  const { data: instanceTypes = [], isLoading: isLoadingInstanceTypes } = useInstanceTypes({
    provider: 'aws',
    credentialId: selectedCredentialId,
    region: selectedRegion,
  });

  const { data: amiTypes = [], isLoading: isLoadingAMITypes } = useEKSAmitTypes({
    provider: 'aws',
  });

  // Check if selected instance type is GPU instance
  const selectedInstanceInfo = React.useMemo(() => {
    if (!selectedInstanceType || instanceTypes.length === 0) {
      return null;
    }
    return instanceTypes.find(it => it.instance_type === selectedInstanceType) || null;
  }, [selectedInstanceType, instanceTypes]);

  const isGPUInstance = selectedInstanceInfo?.has_gpu || false;

  // Debounced GPU quota check
  const [debouncedInstanceType, setDebouncedInstanceType] = React.useState<string | null>(null);
  const [debouncedRequiredCount, setDebouncedRequiredCount] = React.useState<number>(1);

  React.useEffect(() => {
    if (!isGPUInstance || !selectedInstanceType || !selectedCredentialId || !selectedRegion) {
      setDebouncedInstanceType(null);
      return;
    }

    const timer = setTimeout(() => {
      setDebouncedInstanceType(selectedInstanceType);
      setDebouncedRequiredCount(desiredSize);
    }, 500); // 500ms debounce

    return () => clearTimeout(timer);
  }, [selectedInstanceType, selectedCredentialId, selectedRegion, desiredSize, isGPUInstance]);

  // GPU quota check query
  const { data: quotaAvailability, isLoading: isLoadingQuota } = useQuery({
    queryKey: queryKeys.kubernetesMetadata.gpuQuota(
      'aws',
      selectedCredentialId,
      selectedRegion,
      debouncedInstanceType || undefined,
      debouncedRequiredCount
    ),
    queryFn: async () => {
      if (!debouncedInstanceType || !selectedCredentialId || !selectedRegion) {
        return null;
      }
      return kubernetesService.checkGPUQuota(
        'aws',
        selectedCredentialId,
        selectedRegion,
        debouncedInstanceType,
        debouncedRequiredCount
      );
    },
    enabled: !!debouncedInstanceType && !!selectedCredentialId && !!selectedRegion && isGPUInstance,
    staleTime: CACHE_TIMES.MONITORING, // 5분 - quota는 자주 변하지 않지만 사용량은 변할 수 있음
    gcTime: CACHE_TIMES.STABLE, // 30분
  });

  // Fetch availability zones for the selected region
  const { data: regionAvailabilityZones = [], isLoading: isLoadingRegionAZs } = useAvailabilityZones({
    provider: 'aws',
    credentialId: selectedCredentialId,
    region: selectedRegion,
  });

  // Instance type offerings query (인스턴스 타입이 사용 가능한 AZ 조회)
  const { 
    data: instanceTypeOfferings, 
    isLoading: isLoadingOfferings,
    isError: isOfferingsError,
    error: offeringsError
  } = useQuery({
    queryKey: queryKeys.kubernetesMetadata.instanceTypeOfferings(
      'aws',
      selectedCredentialId,
      selectedRegion,
      selectedInstanceType || undefined
    ),
    queryFn: async () => {
      if (!selectedInstanceType || !selectedCredentialId || !selectedRegion) {
        return null;
      }
      return kubernetesService.getInstanceTypeOfferings(
        'aws',
        selectedCredentialId,
        selectedRegion,
        selectedInstanceType
      );
    },
    enabled: !!selectedInstanceType && !!selectedCredentialId && !!selectedRegion,
    staleTime: CACHE_TIMES.STABLE, // 24시간 - AZ 가용성은 자주 변하지 않음
    gcTime: CACHE_TIMES.STABLE,
  });

  // 사용 가능한 AZ 목록 (Region AZ와 Instance Type Offerings의 교집합)
  const availableAZs = React.useMemo(() => {
    // 인스턴스 타입이 선택되지 않았으면 빈 배열 반환 (AZ 선택 UI가 표시되지 않음)
    if (!selectedInstanceType) {
      return [];
    }

    // instanceTypeOfferings가 아직 로드되지 않았거나 로딩 중이면 빈 배열 반환
    if (!instanceTypeOfferings) {
      return [];
    }

    // instanceTypeOfferings가 빈 배열이면 해당 인스턴스 타입이 모든 AZ에서 사용 불가능
    if (instanceTypeOfferings.length === 0) {
      return [];
    }

    const instanceTypeAZs = new Set(
      instanceTypeOfferings.map(offering => offering.availability_zone)
    );

    // Region AZ와 Instance Type AZ의 교집합
    return regionAvailabilityZones.filter(az => instanceTypeAZs.has(az));
  }, [regionAvailabilityZones, instanceTypeOfferings, selectedInstanceType]);

  // Fetch subnets for the selected VPC
  const { subnets = [], isLoadingSubnets = false, setSelectedVPCId: setSubnetVPCId } = useNetworkResources({
    resourceType: 'subnets',
    requireVPC: true,
  });

  // Instance type offerings query (AZ 호환성 검증용 - 기존 로직 유지)
  const selectedSubnetIds = form.watch('subnet_ids') || [];
  const selectedSubnets = React.useMemo(() => {
    return subnets.filter(s => selectedSubnetIds.includes(s.id));
  }, [subnets, selectedSubnetIds]);

  const selectedSubnetAZs = React.useMemo(() => {
    return selectedSubnets
      .map(s => s.availability_zone)
      .filter((az): az is string => !!az);
  }, [selectedSubnets]);

  // AZ 호환성 검증
  const azCompatibilityIssues = React.useMemo(() => {
    if (!selectedInstanceType || !instanceTypeOfferings || selectedSubnetAZs.length === 0) {
      return null;
    }

    const availableAZs = new Set(
      instanceTypeOfferings.map(offering => offering.availability_zone)
    );

    const incompatibleAZs = selectedSubnetAZs.filter(az => !availableAZs.has(az));
    const incompatibleSubnets = selectedSubnets.filter(s => 
      s.availability_zone && incompatibleAZs.includes(s.availability_zone)
    );

    if (incompatibleAZs.length > 0) {
      return {
        instanceType: selectedInstanceType,
        incompatibleAZs,
        incompatibleSubnets: incompatibleSubnets.map(s => ({ id: s.id, name: s.name, az: s.availability_zone })),
        availableAZs: Array.from(availableAZs),
      };
    }

    return null;
  }, [selectedInstanceType, instanceTypeOfferings, selectedSubnetAZs, selectedSubnets]);

  // VPC ID가 변경되면 서브넷 목록 업데이트
  React.useEffect(() => {
    if (selectedVPCId && setSubnetVPCId) {
      setSubnetVPCId(selectedVPCId);
    }
  }, [selectedVPCId, setSubnetVPCId]);

  // 클러스터 정보가 로드되면 VPC ID 설정
  React.useEffect(() => {
    if (clusterVPCId && !selectedVPCId) {
      setSelectedVPCId(clusterVPCId);
    }
  }, [clusterVPCId, selectedVPCId]);

  // 기본 서브넷 ID 설정 (다이얼로그가 열릴 때 또는 클러스터 정보가 로드될 때)
  React.useEffect(() => {
    if (defaultSubnetIds.length > 0 && open && selectedVPCId) {
      // 다이얼로그가 열릴 때 기본 서브넷 선택
      const currentSubnetIds = form.getValues('subnet_ids') || [];
      if (currentSubnetIds.length === 0) {
        form.setValue('subnet_ids', defaultSubnetIds);
      }
    }
  }, [defaultSubnetIds, open, selectedVPCId, form]);

  // Filter instance types based on GPU requirement
  const filteredInstanceTypes = React.useMemo(() => {
    if (instanceTypes.length === 0) return [];
    return instanceTypes.filter(it => it.has_gpu === useGPU);
  }, [instanceTypes, useGPU]);

  // Get compatible AMI types
  const compatibleAMITypes = React.useMemo(() => {
    return getCompatibleAMITypes(selectedInstanceType, instanceTypes, amiTypes, useGPU);
  }, [selectedInstanceType, instanceTypes, amiTypes, useGPU]);

  // Get recommended instance type
  const recommendedInstanceType = React.useMemo(() => {
    if (filteredInstanceTypes.length === 0) return null;
    const selectedInstance = instanceTypes.find(it => it.instance_type === selectedInstanceType);
    const architecture = selectedInstance?.architecture || 'x86_64';
    return getRecommendedInstanceType(filteredInstanceTypes, useGPU, architecture);
  }, [filteredInstanceTypes, useGPU, selectedInstanceType, instanceTypes]);

  // Get recommended AMI type
  const recommendedAMIType = React.useMemo(() => {
    if (compatibleAMITypes.length === 0) return null;
    const selectedInstance = instanceTypes.find(it => it.instance_type === selectedInstanceType);
    const architecture = selectedInstance?.architecture || 'x86_64';
    const hasGPU = selectedInstance?.has_gpu || useGPU;
    return getRecommendedAMIType(compatibleAMITypes, hasGPU, architecture);
  }, [compatibleAMITypes, selectedInstanceType, instanceTypes, useGPU]);

  // Auto-select recommended instance type when GPU preference changes
  React.useEffect(() => {
    if (recommendedInstanceType && !selectedInstanceType) {
      form.setValue('instance_types', [recommendedInstanceType]);
    }
  }, [recommendedInstanceType, selectedInstanceType, form]);

  // Auto-select recommended AMI type when instance type changes
  React.useEffect(() => {
    if (recommendedAMIType && selectedInstanceType) {
      form.setValue('ami_type', recommendedAMIType);
    }
  }, [recommendedAMIType, selectedInstanceType, form]);

  // Auto-select first available AZ when instance type changes and no AZ is selected
  React.useEffect(() => {
    if (availableAZs.length > 0 && selectedInstanceType) {
      const currentAZ = form.watch('availability_zone');
      if (!currentAZ) {
        // 첫 번째 사용 가능한 AZ를 자동 선택
        form.setValue('availability_zone', availableAZs[0]);
      } else {
        // 선택된 AZ가 사용 가능한 AZ 목록에 없으면 첫 번째 사용 가능한 AZ로 변경
        if (!availableAZs.includes(currentAZ)) {
          form.setValue('availability_zone', availableAZs[0]);
        }
      }
    }
  }, [availableAZs, selectedInstanceType, form]);

  // Sync form values when props change
  React.useEffect(() => {
    if (defaultCredentialId) {
      form.setValue('credential_id', defaultCredentialId);
      onCredentialIdChange(defaultCredentialId);
    }
  }, [defaultCredentialId]); // eslint-disable-line react-hooks/exhaustive-deps

  React.useEffect(() => {
    if (defaultRegion) {
      form.setValue('region', defaultRegion);
      onRegionChange(defaultRegion);
    }
  }, [defaultRegion]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleSubmit = form.handleSubmit((data) => {
    // Ensure node_group_name is set
    const submitData: CreateNodeGroupForm = {
      ...data,
      name: data.name || `node-group-${Date.now()}`,
    };
    onSubmit(submitData);
    form.reset({
      cluster_name: clusterName,
      region: defaultRegion,
      credential_id: defaultCredentialId,
      instance_types: [],
      availability_zone: undefined,
      min_size: 1,
      max_size: 10,
      desired_size: 1,
      capacity_type: 'ON_DEMAND',
      subnet_ids: defaultSubnetIds,
    });
    setUseGPU(false);
    if (clusterVPCId) {
      setSelectedVPCId(clusterVPCId);
    }
  });

  const handleGPUChange = (checked: boolean) => {
    setUseGPU(checked);
    // Clear instance type and AMI type when GPU preference changes
    form.setValue('instance_types', []);
    form.setValue('ami_type', undefined);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[95vw] sm:max-w-2xl md:max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEditMode ? 'Edit Node Group' : 'Create Node Group'}</DialogTitle>
          <DialogDescription>
            {isEditMode 
              ? 'Update the node group configuration. Some fields cannot be changed after creation.'
              : 'Create a new node group for this cluster. GPU support can be enabled to filter compatible instance types and AMI types.'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* GPU Support Checkbox */}
          <div className="flex items-center space-x-2">
            <Checkbox
              id="use-gpu"
              checked={useGPU}
              onCheckedChange={handleGPUChange}
            />
            <Label htmlFor="use-gpu" className="font-normal cursor-pointer">
              Enable GPU support
            </Label>
            {recommendedInstanceType && (
              <Badge variant="outline" className="ml-2">
                <Sparkles className="mr-1 h-3 w-3" />
                Recommended: {recommendedInstanceType}
              </Badge>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="ng-name">Name *</Label>
              <Input 
                id="ng-name" 
                {...form.register('name')} 
                disabled={isEditMode}
              />
              {isEditMode && (
                <p className="text-xs text-muted-foreground">Node group name cannot be changed after creation</p>
              )}
              {form.formState.errors.name && (
                <p className="text-sm text-red-600">{form.formState.errors.name.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="ng-instance-type">Instance Type *</Label>
              <Select
                value={selectedInstanceType || ''}
                onValueChange={(value) => {
                  form.setValue('instance_types', [value]);
                  // Clear AMI type to trigger auto-selection
                  form.setValue('ami_type', undefined);
                }}
                disabled={isLoadingInstanceTypes || filteredInstanceTypes.length === 0}
              >
                <SelectTrigger id="ng-instance-type" className="w-full">
                  <SelectValue placeholder={isLoadingInstanceTypes ? 'Loading...' : 'Select instance type'} />
                </SelectTrigger>
                <SelectContent className="min-w-[280px] sm:min-w-[400px] max-w-[90vw] sm:max-w-[500px]">
                  {filteredInstanceTypes.map((it) => (
                    <SelectItem key={it.instance_type} value={it.instance_type} className="py-3">
                      <div className="flex flex-col gap-1 w-full">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{it.instance_type}</span>
                          {it.has_gpu && (
                            <Badge variant="secondary" className="text-xs shrink-0">
                              GPU
                            </Badge>
                          )}
                        </div>
                        <span className="text-xs text-muted-foreground">
                          {it.vcpu} vCPU · {Math.round(it.memory_in_mib / 1024)} GB RAM
                          {it.gpu_count && it.gpu_count > 0 && ` · ${it.gpu_count} GPU${it.gpu_count > 1 ? 's' : ''}`}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {form.formState.errors.instance_types && (
                <p className="text-sm text-red-600">{form.formState.errors.instance_types.message}</p>
              )}
              {/* GPU Quota Warning */}
              {isGPUInstance && selectedInstanceType && (
                <div className="mt-2">
                  {isLoadingQuota ? (
                    <p className="text-xs text-muted-foreground">GPU quota 확인 중...</p>
                  ) : quotaAvailability && quotaAvailability.quota_insufficient ? (
                    <Alert variant="destructive" className="py-2">
                      <AlertCircle className="h-3 w-3" />
                      <AlertDescription className="text-xs">
                        <div className="space-y-1">
                          <p className="font-medium">{quotaAvailability.message || 'GPU quota가 부족합니다.'}</p>
                          <div className="flex flex-wrap gap-1 text-xs">
                            <span>사용 가능: {Math.floor(quotaAvailability.available_quota)}</span>
                            <span>·</span>
                            <span>필요: {quotaAvailability.required_count}</span>
                            {quotaAvailability.current_usage !== undefined && (
                              <>
                                <span>·</span>
                                <span>현재 사용량: {Math.floor(quotaAvailability.current_usage)}</span>
                              </>
                            )}
                          </div>
                        </div>
                      </AlertDescription>
                    </Alert>
                  ) : quotaAvailability && !quotaAvailability.quota_insufficient ? (
                    <p className="text-xs text-green-600">
                      ✓ GPU quota 사용 가능 (사용 가능: {Math.floor(quotaAvailability.available_quota)}, 필요: {quotaAvailability.required_count})
                    </p>
                  ) : null}
                </div>
              )}
            </div>
          </div>

          {/* Availability Zone Single Select */}
          {selectedInstanceType && (
            <div className="space-y-2">
              <Label htmlFor="ng-availability-zone">Availability Zone *</Label>
              {isLoadingRegionAZs || isLoadingOfferings ? (
                <div className="text-sm text-muted-foreground">Loading availability zones...</div>
              ) : isOfferingsError ? (
                <div className="space-y-2">
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>
                      <div className="space-y-1">
                        <p className="font-medium">Availability Zone 조회 실패</p>
                        <p className="text-sm">
                          선택한 인스턴스 타입({selectedInstanceType})의 사용 가능한 Availability Zone을 조회하는 중 오류가 발생했습니다.
                        </p>
                        {offeringsError && (
                          <p className="text-xs text-muted-foreground mt-1">
                            {offeringsError instanceof Error ? offeringsError.message : String(offeringsError)}
                          </p>
                        )}
                        <p className="text-xs text-muted-foreground mt-2">
                          IAM 권한을 확인하거나 잠시 후 다시 시도해주세요.
                        </p>
                      </div>
                    </AlertDescription>
                  </Alert>
                </div>
              ) : availableAZs.length === 0 ? (
                <div className="space-y-2">
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>
                      <div className="space-y-1">
                        <p className="font-medium">사용 가능한 Availability Zone 없음</p>
                        <p className="text-sm">
                          선택한 인스턴스 타입({selectedInstanceType})이 이 region({selectedRegion})에서 사용 가능한 Availability Zone이 없습니다.
                        </p>
                        <p className="text-xs text-muted-foreground mt-2">
                          다른 인스턴스 타입을 선택하거나 다른 region을 선택해주세요.
                        </p>
                      </div>
                    </AlertDescription>
                  </Alert>
                </div>
              ) : (
                <Select
                  value={form.watch('availability_zone') || ''}
                  onValueChange={(value) => form.setValue('availability_zone', value)}
                  disabled={isLoadingRegionAZs || isLoadingOfferings || availableAZs.length === 0}
                >
                  <SelectTrigger id="ng-availability-zone" className="w-full">
                    <SelectValue placeholder="Select availability zone" />
                  </SelectTrigger>
                  <SelectContent>
                    {availableAZs.map((az) => (
                      <SelectItem key={az} value={az} className="font-mono">
                        {az}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
              {form.formState.errors.availability_zone && (
                <p className="text-sm text-red-600">{form.formState.errors.availability_zone.message}</p>
              )}
              {availableAZs.length > 0 && !isOfferingsError && (
                <p className="text-xs text-muted-foreground">
                  선택한 인스턴스 타입({selectedInstanceType})이 사용 가능한 Availability Zone입니다.
                </p>
              )}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="ng-ami-type">AMI Type</Label>
            <Select
              value={form.watch('ami_type') || ''}
              onValueChange={(value) => form.setValue('ami_type', value)}
              disabled={isLoadingAMITypes || compatibleAMITypes.length === 0 || !selectedInstanceType}
            >
              <SelectTrigger id="ng-ami-type" className="w-full">
                <SelectValue placeholder={isLoadingAMITypes ? 'Loading...' : compatibleAMITypes.length === 0 ? 'No compatible AMI types' : 'Select AMI type'} />
              </SelectTrigger>
              <SelectContent className="min-w-[280px] sm:min-w-[400px] max-w-[90vw] sm:max-w-[500px]">
                {compatibleAMITypes.map((amiType) => (
                  <SelectItem key={amiType} value={amiType} className="py-2">
                    <div className="flex items-center gap-2 w-full">
                      <span className="truncate">{amiType}</span>
                      {recommendedAMIType === amiType && (
                        <Badge variant="outline" className="shrink-0 text-xs">
                          <Sparkles className="mr-1 h-3 w-3" />
                          Recommended
                        </Badge>
                      )}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {form.formState.errors.ami_type && (
              <p className="text-sm text-red-600">{form.formState.errors.ami_type.message}</p>
            )}
          </div>

          {/* VPC Display (Read-only, from cluster) */}
          {clusterVPCId && (
            <div className="space-y-2">
              <Label>VPC ID</Label>
              <Input
                value={clusterVPCId}
                disabled
                className="bg-muted"
              />
              <p className="text-xs text-muted-foreground">
                VPC ID는 클러스터의 네트워크 설정에서 자동으로 가져옵니다.
              </p>
            </div>
          )}

          {/* Subnet Multi-Select */}
          <div className="space-y-2">
            <Label htmlFor="ng-subnets">Subnets *</Label>
            {!selectedVPCId ? (
              <div className="text-sm text-muted-foreground">
                VPC ID를 불러오는 중...
              </div>
            ) : (
              <div className="space-y-2">
                <div className="border rounded-md p-2 sm:p-3 max-h-[200px] overflow-y-auto">
                  {isLoadingSubnets ? (
                    <div className="text-sm text-muted-foreground text-center py-4">서브넷 목록을 불러오는 중...</div>
                  ) : subnets.length === 0 ? (
                    <div className="text-sm text-muted-foreground text-center py-4">선택 가능한 서브넷이 없습니다.</div>
                  ) : (
                    <div className="space-y-2">
                      {subnets.map((subnet) => {
                        const currentSelectedSubnetIds = form.watch('subnet_ids') || [];
                        const isSelected = currentSelectedSubnetIds.includes(subnet.id);
                        
                        // AZ 호환성 확인
                        const isIncompatible = selectedInstanceType && instanceTypeOfferings && subnet.availability_zone
                          ? !instanceTypeOfferings.some(offering => offering.availability_zone === subnet.availability_zone)
                          : false;

                        return (
                          <div 
                            key={subnet.id} 
                            className={`flex items-start sm:items-center space-x-2 p-2 rounded-md transition-colors ${
                              isIncompatible && selectedInstanceType
                                ? 'bg-red-50 border border-red-200'
                                : 'hover:bg-muted/50'
                            }`}
                          >
                            <Checkbox
                              id={`subnet-${subnet.id}`}
                              checked={isSelected}
                              disabled={!!(isIncompatible && selectedInstanceType)}
                              onCheckedChange={(checked) => {
                                const currentIds = form.getValues('subnet_ids') || [];
                                if (checked) {
                                  form.setValue('subnet_ids', [...currentIds, subnet.id]);
                                } else {
                                  form.setValue('subnet_ids', currentIds.filter(id => id !== subnet.id));
                                }
                              }}
                              className="mt-1 sm:mt-0"
                            />
                            <Label
                              htmlFor={`subnet-${subnet.id}`}
                              className={`flex-1 cursor-pointer font-normal ${
                                isIncompatible && selectedInstanceType ? 'opacity-60' : ''
                              }`}
                            >
                              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 sm:gap-0">
                                <div className="flex items-center gap-2 flex-wrap">
                                  <span className="font-mono text-xs sm:text-sm break-all">{subnet.id}</span>
                                  {subnet.name && (
                                    <span className="text-xs text-muted-foreground">{subnet.name}</span>
                                  )}
                                </div>
                                {subnet.availability_zone && (
                                  <Badge 
                                    variant={isIncompatible && selectedInstanceType ? "destructive" : "outline"} 
                                    className="text-xs shrink-0 w-fit"
                                  >
                                    {subnet.availability_zone}
                                    {isIncompatible && selectedInstanceType && (
                                      <span className="ml-1">⚠️</span>
                                    )}
                                  </Badge>
                                )}
                              </div>
                              {subnet.cidr_block && (
                                <span className="text-xs text-muted-foreground block mt-1">
                                  {subnet.cidr_block}
                                </span>
                              )}
                              {isIncompatible && selectedInstanceType && (
                                <span className="text-xs text-red-600 block mt-1">
                                  {selectedInstanceType} 인스턴스 타입이 이 AZ에서 지원되지 않습니다.
                                </span>
                              )}
                            </Label>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
                <p className="text-xs text-muted-foreground">
                  최소 2개의 서브넷을 선택해야 하며, 서로 다른 Availability Zone에 있어야 합니다.
                </p>
                {azCompatibilityIssues && (
                  <Alert variant="destructive" className="mt-2">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>
                      <div className="space-y-1">
                        <p className="font-medium">
                          선택한 인스턴스 타입({azCompatibilityIssues.instanceType})이 다음 Availability Zone에서 지원되지 않습니다:
                        </p>
                        <ul className="list-disc list-inside text-sm space-y-1">
                          {azCompatibilityIssues.incompatibleSubnets.map(subnet => (
                            <li key={subnet.id}>
                              {subnet.name || subnet.id} ({subnet.az})
                            </li>
                          ))}
                        </ul>
                        {azCompatibilityIssues.availableAZs.length > 0 && (
                          <p className="text-sm mt-2">
                            사용 가능한 AZ: {azCompatibilityIssues.availableAZs.join(', ')}
                          </p>
                        )}
                      </div>
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            )}
            {form.formState.errors.subnet_ids && (
              <p className="text-sm text-red-600">{form.formState.errors.subnet_ids.message}</p>
            )}
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="ng-min-size">Min Size</Label>
              <Input
                id="ng-min-size"
                type="number"
                {...form.register('min_size', { valueAsNumber: true })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ng-max-size">Max Size</Label>
              <Input
                id="ng-max-size"
                type="number"
                {...form.register('max_size', { valueAsNumber: true })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ng-desired-size">Desired Size</Label>
              <Input
                id="ng-desired-size"
                type="number"
                {...form.register('desired_size', { valueAsNumber: true })}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="ng-disk-size">Disk Size (GB)</Label>
              <Input
                id="ng-disk-size"
                type="number"
                {...form.register('disk_size', { valueAsNumber: true })}
                placeholder="20"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ng-capacity-type">Capacity Type</Label>
              <Select
                value={form.watch('capacity_type') || 'ON_DEMAND'}
                onValueChange={(value) => form.setValue('capacity_type', value as 'ON_DEMAND' | 'SPOT')}
              >
                <SelectTrigger id="ng-capacity-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ON_DEMAND">On-Demand</SelectItem>
                  <SelectItem value="SPOT">Spot</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex flex-col-reverse sm:flex-row justify-end gap-2 sm:space-x-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isPending} className="w-full sm:w-auto">
              Cancel
            </Button>
            <Button type="submit" disabled={isPending} className="w-full sm:w-auto">
              {isPending ? 'Creating...' : 'Create Node Group'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

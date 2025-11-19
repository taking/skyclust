/**
 * Azure Advanced Configuration Component
 * Azure AKS 고급 설정: Node Pool Config, Security Config
 */

'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { FormField, FormItem, FormLabel, FormControl, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Search, X, Sparkles } from 'lucide-react';
import type { CreateClusterForm } from '@/lib/types';
import { useAzureVMSizes } from '@/features/kubernetes/hooks/use-kubernetes-metadata';
import { DataProcessor } from '@/lib/data';

interface AzureAdvancedConfigProps {
  form: UseFormReturn<CreateClusterForm>;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
  deploymentMode?: 'auto' | 'custom';
}

// 추천 VM Size 목록 (일반적으로 사용되는 VM Size 우선순위)
const RECOMMENDED_VM_SIZES = [
  'Standard_D2s_v3',
  'Standard_D4s_v3',
  'Standard_D8s_v3',
  'Standard_DS2_v2',
  'Standard_DS3_v2',
  'Standard_DS4_v2',
  'Standard_B2s',
  'Standard_B4ms',
  'Standard_B8ms',
];

// Azure Node Pool 기본값 상수
const DEFAULT_AZURE_NODE_POOL: NonNullable<CreateClusterForm['node_pool']> = {
  name: 'nodepool1',
  vm_size: '',
  node_count: 3,
  min_count: 1,
  max_count: 10,
  enable_auto_scaling: false,
  os_disk_size_gb: 128,
  os_disk_type: 'Managed',
  os_type: 'Linux',
  os_sku: 'Ubuntu',
  max_pods: 30,
  mode: 'System',
};

// Azure Security 기본값 상수
const DEFAULT_AZURE_SECURITY: NonNullable<CreateClusterForm['security']> = {
  enable_rbac: true,
  enable_pod_security_policy: false,
  enable_private_cluster: false,
  enable_azure_policy: false,
  enable_workload_identity: false,
};

// VM Size 추천 함수
function getRecommendedVMSize(vmSizes: string[]): string | null {
  if (vmSizes.length === 0) return null;
  
  // 추천 목록에서 사용 가능한 첫 번째 VM Size 찾기
  for (const recommended of RECOMMENDED_VM_SIZES) {
    if (vmSizes.includes(recommended)) {
      return recommended;
    }
  }
  
  // 추천 목록에 없으면 Standard_로 시작하는 첫 번째 VM Size
  const standardSize = vmSizes.find(size => size.startsWith('Standard_'));
  if (standardSize) return standardSize;
  
  // 그 외에는 첫 번째 VM Size
  return vmSizes[0];
}

export function AzureAdvancedConfig({
  form,
  onDataChange,
  deploymentMode = 'auto',
}: AzureAdvancedConfigProps) {
  // 자동 모드일 때만 Node Pool Configuration 표시
  const showNodePoolConfig = deploymentMode === 'auto';
  
  // Form에서 credentialId와 region 가져오기
  const credentialId = form.watch('credential_id');
  const region = form.watch('region') || form.watch('location') || '';
  
  // 검색 상태
  const [vmSizeSearchQuery, setVmSizeSearchQuery] = useState('');
  const [vmSizeSelectOpen, setVmSizeSelectOpen] = useState(false);
  
  // Azure VM Sizes 조회
  const {
    data: vmSizes = [],
    isLoading: isLoadingVMSizes,
    error: vmSizesError,
  } = useAzureVMSizes({
    provider: 'azure',
    credentialId: credentialId || '',
    region,
  });
  
  // 검색 필터링된 VM Sizes
  const filteredVMSizes = useMemo(() => {
    if (!vmSizeSearchQuery.trim()) {
      return vmSizes;
    }
    return DataProcessor.search(vmSizes, vmSizeSearchQuery, {
      keys: [], // 문자열 배열이므로 keys 없이 직접 검색
      threshold: 0.3,
    });
  }, [vmSizes, vmSizeSearchQuery]);
  
  // 추천 VM Size 계산
  const recommendedVMSize = useMemo(() => {
    return getRecommendedVMSize(vmSizes);
  }, [vmSizes]);
  
  // Node Pool 기본값 업데이트 함수
  const updateNodePool = useCallback((updates: Partial<NonNullable<CreateClusterForm['node_pool']>>) => {
    const current = form.getValues('node_pool') || {};
    const updated: NonNullable<CreateClusterForm['node_pool']> = {
      ...DEFAULT_AZURE_NODE_POOL,
      ...current,
      ...updates,
    };
    form.setValue('node_pool', updated);
    onDataChange({ node_pool: updated });
  }, [form, onDataChange]);
  
  // Security 기본값 업데이트 함수
  const updateSecurity = useCallback((updates: Partial<NonNullable<CreateClusterForm['security']>>) => {
    const current = form.getValues('security') || {};
    const updated: NonNullable<CreateClusterForm['security']> = {
      ...DEFAULT_AZURE_SECURITY,
      ...current,
      ...updates,
    };
    form.setValue('security', updated);
    onDataChange({ security: updated });
  }, [form, onDataChange]);
  
  // Node Pool 기본값 초기화 (자동 모드일 때만)
  useEffect(() => {
    if (!showNodePoolConfig) return;
    
    const currentNodePool = form.getValues('node_pool');
    const needsInitialization = !currentNodePool || !currentNodePool.name;
    
    if (needsInitialization) {
      const initialNodePool: NonNullable<CreateClusterForm['node_pool']> = {
        ...DEFAULT_AZURE_NODE_POOL,
        ...currentNodePool,
        // VM Size는 추천값이 있으면 사용, 없으면 빈 값 유지
        vm_size: currentNodePool?.vm_size || recommendedVMSize || '',
      };
      form.setValue('node_pool', initialNodePool);
      onDataChange({ node_pool: initialNodePool });
    } else if (recommendedVMSize && !currentNodePool.vm_size && !isLoadingVMSizes) {
      // VM Size만 업데이트 (다른 필드는 유지)
      updateNodePool({ vm_size: recommendedVMSize });
    }
  }, [showNodePoolConfig, recommendedVMSize, isLoadingVMSizes, form, onDataChange, updateNodePool]);
  
  // Security 기본값 초기화
  useEffect(() => {
    const currentSecurity = form.getValues('security');
    if (!currentSecurity || Object.keys(currentSecurity).length === 0) {
      form.setValue('security', DEFAULT_AZURE_SECURITY);
      onDataChange({ security: DEFAULT_AZURE_SECURITY });
    }
  }, [form, onDataChange]);
  
  return (
    <>
      {/* Node Pool Configuration - 자동 모드일 때만 표시 */}
      {showNodePoolConfig && (
        <div className="space-y-6 mt-6 pt-6 border-t">
          <h3 className="text-lg font-semibold">Azure Node Pool Configuration</h3>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="node_pool.name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Node Pool Name *</FormLabel>
                <FormControl>
                  <Input
                    placeholder="nodepool1"
                    value={field.value || ''}
                    onChange={(e) => {
                      field.onChange(e.target.value);
                      updateNodePool({ name: e.target.value });
                    }}
                  />
                </FormControl>
                <FormDescription>
                  Name for the node pool
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.vm_size"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="flex items-center gap-2">
                  VM Size *
                  {recommendedVMSize && recommendedVMSize === field.value && (
                    <span className="flex items-center gap-1 text-xs text-muted-foreground">
                      <Sparkles className="h-3 w-3" />
                      Recommended
                    </span>
                  )}
                </FormLabel>
                <FormControl>
                  <Select
                    value={field.value || ''}
                    onValueChange={(value) => {
                      field.onChange(value);
                      updateNodePool({ vm_size: value });
                    }}
                    disabled={isLoadingVMSizes || !credentialId || !region}
                    open={vmSizeSelectOpen}
                    onOpenChange={setVmSizeSelectOpen}
                  >
                    <SelectTrigger>
                      <SelectValue 
                        placeholder={
                          isLoadingVMSizes 
                            ? 'Loading VM sizes...' 
                            : !credentialId || !region
                            ? 'Select credential and region first'
                            : 'Select VM size'
                        } 
                      />
                    </SelectTrigger>
                    <SelectContent className="p-0" onCloseAutoFocus={(e) => e.preventDefault()}>
                      {/* 검색 입력 필드 */}
                      <div className="p-2 border-b sticky top-0 bg-background z-10">
                        <div className="relative">
                          <Search className="absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                          <Input
                            placeholder="Search VM sizes..."
                            value={vmSizeSearchQuery}
                            onChange={(e) => setVmSizeSearchQuery(e.target.value)}
                            className="pl-8 pr-8 h-8"
                            onClick={(e) => e.stopPropagation()}
                            onKeyDown={(e) => e.stopPropagation()}
                          />
                          {vmSizeSearchQuery && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6 p-0"
                              onClick={(e) => {
                                e.stopPropagation();
                                setVmSizeSearchQuery('');
                              }}
                            >
                              <X className="h-3 w-3" />
                            </Button>
                          )}
                        </div>
                      </div>

                      {/* 로딩 상태 */}
                      {isLoadingVMSizes && (
                        <div className="p-4 text-center text-sm text-muted-foreground">
                          Loading VM sizes...
                        </div>
                      )}

                      {/* 빈 상태 */}
                      {!isLoadingVMSizes && filteredVMSizes.length === 0 && (
                        <div className="p-4 text-center text-sm text-muted-foreground">
                          {vmSizeSearchQuery ? 'No VM sizes found' : 'No VM sizes available'}
                        </div>
                      )}

                      {/* 리스트 */}
                      {!isLoadingVMSizes && filteredVMSizes.length > 0 && (
                        <div className="max-h-[300px] overflow-y-auto">
                          {filteredVMSizes.map((vmSize) => {
                            const isRecommended = vmSize === recommendedVMSize;
                            return (
                              <SelectItem key={vmSize} value={vmSize}>
                                <div className="flex items-center gap-2">
                                  <span>{vmSize}</span>
                                  {isRecommended && (
                                    <span className="flex items-center gap-1 text-xs text-muted-foreground">
                                      <Sparkles className="h-3 w-3" />
                                      Recommended
                                    </span>
                                  )}
                                </div>
                              </SelectItem>
                            );
                          })}
                        </div>
                      )}
                    </SelectContent>
                  </Select>
                </FormControl>
                <FormDescription>
                  Azure VM size (e.g., Standard_D2s_v3)
                  {recommendedVMSize && (
                    <span className="ml-2 text-xs text-muted-foreground">
                      • Recommended: {recommendedVMSize}
                    </span>
                  )}
                </FormDescription>
                {vmSizesError && (
                  <FormDescription className="text-destructive mt-1">
                    Failed to load VM sizes: {vmSizesError.message}
                  </FormDescription>
                )}
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <FormField
            control={form.control}
            name="node_pool.node_count"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Node Count *</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="1"
                    placeholder="3"
                    value={field.value ?? DEFAULT_AZURE_NODE_POOL.node_count}
                    onChange={(e) => {
                      const value = parseInt(e.target.value, 10) || DEFAULT_AZURE_NODE_POOL.node_count;
                      field.onChange(value);
                      updateNodePool({ node_count: value });
                    }}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.min_count"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Min Count</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="0"
                    placeholder="1"
                    value={field.value ?? DEFAULT_AZURE_NODE_POOL.min_count}
                    onChange={(e) => {
                      const value = parseInt(e.target.value, 10) || DEFAULT_AZURE_NODE_POOL.min_count;
                      field.onChange(value);
                      updateNodePool({ min_count: value });
                    }}
                  />
                </FormControl>
                <FormDescription>Minimum nodes (for auto-scaling)</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.max_count"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Max Count</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="1"
                    placeholder="10"
                    value={field.value ?? DEFAULT_AZURE_NODE_POOL.max_count}
                    onChange={(e) => {
                      const value = parseInt(e.target.value, 10) || DEFAULT_AZURE_NODE_POOL.max_count;
                      field.onChange(value);
                      updateNodePool({ max_count: value });
                    }}
                  />
                </FormControl>
                <FormDescription>Maximum nodes (for auto-scaling)</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <FormField
          control={form.control}
          name="node_pool.enable_auto_scaling"
          render={({ field }) => (
            <div className="flex items-center space-x-2">
              <Checkbox
                id="enable-auto-scaling"
                checked={field.value ?? DEFAULT_AZURE_NODE_POOL.enable_auto_scaling}
                onCheckedChange={(checked) => {
                  const checkedValue = checked === true;
                  field.onChange(checkedValue);
                  updateNodePool({ enable_auto_scaling: checkedValue });
                }}
              />
              <Label
                htmlFor="enable-auto-scaling"
                className="text-sm font-normal cursor-pointer"
              >
                Enable Auto Scaling
              </Label>
            </div>
          )}
        />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
          <FormField
            control={form.control}
            name="node_pool.os_disk_size_gb"
            render={({ field }) => (
              <FormItem>
                <FormLabel>OS Disk Size (GB)</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="30"
                    placeholder="128"
                    value={field.value ?? DEFAULT_AZURE_NODE_POOL.os_disk_size_gb}
                    onChange={(e) => {
                      const value = parseInt(e.target.value, 10) || DEFAULT_AZURE_NODE_POOL.os_disk_size_gb;
                      field.onChange(value);
                      updateNodePool({ os_disk_size_gb: value });
                    }}
                  />
                </FormControl>
                <FormDescription>OS disk size in GB</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.os_disk_type"
            render={({ field }) => (
              <FormItem>
                <FormLabel>OS Disk Type</FormLabel>
                <Select
                  value={field.value || DEFAULT_AZURE_NODE_POOL.os_disk_type}
                  onValueChange={(value) => {
                    field.onChange(value);
                    updateNodePool({ os_disk_type: value });
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select disk type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="Managed">Managed</SelectItem>
                    <SelectItem value="Ephemeral">Ephemeral</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>OS disk type</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.os_type"
            render={({ field }) => (
              <FormItem>
                <FormLabel>OS Type</FormLabel>
                <Select
                  value={field.value || DEFAULT_AZURE_NODE_POOL.os_type}
                  onValueChange={(value) => {
                    field.onChange(value);
                    updateNodePool({ os_type: value });
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select OS type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="Linux">Linux</SelectItem>
                    <SelectItem value="Windows">Windows</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>Operating system type</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.os_sku"
            render={({ field }) => (
              <FormItem>
                <FormLabel>OS SKU</FormLabel>
                <Select
                  value={field.value || DEFAULT_AZURE_NODE_POOL.os_sku}
                  onValueChange={(value) => {
                    field.onChange(value);
                    updateNodePool({ os_sku: value });
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select OS SKU" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="Ubuntu">Ubuntu</SelectItem>
                    <SelectItem value="CBLMariner">CBLMariner</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>OS SKU</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.max_pods"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Max Pods</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="10"
                    placeholder="30"
                    value={field.value ?? DEFAULT_AZURE_NODE_POOL.max_pods}
                    onChange={(e) => {
                      const value = parseInt(e.target.value, 10) || DEFAULT_AZURE_NODE_POOL.max_pods;
                      field.onChange(value);
                      updateNodePool({ max_pods: value });
                    }}
                  />
                </FormControl>
                <FormDescription>Maximum number of pods per node</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.mode"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Mode</FormLabel>
                <Select
                  value={field.value || DEFAULT_AZURE_NODE_POOL.mode}
                  onValueChange={(value) => {
                    field.onChange(value);
                    updateNodePool({ mode: value });
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select mode" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="System">System</SelectItem>
                    <SelectItem value="User">User</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>Node pool mode</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        </div>
      )}

      {/* Security Configuration */}
      <div className="space-y-6 mt-6 pt-6 border-t">
        <h3 className="text-lg font-semibold">Azure Security Configuration</h3>
        
        <div className="space-y-4">
          <FormField
            control={form.control}
            name="security.enable_rbac"
            render={({ field }) => (
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="enable-rbac"
                  checked={field.value ?? DEFAULT_AZURE_SECURITY.enable_rbac}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    updateSecurity({ enable_rbac: checkedValue });
                  }}
                />
                <Label
                  htmlFor="enable-rbac"
                  className="text-sm font-normal cursor-pointer"
                >
                  Enable RBAC
                </Label>
              </div>
            )}
          />

          <FormField
            control={form.control}
            name="security.enable_pod_security_policy"
            render={({ field }) => (
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="enable-pod-security-policy"
                  checked={field.value ?? DEFAULT_AZURE_SECURITY.enable_pod_security_policy}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    updateSecurity({ enable_pod_security_policy: checkedValue });
                  }}
                />
                <Label
                  htmlFor="enable-pod-security-policy"
                  className="text-sm font-normal cursor-pointer"
                >
                  Enable Pod Security Policy
                </Label>
              </div>
            )}
          />

          <FormField
            control={form.control}
            name="security.enable_private_cluster"
            render={({ field }) => (
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="enable-private-cluster"
                  checked={field.value ?? DEFAULT_AZURE_SECURITY.enable_private_cluster}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    updateSecurity({ enable_private_cluster: checkedValue });
                  }}
                />
                <Label
                  htmlFor="enable-private-cluster"
                  className="text-sm font-normal cursor-pointer"
                >
                  Enable Private Cluster
                </Label>
              </div>
            )}
          />

          <FormField
            control={form.control}
            name="security.enable_azure_policy"
            render={({ field }) => (
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="enable-azure-policy"
                  checked={field.value ?? DEFAULT_AZURE_SECURITY.enable_azure_policy}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    updateSecurity({ enable_azure_policy: checkedValue });
                  }}
                />
                <Label
                  htmlFor="enable-azure-policy"
                  className="text-sm font-normal cursor-pointer"
                >
                  Enable Azure Policy
                </Label>
              </div>
            )}
          />

          <FormField
            control={form.control}
            name="security.enable_workload_identity"
            render={({ field }) => (
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="enable-workload-identity-azure"
                  checked={field.value ?? DEFAULT_AZURE_SECURITY.enable_workload_identity}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    updateSecurity({ enable_workload_identity: checkedValue });
                  }}
                />
                <Label
                  htmlFor="enable-workload-identity-azure"
                  className="text-sm font-normal cursor-pointer"
                >
                  Enable Workload Identity
                </Label>
              </div>
            )}
          />
        </div>
      </div>
    </>
  );
}


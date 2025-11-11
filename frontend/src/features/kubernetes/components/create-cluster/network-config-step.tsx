/**
 * Network Configuration Step
 * Step 2: VPC 및 Subnet 선택
 * 
 * 레이아웃:
 * 1row 1column 방식으로 세로 나열
 * - VPC
 * - Subnets
 */

'use client';

import { useState, useEffect } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { useVPCs } from '@/features/networks/hooks/use-vpcs';
import { useSubnets } from '@/features/networks/hooks/use-subnets';
import { CreateVPCDialog } from '@/features/networks/components/create-vpc-dialog';
import { CreateSubnetDialog } from '@/features/networks/components/create-subnet-dialog';
import { RefreshCw, Plus, X } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import type { CreateClusterForm, CloudProvider, CreateVPCForm, CreateSubnetForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface NetworkConfigStepProps {
  form: UseFormReturn<CreateClusterForm>;
  selectedProvider?: CloudProvider;
  selectedCredentialId: string;
  selectedRegion: string;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
}

export function NetworkConfigStep({
  form,
  selectedProvider,
  selectedCredentialId,
  selectedRegion,
  onDataChange,
}: NetworkConfigStepProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [selectedVPCId, setSelectedVPCId] = useState<string>('');
  const [isCreateVPCDialogOpen, setIsCreateVPCDialogOpen] = useState(false);
  const [isCreateSubnetDialogOpen, setIsCreateSubnetDialogOpen] = useState(false);

  const { vpcs, isLoadingVPCs } = useVPCs();
  const { subnets, isLoadingSubnets, setSelectedVPCId: setSubnetVPCId } = useSubnets();

  // VPC 목록 새로고침
  const handleRefreshVPCs = () => {
    if (selectedProvider && selectedCredentialId) {
      queryClient.invalidateQueries({
        queryKey: queryKeys.vpcs.list(selectedProvider, selectedCredentialId, selectedRegion),
      });
    }
  };

  // Subnet 목록 새로고침
  const handleRefreshSubnets = () => {
    if (selectedProvider && selectedCredentialId && selectedVPCId) {
      queryClient.invalidateQueries({
        queryKey: queryKeys.subnets.list(selectedProvider, selectedCredentialId, selectedVPCId, selectedRegion),
      });
    }
  };

  // Subnet actions는 create-subnet-dialog에서 직접 처리

  // Form에서 VPC ID 가져오기 (초기값)
  const formVPCId = form.watch('vpc_id');
  useEffect(() => {
    if (formVPCId && !selectedVPCId) {
      setSelectedVPCId(formVPCId);
      setSubnetVPCId(formVPCId);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formVPCId, selectedVPCId]);

  // VPC 선택 시 Subnet 목록 로드
  const handleVPCChange = (vpcId: string) => {
    setSelectedVPCId(vpcId);
    setSubnetVPCId(vpcId);
    form.setValue('vpc_id', vpcId);
    form.setValue('subnet_ids', []); // VPC 변경 시 Subnet 초기화
    
    // Azure의 경우 network 객체에 virtual_network_id 설정
    if (selectedProvider === 'azure') {
      const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
      const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
        virtual_network_id: vpcId,
        subnet_id: currentNetwork?.subnet_id || '',
        network_plugin: currentNetwork?.network_plugin,
        network_policy: currentNetwork?.network_policy,
        pod_cidr: currentNetwork?.pod_cidr,
        service_cidr: currentNetwork?.service_cidr,
        dns_service_ip: currentNetwork?.dns_service_ip,
        docker_bridge_cidr: currentNetwork?.docker_bridge_cidr,
      };
      form.setValue('network', updatedNetwork);
      onDataChange({ 
        vpc_id: vpcId, 
        subnet_ids: [],
        network: updatedNetwork,
      });
    } else {
      onDataChange({ vpc_id: vpcId, subnet_ids: [] });
    }
  };

  // VPC 생성 성공 핸들러 (모달에서 호출)
  const handleVPCCreated = (vpcId: string) => {
    // VPC 생성 후 자동 선택 및 목록 갱신
    setTimeout(() => {
      handleVPCChange(vpcId);
      handleRefreshVPCs();
    }, 500);
  };

  // Subnet 생성 성공 핸들러 (모달에서 호출)
  const handleSubnetCreated = (subnetId: string) => {
    // Subnet 생성 후 자동 선택
    setTimeout(() => {
      handleRefreshSubnets();
      // 생성된 Subnet을 자동으로 선택 목록에 추가
      const currentIds = form.getValues('subnet_ids') || [];
      if (!currentIds.includes(subnetId)) {
        handleSubnetToggle(subnetId, true);
      }
    }, 500);
  };

  // Subnet 선택 (Multi-select with checkbox)
  const selectedSubnetIds = form.watch('subnet_ids') || [];
  const handleSubnetToggle = (subnetId: string, checked: boolean) => {
    let newIds: string[];
    
    if (checked) {
      // 체크박스 선택
      if (selectedProvider === 'gcp' || selectedProvider === 'azure') {
        // GCP/Azure는 단일 선택만 허용 (기존 선택 해제 후 새로 선택)
        newIds = [subnetId];
      } else {
        // AWS는 다중 선택 가능
        newIds = selectedSubnetIds.includes(subnetId)
          ? selectedSubnetIds
          : [...selectedSubnetIds, subnetId];
        
        // AWS EKS는 최소 2개의 다른 AZ에 서브넷이 필요
        if (selectedProvider === 'aws' && newIds.length > 0) {
          const selectedSubnetsData = newIds
            .map(id => subnets.find(s => s.id === id))
            .filter(Boolean);
          
          const uniqueAZs = new Set(
            selectedSubnetsData
              .map(s => s?.availability_zone)
              .filter(Boolean)
          );
          
          // 최소 2개의 다른 AZ가 필요하지만, 사용자가 선택하는 동안은 경고만 표시
          // 실제 validation은 form submit 시점에 수행
        }
      }
    } else {
      // 체크박스 해제
      newIds = selectedSubnetIds.filter(id => id !== subnetId);
    }
    
    form.setValue('subnet_ids', newIds);
    
    // Azure의 경우 network 객체에 subnet_id 설정 (첫 번째 subnet 사용)
    if (selectedProvider === 'azure' && newIds.length > 0) {
      const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
      const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
        virtual_network_id: currentNetwork?.virtual_network_id || selectedVPCId || '',
        subnet_id: newIds[0],
        network_plugin: currentNetwork?.network_plugin,
        network_policy: currentNetwork?.network_policy,
        pod_cidr: currentNetwork?.pod_cidr,
        service_cidr: currentNetwork?.service_cidr,
        dns_service_ip: currentNetwork?.dns_service_ip,
        docker_bridge_cidr: currentNetwork?.docker_bridge_cidr,
      };
      form.setValue('network', updatedNetwork);
      onDataChange({ 
        subnet_ids: newIds,
        network: updatedNetwork,
      });
    } else {
      onDataChange({ subnet_ids: newIds });
    }
  };

  const handleRemoveSubnet = (subnetId: string) => {
    const newIds = selectedSubnetIds.filter(id => id !== subnetId);
    form.setValue('subnet_ids', newIds);
    onDataChange({ subnet_ids: newIds });
  };

  return (
    <Form {...form}>
      <div className="space-y-6">
        {/* VPC Selection */}
        <FormField
          control={form.control}
          name="vpc_id"
          render={({ field }) => (
            <FormItem>
              <FormLabel>VPC *</FormLabel>
              <div className="flex gap-2">
                <FormControl className="flex-1">
                  <Select
                    value={selectedVPCId || field.value || ''}
                    onValueChange={handleVPCChange}
                    disabled={!selectedProvider || !selectedCredentialId || isLoadingVPCs}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={isLoadingVPCs ? 'Loading VPCs...' : 'Select VPC'} />
                    </SelectTrigger>
                    <SelectContent>
                      {vpcs.length === 0 && !isLoadingVPCs ? (
                        <div className="p-2 text-sm text-muted-foreground">No VPCs found</div>
                      ) : (
                        vpcs.map((vpc) => (
                          <SelectItem key={vpc.id} value={vpc.id}>
                            {vpc.name || vpc.id} {vpc.cidr_block && `(${vpc.cidr_block})`}
                          </SelectItem>
                        ))
                      )}
                    </SelectContent>
                  </Select>
                </FormControl>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={handleRefreshVPCs}
                  disabled={isLoadingVPCs || !selectedProvider || !selectedCredentialId}
                  title="Refresh VPC list"
                >
                  <RefreshCw className={`h-4 w-4 ${isLoadingVPCs ? 'animate-spin' : ''}`} />
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setIsCreateVPCDialogOpen(true)}
                  disabled={!selectedProvider || !selectedCredentialId}
                  title="Create new VPC"
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Create VPC
                </Button>
              </div>
              <FormDescription>
                Select a VPC for your cluster. Subnets will be loaded after selection.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Subnet Selection with Checkboxes */}
        <FormField
          control={form.control}
          name="subnet_ids"
          render={({ field: _field }) => (
            <FormItem>
              <FormLabel>Subnets *</FormLabel>
              <div className="flex gap-2 mb-2">
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={handleRefreshSubnets}
                  disabled={isLoadingSubnets || !selectedVPCId}
                  title="Refresh subnet list"
                >
                  <RefreshCw className={`h-4 w-4 ${isLoadingSubnets ? 'animate-spin' : ''}`} />
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setIsCreateSubnetDialogOpen(true)}
                  disabled={!selectedVPCId}
                  title="Create new subnet"
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Create Subnet
                </Button>
              </div>
              
              {/* Checkbox List */}
              <FormControl>
                <div className="border rounded-md p-4 max-h-60 overflow-y-auto space-y-3">
                  {!selectedVPCId ? (
                    <div className="text-sm text-muted-foreground text-center py-4">
                      {t('network.selectVPCToViewSubnets')}
                    </div>
                  ) : isLoadingSubnets ? (
                    <div className="text-sm text-muted-foreground text-center py-4">
                      Loading subnets...
                    </div>
                  ) : subnets.length === 0 ? (
                    <div className="text-sm text-muted-foreground text-center py-4">
                      No subnets found in the selected VPC
                    </div>
                  ) : (
                    subnets.map((subnet) => {
                      const isSelected = selectedSubnetIds.includes(subnet.id);
                      return (
                        <div
                          key={subnet.id}
                          className="flex items-center space-x-3 p-2 rounded-md hover:bg-muted/50 transition-colors"
                        >
                          <Checkbox
                            id={`subnet-${subnet.id}`}
                            checked={isSelected}
                            onCheckedChange={(checked) => {
                              handleSubnetToggle(subnet.id, checked === true);
                            }}
                            disabled={!selectedVPCId || isLoadingSubnets}
                          />
                          <label
                            htmlFor={`subnet-${subnet.id}`}
                            className="flex-1 text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
                          >
                            <div className="flex items-center justify-between">
                              <span>
                                {subnet.name || subnet.id}
                              </span>
                              {subnet.cidr_block && (
                                <span className="text-xs text-muted-foreground ml-2">
                                  {subnet.cidr_block}
                                </span>
                              )}
                            </div>
                            {subnet.availability_zone && (
                              <div className="text-xs text-muted-foreground mt-1">
                                Zone: {subnet.availability_zone}
                              </div>
                            )}
                          </label>
                        </div>
                      );
                    })
                  )}
                </div>
              </FormControl>
              
              <FormDescription>
                {selectedProvider === 'aws' 
                  ? 'Select at least one subnet. You can select multiple subnets for high availability.'
                  : selectedProvider === 'gcp' || selectedProvider === 'azure'
                  ? 'Select one subnet for your cluster.'
                  : 'Select at least one subnet.'
                }
              </FormDescription>
              
              {/* Selected Subnets Display */}
              {selectedSubnetIds.length > 0 && (
                <div className="flex flex-wrap gap-2 mt-2">
                  {selectedSubnetIds.map((subnetId) => {
                    const subnet = subnets.find(s => s.id === subnetId);
                    return (
                      <Badge key={subnetId} variant="secondary" className="flex items-center gap-1">
                        {subnet?.name || subnetId}
                        <button
                          type="button"
                          onClick={() => handleRemoveSubnet(subnetId)}
                          className="ml-1 hover:bg-destructive/20 rounded-full p-0.5"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </Badge>
                    );
                  })}
                </div>
              )}
              
              <FormMessage />
            </FormItem>
          )}
        />

        {!selectedVPCId && (
          <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
            <p className="text-sm text-blue-800">
              Please select a VPC first to load available subnets.
            </p>
          </div>
        )}

        {selectedVPCId && subnets.length === 0 && !isLoadingSubnets && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
            <p className="text-sm text-yellow-800">
              No subnets found in the selected VPC. Please create a subnet first.
            </p>
          </div>
        )}

        {/* VPC Creation Dialog */}
        <CreateVPCDialog
          open={isCreateVPCDialogOpen}
          onOpenChange={setIsCreateVPCDialogOpen}
          selectedProvider={selectedProvider}
          selectedCredentialId={selectedCredentialId}
          selectedRegion={form.watch('region') || selectedRegion}
          onSuccess={handleVPCCreated}
          disabled={!selectedProvider || !selectedCredentialId}
        />

        {/* Subnet Creation Dialog */}
        {selectedVPCId && (
          <CreateSubnetDialog
            open={isCreateSubnetDialogOpen}
            onOpenChange={setIsCreateSubnetDialogOpen}
            selectedProvider={selectedProvider}
            selectedCredentialId={selectedCredentialId}
            selectedRegion={form.watch('region') || selectedRegion}
            selectedZone={form.watch('zone')}
            selectedVPCId={selectedVPCId}
            vpcs={vpcs}
            onVPCChange={handleVPCChange}
            onSuccess={handleSubnetCreated}
            disabled={!selectedVPCId}
          />
        )}

        {/* Azure specific: Network and Node Pool Configuration */}
        {selectedProvider === 'azure' && selectedVPCId && selectedSubnetIds.length > 0 && (
          <div className="space-y-6 mt-6 pt-6 border-t">
            <h3 className="text-lg font-semibold">Azure Network Configuration</h3>
            
            {/* Network Plugin */}
            <FormField
              control={form.control}
              name="network.network_plugin"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Network Plugin</FormLabel>
                  <Select
                    value={field.value || 'azure'}
                    onValueChange={(value) => {
                      field.onChange(value);
                      const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
                      const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                        virtual_network_id: currentNetwork?.virtual_network_id || selectedVPCId || '',
                        subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                        network_plugin: value,
                        network_policy: currentNetwork?.network_policy,
                        pod_cidr: currentNetwork?.pod_cidr,
                        service_cidr: currentNetwork?.service_cidr,
                        dns_service_ip: currentNetwork?.dns_service_ip,
                        docker_bridge_cidr: currentNetwork?.docker_bridge_cidr,
                      };
                      form.setValue('network', updatedNetwork);
                      onDataChange({ network: updatedNetwork });
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select network plugin" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="azure">Azure</SelectItem>
                      <SelectItem value="kubenet">Kubenet</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Network plugin to use for the cluster
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Network Policy */}
            <FormField
              control={form.control}
              name="network.network_policy"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Network Policy</FormLabel>
                  <Select
                    value={field.value || ''}
                    onValueChange={(value) => {
                      field.onChange(value);
                      const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
                      const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                        virtual_network_id: currentNetwork?.virtual_network_id || selectedVPCId || '',
                        subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                        network_plugin: currentNetwork?.network_plugin || 'azure',
                        network_policy: value,
                        pod_cidr: currentNetwork?.pod_cidr,
                        service_cidr: currentNetwork?.service_cidr,
                        dns_service_ip: currentNetwork?.dns_service_ip,
                        docker_bridge_cidr: currentNetwork?.docker_bridge_cidr,
                      };
                      form.setValue('network', updatedNetwork);
                      onDataChange({ network: updatedNetwork });
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select network policy (optional)" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">None</SelectItem>
                      <SelectItem value="azure">Azure</SelectItem>
                      <SelectItem value="calico">Calico</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Network policy to use for the cluster (optional)
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Virtual Network ID and Subnet ID (from selected VPC/Subnet) */}
            {selectedVPCId && selectedSubnetIds.length > 0 && (
              <div className="space-y-2">
                <FormDescription>
                  Virtual Network ID and Subnet ID will be set automatically from your selection above.
                </FormDescription>
                <div className="text-sm text-muted-foreground">
                  <p>Virtual Network ID: {selectedVPCId}</p>
                  <p>Subnet ID: {selectedSubnetIds[0]}</p>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </Form>
  );
}


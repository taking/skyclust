/**
 * Network Configuration Step
 * Step 2: VPC 및 Subnet 선택
 */

'use client';

import { useState, useEffect } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useVPCs } from '@/features/networks/hooks/use-vpcs';
import { useSubnets } from '@/features/networks/hooks/use-subnets';
import { useVPCActions } from '@/features/networks/hooks/use-vpc-actions';
import { useSubnetActions } from '@/features/networks/hooks/use-subnet-actions';
import { CreateVPCDialog } from '@/features/networks/components/create-vpc-dialog';
import { CreateSubnetDialog } from '@/features/networks/components/create-subnet-dialog';
import { RefreshCw, Plus, X } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query-keys';
import type { CreateClusterForm, CloudProvider, CreateVPCForm, CreateSubnetForm } from '@/lib/types';

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

  // VPC and Subnet actions
  const { createVPCMutation } = useVPCActions({
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
    onSuccess: () => {
      handleRefreshVPCs();
    },
  });

  const { createSubnetMutation } = useSubnetActions({
    selectedProvider,
    selectedCredentialId,
    onSuccess: () => {
      handleRefreshSubnets();
    },
  });

  // Form에서 VPC ID 가져오기 (초기값)
  const formVPCId = form.watch('vpc_id');
  useEffect(() => {
    if (formVPCId && !selectedVPCId) {
      setSelectedVPCId(formVPCId);
      setSubnetVPCId(formVPCId);
    }
  }, [formVPCId, selectedVPCId, setSubnetVPCId]);

  // VPC 선택 시 Subnet 목록 로드
  const handleVPCChange = (vpcId: string) => {
    setSelectedVPCId(vpcId);
    setSubnetVPCId(vpcId);
    form.setValue('vpc_id', vpcId);
    form.setValue('subnet_ids', []); // VPC 변경 시 Subnet 초기화
    onDataChange({ vpc_id: vpcId, subnet_ids: [] });
  };

  // VPC 생성 핸들러 (모달에서 호출)
  const handleCreateVPC = async (data: CreateVPCForm) => {
    try {
      // credential_id 자동 추가
      const vpcData: CreateVPCForm = {
        ...data,
        credential_id: selectedCredentialId,
        region: selectedRegion,
      };
      
      const result = await createVPCMutation.mutateAsync(vpcData);
      setIsCreateVPCDialogOpen(false);
      
      // VPC 생성 후 자동 선택 및 목록 갱신
      if (result?.id) {
        // 목록이 갱신될 때까지 잠시 대기
        setTimeout(() => {
          handleVPCChange(result.id);
          handleRefreshVPCs();
        }, 500);
      }
    } catch (_error) {
      // 에러는 mutation에서 자동 처리됨
    }
  };

  // Subnet 생성 핸들러 (모달에서 호출)
  const handleCreateSubnet = async (data: CreateSubnetForm) => {
    if (!selectedVPCId) return;
    
    try {
      // credential_id 자동 추가
      const subnetData: CreateSubnetForm = {
        ...data,
        credential_id: selectedCredentialId,
        vpc_id: selectedVPCId,
        region: selectedRegion,
      };
      
      const result = await createSubnetMutation.mutateAsync(subnetData);
      setIsCreateSubnetDialogOpen(false);
      
      // Subnet 생성 후 자동 선택
      if (result?.id) {
        setTimeout(() => {
          handleRefreshSubnets();
          // 생성된 Subnet을 자동으로 선택 목록에 추가
          const currentIds = form.getValues('subnet_ids') || [];
          if (!currentIds.includes(result.id)) {
            handleSubnetToggle(result.id);
          }
        }, 500);
      }
    } catch (_error) {
      // 에러는 mutation에서 자동 처리됨
    }
  };

  // Subnet 선택 (Multi-select)
  const selectedSubnetIds = form.watch('subnet_ids') || [];
  const handleSubnetToggle = (subnetId: string) => {
    const currentIds = selectedSubnetIds;
    const newIds = currentIds.includes(subnetId)
      ? currentIds.filter(id => id !== subnetId)
      : [...currentIds, subnetId];
    
    form.setValue('subnet_ids', newIds);
    onDataChange({ subnet_ids: newIds });
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

        {/* Subnet Selection */}
        <FormField
          control={form.control}
          name="subnet_ids"
          render={({ field: _field }) => (
            <FormItem>
              <FormLabel>Subnets *</FormLabel>
              <div className="flex gap-2">
                <FormControl className="flex-1">
                  <Select
                    value={undefined}
                    onValueChange={handleSubnetToggle}
                    disabled={!selectedVPCId || isLoadingSubnets}
                  >
                    <SelectTrigger>
                      <SelectValue 
                        placeholder={
                          !selectedVPCId 
                            ? 'Select VPC first' 
                            : isLoadingSubnets 
                            ? 'Loading subnets...' 
                            : 'Select subnets'
                        } 
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {subnets.length === 0 && !isLoadingSubnets && selectedVPCId ? (
                        <div className="p-2 text-sm text-muted-foreground">No subnets found</div>
                      ) : (
                        subnets.map((subnet) => {
                          const isSelected = selectedSubnetIds.includes(subnet.id);
                          return (
                            <SelectItem 
                              key={subnet.id} 
                              value={subnet.id}
                              className={isSelected ? 'bg-muted' : ''}
                            >
                              <div className="flex items-center justify-between w-full">
                                <span>
                                  {subnet.name || subnet.id} {subnet.cidr_block && `(${subnet.cidr_block})`}
                                </span>
                                {isSelected && <span className="ml-2 text-xs">✓</span>}
                              </div>
                            </SelectItem>
                          );
                        })
                      )}
                    </SelectContent>
                  </Select>
                </FormControl>
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
              <FormDescription>
                Select at least one subnet. You can select multiple subnets.
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
          onSubmit={handleCreateVPC}
          selectedProvider={selectedProvider}
          selectedRegion={selectedRegion}
          isPending={createVPCMutation.isPending}
          disabled={!selectedProvider || !selectedCredentialId}
        />

        {/* Subnet Creation Dialog */}
        {selectedVPCId && (
          <CreateSubnetDialog
            open={isCreateSubnetDialogOpen}
            onOpenChange={setIsCreateSubnetDialogOpen}
            onSubmit={handleCreateSubnet}
            selectedProvider={selectedProvider}
            selectedRegion={selectedRegion}
            selectedVPCId={selectedVPCId}
            vpcs={vpcs}
            onVPCChange={handleVPCChange}
            isPending={createSubnetMutation.isPending}
            disabled={!selectedVPCId}
          />
        )}
      </div>
    </Form>
  );
}


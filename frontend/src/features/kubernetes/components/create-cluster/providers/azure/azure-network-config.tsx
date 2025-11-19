/**
 * Azure Network Configuration Component
 * Azure AKS 네트워크 설정: Network Plugin, Network Policy
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormField, FormItem, FormLabel, FormControl, FormMessage, FormDescription } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateClusterForm } from '@/lib/types';

interface AzureNetworkConfigProps {
  form: UseFormReturn<CreateClusterForm>;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
  selectedVPCId: string;
  selectedSubnetIds: string[];
}

export function AzureNetworkConfig({
  form,
  onDataChange,
  selectedVPCId,
  selectedSubnetIds,
}: AzureNetworkConfigProps) {
  if (!selectedVPCId || !selectedSubnetIds || selectedSubnetIds.length === 0) {
    return null;
  }

  return (
    <div className="space-y-6 mt-6 pt-6 border-t">
      <h3 className="text-lg font-semibold">Azure Network Configuration</h3>
      
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
  );
}


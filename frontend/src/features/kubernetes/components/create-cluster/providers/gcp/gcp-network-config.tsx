/**
 * GCP Network Configuration Component
 * GCP GKE 네트워크 설정: Pod CIDR, Service CIDR, Master Authorized Networks, Private Endpoint/Nodes
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormField, FormItem, FormLabel, FormControl, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Plus, X } from 'lucide-react';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface GCPNetworkConfigProps {
  form: UseFormReturn<CreateClusterForm>;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
  selectedVPCId: string;
  selectedSubnetIds: string[];
  selectedProjectId?: string;
}

export function GCPNetworkConfig({
  form,
  onDataChange,
  selectedVPCId: _selectedVPCId,
  selectedSubnetIds,
  selectedProjectId: _selectedProjectId,
}: GCPNetworkConfigProps) {
  const { t } = useTranslation();

  if (!selectedSubnetIds || selectedSubnetIds.length === 0) {
    return null;
  }

  return (
    <div className="space-y-6 mt-6 pt-6 border-t">
      <h3 className="text-lg font-semibold">GCP Network Configuration</h3>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <FormField
          control={form.control}
          name="network.pod_cidr"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('kubernetes.gcp.podCidr')}</FormLabel>
              <FormControl>
                <Input
                  placeholder={t('kubernetes.gcp.podCidrDefault')}
                  value={field.value || ''}
                  onChange={(e) => {
                    field.onChange(e.target.value);
                    const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
                    const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                      subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                      pod_cidr: e.target.value || t('kubernetes.gcp.podCidrDefault'),
                      service_cidr: currentNetwork?.service_cidr || t('kubernetes.gcp.serviceCidrDefault'),
                      private_endpoint: currentNetwork?.private_endpoint,
                      private_nodes: currentNetwork?.private_nodes,
                      master_authorized_networks: currentNetwork?.master_authorized_networks || [],
                      ...currentNetwork,
                    };
                    form.setValue('network', updatedNetwork);
                    onDataChange({ network: updatedNetwork });
                  }}
                />
              </FormControl>
              <FormDescription>
                {t('kubernetes.gcp.podCidrDescription')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="network.service_cidr"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('kubernetes.gcp.serviceCidr')}</FormLabel>
              <FormControl>
                <Input
                  placeholder={t('kubernetes.gcp.serviceCidrDefault')}
                  value={field.value || ''}
                  onChange={(e) => {
                    field.onChange(e.target.value);
                    const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
                    const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                      subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                      pod_cidr: currentNetwork?.pod_cidr || t('kubernetes.gcp.podCidrDefault'),
                      service_cidr: e.target.value || t('kubernetes.gcp.serviceCidrDefault'),
                      private_endpoint: currentNetwork?.private_endpoint,
                      private_nodes: currentNetwork?.private_nodes,
                      master_authorized_networks: currentNetwork?.master_authorized_networks || [],
                      ...currentNetwork,
                    };
                    form.setValue('network', updatedNetwork);
                    onDataChange({ network: updatedNetwork });
                  }}
                />
              </FormControl>
              <FormDescription>
                {t('kubernetes.gcp.serviceCidrDescription')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <div className="space-y-2">
        <Label>{t('kubernetes.gcp.masterAuthorizedNetworks')}</Label>
        <FormDescription className="mb-2">
          {t('kubernetes.gcp.masterAuthorizedNetworksDescription')}
        </FormDescription>
        <div className="space-y-2">
          {(form.watch('network.master_authorized_networks') || []).map((network: string, index: number) => (
            <div key={index} className="flex items-center gap-2">
              <Input
                value={network}
                onChange={(e) => {
                  const currentNetworks = form.watch('network.master_authorized_networks') || [];
                  const updatedNetworks = [...currentNetworks];
                  updatedNetworks[index] = e.target.value;
                  const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
                  const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                    subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                    pod_cidr: currentNetwork?.pod_cidr || t('kubernetes.gcp.podCidrDefault'),
                    service_cidr: currentNetwork?.service_cidr || t('kubernetes.gcp.serviceCidrDefault'),
                    private_endpoint: currentNetwork?.private_endpoint,
                    private_nodes: currentNetwork?.private_nodes,
                    master_authorized_networks: updatedNetworks,
                    ...currentNetwork,
                  };
                  form.setValue('network', updatedNetwork);
                  onDataChange({ network: updatedNetwork });
                }}
                placeholder="0.0.0.0/0"
              />
              {network === '0.0.0.0/0' && (
                <span className="text-xs text-yellow-600 dark:text-yellow-400">
                  {t('kubernetes.gcp.masterAuthorizedNetworksWarning')}
                </span>
              )}
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => {
                  const currentNetworks = form.watch('network.master_authorized_networks') || [];
                  const updatedNetworks = currentNetworks.filter((_: string, i: number) => i !== index);
                  const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
                  const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                    subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                    pod_cidr: currentNetwork?.pod_cidr || t('kubernetes.gcp.podCidrDefault'),
                    service_cidr: currentNetwork?.service_cidr || t('kubernetes.gcp.serviceCidrDefault'),
                    private_endpoint: currentNetwork?.private_endpoint,
                    private_nodes: currentNetwork?.private_nodes,
                    master_authorized_networks: updatedNetworks,
                    ...currentNetwork,
                  };
                  form.setValue('network', updatedNetwork);
                  onDataChange({ network: updatedNetwork });
                }}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => {
              const currentNetworks = form.watch('network.master_authorized_networks') || [];
              const updatedNetworks = [...currentNetworks, ''];
              const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
              const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                pod_cidr: currentNetwork?.pod_cidr || t('kubernetes.gcp.podCidrDefault'),
                service_cidr: currentNetwork?.service_cidr || t('kubernetes.gcp.serviceCidrDefault'),
                private_endpoint: currentNetwork?.private_endpoint,
                private_nodes: currentNetwork?.private_nodes,
                master_authorized_networks: updatedNetworks,
                ...currentNetwork,
              };
              form.setValue('network', updatedNetwork);
              onDataChange({ network: updatedNetwork });
            }}
          >
            <Plus className="mr-2 h-4 w-4" />
            {t('common.add')} Network
          </Button>
        </div>
      </div>

      <div className="space-y-4">
        <FormField
          control={form.control}
          name="network.private_endpoint"
          render={({ field }) => (
            <div className="space-y-1">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="private-endpoint"
                  checked={field.value || false}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
                    const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                      subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                      pod_cidr: currentNetwork?.pod_cidr || t('kubernetes.gcp.podCidrDefault'),
                      service_cidr: currentNetwork?.service_cidr || t('kubernetes.gcp.serviceCidrDefault'),
                      private_endpoint: checkedValue,
                      private_nodes: currentNetwork?.private_nodes,
                      master_authorized_networks: currentNetwork?.master_authorized_networks || [],
                      ...currentNetwork,
                    };
                    form.setValue('network', updatedNetwork);
                    onDataChange({ network: updatedNetwork });
                  }}
                />
                <Label
                  htmlFor="private-endpoint"
                  className="text-sm font-normal cursor-pointer"
                >
                  {t('kubernetes.gcp.privateEndpoint')}
                </Label>
              </div>
              <FormDescription className="text-xs text-muted-foreground ml-6">
                {t('kubernetes.gcp.privateEndpointDescription')}
              </FormDescription>
            </div>
          )}
        />

        <FormField
          control={form.control}
          name="network.private_nodes"
          render={({ field }) => (
            <div className="space-y-1">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="private-nodes"
                  checked={field.value || false}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    const currentNetwork = form.getValues('network') || {} as CreateClusterForm['network'];
                    const updatedNetwork: NonNullable<CreateClusterForm['network']> = {
                      subnet_id: currentNetwork?.subnet_id || selectedSubnetIds[0] || '',
                      pod_cidr: currentNetwork?.pod_cidr || t('kubernetes.gcp.podCidrDefault'),
                      service_cidr: currentNetwork?.service_cidr || t('kubernetes.gcp.serviceCidrDefault'),
                      private_endpoint: currentNetwork?.private_endpoint,
                      private_nodes: checkedValue,
                      master_authorized_networks: currentNetwork?.master_authorized_networks || [],
                      ...currentNetwork,
                    };
                    form.setValue('network', updatedNetwork);
                    onDataChange({ network: updatedNetwork });
                  }}
                />
                <Label
                  htmlFor="private-nodes"
                  className="text-sm font-normal cursor-pointer"
                >
                  {t('kubernetes.gcp.privateNodes')}
                </Label>
              </div>
              <FormDescription className="text-xs text-muted-foreground ml-6">
                {t('kubernetes.gcp.privateNodesDescription')}
              </FormDescription>
            </div>
          )}
        />
      </div>
    </div>
  );
}


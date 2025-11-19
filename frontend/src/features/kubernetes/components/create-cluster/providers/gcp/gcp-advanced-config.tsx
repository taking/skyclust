/**
 * GCP Advanced Configuration Component
 * GCP GKE 고급 설정: Project ID, Cluster Mode, Node Pool Config, Security Config
 */

'use client';

import { UseFormReturn } from 'react-hook-form';
import { FormField, FormItem, FormLabel, FormControl, FormMessage, FormDescription } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { CreateClusterForm } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface GCPAdvancedConfigProps {
  form: UseFormReturn<CreateClusterForm>;
  selectedCredentialId?: string;
  selectedProjectId?: string;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
}

export function GCPAdvancedConfig({
  form,
  selectedCredentialId: _selectedCredentialId,
  selectedProjectId,
  onDataChange,
}: GCPAdvancedConfigProps) {
  const { t } = useTranslation();

  // project_id는 props로 전달받은 값 우선, 없으면 form에서 가져오기
  const projectId = selectedProjectId || form.watch('project_id') || '';

  return (
    <>
      {/* Project ID and Cluster Mode */}
      <div className="space-y-6 mt-6 pt-6 border-t">
        <h3 className="text-lg font-semibold">GCP GKE Configuration</h3>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="project_id"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('kubernetes.gcp.projectId')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder="my-gcp-project"
                    value={projectId}
                    readOnly
                    disabled
                    className="bg-muted cursor-not-allowed"
                  />
                </FormControl>
                <FormDescription>
                  {t('kubernetes.gcp.projectIdDescription')} (자동 설정됨)
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="cluster_mode.type"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('kubernetes.gcp.clusterMode')}</FormLabel>
                <Select
                  value={field.value || 'standard'}
                  onValueChange={(value) => {
                    field.onChange(value);
                    const currentClusterMode = form.getValues('cluster_mode') || {};
                    const updatedClusterMode = {
                      ...currentClusterMode,
                      type: value,
                    };
                    form.setValue('cluster_mode', updatedClusterMode);
                    onDataChange({ cluster_mode: updatedClusterMode });
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={t('kubernetes.gcp.clusterMode')} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="standard">Standard</SelectItem>
                    <SelectItem value="autopilot">Autopilot</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>
                  {t('kubernetes.gcp.clusterModeDescription')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <FormField
          control={form.control}
          name="cluster_mode.remove_default_node_pool"
          render={({ field }) => (
            <div className="flex items-center space-x-2">
              <Checkbox
                id="remove-default-node-pool"
                checked={field.value || false}
                onCheckedChange={(checked) => {
                  const checkedValue = checked === true;
                  field.onChange(checkedValue);
                  const currentClusterMode = form.getValues('cluster_mode') || {};
                  const updatedClusterMode = {
                    ...currentClusterMode,
                    remove_default_node_pool: checkedValue,
                  };
                  form.setValue('cluster_mode', updatedClusterMode);
                  onDataChange({ cluster_mode: updatedClusterMode });
                }}
              />
              <Label
                htmlFor="remove-default-node-pool"
                className="text-sm font-normal cursor-pointer"
              >
                {t('kubernetes.gcp.removeDefaultNodePool')}
              </Label>
            </div>
          )}
        />
      </div>

      {/* Node Pool Configuration */}
      <div className="space-y-6 mt-6 pt-6 border-t">
        <h3 className="text-lg font-semibold">GCP Node Pool Configuration</h3>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="node_pool.name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('kubernetes.gcp.nodePoolName')} *</FormLabel>
                <FormControl>
                  <Input
                    placeholder="default-pool"
                    value={field.value || ''}
                    onChange={(e) => {
                      field.onChange(e);
                      const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                      const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                        name: e.target.value,
                        machine_type: currentNodePool.machine_type || '',
                        node_count: currentNodePool.node_count || 2,
                        ...currentNodePool,
                      };
                      form.setValue('node_pool', updatedNodePool);
                      onDataChange({ node_pool: updatedNodePool });
                    }}
                    required
                    aria-required="true"
                  />
                </FormControl>
                <FormDescription>
                  {t('kubernetes.gcp.nodePoolNameDescription')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.machine_type"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('kubernetes.gcp.machineType')} *</FormLabel>
                <FormControl>
                  <Input
                    placeholder="e2-medium"
                    value={field.value || ''}
                    onChange={(e) => {
                      field.onChange(e);
                      const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                      const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                        name: currentNodePool.name || '',
                        machine_type: e.target.value,
                        node_count: currentNodePool.node_count || 2,
                        ...currentNodePool,
                      };
                      form.setValue('node_pool', updatedNodePool);
                      onDataChange({ node_pool: updatedNodePool });
                    }}
                    required
                    aria-required="true"
                  />
                </FormControl>
                <FormDescription>
                  {t('kubernetes.gcp.machineTypeDescription')}
                </FormDescription>
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
                <FormLabel>{t('kubernetes.gcp.nodeCount')} *</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="1"
                    placeholder={t('kubernetes.gcp.nodeCountDefault')}
                    value={field.value || ''}
                    onChange={(e) => {
                      const value = parseInt(e.target.value, 10) || 0;
                      field.onChange(value);
                      const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                      const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                        name: currentNodePool.name || '',
                        machine_type: currentNodePool.machine_type || '',
                        node_count: value,
                        ...currentNodePool,
                      };
                      form.setValue('node_pool', updatedNodePool);
                      onDataChange({ node_pool: updatedNodePool });
                    }}
                    required
                    aria-required="true"
                  />
                </FormControl>
                <FormDescription>
                  {t('kubernetes.gcp.nodeCountDescription')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.disk_size_gb"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('kubernetes.gcp.diskSize')}</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="10"
                    placeholder={t('kubernetes.gcp.diskSizeDefault')}
                    value={field.value || ''}
                    onChange={(e) => {
                      const value = parseInt(e.target.value, 10) || 0;
                      field.onChange(value);
                      const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                      const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                        name: currentNodePool.name || '',
                        machine_type: currentNodePool.machine_type || '',
                        node_count: currentNodePool.node_count || 2,
                        disk_size_gb: value,
                        ...currentNodePool,
                      };
                      form.setValue('node_pool', updatedNodePool);
                      onDataChange({ node_pool: updatedNodePool });
                    }}
                  />
                </FormControl>
                <FormDescription>
                  {t('kubernetes.gcp.diskSizeDescription')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="node_pool.disk_type"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('kubernetes.gcp.diskType')}</FormLabel>
                <Select
                  value={field.value || 'pd-standard'}
                  onValueChange={(value) => {
                    field.onChange(value);
                    const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                    const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                      name: currentNodePool.name || '',
                      machine_type: currentNodePool.machine_type || '',
                      node_count: currentNodePool.node_count || 2,
                      disk_type: value,
                      ...currentNodePool,
                    };
                    form.setValue('node_pool', updatedNodePool);
                    onDataChange({ node_pool: updatedNodePool });
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={t('kubernetes.gcp.diskType')} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="pd-standard">pd-standard</SelectItem>
                    <SelectItem value="pd-ssd">pd-ssd</SelectItem>
                    <SelectItem value="pd-balanced">pd-balanced</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>
                  {t('kubernetes.gcp.diskTypeDescription')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <div className="space-y-4 p-4 border rounded-md">
          <div className="flex items-center space-x-2">
            <Checkbox
              id="gcp-enable-auto-scaling"
              checked={form.watch('node_pool.auto_scaling.enabled') || false}
              onCheckedChange={(checked) => {
                const checkedValue = checked === true;
                const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                const currentAutoScaling = currentNodePool.auto_scaling || {};
                const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                  name: currentNodePool.name || '',
                  machine_type: currentNodePool.machine_type || '',
                  node_count: currentNodePool.node_count || 2,
                  auto_scaling: {
                    ...currentAutoScaling,
                    enabled: checkedValue,
                  },
                  ...currentNodePool,
                };
                form.setValue('node_pool', updatedNodePool);
                onDataChange({ node_pool: updatedNodePool });
              }}
            />
            <Label
              htmlFor="gcp-enable-auto-scaling"
              className="text-sm font-normal cursor-pointer"
            >
              {t('kubernetes.gcp.autoScaling')}
            </Label>
          </div>
          <FormDescription className="text-xs text-muted-foreground">
            {t('kubernetes.gcp.autoScalingDescription')}
          </FormDescription>

          {form.watch('node_pool.auto_scaling.enabled') && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
              <FormField
                control={form.control}
                name="node_pool.auto_scaling.min_node_count"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('kubernetes.gcp.minNodeCount')}</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min="1"
                        placeholder={t('kubernetes.gcp.minNodeCountDefault')}
                        value={field.value || ''}
                        onChange={(e) => {
                          const value = parseInt(e.target.value, 10) || 0;
                          field.onChange(value);
                          const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                          const currentAutoScaling = currentNodePool.auto_scaling || {};
                          const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                            name: currentNodePool.name || '',
                            machine_type: currentNodePool.machine_type || '',
                            node_count: currentNodePool.node_count || 2,
                            auto_scaling: {
                              ...currentAutoScaling,
                              min_node_count: value,
                            },
                            ...currentNodePool,
                          };
                          form.setValue('node_pool', updatedNodePool);
                          onDataChange({ node_pool: updatedNodePool });
                        }}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('kubernetes.gcp.minNodeCountDescription')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="node_pool.auto_scaling.max_node_count"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('kubernetes.gcp.maxNodeCount')}</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min="1"
                        placeholder={t('kubernetes.gcp.maxNodeCountDefault')}
                        value={field.value || ''}
                        onChange={(e) => {
                          const value = parseInt(e.target.value, 10) || 0;
                          field.onChange(value);
                          const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                          const currentAutoScaling = currentNodePool.auto_scaling || {};
                          const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                            name: currentNodePool.name || '',
                            machine_type: currentNodePool.machine_type || '',
                            node_count: currentNodePool.node_count || 2,
                            auto_scaling: {
                              ...currentAutoScaling,
                              max_node_count: value,
                            },
                            ...currentNodePool,
                          };
                          form.setValue('node_pool', updatedNodePool);
                          onDataChange({ node_pool: updatedNodePool });
                        }}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('kubernetes.gcp.maxNodeCountDescription')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          )}
        </div>

        <div className="flex items-center space-x-4">
          <FormField
            control={form.control}
            name="node_pool.preemptible"
            render={({ field }) => (
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="preemptible"
                  checked={field.value || false}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                    const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                      name: currentNodePool.name || '',
                      machine_type: currentNodePool.machine_type || '',
                      node_count: currentNodePool.node_count || 2,
                      preemptible: checkedValue,
                      ...currentNodePool,
                    };
                    form.setValue('node_pool', updatedNodePool);
                    onDataChange({ node_pool: updatedNodePool });
                  }}
                />
                <Label
                  htmlFor="preemptible"
                  className="text-sm font-normal cursor-pointer"
                >
                  {t('kubernetes.gcp.preemptible')}
                </Label>
              </div>
            )}
          />
          <FormDescription className="text-xs text-muted-foreground">
            {t('kubernetes.gcp.preemptibleDescription')}
          </FormDescription>

          <FormField
            control={form.control}
            name="node_pool.spot"
            render={({ field }) => (
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="spot"
                  checked={field.value || false}
                  onCheckedChange={(checked) => {
                    const checkedValue = checked === true;
                    field.onChange(checkedValue);
                    const currentNodePool = (form.getValues('node_pool') || {}) as Partial<NonNullable<CreateClusterForm['node_pool']>>;
                    const updatedNodePool: NonNullable<CreateClusterForm['node_pool']> = { 
                      name: currentNodePool.name || '',
                      machine_type: currentNodePool.machine_type || '',
                      node_count: currentNodePool.node_count || 2,
                      spot: checkedValue,
                      ...currentNodePool,
                    };
                    form.setValue('node_pool', updatedNodePool);
                    onDataChange({ node_pool: updatedNodePool });
                  }}
                />
                <Label
                  htmlFor="spot"
                  className="text-sm font-normal cursor-pointer"
                >
                  {t('kubernetes.gcp.spot')}
                </Label>
              </div>
            )}
          />
          <FormDescription className="text-xs text-muted-foreground">
            {t('kubernetes.gcp.spotDescription')}
          </FormDescription>
        </div>
      </div>

      {/* Security Configuration */}
      <div className="space-y-6 mt-6 pt-6 border-t">
        <h3 className="text-lg font-semibold">{t('kubernetes.gcp.binaryAuthorization')} & Security</h3>
        
        <div className="space-y-4">
          <FormField
            control={form.control}
            name="security.binary_authorization"
            render={({ field }) => (
              <div className="space-y-1">
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="binary-authorization"
                    checked={field.value || false}
                    onCheckedChange={(checked) => {
                      const checkedValue = checked === true;
                      field.onChange(checkedValue);
                      const currentSecurity = form.getValues('security') || {};
                      const updatedSecurity = {
                        ...currentSecurity,
                        binary_authorization: checkedValue,
                      };
                      form.setValue('security', updatedSecurity);
                      onDataChange({ security: updatedSecurity });
                    }}
                  />
                  <Label
                    htmlFor="binary-authorization"
                    className="text-sm font-normal cursor-pointer"
                  >
                    {t('kubernetes.gcp.binaryAuthorization')}
                  </Label>
                </div>
                <FormDescription className="text-xs text-muted-foreground ml-6">
                  {t('kubernetes.gcp.binaryAuthorizationDescription')}
                </FormDescription>
              </div>
            )}
          />

          <FormField
            control={form.control}
            name="security.network_policy"
            render={({ field }) => (
              <div className="space-y-1">
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="network-policy"
                    checked={field.value || false}
                    onCheckedChange={(checked) => {
                      const checkedValue = checked === true;
                      field.onChange(checkedValue);
                      const currentSecurity = form.getValues('security') || {};
                      const updatedSecurity = {
                        ...currentSecurity,
                        network_policy: checkedValue,
                      };
                      form.setValue('security', updatedSecurity);
                      onDataChange({ security: updatedSecurity });
                    }}
                  />
                  <Label
                    htmlFor="network-policy"
                    className="text-sm font-normal cursor-pointer"
                  >
                    {t('kubernetes.gcp.networkPolicy')}
                  </Label>
                </div>
                <FormDescription className="text-xs text-muted-foreground ml-6">
                  {t('kubernetes.gcp.networkPolicyDescription')}
                </FormDescription>
              </div>
            )}
          />

          <FormField
            control={form.control}
            name="security.pod_security_policy"
            render={({ field }) => (
              <div className="space-y-1">
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="pod-security-policy"
                    checked={field.value || false}
                    onCheckedChange={(checked) => {
                      const checkedValue = checked === true;
                      field.onChange(checkedValue);
                      const currentSecurity = form.getValues('security') || {};
                      const updatedSecurity = {
                        ...currentSecurity,
                        pod_security_policy: checkedValue,
                      };
                      form.setValue('security', updatedSecurity);
                      onDataChange({ security: updatedSecurity });
                    }}
                  />
                  <Label
                    htmlFor="pod-security-policy"
                    className="text-sm font-normal cursor-pointer"
                  >
                    {t('kubernetes.gcp.podSecurityPolicy')}
                  </Label>
                </div>
                <FormDescription className="text-xs text-muted-foreground ml-6">
                  {t('kubernetes.gcp.podSecurityPolicyDescription')}
                </FormDescription>
              </div>
            )}
          />

          <FormField
            control={form.control}
            name="security.enable_workload_identity"
            render={({ field }) => (
              <div className="space-y-1">
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="workload-identity"
                    checked={field.value || false}
                    onCheckedChange={(checked) => {
                      const checkedValue = checked === true;
                      field.onChange(checkedValue);
                      const currentSecurity = form.getValues('security') || {};
                      const updatedSecurity = {
                        ...currentSecurity,
                        enable_workload_identity: checkedValue,
                      };
                      form.setValue('security', updatedSecurity);
                      onDataChange({ security: updatedSecurity });
                    }}
                  />
                  <Label
                    htmlFor="workload-identity"
                    className="text-sm font-normal cursor-pointer"
                  >
                    {t('kubernetes.gcp.workloadIdentity')}
                  </Label>
                </div>
                <FormDescription className="text-xs text-muted-foreground ml-6">
                  {t('kubernetes.gcp.workloadIdentityDescription')}
                </FormDescription>
              </div>
            )}
          />
        </div>
      </div>
    </>
  );
}


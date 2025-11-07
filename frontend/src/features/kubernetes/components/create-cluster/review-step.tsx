/**
 * Review Step
 * Step 4: 최종 확인 및 생성
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { CreateClusterForm, CloudProvider } from '@/lib/types';

interface ReviewStepProps {
  formData: CreateClusterForm;
  selectedProvider?: CloudProvider;
  onCreate: () => void;
  isPending: boolean;
}

export function ReviewStep({
  formData,
  selectedProvider,
  onCreate: _onCreate,
  isPending: _isPending,
}: ReviewStepProps) {
  return (
    <div className="space-y-6">
      <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
        <p className="text-sm text-blue-800">
          Please review all settings before creating the cluster. Once created, some settings cannot be changed.
        </p>
      </div>

      {/* Basic Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>Basic Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Cluster Name</p>
              <p className="text-sm">{formData.name}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Kubernetes Version</p>
              <p className="text-sm">{formData.version}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Region</p>
              <p className="text-sm">{formData.region}</p>
            </div>
            {formData.zone && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Zone</p>
                <p className="text-sm">{formData.zone}</p>
              </div>
            )}
            <div>
              <p className="text-sm font-medium text-muted-foreground">Provider</p>
              <p className="text-sm uppercase">{selectedProvider || 'N/A'}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Network Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>Network Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">Selected Subnets</p>
            {formData.subnet_ids && formData.subnet_ids.length > 0 ? (
              <div className="flex flex-wrap gap-2">
                {formData.subnet_ids.map((subnetId) => (
                  <Badge key={subnetId} variant="secondary">
                    {subnetId}
                  </Badge>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No subnets selected</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Advanced Settings */}
      <Card>
        <CardHeader>
          <CardTitle>Advanced Settings</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {formData.role_arn && (
            <div>
              <p className="text-sm font-medium text-muted-foreground">Role ARN</p>
              <p className="text-sm break-all">{formData.role_arn}</p>
            </div>
          )}

          {formData.tags && Object.keys(formData.tags).length > 0 && (
            <div>
              <p className="text-sm font-medium text-muted-foreground mb-2">Tags</p>
              <div className="space-y-1">
                {Object.entries(formData.tags).map(([key, value]) => (
                  <div key={key} className="flex items-center gap-2">
                    <span className="text-sm font-medium">{key}:</span>
                    <span className="text-sm text-muted-foreground">{value}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {formData.access_config && (
            <div>
              <p className="text-sm font-medium text-muted-foreground mb-2">Access Configuration</p>
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="text-sm">Authentication Mode:</span>
                  <span className="text-sm text-muted-foreground">
                    {formData.access_config.authentication_mode || 'API'}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm">Bootstrap Admin Permissions:</span>
                  <span className="text-sm text-muted-foreground">
                    {formData.access_config.bootstrap_cluster_creator_admin_permissions ? 'Yes' : 'No'}
                  </span>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* API Request Preview */}
      <Card>
        <CardHeader>
          <CardTitle>API Request Preview</CardTitle>
          <CardDescription>This is what will be sent to the API</CardDescription>
        </CardHeader>
        <CardContent>
          <pre className="bg-muted p-4 rounded-md text-xs overflow-auto max-h-64">
            {JSON.stringify(
              {
                credential_id: formData.credential_id,
                name: formData.name,
                version: formData.version,
                region: formData.region,
                subnet_ids: formData.subnet_ids,
                role_arn: formData.role_arn || undefined,
                tags: formData.tags || undefined,
                access_config: formData.access_config || undefined,
              },
              null,
              2
            )}
          </pre>
        </CardContent>
      </Card>
    </div>
  );
}


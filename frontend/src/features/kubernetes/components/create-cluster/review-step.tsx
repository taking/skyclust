/**
 * Review Step
 * Step 4: 최종 확인 및 생성
 */

'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { 
  FileText, 
  Server, 
  Network, 
  Settings, 
  Code, 
  Globe, 
  MapPin, 
  Tag, 
  Shield, 
  CheckCircle2, 
  ChevronRight,
  Cloud,
  Layers
} from 'lucide-react';
import { useVPCs } from '@/features/networks/hooks/use-vpcs';
import { useSubnets } from '@/features/networks/hooks/use-subnets';
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
  // VPC와 Subnet 정보를 가져오기 위해 사용
  const { vpcs } = useVPCs();
  const { subnets } = useSubnets();

  // 선택된 Subnet 정보 찾기
  const selectedSubnets = formData.subnet_ids
    ? formData.subnet_ids.map(id => subnets.find(s => s.id === id)).filter(Boolean)
    : [];

  // 선택된 VPC 찾기 (subnet_ids에서 첫 번째 subnet의 vpc_id 사용)
  const firstSubnet = selectedSubnets.length > 0 ? selectedSubnets[0] : null;
  const vpcIdFromSubnet = firstSubnet?.vpc_id;
  const selectedVPC = vpcIdFromSubnet 
    ? vpcs.find(v => v.id === vpcIdFromSubnet) 
    : (formData.vpc_id ? vpcs.find(v => v.id === formData.vpc_id) : null);

  // Provider별 색상 설정
  const getProviderColor = (provider?: CloudProvider) => {
    switch (provider) {
      case 'aws':
        return 'bg-orange-100 text-orange-800 border-orange-200';
      case 'gcp':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'azure':
        return 'bg-cyan-100 text-cyan-800 border-cyan-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const [isApiPreviewOpen, setIsApiPreviewOpen] = useState(false);

  return (
    <div className="space-y-6">
      {/* Alert Banner */}
      <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4 flex items-start gap-3">
        <Shield className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
        <p className="text-sm text-blue-800 dark:text-blue-200">
          Please review all settings before creating the cluster. Once created, some settings cannot be changed.
        </p>
      </div>

      <Separator />

      {/* Basic Configuration */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <FileText className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg font-semibold">Basic Configuration</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Server className="h-4 w-4" />
                Cluster Name
              </div>
              <p className="text-sm font-semibold ml-6">{formData.name || 'N/A'}</p>
            </div>
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Layers className="h-4 w-4" />
                Kubernetes Version
              </div>
              <p className="text-sm font-semibold ml-6">{formData.version || 'N/A'}</p>
            </div>
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Globe className="h-4 w-4" />
                Region
              </div>
              <p className="text-sm font-semibold ml-6">{formData.region || 'N/A'}</p>
            </div>
            {formData.zone && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <MapPin className="h-4 w-4" />
                  Zone
                </div>
                <p className="text-sm font-semibold ml-6">{formData.zone}</p>
              </div>
            )}
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Cloud className="h-4 w-4" />
                Provider
              </div>
              <div className="ml-6">
                <Badge className={getProviderColor(selectedProvider)}>
                  {selectedProvider?.toUpperCase() || 'N/A'}
                </Badge>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Separator />

      {/* Network Configuration */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Network className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg font-semibold">Network Configuration</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {selectedVPC ? (
            <div className="space-y-3">
              {/* VPC 트리 구조 */}
              <div className="space-y-2">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  Virtual Private Cloud (VPC)
                </div>
                <div className="ml-6 space-y-1">
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary" className="font-medium">
                      {selectedVPC.name || selectedVPC.id}
                    </Badge>
                    {selectedVPC.cidr_block && (
                      <span className="text-xs text-muted-foreground">
                        {selectedVPC.cidr_block}
                      </span>
                    )}
                  </div>
                  
                  {/* Subnets 트리 구조 */}
                  {selectedSubnets.length > 0 && (
                    <div className="mt-3 space-y-2">
                      <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
                        <ChevronRight className="h-3 w-3" />
                        Subnets ({selectedSubnets.length})
                      </div>
                      <div className="ml-4 space-y-1.5">
                        {selectedSubnets.map((subnet, index) => (
                          <div key={subnet?.id} className="flex items-center gap-2">
                            <div className="flex items-center gap-1.5">
                              <div className="h-1.5 w-1.5 rounded-full bg-muted-foreground/40" />
                              <Badge variant="outline" className="text-xs">
                                {subnet?.name || subnet?.id}
                              </Badge>
                            </div>
                            {subnet?.availability_zone && (
                              <Badge variant="secondary" className="text-xs">
                                {subnet.availability_zone}
                              </Badge>
                            )}
                            {subnet?.cidr_block && (
                              <span className="text-xs text-muted-foreground">
                                {subnet.cidr_block}
                              </span>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          ) : selectedSubnets.length > 0 ? (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Layers className="h-4 w-4" />
                Selected Subnets ({selectedSubnets.length})
              </div>
              <div className="ml-6 space-y-1.5">
                {selectedSubnets.map((subnet) => (
                  <div key={subnet?.id} className="flex items-center gap-2">
                    <Badge variant="outline">
                      {subnet?.name || subnet?.id}
                    </Badge>
                    {subnet?.availability_zone && (
                      <Badge variant="secondary" className="text-xs">
                        {subnet.availability_zone}
                      </Badge>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ) : formData.subnet_ids && formData.subnet_ids.length > 0 ? (
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
        </CardContent>
      </Card>

      <Separator />

      {/* Advanced Settings */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Settings className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg font-semibold">Advanced Settings</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <Accordion type="single" collapsible className="w-full">
            {formData.tags && Object.keys(formData.tags).length > 0 && (
              <AccordionItem value="tags">
                <AccordionTrigger>
                  <div className="flex items-center gap-2">
                    <Tag className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Tags ({Object.keys(formData.tags).length})</span>
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="flex flex-wrap gap-2 pt-2">
                    {Object.entries(formData.tags).map(([key, value]) => (
                      <Badge key={key} variant="secondary" className="text-xs">
                        {key}: {value}
                      </Badge>
                    ))}
                  </div>
                </AccordionContent>
              </AccordionItem>
            )}

            {formData.access_config && (
              <AccordionItem value="access-config">
                <AccordionTrigger>
                  <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Access Configuration</span>
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-3 pt-2">
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                      <div className="flex-1">
                        <p className="text-sm font-medium">Authentication Mode</p>
                        <p className="text-sm text-muted-foreground">
                          {formData.access_config.authentication_mode || 'API'}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className={`h-4 w-4 ${formData.access_config.bootstrap_cluster_creator_admin_permissions ? 'text-green-600' : 'text-gray-400'}`} />
                      <div className="flex-1">
                        <p className="text-sm font-medium">Bootstrap Admin Permissions</p>
                        <p className="text-sm text-muted-foreground">
                          {formData.access_config.bootstrap_cluster_creator_admin_permissions ? 'Enabled' : 'Disabled'}
                        </p>
                      </div>
                    </div>
                  </div>
                </AccordionContent>
              </AccordionItem>
            )}

            {(!formData.tags || Object.keys(formData.tags).length === 0) && !formData.access_config && (
              <div className="py-4 text-sm text-muted-foreground text-center">
                No advanced settings configured
              </div>
            )}
          </Accordion>
        </CardContent>
      </Card>

      <Separator />

      {/* API Request Preview */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Code className="h-5 w-5 text-muted-foreground" />
              <div>
                <CardTitle className="text-lg font-semibold">API Request Preview</CardTitle>
                <CardDescription className="text-xs mt-1">
                  This is what will be sent to the API (Developer reference)
                </CardDescription>
              </div>
            </div>
            <button
              type="button"
              onClick={() => setIsApiPreviewOpen(!isApiPreviewOpen)}
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              {isApiPreviewOpen ? 'Hide' : 'Show'}
            </button>
          </div>
        </CardHeader>
        {isApiPreviewOpen && (
          <CardContent>
            <pre className="bg-muted dark:bg-muted/50 p-4 rounded-md text-xs overflow-auto max-h-64 border">
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
        )}
      </Card>
    </div>
  );
}


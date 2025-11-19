/**
 * Azure Review Configuration Component
 * Azure AKS 클러스터 생성 Review Step의 Provider별 설정 표시
 */

'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Settings, Code, Tag, Shield, CheckCircle2, Layers } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import type { CreateClusterForm } from '@/lib/types';
import { buildAzurePayload } from './azure-review-payload';

interface AzureReviewConfigProps {
  formData: CreateClusterForm;
}

export function AzureReviewConfig({ formData }: AzureReviewConfigProps) {
  const { t } = useTranslation();
  const [isApiPreviewOpen, setIsApiPreviewOpen] = useState(false);

  return (
    <>
      <Separator />

      {/* Advanced Settings */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Settings className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg font-semibold">{t('kubernetes.review.advancedSettings')}</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <Accordion type="single" collapsible className="w-full">
            {formData.node_pool && (
              <AccordionItem value="node-pool">
                <AccordionTrigger>
                  <div className="flex items-center gap-2">
                    <Layers className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Node Pool Configuration</span>
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-3 pt-2">
                    {formData.node_pool.name && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Name</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.name}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.vm_size && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">VM Size</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.vm_size}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.node_count !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Node Count</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.node_count}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.os_disk_size_gb !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">OS Disk Size (GB)</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.os_disk_size_gb}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.os_disk_type && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">OS Disk Type</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.os_disk_type}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.os_type && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">OS Type</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.os_type}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.os_sku && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">OS SKU</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.os_sku}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.enable_auto_scaling !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.node_pool.enable_auto_scaling ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Auto Scaling</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.node_pool.enable_auto_scaling ? 'Enabled' : 'Disabled'}
                            {formData.node_pool.enable_auto_scaling && (
                              <span className="ml-2">
                                ({formData.node_pool.min_count || 1} - {formData.node_pool.max_count || 10} nodes)
                              </span>
                            )}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.max_pods !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Max Pods</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.max_pods}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.mode && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Mode</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.mode}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.labels && Object.keys(formData.node_pool.labels).length > 0 && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Labels</p>
                          <div className="flex flex-wrap gap-2 mt-1">
                            {Object.entries(formData.node_pool.labels).map(([key, value]) => (
                              <Badge key={key} variant="secondary" className="text-xs">
                                {key}: {value}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.availability_zones && formData.node_pool.availability_zones.length > 0 && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Availability Zones</p>
                          <div className="flex flex-wrap gap-2 mt-1">
                            {formData.node_pool.availability_zones.map((zone) => (
                              <Badge key={zone} variant="secondary" className="text-xs">
                                {zone}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                </AccordionContent>
              </AccordionItem>
            )}

            {formData.security && (
              <AccordionItem value="security">
                <AccordionTrigger>
                  <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Security Configuration</span>
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-3 pt-2">
                    {formData.security.enable_rbac !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.security.enable_rbac ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">RBAC</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.security.enable_rbac ? 'Enabled' : 'Disabled'}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.security.enable_pod_security_policy !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.security.enable_pod_security_policy ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Pod Security Policy</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.security.enable_pod_security_policy ? 'Enabled' : 'Disabled'}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.security.enable_private_cluster !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.security.enable_private_cluster ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Private Cluster</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.security.enable_private_cluster ? 'Enabled' : 'Disabled'}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.security.api_server_authorized_ip_ranges && formData.security.api_server_authorized_ip_ranges.length > 0 && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">API Server Authorized IP Ranges</p>
                          <div className="flex flex-wrap gap-2 mt-1">
                            {formData.security.api_server_authorized_ip_ranges.map((ipRange) => (
                              <Badge key={ipRange} variant="secondary" className="text-xs">
                                {ipRange}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                    {formData.security.enable_azure_policy !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.security.enable_azure_policy ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Azure Policy</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.security.enable_azure_policy ? 'Enabled' : 'Disabled'}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.security.enable_workload_identity !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.security.enable_workload_identity ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Workload Identity</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.security.enable_workload_identity ? 'Enabled' : 'Disabled'}
                          </p>
                        </div>
                      </div>
                    )}
                  </div>
                </AccordionContent>
              </AccordionItem>
            )}

            {formData.tags && Object.keys(formData.tags).length > 0 && (
              <AccordionItem value="tags">
                <AccordionTrigger>
                  <div className="flex items-center gap-2">
                    <Tag className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('kubernetes.review.tags')} ({Object.keys(formData.tags).length})</span>
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

            {!formData.node_pool && !formData.security && (!formData.tags || Object.keys(formData.tags).length === 0) && (
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
                <CardTitle className="text-lg font-semibold">{t('kubernetes.review.apiRequestPreview')}</CardTitle>
                <CardDescription className="text-xs mt-1">
                  {t('kubernetes.review.apiRequestPreviewDescription')}
                </CardDescription>
              </div>
            </div>
            <button
              type="button"
              onClick={() => setIsApiPreviewOpen(!isApiPreviewOpen)}
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              {isApiPreviewOpen ? t('kubernetes.review.hide') : t('kubernetes.review.show')}
            </button>
          </div>
        </CardHeader>
        {isApiPreviewOpen && (
          <CardContent>
            <pre className="bg-muted dark:bg-muted/50 p-4 rounded-md text-xs overflow-auto max-h-96 border">
              {JSON.stringify(buildAzurePayload(formData), null, 2)}
            </pre>
          </CardContent>
        )}
      </Card>
    </>
  );
}


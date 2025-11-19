/**
 * GCP Review Configuration Component
 * GCP GKE 클러스터 생성 Review Step의 Provider별 설정 표시
 */

'use client';

import { useState, useMemo } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Settings, Code, Tag, Shield, CheckCircle2, Layers, Database } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import type { CreateClusterForm } from '@/lib/types';
import { buildGCPPayload } from './gcp-review-payload';

interface GCPReviewConfigProps {
  formData: CreateClusterForm;
  selectedProjectId?: string;
}

export function GCPReviewConfig({ formData, selectedProjectId }: GCPReviewConfigProps) {
  const { t } = useTranslation();
  const [isApiPreviewOpen, setIsApiPreviewOpen] = useState(false);

  // project_id는 props로 전달받거나 formData에서 가져오기
  const projectId = useMemo(() => {
    if (selectedProjectId && selectedProjectId.trim() !== '') {
      return selectedProjectId;
    }
    if (formData.project_id && formData.project_id.trim() !== '') {
      return formData.project_id;
    }
    return undefined;
  }, [selectedProjectId, formData.project_id]);

  // payload 생성 (project_id 포함)
  const payload = useMemo(() => {
    if (!projectId || projectId.trim() === '') {
      return null;
    }
    try {
      const formDataWithProjectId = { ...formData, project_id: projectId };
      return buildGCPPayload(formDataWithProjectId);
    } catch (error) {
      console.error('Failed to build GCP payload:', error);
      return null;
    }
  }, [formData, projectId]);

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
            {formData.cluster_mode && (
              <AccordionItem value="cluster-mode">
                <AccordionTrigger>
                  <div className="flex items-center gap-2">
                    <Database className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Cluster Mode</span>
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-3 pt-2">
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                      <div className="flex-1">
                        <p className="text-sm font-medium">Type</p>
                        <p className="text-sm text-muted-foreground">
                          {formData.cluster_mode.type || 'standard'}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className={`h-4 w-4 ${formData.cluster_mode.remove_default_node_pool ? 'text-green-600' : 'text-gray-400'}`} />
                      <div className="flex-1">
                        <p className="text-sm font-medium">Remove Default Node Pool</p>
                        <p className="text-sm text-muted-foreground">
                          {formData.cluster_mode.remove_default_node_pool ? 'Enabled' : 'Disabled'}
                        </p>
                      </div>
                    </div>
                  </div>
                </AccordionContent>
              </AccordionItem>
            )}

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
                    {formData.node_pool.machine_type && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Machine Type</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.machine_type}</p>
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
                    {formData.node_pool.disk_size_gb !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Disk Size (GB)</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.disk_size_gb}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.disk_type && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Disk Type</p>
                          <p className="text-sm text-muted-foreground">{formData.node_pool.disk_type}</p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.auto_scaling && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.node_pool.auto_scaling.enabled ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Auto Scaling</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.node_pool.auto_scaling.enabled ? 'Enabled' : 'Disabled'}
                            {formData.node_pool.auto_scaling.enabled && (
                              <span className="ml-2">
                                ({formData.node_pool.auto_scaling.min_node_count || 1} - {formData.node_pool.auto_scaling.max_node_count || 10} nodes)
                              </span>
                            )}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.preemptible !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.node_pool.preemptible ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Preemptible</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.node_pool.preemptible ? 'Enabled' : 'Disabled'}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.node_pool.spot !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.node_pool.spot ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Spot</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.node_pool.spot ? 'Enabled' : 'Disabled'}
                          </p>
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
                    {formData.security.binary_authorization !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.security.binary_authorization ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Binary Authorization</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.security.binary_authorization ? 'Enabled' : 'Disabled'}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.security.network_policy !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.security.network_policy ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Network Policy</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.security.network_policy ? 'Enabled' : 'Disabled'}
                          </p>
                        </div>
                      </div>
                    )}
                    {formData.security.pod_security_policy !== undefined && (
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className={`h-4 w-4 ${formData.security.pod_security_policy ? 'text-green-600' : 'text-gray-400'}`} />
                        <div className="flex-1">
                          <p className="text-sm font-medium">Pod Security Policy</p>
                          <p className="text-sm text-muted-foreground">
                            {formData.security.pod_security_policy ? 'Enabled' : 'Disabled'}
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

            {projectId && (
              <AccordionItem value="project-id">
                <AccordionTrigger>
                  <div className="flex items-center gap-2">
                    <Database className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('kubernetes.gcp.projectId')}</span>
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-3 pt-2">
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                      <div className="flex-1">
                        <p className="text-sm font-medium">{t('kubernetes.gcp.projectId')}</p>
                        <p className="text-sm text-muted-foreground">
                          {projectId}
                        </p>
                      </div>
                    </div>
                  </div>
                </AccordionContent>
              </AccordionItem>
            )}

            {(!projectId && !formData.cluster_mode && !formData.node_pool && !formData.security && (!formData.tags || Object.keys(formData.tags).length === 0)) && (
              <div className="py-4 text-sm text-muted-foreground text-center">
                {t('kubernetes.review.noAdvancedSettings')}
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
            {payload ? (
              <pre className="bg-muted dark:bg-muted/50 p-4 rounded-md text-xs overflow-auto max-h-96 border">
                {JSON.stringify(payload, null, 2)}
              </pre>
            ) : (
              <div className="p-4 text-sm text-destructive">
                {!projectId ? (
                  <div>
                    <p>{t('kubernetes.review.missingProjectId')}</p>
                    <p className="mt-2 text-xs text-muted-foreground">
                      {t('kubernetes.review.payloadError')}
                    </p>
                  </div>
                ) : (
                  <p>{t('kubernetes.review.payloadError')}</p>
                )}
              </div>
            )}
          </CardContent>
        )}
      </Card>
    </>
  );
}


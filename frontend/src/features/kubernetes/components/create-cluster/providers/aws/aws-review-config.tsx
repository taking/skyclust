/**
 * AWS Review Configuration Component
 * AWS EKS 클러스터 생성 Review Step의 Provider별 설정 표시
 */

'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Settings, Code, Tag, Shield, CheckCircle2 } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import type { CreateClusterForm } from '@/lib/types';
import { buildAWSPayload } from './aws-review-payload';

interface AWSReviewConfigProps {
  formData: CreateClusterForm;
}

export function AWSReviewConfig({ formData }: AWSReviewConfigProps) {
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
              {JSON.stringify(buildAWSPayload(formData), null, 2)}
            </pre>
          </CardContent>
        )}
      </Card>
    </>
  );
}


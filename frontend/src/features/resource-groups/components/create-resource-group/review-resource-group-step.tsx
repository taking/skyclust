/**
 * Review Resource Group Step
 * Step 2: Resource Group 생성 전 최종 확인
 */

'use client';

import * as React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { CreateResourceGroupForm } from '@/features/resource-groups/hooks/use-resource-group-actions';

interface ReviewResourceGroupStepProps {
  formData: CreateResourceGroupForm;
}

export function ReviewResourceGroupStep({
  formData,
}: ReviewResourceGroupStepProps) {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Resource Group Configuration</CardTitle>
          <CardDescription>
            Review your resource group settings before creating
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-muted-foreground">Name</label>
              <p className="text-sm font-medium mt-1">{formData.name || '-'}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-muted-foreground">Location</label>
              <p className="text-sm font-medium mt-1">{formData.location || '-'}</p>
            </div>
          </div>

          {formData.tags && Object.keys(formData.tags).length > 0 && (
            <div>
              <label className="text-sm font-medium text-muted-foreground">Tags</label>
              <div className="flex flex-wrap gap-2 mt-2">
                {Object.entries(formData.tags).map(([key, value]) => (
                  <Badge key={key} variant="outline">
                    {key}: {value}
                  </Badge>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}


/**
 * Subnets Page Header Component
 * Subnets 페이지 헤더 컴포넌트
 */

'use client';

import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';

interface SubnetsPageHeaderProps {
  selectedProvider?: string;
  selectedCredentialId?: string;
  selectedVPCId?: string;
  vpcs: Array<{ id: string; name?: string }>;
  onVPCChange: (vpcId: string) => void;
  onCreateClick?: () => void;
  disabled?: boolean;
}

export function SubnetsPageHeader({
  selectedProvider,
  selectedCredentialId,
  selectedVPCId,
  vpcs,
  onVPCChange,
  onCreateClick,
  disabled = false,
}: SubnetsPageHeaderProps) {
  const { currentWorkspace } = useWorkspaceStore();
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{t('network.subnets')}</h1>
          <p className="text-gray-600 mt-1">
            {currentWorkspace 
              ? t('network.manageSubnetsWithWorkspace', { workspaceName: currentWorkspace.name }) 
              : t('network.manageSubnets')
            }
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {onCreateClick && (
            <Button
              onClick={onCreateClick}
              disabled={disabled || !selectedVPCId}
            >
              <Plus className="mr-2 h-4 w-4" />
              {t('network.createSubnet') || 'Create Subnet'}
            </Button>
          )}
        </div>
      </div>
      
      {/* Configuration Card - VPC Selection */}
      {selectedProvider && selectedCredentialId && (
        <Card>
          <CardHeader>
            <CardTitle>{t('common.configuration')}</CardTitle>
            <CardDescription>{t('network.selectVPCToViewSubnets')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Label>VPC *</Label>
              <Select
                value={selectedVPCId}
                onValueChange={onVPCChange}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t('network.selectVPC')} />
                </SelectTrigger>
                <SelectContent>
                  {vpcs.map((vpc) => (
                    <SelectItem key={vpc.id} value={vpc.id}>
                      {vpc.name || vpc.id}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

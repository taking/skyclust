/**
 * Security Groups Page Header Component
 * Security Groups 페이지 헤더 컴포넌트
 */

'use client';

import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface SecurityGroupsPageHeaderProps {
  selectedProvider?: string;
  selectedCredentialId?: string;
  selectedVPCId?: string;
  vpcs: Array<{ id: string; name?: string }>;
  onVPCChange: (vpcId: string) => void;
}

export function SecurityGroupsPageHeader({
  selectedProvider,
  selectedCredentialId,
  selectedVPCId,
  vpcs,
  onVPCChange,
}: SecurityGroupsPageHeaderProps) {
  const { currentWorkspace } = useWorkspaceStore();
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{t('network.securityGroups')}</h1>
          <p className="text-gray-600 mt-1">
            {currentWorkspace 
              ? t('network.manageSecurityGroupsWithWorkspace', { workspaceName: currentWorkspace.name }) 
              : t('network.manageSecurityGroups')
            }
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {/* Credential selection is now handled in Header */}
        </div>
      </div>
      
      {/* Configuration Card - VPC Selection */}
      {selectedProvider && selectedCredentialId && (
        <Card>
          <CardHeader>
            <CardTitle>{t('common.configuration')}</CardTitle>
            <CardDescription>{t('network.selectVPCToViewSecurityGroups')}</CardDescription>
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

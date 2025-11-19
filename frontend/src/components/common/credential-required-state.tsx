/**
 * Credential Required State Component
 * Credential이 필요할 때 표시되는 안내 컴포넌트
 */

'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Key, ExternalLink } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';
import { useWorkspaceStore } from '@/store/workspace';
import { buildManagementPath } from '@/lib/routing/helpers';

interface CredentialRequiredStateProps {
  title?: string;
  description?: string;
  serviceName?: string;
}

function CredentialRequiredStateComponent({
  title,
  description,
  serviceName,
}: CredentialRequiredStateProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const { currentWorkspace } = useWorkspaceStore();

  const defaultTitle = title || t('components.credentialRequired.title');
  const defaultDescription = description || t('components.credentialRequired.description', { 
    serviceName: serviceName || t('credential.title') 
  });

  const handleGoToCredentials = () => {
    if (currentWorkspace?.id) {
      router.push(buildManagementPath(currentWorkspace.id, 'credentials'));
    } else {
      router.push('/credentials');
    }
  };

  return (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Key className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">{defaultTitle}</h3>
        <p className="text-sm text-gray-500 text-center mb-4">{defaultDescription}</p>
        <Button
          onClick={handleGoToCredentials}
          variant="default"
        >
          <Key className="mr-2 h-4 w-4" />
          {t('components.credentialRequired.registerButton')}
          <ExternalLink className="ml-2 h-4 w-4" />
        </Button>
      </CardContent>
    </Card>
  );
}

export const CredentialRequiredState = React.memo(CredentialRequiredStateComponent);


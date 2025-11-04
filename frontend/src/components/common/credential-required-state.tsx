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

interface CredentialRequiredStateProps {
  title?: string;
  description?: string;
  serviceName?: string;
}

function CredentialRequiredStateComponent({
  title,
  description,
  serviceName = 'this service',
}: CredentialRequiredStateProps) {
  const router = useRouter();

  const defaultTitle = title || 'No Credentials Found';
  const defaultDescription = description || `To use ${serviceName}, you need to register cloud credentials first.`;

  return (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Key className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">{defaultTitle}</h3>
        <p className="text-sm text-gray-500 text-center mb-4">{defaultDescription}</p>
        <Button
          onClick={() => router.push('/credentials')}
          variant="default"
        >
          <Key className="mr-2 h-4 w-4" />
          Register Credentials
          <ExternalLink className="ml-2 h-4 w-4" />
        </Button>
      </CardContent>
    </Card>
  );
}

export const CredentialRequiredState = React.memo(CredentialRequiredStateComponent);


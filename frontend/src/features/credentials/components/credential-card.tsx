/**
 * Credential Card Component
 * 개별 Credential 카드 컴포넌트
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Eye, EyeOff, Edit, Trash2 } from 'lucide-react';
import type { Credential } from '@/lib/types';

interface CredentialCardProps {
  credential: Credential;
  showCredentials: boolean;
  onToggleShow: () => void;
  onEdit: () => void;
  onDelete: () => void;
  isDeleting?: boolean;
}

function getProviderIcon(provider: string): string {
  switch (provider.toLowerCase()) {
    case 'aws':
      return '☁️';
    case 'gcp':
      return '🌐';
    case 'azure':
      return '🔷';
    default:
      return '🔑';
  }
}

function getProviderBadgeVariant(provider: string): 'default' | 'secondary' | 'outline' {
  switch (provider.toLowerCase()) {
    case 'aws':
      return 'default';
    case 'gcp':
      return 'secondary';
    case 'azure':
      return 'outline';
    default:
      return 'outline';
  }
}

export function CredentialCard({
  credential,
  showCredentials,
  onToggleShow,
  onEdit,
  onDelete,
  isDeleting = false,
}: CredentialCardProps) {
  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex items-center space-x-2">
            <span className="text-2xl">{getProviderIcon(credential.provider)}</span>
            <div>
              <CardTitle className="text-lg">{credential.provider.toUpperCase()}</CardTitle>
              <CardDescription>
                Cloud provider credentials
              </CardDescription>
            </div>
          </div>
          <Badge variant={getProviderBadgeVariant(credential.provider)}>
            {credential.provider}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          <div className="text-sm text-gray-500">
            Created {new Date(credential.created_at).toLocaleDateString()}
          </div>
          <div className="flex items-center justify-between">
            <Button
              variant="outline"
              size="sm"
              onClick={onToggleShow}
              className="flex-1 min-w-0"
            >
              {showCredentials ? (
                <>
                  <EyeOff className="mr-1 h-3 w-3 md:mr-2 md:h-4 md:w-4" />
                  <span className="hidden sm:inline">Hide</span>
                </>
              ) : (
                <>
                  <Eye className="mr-1 h-3 w-3 md:mr-2 md:h-4 md:w-4" />
                  <span className="hidden sm:inline">Show</span>
                </>
              )}
            </Button>
            <div className="flex space-x-1">
              <Button
                variant="ghost"
                size="sm"
                onClick={onEdit}
                className="h-8 w-8 p-0"
              >
                <Edit className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={onDelete}
                className="h-8 w-8 p-0"
                disabled={isDeleting}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
          {showCredentials && (
            <div className="mt-4 p-3 bg-gray-50 rounded-md">
              <div className="text-xs text-gray-600">
                <strong>Encrypted credentials stored securely</strong>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}


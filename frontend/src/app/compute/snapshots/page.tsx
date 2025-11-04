/**
 * VM Snapshots Page
 * VM 스냅샷 관리 페이지
 */

'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { HardDrive } from 'lucide-react';
import { useRequireAuth } from '@/hooks/use-auth';
import { useWorkspaceStore } from '@/store/workspace';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { Layout } from '@/components/layout/layout';
import { useCredentials } from '@/hooks/use-credentials';

export default function SnapshotsPage() {
  const { currentWorkspace } = useWorkspaceStore();
  const { isLoading: authLoading } = useRequireAuth();
  const [selectedCredentialId, setSelectedCredentialId] = useState<string>('');

  // Fetch credentials using unified hook
  const { credentials } = useCredentials({
    workspaceId: currentWorkspace?.id,
  });

  if (authLoading) {
    return (
      <WorkspaceRequired>
        <Layout>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
              <p className="mt-2 text-gray-600">Loading...</p>
            </div>
          </div>
        </Layout>
      </WorkspaceRequired>
    );
  }

  return (
    <WorkspaceRequired>
      <Layout>
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">VM Snapshots</h1>
              <p className="text-gray-600 mt-1">
                Manage VM Snapshots{currentWorkspace ? ` for ${currentWorkspace.name}` : ''}
              </p>
            </div>
            <div className="flex items-center space-x-2">
              <Select
                value={selectedCredentialId || ''}
                onValueChange={setSelectedCredentialId}
              >
                <SelectTrigger className="w-[250px]">
                  <SelectValue placeholder="Select Credential" />
                </SelectTrigger>
                <SelectContent>
                  {credentials.map((credential) => (
                    <SelectItem key={credential.id} value={credential.id}>
                      {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Empty State - API not implemented yet */}
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <HardDrive className="h-12 w-12 text-gray-400 mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">Coming Soon</h3>
              <p className="text-sm text-gray-500 text-center">
                VM Snapshots management feature is coming soon. This page will allow you to manage and view VM snapshots across different cloud providers.
              </p>
            </CardContent>
          </Card>
        </div>
      </Layout>
    </WorkspaceRequired>
  );
}


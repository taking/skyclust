/**
 * Workspace Members Tab
 * Workspace 상세 페이지의 Members 탭
 * 
 * 멤버 관리 기능 제공
 */

'use client';

import { Suspense } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Users, Plus, Trash2 } from 'lucide-react';
import { workspaceService } from '../services/workspace';
import { queryKeys } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handling';
import { Spinner } from '@/components/ui/spinner';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { useState, useCallback } from 'react';
import type { WorkspaceMember } from '@/lib/types';

interface WorkspaceMembersTabProps {
  workspaceId: string;
}

export function WorkspaceMembersTab({ workspaceId }: WorkspaceMembersTabProps) {
  const { t } = useTranslation();
  const { success, error: showError } = useToast();
  const queryClient = useQueryClient();
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    userId: string | null;
    userName?: string;
  }>({
    open: false,
    userId: null,
    userName: undefined,
  });

  const { data: members = [], isLoading } = useQuery({
    queryKey: queryKeys.workspaces.members(workspaceId),
    queryFn: () => workspaceService.getMembers(workspaceId),
    enabled: !!workspaceId,
  });

  const removeMemberMutation = useMutation({
    mutationFn: (userId: string) => workspaceService.removeMember(workspaceId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.members(workspaceId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.detail(workspaceId) });
      success(t('workspace.memberRemoved') || 'Member removed successfully');
      setDeleteDialogState({ open: false, userId: null });
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'removeMember', source: 'workspace-members-tab' });
      showError(t('workspace.memberRemoveFailed') || 'Failed to remove member');
    },
  });

  const handleDeleteMember = useCallback((userId: string, userName?: string) => {
    setDeleteDialogState({
      open: true,
      userId,
      userName,
    });
  }, []);

  const handleConfirmDelete = useCallback(() => {
    if (!deleteDialogState.userId) return;
    removeMemberMutation.mutate(deleteDialogState.userId);
  }, [deleteDialogState.userId, removeMemberMutation]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" label="Loading members..." />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center">
                <Users className="mr-2 h-5 w-5" />
                {t('workspace.members') || 'Members'}
              </CardTitle>
              <CardDescription>
                {t('workspace.membersDescription') || 'Manage workspace members and their roles'}
              </CardDescription>
            </div>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              {t('workspace.addMember') || 'Add Member'}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {members.length === 0 ? (
            <div className="text-center py-12">
              <Users className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">{t('workspace.noMembers') || 'No members found'}</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('workspace.memberName') || 'Name'}</TableHead>
                  <TableHead>{t('workspace.memberEmail') || 'Email'}</TableHead>
                  <TableHead>{t('workspace.memberRole') || 'Role'}</TableHead>
                  <TableHead>{t('workspace.memberJoinedAt') || 'Joined At'}</TableHead>
                  <TableHead className="text-right">{t('common.actions') || 'Actions'}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {members.map((member) => (
                  <TableRow key={member.user_id}>
                    <TableCell className="font-medium">{member.name || member.user_id}</TableCell>
                    <TableCell>{member.email || '-'}</TableCell>
                    <TableCell>
                      <Badge variant={member.role === 'owner' ? 'default' : 'secondary'}>
                        {member.role}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {member.joined_at ? new Date(member.joined_at).toLocaleDateString() : '-'}
                    </TableCell>
                    <TableCell className="text-right">
                      {member.role !== 'owner' && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteMember(member.user_id, member.name)}
                        >
                          <Trash2 className="h-4 w-4 text-red-600" />
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ open, userId: deleteDialogState.userId })}
        onConfirm={handleConfirmDelete}
        title={t('workspace.removeMember') || 'Remove Member'}
        description={t('workspace.confirmRemoveMember', { userName: deleteDialogState.userName || deleteDialogState.userId || '' }) || 'Are you sure you want to remove this member?'}
        isLoading={removeMemberMutation.isPending}
        resourceName={deleteDialogState.userName || deleteDialogState.userId || undefined}
        resourceNameLabel={t('workspace.memberName') || 'Member Name'}
      />
    </div>
  );
}


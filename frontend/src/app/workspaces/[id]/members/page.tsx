'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { workspaceService } from '@/features/workspaces';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { ArrowLeft, UserPlus, Trash2, Crown, Shield, User, Users } from 'lucide-react';
import { queryKeys } from '@/lib/query';
import { createValidationSchemas } from '@/lib/validation';
import { useTranslation } from '@/hooks/use-translation';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';

interface WorkspaceMember {
  user_id: string;
  workspace_id: string;
  role: 'owner' | 'admin' | 'member';
  joined_at: string;
  user: {
    id: string;
    username: string;
    email: string;
  };
}

export default function WorkspaceMembersPage() {
  const { t } = useTranslation();
  const { addMemberSchema } = createValidationSchemas(t);
  const params = useParams();
  const router = useRouter();
  const workspaceId = params.id as string;
  const [isAddMemberDialogOpen, setIsAddMemberDialogOpen] = useState(false);
  const [deleteMemberDialogState, setDeleteMemberDialogState] = useState<{
    open: boolean;
    userId: string | null;
  }>({
    open: false,
    userId: null,
  });
  const queryClient = useQueryClient();
  const { success, error } = useToast();
  const { handleError } = useErrorHandler();

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
    // watch,
  } = useForm<z.infer<typeof addMemberSchema>>({
    resolver: zodResolver(addMemberSchema),
  });

  // const selectedRole = watch('role');

  // Fetch workspace details
  const { data: workspace } = useQuery({
    queryKey: queryKeys.workspaces.detail(workspaceId),
    queryFn: () => workspaceService.getWorkspace(workspaceId),
  });

  // Fetch workspace members
  const { data: members = [], isLoading, error: membersError } = useQuery({
    queryKey: queryKeys.workspaces.members(workspaceId),
    queryFn: () => workspaceService.getMembers(workspaceId),
    enabled: !!workspaceId,
    retry: 1,
  });

  // Add member mutation
  const addMemberMutation = useMutation({
    mutationFn: async (data: { email: string; role: string }) => {
      await workspaceService.addMember(workspaceId, data.email, data.role);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.members(workspaceId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.detail(workspaceId) });
      setIsAddMemberDialogOpen(false);
      reset();
      success('Member added successfully');
    },
    onError: (err) => {
      handleError(err, { operation: 'addMember', workspaceId });
    },
  });

  // Remove member mutation
  const removeMemberMutation = useMutation({
    mutationFn: async (userId: string) => {
      await workspaceService.removeMember(workspaceId, userId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.members(workspaceId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.detail(workspaceId) });
      success('Member removed successfully');
    },
    onError: (err) => {
      handleError(err, { operation: 'removeMember', workspaceId });
    },
  });

  const handleAddMember = (data: z.infer<typeof addMemberSchema>) => {
    addMemberMutation.mutate(data);
  };

  const handleRemoveMember = (userId: string) => {
    setDeleteMemberDialogState({ open: true, userId });
  };

  const handleConfirmRemoveMember = () => {
    if (deleteMemberDialogState.userId) {
      removeMemberMutation.mutate(deleteMemberDialogState.userId);
      setDeleteMemberDialogState({ open: false, userId: null });
    }
  };

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'owner':
        return <Crown className="h-4 w-4 text-yellow-500" />;
      case 'admin':
        return <Shield className="h-4 w-4 text-blue-500" />;
      default:
        return <User className="h-4 w-4 text-gray-500" />;
    }
  };

  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case 'owner':
        return 'default';
      case 'admin':
        return 'secondary';
      default:
        return 'outline';
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading members...</p>
        </div>
      </div>
    );
  }

  if (membersError) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 text-lg font-medium">Failed to load members</p>
          <p className="text-gray-500 mt-2">Please try again later</p>
          <Button
            variant="outline"
            onClick={() => queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.members(workspaceId) })}
            className="mt-4"
          >
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <Button
            variant="ghost"
            onClick={() => router.back()}
            className="mb-4"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back
          </Button>
          <h1 className="text-3xl font-bold text-gray-900">
            {workspace?.name} Members
          </h1>
          <p className="text-gray-600">
            Manage workspace members and their permissions
          </p>
        </div>

        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-semibold">Members ({members.length})</h2>
          <Dialog open={isAddMemberDialogOpen} onOpenChange={setIsAddMemberDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <UserPlus className="mr-2 h-4 w-4" />
                Add Member
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add Member</DialogTitle>
                <DialogDescription>
                  Invite a new member to this workspace by email.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit(handleAddMember)} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="email">Email Address</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="Enter member email"
                    {...register('email')}
                  />
                  {errors.email && (
                    <p className="text-sm text-red-600">{errors.email.message}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="role">Role</Label>
                  <Select onValueChange={(value) => setValue('role', value as 'admin' | 'member')}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="member">Member</SelectItem>
                      <SelectItem value="admin">Admin</SelectItem>
                    </SelectContent>
                  </Select>
                  {errors.role && (
                    <p className="text-sm text-red-600">{errors.role.message}</p>
                  )}
                </div>
                {addMemberMutation.isError && (
                  <div className="bg-red-50 border border-red-200 rounded-md p-3">
                    <p className="text-sm text-red-800">
                      {addMemberMutation.error instanceof Error
                        ? addMemberMutation.error.message
                        : 'Failed to add member. Please try again.'}
                    </p>
                  </div>
                )}
                <div className="flex justify-end space-x-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setIsAddMemberDialogOpen(false);
                      reset();
                    }}
                    disabled={addMemberMutation.isPending}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={addMemberMutation.isPending}>
                    {addMemberMutation.isPending ? 'Adding...' : 'Add Member'}
                  </Button>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        <div className="space-y-4">
          {members.map((member: WorkspaceMember) => (
            <Card key={member.user_id}>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <Avatar>
                      <AvatarImage src="" alt={member.user.username} />
                      <AvatarFallback>
                        {member.user.username.charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <h3 className="font-medium text-gray-900">{member.user.username}</h3>
                      <p className="text-sm text-gray-500">{member.user.email}</p>
                      <p className="text-xs text-gray-400">
                        Joined {new Date(member.joined_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-3">
                    <Badge variant={getRoleBadgeVariant(member.role)}>
                      <div className="flex items-center space-x-1">
                        {getRoleIcon(member.role)}
                        <span className="capitalize">{member.role}</span>
                      </div>
                    </Badge>
                    {member.role !== 'owner' && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveMember(member.user_id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {members.length === 0 && (
          <div className="text-center py-12">
            <div className="mx-auto h-12 w-12 text-gray-400">
              <Users className="h-12 w-12" />
            </div>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No members</h3>
            <p className="mt-1 text-sm text-gray-500">
              This workspace doesn&apos;t have any members yet.
            </p>
          </div>
        )}
      </div>

      {/* Delete Member Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteMemberDialogState.open}
        onOpenChange={(open) => setDeleteMemberDialogState({ ...deleteMemberDialogState, open })}
        onConfirm={handleConfirmRemoveMember}
        title={t('workspace.removeMember')}
        description="이 멤버를 워크스페이스에서 제거하시겠습니까?"
        isLoading={removeMemberMutation.isPending}
      />
    </div>
  );
}

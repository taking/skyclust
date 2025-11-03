'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
// import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { credentialService } from '@/services/credential';
import { useWorkspaceStore } from '@/store/workspace';
import { useRouter } from 'next/navigation';
import { Plus, Key, Trash2, Edit, Eye, EyeOff, Home } from 'lucide-react';
import { CreateCredentialForm } from '@/lib/types';
import { WorkspaceRequired } from '@/components/common/workspace-required';

const createCredentialSchema = z.object({
  name: z.string().optional(),
  provider: z.string().min(1, 'Provider is required'),
  credentials: z.record(z.string(), z.unknown()),
});

export default function CredentialsPage() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [editingCredential, setEditingCredential] = useState<any>(null); // eslint-disable-line @typescript-eslint/no-explicit-any
  const [showCredentials, setShowCredentials] = useState<Record<string, boolean>>({});
  const [gcpInputMode, setGcpInputMode] = useState<'json' | 'file'>('json');
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const queryClient = useQueryClient();

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
    watch,
  } = useForm<CreateCredentialForm>({
    resolver: zodResolver(createCredentialSchema),
  });

  const selectedProvider = watch('provider');

  // Fetch credentials
  const { data: credentialsData, isLoading } = useQuery({
    queryKey: ['credentials', currentWorkspace?.id],
    queryFn: () => currentWorkspace ? credentialService.getCredentials(currentWorkspace.id) : Promise.resolve([]),
    enabled: !!currentWorkspace,
  });

  // Ensure credentials is always an array
  const credentials = Array.isArray(credentialsData) ? credentialsData : [];

  // Create credential mutation
  const createCredentialMutation = useMutation({
    mutationFn: credentialService.createCredential,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials', currentWorkspace?.id] });
      setIsCreateDialogOpen(false);
      reset();
      setGcpInputMode('json');
    },
  });

  // Create credential from file mutation (for GCP)
  const createCredentialFromFileMutation = useMutation({
    mutationFn: ({ workspaceId, name, provider, file }: { workspaceId: string; name: string; provider: string; file: File }) =>
      credentialService.createCredentialFromFile(workspaceId, name, provider, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials', currentWorkspace?.id] });
      setIsCreateDialogOpen(false);
      reset();
      setGcpInputMode('json');
    },
  });

  // Update credential mutation
  const updateCredentialMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateCredentialForm> }) =>
      credentialService.updateCredential(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials', currentWorkspace?.id] });
      setIsEditDialogOpen(false);
      setEditingCredential(null);
      reset();
    },
  });

  // Delete credential mutation
  const deleteCredentialMutation = useMutation({
    mutationFn: credentialService.deleteCredential,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials', currentWorkspace?.id] });
    },
  });

  const handleCreateCredential = async (data: CreateCredentialForm) => {
    if (!currentWorkspace) return;
    
    // Handle GCP file upload
    if (data.provider === 'gcp' && gcpInputMode === 'file') {
      const file = (data.credentials as any)?._file as File; // eslint-disable-line @typescript-eslint/no-explicit-any
      if (!file) {
        alert('Please select a GCP service account JSON file');
        return;
      }
      
      createCredentialFromFileMutation.mutate({
        workspaceId: currentWorkspace.id,
        name: data.name || 'GCP Production',
        provider: 'gcp',
        file,
      });
      return;
    }
    
    // Transform credentials object to data field (remove _file if present)
    const credentials = { ...data.credentials };
    delete (credentials as any)._file; // eslint-disable-line @typescript-eslint/no-explicit-any
    
    const requestData = {
      workspace_id: currentWorkspace.id,
      name: data.name || `${data.provider.toUpperCase()} Credential`,
      provider: data.provider,
      data: credentials || {},
    };
    createCredentialMutation.mutate(requestData as any); // eslint-disable-line @typescript-eslint/no-explicit-any
  };

  const handleEditCredential = (credential: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
    setEditingCredential(credential);
    setValue('provider', credential.provider);
    setIsEditDialogOpen(true);
  };

  const handleUpdateCredential = (data: CreateCredentialForm) => {
    if (!editingCredential) return;
    updateCredentialMutation.mutate({
      id: editingCredential.id,
      data,
    });
  };

  const handleDeleteCredential = (credentialId: string) => {
    if (confirm('Are you sure you want to delete this credential?')) {
      deleteCredentialMutation.mutate(credentialId);
    }
  };

  const toggleShowCredentials = (credentialId: string) => {
    setShowCredentials(prev => ({
      ...prev,
      [credentialId]: !prev[credentialId]
    }));
  };

  const getProviderIcon = (provider: string) => {
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
  };

  const getProviderBadgeVariant = (provider: string) => {
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
  };

  if (!currentWorkspace) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">
            No Workspace Selected
          </h2>
          <p className="text-gray-600 mb-6">
            Please select a workspace to manage credentials.
          </p>
          <Button onClick={() => router.push('/workspaces')}>
            Select Workspace
          </Button>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <WorkspaceRequired>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">Loading credentials...</p>
          </div>
        </div>
      </WorkspaceRequired>
    );
  }

  return (
    <WorkspaceRequired>
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col space-y-4 md:flex-row md:justify-between md:items-center md:space-y-0 mb-6 md:mb-8">
              <div>
                <h1 className="text-2xl md:text-3xl font-bold text-gray-900">Credentials</h1>
                <p className="text-sm md:text-base text-gray-600">
                  {currentWorkspace ? `Manage cloud provider credentials for ${currentWorkspace.name}` : 'Manage cloud provider credentials'}
                </p>
              </div>
              <div className="flex items-center space-x-2">
                <Button variant="outline" onClick={() => router.push('/dashboard')}>
                  <Home className="mr-2 h-4 w-4" />
                  Home
                </Button>
                <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
                  <DialogTrigger asChild>
                    <Button className="w-full md:w-auto">
                      <Plus className="mr-2 h-4 w-4" />
                      Add Credentials
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="max-w-2xl">
                    <DialogHeader>
                      <DialogTitle>Add New Credentials</DialogTitle>
                      <DialogDescription>
                        Add credentials for a cloud provider to enable VM management.
                      </DialogDescription>
                    </DialogHeader>
                    <form onSubmit={handleSubmit(handleCreateCredential)} className="space-y-4">
                      <div className="space-y-2">
                        <Label htmlFor="name">Name</Label>
                        <Input
                          id="name"
                          placeholder="e.g., AWS Production"
                          {...register('name')}
                        />
                      </div>
                      
                      <div className="space-y-2">
                        <Label htmlFor="provider">Provider</Label>
                        <Select onValueChange={(value) => setValue('provider', value)}>
                          <SelectTrigger>
                            <SelectValue placeholder="Select provider" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="aws">AWS</SelectItem>
                            <SelectItem value="gcp">Google Cloud</SelectItem>
                            <SelectItem value="azure">Azure</SelectItem>
                          </SelectContent>
                        </Select>
                        {errors.provider && (
                          <p className="text-sm text-red-600">{errors.provider.message}</p>
                        )}
                      </div>
                      
                      {selectedProvider && (
                        <div className="space-y-4">
                          <div className="text-sm text-gray-600">
                            Enter your {selectedProvider.toUpperCase()} credentials:
                          </div>
                          
                          {selectedProvider === 'aws' && (
                            <div className="grid grid-cols-2 gap-4">
                              <div className="space-y-2">
                                <Label htmlFor="access_key">Access Key ID *</Label>
                                <Input
                                  id="access_key"
                                  placeholder="AKIA..."
                                  {...register('credentials.access_key')}
                                />
                              </div>
                              <div className="space-y-2">
                                <Label htmlFor="secret_key">Secret Access Key *</Label>
                                <Input
                                  id="secret_key"
                                  type="password"
                                  placeholder="Enter secret key"
                                  {...register('credentials.secret_key')}
                                />
                              </div>
                              <div className="space-y-2">
                                <Label htmlFor="region">Region</Label>
                                <Input
                                  id="region"
                                  placeholder="us-east-1"
                                  {...register('credentials.region')}
                                />
                              </div>
                            </div>
                          )}
                          
                          {selectedProvider === 'gcp' && (
                            <div className="space-y-4">
                              <div className="flex items-center space-x-4">
                                <Button
                                  type="button"
                                  variant={gcpInputMode === 'json' ? 'default' : 'outline'}
                                  onClick={() => setGcpInputMode('json')}
                                >
                                  JSON Input
                                </Button>
                                <Button
                                  type="button"
                                  variant={gcpInputMode === 'file' ? 'default' : 'outline'}
                                  onClick={() => setGcpInputMode('file')}
                                >
                                  File Upload
                                </Button>
                              </div>
                              
                              {gcpInputMode === 'json' ? (
                                <>
                                  <div className="text-sm font-medium">GCP Service Account JSON Fields:</div>
                                  <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                      <Label htmlFor="type">Type *</Label>
                                      <Input
                                        id="type"
                                        placeholder="service_account"
                                        defaultValue="service_account"
                                        {...register('credentials.type')}
                                      />
                                    </div>
                                    <div className="space-y-2">
                                      <Label htmlFor="project_id">Project ID *</Label>
                                      <Input
                                        id="project_id"
                                        placeholder="my-project-123"
                                        {...register('credentials.project_id')}
                                      />
                                    </div>
                                    <div className="space-y-2">
                                      <Label htmlFor="private_key_id">Private Key ID *</Label>
                                      <Input
                                        id="private_key_id"
                                        placeholder="Enter private key ID"
                                        {...register('credentials.private_key_id')}
                                      />
                                    </div>
                                    <div className="space-y-2">
                                      <Label htmlFor="private_key">Private Key *</Label>
                                      <Input
                                        id="private_key"
                                        type="password"
                                        placeholder="-----BEGIN PRIVATE KEY-----"
                                        {...register('credentials.private_key')}
                                      />
                                    </div>
                                    <div className="space-y-2">
                                      <Label htmlFor="client_email">Client Email *</Label>
                                      <Input
                                        id="client_email"
                                        type="email"
                                        placeholder="service-account@project.iam.gserviceaccount.com"
                                        {...register('credentials.client_email')}
                                      />
                                    </div>
                                    <div className="space-y-2">
                                      <Label htmlFor="client_id">Client ID *</Label>
                                      <Input
                                        id="client_id"
                                        placeholder="Enter client ID"
                                        {...register('credentials.client_id')}
                                      />
                                    </div>
                                    <div className="space-y-2 col-span-2">
                                      <Label htmlFor="auth_uri">Auth URI</Label>
                                      <Input
                                        id="auth_uri"
                                        defaultValue="https://accounts.google.com/o/oauth2/auth"
                                        {...register('credentials.auth_uri')}
                                      />
                                    </div>
                                    <div className="space-y-2 col-span-2">
                                      <Label htmlFor="token_uri">Token URI</Label>
                                      <Input
                                        id="token_uri"
                                        defaultValue="https://oauth2.googleapis.com/token"
                                        {...register('credentials.token_uri')}
                                      />
                                    </div>
                                    <div className="space-y-2 col-span-2">
                                      <Label htmlFor="auth_provider_x509_cert_url">Auth Provider X509 Cert URL</Label>
                                      <Input
                                        id="auth_provider_x509_cert_url"
                                        defaultValue="https://www.googleapis.com/oauth2/v1/certs"
                                        {...register('credentials.auth_provider_x509_cert_url')}
                                      />
                                    </div>
                                    <div className="space-y-2 col-span-2">
                                      <Label htmlFor="client_x509_cert_url">Client X509 Cert URL</Label>
                                      <Input
                                        id="client_x509_cert_url"
                                        placeholder="Enter client x509 cert URL"
                                        {...register('credentials.client_x509_cert_url')}
                                      />
                                    </div>
                                    <div className="space-y-2 col-span-2">
                                      <Label htmlFor="universe_domain">Universe Domain</Label>
                                      <Input
                                        id="universe_domain"
                                        defaultValue="googleapis.com"
                                        {...register('credentials.universe_domain')}
                                      />
                                    </div>
                                  </div>
                                </>
                              ) : (
                                <div className="space-y-2">
                                  <Label htmlFor="gcp_file">Service Account JSON File *</Label>
                                  <Input
                                    id="gcp_file"
                                    type="file"
                                    accept=".json"
                                    onChange={(e) => {
                                      const file = e.target.files?.[0];
                                      if (file) {
                                        setValue('credentials._file', file as any); // eslint-disable-line @typescript-eslint/no-explicit-any
                                      }
                                    }}
                                  />
                                  <p className="text-sm text-gray-500">
                                    Upload your GCP service account JSON file
                                  </p>
                                </div>
                              )}
                            </div>
                          )}
                          
                          {selectedProvider === 'azure' && (
                            <div className="grid grid-cols-2 gap-4">
                              <div className="space-y-2">
                                <Label htmlFor="subscription_id">Subscription ID</Label>
                                <Input
                                  id="subscription_id"
                                  placeholder="12345678-1234-1234-1234-123456789012"
                                  {...register('credentials.subscription_id')}
                                />
                              </div>
                              <div className="space-y-2">
                                <Label htmlFor="client_id">Client ID</Label>
                                <Input
                                  id="client_id"
                                  placeholder="12345678-1234-1234-1234-123456789012"
                                  {...register('credentials.client_id')}
                                />
                              </div>
                              <div className="space-y-2">
                                <Label htmlFor="client_secret">Client Secret</Label>
                                <Input
                                  id="client_secret"
                                  type="password"
                                  placeholder="Enter client secret"
                                  {...register('credentials.client_secret')}
                                />
                              </div>
                              <div className="space-y-2">
                                <Label htmlFor="tenant_id">Tenant ID</Label>
                                <Input
                                  id="tenant_id"
                                  placeholder="12345678-1234-1234-1234-123456789012"
                                  {...register('credentials.tenant_id')}
                                />
                              </div>
                            </div>
                          )}
                        </div>
                      )}
                      
                      <div className="flex justify-end space-x-2">
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() => setIsCreateDialogOpen(false)}
                        >
                          Cancel
                        </Button>
                        <Button type="submit" disabled={createCredentialMutation.isPending || createCredentialFromFileMutation.isPending}>
                          {(createCredentialMutation.isPending || createCredentialFromFileMutation.isPending) ? 'Adding...' : 'Add Credentials'}
                        </Button>
                      </div>
                  </form>
                </DialogContent>
              </Dialog>
            </div>
          </div>

        {credentials.length === 0 ? (
          <div className="text-center py-8 md:py-12">
            <div className="mx-auto h-10 w-10 md:h-12 md:w-12 text-gray-400">
              <Key className="h-10 w-10 md:h-12 md:w-12" />
            </div>
            <h3 className="mt-2 text-sm md:text-base font-medium text-gray-900">No credentials</h3>
            <p className="mt-1 text-xs md:text-sm text-gray-500">
              Add cloud provider credentials to start managing VMs.
            </p>
            <div className="mt-4 md:mt-6">
              <Button onClick={() => setIsCreateDialogOpen(true)} className="w-full md:w-auto">
                <Plus className="mr-2 h-4 w-4" />
                Add Credentials
              </Button>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6">
            {credentials.map((credential) => (
              <Card key={credential.id} className="hover:shadow-lg transition-shadow">
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
                        onClick={() => toggleShowCredentials(credential.id)}
                        className="flex-1 min-w-0"
                      >
                        {showCredentials[credential.id] ? (
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
                          onClick={() => handleEditCredential(credential)}
                          className="h-8 w-8 p-0"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteCredential(credential.id)}
                          className="h-8 w-8 p-0"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                    {showCredentials[credential.id] && (
                      <div className="mt-4 p-3 bg-gray-50 rounded-md">
                        <div className="text-xs text-gray-600">
                          <strong>Encrypted credentials stored securely</strong>
                        </div>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Edit Dialog */}
        <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Edit Credentials</DialogTitle>
              <DialogDescription>
                Update your cloud provider credentials.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleSubmit(handleUpdateCredential)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-provider">Provider</Label>
                <Input
                  id="edit-provider"
                  value={editingCredential?.provider || ''}
                  disabled
                />
              </div>
              
              <div className="space-y-4">
                <div className="text-sm text-gray-600">
                  Update your {editingCredential?.provider?.toUpperCase()} credentials:
                </div>
                
                {editingCredential?.provider === 'aws' && (
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="edit-access_key">Access Key ID</Label>
                      <Input
                        id="edit-access_key"
                        placeholder="AKIA..."
                        {...register('credentials.access_key')}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="edit-secret_key">Secret Access Key</Label>
                      <Input
                        id="edit-secret_key"
                        type="password"
                        placeholder="Enter secret key"
                        {...register('credentials.secret_key')}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="edit-region">Region</Label>
                      <Input
                        id="edit-region"
                        placeholder="us-east-1"
                        {...register('credentials.region')}
                      />
                    </div>
                  </div>
                )}
              </div>
              
              <div className="flex justify-end space-x-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setIsEditDialogOpen(false)}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={updateCredentialMutation.isPending}>
                  {updateCredentialMutation.isPending ? 'Updating...' : 'Update Credentials'}
                </Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
        </div>
      </div>
    </WorkspaceRequired>
  );
}

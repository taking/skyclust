'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Layout } from '@/components/layout/layout';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import * as z from 'zod';
import { networkService } from '@/services/network';
import { credentialService } from '@/services/credential';
import { useWorkspaceStore } from '@/store/workspace';
import { Plus, Trash2, Edit, Network, Shield, Layers, Search, Filter } from 'lucide-react';
import { CreateVPCForm, CreateSubnetForm, CreateSecurityGroupForm, CloudProvider } from '@/lib/types';
import { useToast } from '@/hooks/useToast';
import { useRequireAuth } from '@/hooks/useAuth';
import { useSearch } from '@/hooks/useSearch';
import { SearchBar } from '@/components/ui/search-bar';
import { FilterPanel, FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { useSSEMonitoring } from '@/hooks/useSSEMonitoring';
import { NetworkTopologyViewer } from '@/components/network/network-topology-viewer';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { Checkbox } from '@/components/ui/checkbox';
import { TagManager } from '@/components/common/tag-manager';
import { Pagination } from '@/components/ui/pagination';
import { usePagination } from '@/hooks/usePagination';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { WorkspaceRequired } from '@/components/common/workspace-required';

const createVPCSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters'),
  description: z.string().optional(),
  cidr_block: z.string().optional(),
  region: z.string().optional(),
  project_id: z.string().optional(),
  auto_create_subnets: z.boolean().optional(),
  routing_mode: z.string().optional(),
  mtu: z.number().optional(),
});

const createSubnetSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters'),
  vpc_id: z.string().min(1, 'VPC ID is required'),
  cidr_block: z.string().min(1, 'CIDR block is required'),
  availability_zone: z.string().min(1, 'Availability zone is required'),
  region: z.string().min(1, 'Region is required'),
  description: z.string().optional(),
  private_ip_google_access: z.boolean().optional(),
  flow_logs: z.boolean().optional(),
});

const createSecurityGroupSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters'),
  description: z.string().min(1, 'Description is required').max(255, 'Description must be less than 255 characters'),
  vpc_id: z.string().min(1, 'VPC ID is required'),
  region: z.string().min(1, 'Region is required'),
  project_id: z.string().optional(),
  direction: z.string().optional(),
  priority: z.number().optional(),
  action: z.string().optional(),
  protocol: z.string().optional(),
  ports: z.array(z.string()).optional(),
  source_ranges: z.array(z.string()).optional(),
  target_tags: z.array(z.string()).optional(),
});

export default function NetworksPage() {
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const { isLoading: authLoading } = useRequireAuth();

  const [activeTab, setActiveTab] = useState<'vpcs' | 'subnets' | 'security-groups'>('vpcs');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [selectedRegion, setSelectedRegion] = useState<string>('');
  const [selectedVPCId, setSelectedVPCId] = useState<string>('');
  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedVPCIds, setSelectedVPCIds] = useState<string[]>([]);
  const [selectedSubnetIds, setSelectedSubnetIds] = useState<string[]>([]);
  const [selectedSecurityGroupIds, setSelectedSecurityGroupIds] = useState<string[]>([]);
  const [isTagDialogOpen, setIsTagDialogOpen] = useState(false);
  const [bulkTagKey, setBulkTagKey] = useState('');
  const [bulkTagValue, setBulkTagValue] = useState('');
  const [pageSize, setPageSize] = useState(20);
  const [selectedCredentialId, setSelectedCredentialId] = useState<string>('');

  // SSE 실시간 업데이트
  useSSEMonitoring();

  // Form hooks for each resource type
  const vpcForm = useForm<CreateVPCForm>({
    resolver: zodResolver(createVPCSchema),
  });

  const subnetForm = useForm<CreateSubnetForm>({
    resolver: zodResolver(createSubnetSchema),
  });

  const securityGroupForm = useForm<CreateSecurityGroupForm>({
    resolver: zodResolver(createSecurityGroupSchema),
  });

  // Fetch credentials
  const { data: credentials = [] } = useQuery({
    queryKey: ['credentials', currentWorkspace?.id],
    queryFn: () => currentWorkspace ? credentialService.getCredentials(currentWorkspace.id) : Promise.resolve([]),
    enabled: !!currentWorkspace,
  });

  // Get selected credential and provider (after credentials is loaded)
  const selectedCredential = credentials.find(c => c.id === selectedCredentialId);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  // Use selectedCredentialId for consistency
  const watchedCredentialId = selectedCredentialId;

  // No need to filter credentials - show all available credentials
  const filteredCredentials = credentials;

  // Fetch VPCs
  const { data: vpcs = [], isLoading: isLoadingVPCs } = useQuery({
    queryKey: ['vpcs', selectedProvider, watchedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId) return [];
      return networkService.listVPCs(selectedProvider, watchedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!watchedCredentialId && !!currentWorkspace,
    refetchInterval: 30000, // Poll every 30 seconds
  });

  // Fetch Subnets (requires VPC ID)
  const { data: subnets = [], isLoading: isLoadingSubnets } = useQuery({
    queryKey: ['subnets', selectedProvider, watchedCredentialId, selectedVPCId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId || !selectedVPCId || !selectedRegion) return [];
      return networkService.listSubnets(selectedProvider, watchedCredentialId, selectedVPCId, selectedRegion);
    },
    enabled: !!selectedProvider && !!watchedCredentialId && !!selectedVPCId && !!selectedRegion && !!currentWorkspace,
    refetchInterval: 30000,
  });

  // Fetch Security Groups
  const { data: securityGroups = [], isLoading: isLoadingSecurityGroups } = useQuery({
    queryKey: ['security-groups', selectedProvider, watchedCredentialId, selectedRegion, selectedVPCId],
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId || !currentWorkspace) {
        return [];
      }
      return networkService.listSecurityGroups(selectedProvider, watchedCredentialId, selectedRegion, selectedVPCId);
    },
    enabled: !!selectedProvider && !!watchedCredentialId && !!currentWorkspace,
    refetchInterval: 30000,
  });

  // Search functionality for each tab
  const {
    query: searchQueryVPCs,
    setQuery: setSearchQueryVPCs,
    results: searchResultsVPCs,
    clearSearch: clearSearchVPCs,
  } = useSearch(vpcs, {
    keys: ['name', 'id', 'state'],
    threshold: 0.3,
  });

  const {
    query: searchQuerySubnets,
    setQuery: setSearchQuerySubnets,
    results: searchResultsSubnets,
    clearSearch: clearSearchSubnets,
  } = useSearch(subnets, {
    keys: ['name', 'id', 'cidr_block', 'state'],
    threshold: 0.3,
  });

  const {
    query: searchQuerySecurityGroups,
    setQuery: setSearchQuerySecurityGroups,
    results: searchResultsSecurityGroups,
    clearSearch: clearSearchSecurityGroups,
  } = useSearch(securityGroups, {
    keys: ['name', 'id', 'description'],
    threshold: 0.3,
  });

  // Filter configurations
  const vpcFilterConfigs: FilterConfig[] = [
    {
      key: 'state',
      label: 'State',
      type: 'select',
      options: [
        { value: 'available', label: 'Available' },
        { value: 'pending', label: 'Pending' },
        { value: 'deleting', label: 'Deleting' },
      ],
    },
    {
      key: 'is_default',
      label: 'Type',
      type: 'select',
      options: [
        { value: 'true', label: 'Default VPC' },
        { value: 'false', label: 'Custom VPC' },
      ],
    },
  ];

  // Apply filters
  const filteredVPCs = searchResultsVPCs.filter((vpc) => {
    if (filters.state && vpc.state !== filters.state) return false;
    if (filters.is_default !== undefined) {
      const isDefault = filters.is_default === 'true';
      if (vpc.is_default !== isDefault) return false;
    }
    return true;
  });

  const filteredSubnets = searchResultsSubnets;
  const filteredSecurityGroups = searchResultsSecurityGroups;

  // Pagination for each resource type
  const {
    page: vpcPage,
    paginatedItems: paginatedVPCs,
    setPage: setVPCPage,
    setPageSize: setVPCPageSize,
  } = usePagination(filteredVPCs, {
    totalItems: filteredVPCs.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const {
    page: subnetPage,
    paginatedItems: paginatedSubnets,
    setPage: setSubnetPage,
    setPageSize: setSubnetPageSize,
  } = usePagination(filteredSubnets, {
    totalItems: filteredSubnets.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const {
    page: sgPage,
    paginatedItems: paginatedSecurityGroups,
    setPage: setSGPage,
    setPageSize: setSGPageSize,
  } = usePagination(filteredSecurityGroups, {
    totalItems: filteredSecurityGroups.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // Create VPC mutation
  const createVPCMutation = useMutation({
    mutationFn: (data: CreateVPCForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.createVPC(selectedProvider, data);
    },
    onSuccess: () => {
      success('VPC creation initiated');
      queryClient.invalidateQueries({ queryKey: ['vpcs'] });
      setIsCreateDialogOpen(false);
      vpcForm.reset();
    },
    onError: (error: Error) => {
      showError(`Failed to create VPC: ${error.message}`);
    },
  });

  // Create Subnet mutation
  const createSubnetMutation = useMutation({
    mutationFn: (data: CreateSubnetForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.createSubnet(selectedProvider, data);
    },
    onSuccess: () => {
      success('Subnet creation initiated');
      queryClient.invalidateQueries({ queryKey: ['subnets'] });
      setIsCreateDialogOpen(false);
      subnetForm.reset();
    },
    onError: (error: Error) => {
      showError(`Failed to create subnet: ${error.message}`);
    },
  });

  // Create Security Group mutation
  const createSecurityGroupMutation = useMutation({
    mutationFn: (data: CreateSecurityGroupForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.createSecurityGroup(selectedProvider, data);
    },
    onSuccess: () => {
      success('Security group creation initiated');
      queryClient.invalidateQueries({ queryKey: ['security-groups'] });
      setIsCreateDialogOpen(false);
      securityGroupForm.reset();
    },
    onError: (error: Error) => {
      showError(`Failed to create security group: ${error.message}`);
    },
  });

  // Delete mutations
  const deleteVPCMutation = useMutation({
    mutationFn: async ({ vpcId, credentialId, region }: { vpcId: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.deleteVPC(selectedProvider, vpcId, credentialId, region);
    },
    onSuccess: () => {
      success('VPC deletion initiated');
      queryClient.invalidateQueries({ queryKey: ['vpcs'] });
    },
    onError: (error: Error) => {
      showError(`Failed to delete VPC: ${error.message}`);
    },
  });

  const deleteSubnetMutation = useMutation({
    mutationFn: async ({ subnetId, credentialId, region }: { subnetId: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.deleteSubnet(selectedProvider, subnetId, credentialId, region);
    },
    onSuccess: () => {
      success('Subnet deletion initiated');
      queryClient.invalidateQueries({ queryKey: ['subnets'] });
    },
    onError: (error: Error) => {
      showError(`Failed to delete subnet: ${error.message}`);
    },
  });

  const deleteSecurityGroupMutation = useMutation({
    mutationFn: async ({ securityGroupId, credentialId, region }: { securityGroupId: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.deleteSecurityGroup(selectedProvider, securityGroupId, credentialId, region);
    },
    onSuccess: () => {
      success('Security group deletion initiated');
      queryClient.invalidateQueries({ queryKey: ['security-groups'] });
    },
    onError: (error: Error) => {
      showError(`Failed to delete security group: ${error.message}`);
    },
  });

  const handleCreateVPC = (data: CreateVPCForm) => {
    createVPCMutation.mutate(data);
  };

  const handleCreateSubnet = (data: CreateSubnetForm) => {
    createSubnetMutation.mutate(data);
  };

  const handleCreateSecurityGroup = (data: CreateSecurityGroupForm) => {
    createSecurityGroupMutation.mutate(data);
  };

  // Bulk operations handlers
  const handleBulkDeleteVPCs = async (vpcIds: string[]) => {
    if (!watchedCredentialId || !selectedProvider) return;
    
    const vpcsToDelete = filteredVPCs.filter(v => vpcIds.includes(v.id));
    const deletePromises = vpcsToDelete.map(vpc =>
      deleteVPCMutation.mutateAsync({
        vpcId: vpc.id,
        credentialId: watchedCredentialId,
        region: vpc.region || '',
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${vpcIds.length} VPC(s)`);
      setSelectedVPCIds([]);
    } catch (error) {
      showError(`Failed to delete some VPCs: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleBulkDeleteSubnets = async (subnetIds: string[]) => {
    if (!watchedCredentialId || !selectedProvider) return;
    
    const subnetsToDelete = filteredSubnets.filter(s => subnetIds.includes(s.id));
    const deletePromises = subnetsToDelete.map(subnet =>
      deleteSubnetMutation.mutateAsync({
        subnetId: subnet.id,
        credentialId: watchedCredentialId,
        region: subnet.region,
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${subnetIds.length} subnet(s)`);
      setSelectedSubnetIds([]);
    } catch (error) {
      showError(`Failed to delete some subnets: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleBulkDeleteSecurityGroups = async (sgIds: string[]) => {
    if (!watchedCredentialId || !selectedProvider) return;
    
    const sgsToDelete = filteredSecurityGroups.filter(sg => sgIds.includes(sg.id));
    const deletePromises = sgsToDelete.map(sg =>
      deleteSecurityGroupMutation.mutateAsync({
        securityGroupId: sg.id,
        credentialId: watchedCredentialId,
        region: sg.region,
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${sgIds.length} security group(s)`);
      setSelectedSecurityGroupIds([]);
    } catch (error) {
      showError(`Failed to delete some security groups: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleBulkTag = (ids: string[]) => {
    setIsTagDialogOpen(true);
  };

  const handleBulkTagSubmit = async () => {
    if (!bulkTagKey.trim() || !bulkTagValue.trim()) return;
    
    const selectedCount = activeTab === 'vpcs' ? selectedVPCIds.length 
      : activeTab === 'subnets' ? selectedSubnetIds.length 
      : selectedSecurityGroupIds.length;
    
    success(`Tag "${bulkTagKey}: ${bulkTagValue}" will be added to ${selectedCount} resource(s)`);
    setIsTagDialogOpen(false);
    setBulkTagKey('');
    setBulkTagValue('');
    if (activeTab === 'vpcs') setSelectedVPCIds([]);
    else if (activeTab === 'subnets') setSelectedSubnetIds([]);
    else setSelectedSecurityGroupIds([]);
  };

  const handleDeleteVPC = (vpcId: string, region?: string) => {
    if (!watchedCredentialId || !region) return;
    if (confirm(`Are you sure you want to delete this VPC? This action cannot be undone.`)) {
      deleteVPCMutation.mutate({ vpcId, credentialId: watchedCredentialId, region });
    }
  };

  const handleDeleteSubnet = (subnetId: string, region: string) => {
    if (!watchedCredentialId) return;
    if (confirm(`Are you sure you want to delete this subnet? This action cannot be undone.`)) {
      deleteSubnetMutation.mutate({ subnetId, credentialId: watchedCredentialId, region });
    }
  };

  const handleDeleteSecurityGroup = (securityGroupId: string, region: string) => {
    if (!watchedCredentialId) return;
    if (confirm(`Are you sure you want to delete this security group? This action cannot be undone.`)) {
      deleteSecurityGroupMutation.mutate({ securityGroupId, credentialId: watchedCredentialId, region });
    }
  };

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
            <h1 className="text-3xl font-bold text-gray-900">Network Management</h1>
            <p className="text-gray-600">
              Manage VPCs, Subnets, and Security Groups for {currentWorkspace.name}
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Select
              value={selectedCredentialId || ''}
              onValueChange={(value) => {
                setSelectedCredentialId(value);
                vpcForm.setValue('credential_id', value);
                subnetForm.setValue('credential_id', value);
                securityGroupForm.setValue('credential_id', value);
              }}
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

        {/* Region and VPC Selection */}
        {selectedProvider && selectedCredentialId && (
          <Card>
            <CardHeader>
              <CardTitle>Configuration</CardTitle>
              <CardDescription>Select region and VPC to view network resources</CardDescription>
            </CardHeader>
            <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Region</Label>
                <Input
                  placeholder="e.g., ap-northeast-2"
                  value={selectedRegion}
                  onChange={(e) => setSelectedRegion(e.target.value)}
                />
              </div>
              {(activeTab === 'subnets' || activeTab === 'security-groups') && (
                <div className="space-y-2">
                  <Label>VPC</Label>
                  <Select
                    value={selectedVPCId}
                    onValueChange={(value) => {
                      setSelectedVPCId(value);
                      subnetForm.setValue('vpc_id', value);
                      securityGroupForm.setValue('vpc_id', value);
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select VPC" />
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
              )}
            </CardContent>
          </Card>
        )}

        {/* Network Topology */}
        {selectedProvider && watchedCredentialId && selectedVPCId && (
          <NetworkTopologyViewer
            vpcs={vpcs}
            subnets={subnets}
            securityGroups={securityGroups}
            selectedVPCId={selectedVPCId}
            onVPCClick={(vpcId) => {
              setSelectedVPCId(vpcId);
              subnetForm.setValue('vpc_id', vpcId);
              securityGroupForm.setValue('vpc_id', vpcId);
            }}
          />
        )}

        {/* Tabs */}
        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)} className="space-y-4">
          <div className="flex items-center justify-between">
            <TabsList>
              <TabsTrigger value="vpcs">
                <Network className="mr-2 h-4 w-4" />
                VPCs
              </TabsTrigger>
              <TabsTrigger value="subnets">
                <Layers className="mr-2 h-4 w-4" />
                Subnets
              </TabsTrigger>
              <TabsTrigger value="security-groups">
                <Shield className="mr-2 h-4 w-4" />
                Security Groups
              </TabsTrigger>
            </TabsList>
            {selectedProvider && selectedCredentialId && (
              <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
                <DialogTrigger asChild>
                  <Button disabled={credentials.length === 0 || (activeTab !== 'vpcs' && !selectedVPCId)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create {activeTab === 'vpcs' ? 'VPC' : activeTab === 'subnets' ? 'Subnet' : 'Security Group'}
                  </Button>
                </DialogTrigger>
                <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                  <DialogHeader>
                    <DialogTitle>
                      Create {activeTab === 'vpcs' ? 'VPC' : activeTab === 'subnets' ? 'Subnet' : 'Security Group'}
                    </DialogTitle>
                    <DialogDescription>
                      Create a new {activeTab === 'vpcs' ? 'VPC' : activeTab === 'subnets' ? 'subnet' : 'security group'} on {selectedProvider?.toUpperCase()}
                    </DialogDescription>
                  </DialogHeader>
                  {activeTab === 'vpcs' && (
                    <form onSubmit={vpcForm.handleSubmit(handleCreateVPC)} className="space-y-4">
                      <div className="space-y-2">
                        <Label htmlFor="vpc-name">Name *</Label>
                        <Input id="vpc-name" {...vpcForm.register('name')} placeholder="my-vpc" />
                        {vpcForm.formState.errors.name && (
                          <p className="text-sm text-red-600">{vpcForm.formState.errors.name.message}</p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="vpc-description">Description</Label>
                        <Input id="vpc-description" {...vpcForm.register('description')} placeholder="VPC description" />
                      </div>
                      <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                          <Label htmlFor="vpc-cidr">CIDR Block</Label>
                          <Input id="vpc-cidr" {...vpcForm.register('cidr_block')} placeholder="10.0.0.0/16" />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="vpc-region">Region</Label>
                          <Input
                            id="vpc-region"
                            {...vpcForm.register('region')}
                            placeholder="ap-northeast-2"
                            onChange={(e) => {
                              vpcForm.setValue('region', e.target.value);
                              setSelectedRegion(e.target.value);
                            }}
                          />
                        </div>
                      </div>
                      <div className="flex justify-end space-x-2">
                        <Button type="button" variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                          Cancel
                        </Button>
                        <Button type="submit" disabled={createVPCMutation.isPending}>
                          {createVPCMutation.isPending ? 'Creating...' : 'Create VPC'}
                        </Button>
                      </div>
                    </form>
                  )}
                  {activeTab === 'subnets' && (
                    <form onSubmit={subnetForm.handleSubmit(handleCreateSubnet)} className="space-y-4">
                      <div className="space-y-2">
                        <Label htmlFor="subnet-name">Name *</Label>
                        <Input id="subnet-name" {...subnetForm.register('name')} placeholder="my-subnet" />
                        {subnetForm.formState.errors.name && (
                          <p className="text-sm text-red-600">{subnetForm.formState.errors.name.message}</p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="subnet-vpc">VPC ID *</Label>
                        <Input id="subnet-vpc" {...subnetForm.register('vpc_id')} placeholder="vpc-12345" />
                        {subnetForm.formState.errors.vpc_id && (
                          <p className="text-sm text-red-600">{subnetForm.formState.errors.vpc_id.message}</p>
                        )}
                      </div>
                      <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                          <Label htmlFor="subnet-cidr">CIDR Block *</Label>
                          <Input id="subnet-cidr" {...subnetForm.register('cidr_block')} placeholder="10.0.1.0/24" />
                          {subnetForm.formState.errors.cidr_block && (
                            <p className="text-sm text-red-600">{subnetForm.formState.errors.cidr_block.message}</p>
                          )}
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="subnet-az">Availability Zone *</Label>
                          <Input id="subnet-az" {...subnetForm.register('availability_zone')} placeholder="ap-northeast-2a" />
                          {subnetForm.formState.errors.availability_zone && (
                            <p className="text-sm text-red-600">{subnetForm.formState.errors.availability_zone.message}</p>
                          )}
                        </div>
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="subnet-region">Region *</Label>
                        <Input
                          id="subnet-region"
                          {...subnetForm.register('region')}
                          placeholder="ap-northeast-2"
                          value={selectedRegion}
                          onChange={(e) => {
                            subnetForm.setValue('region', e.target.value);
                            setSelectedRegion(e.target.value);
                          }}
                        />
                        {subnetForm.formState.errors.region && (
                          <p className="text-sm text-red-600">{subnetForm.formState.errors.region.message}</p>
                        )}
                      </div>
                      <div className="flex justify-end space-x-2">
                        <Button type="button" variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                          Cancel
                        </Button>
                        <Button type="submit" disabled={createSubnetMutation.isPending}>
                          {createSubnetMutation.isPending ? 'Creating...' : 'Create Subnet'}
                        </Button>
                      </div>
                    </form>
                  )}
                  {activeTab === 'security-groups' && (
                    <form onSubmit={securityGroupForm.handleSubmit(handleCreateSecurityGroup)} className="space-y-4">
                      <div className="space-y-2">
                        <Label htmlFor="sg-name">Name *</Label>
                        <Input id="sg-name" {...securityGroupForm.register('name')} placeholder="my-security-group" />
                        {securityGroupForm.formState.errors.name && (
                          <p className="text-sm text-red-600">{securityGroupForm.formState.errors.name.message}</p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="sg-description">Description *</Label>
                        <Input id="sg-description" {...securityGroupForm.register('description')} placeholder="Security group description" />
                        {securityGroupForm.formState.errors.description && (
                          <p className="text-sm text-red-600">{securityGroupForm.formState.errors.description.message}</p>
                        )}
                      </div>
                      <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                          <Label htmlFor="sg-vpc">VPC ID *</Label>
                          <Input id="sg-vpc" {...securityGroupForm.register('vpc_id')} placeholder="vpc-12345" />
                          {securityGroupForm.formState.errors.vpc_id && (
                            <p className="text-sm text-red-600">{securityGroupForm.formState.errors.vpc_id.message}</p>
                          )}
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="sg-region">Region *</Label>
                          <Input
                            id="sg-region"
                            {...securityGroupForm.register('region')}
                            placeholder="ap-northeast-2"
                            value={selectedRegion}
                            onChange={(e) => {
                              securityGroupForm.setValue('region', e.target.value);
                              setSelectedRegion(e.target.value);
                            }}
                          />
                          {securityGroupForm.formState.errors.region && (
                            <p className="text-sm text-red-600">{securityGroupForm.formState.errors.region.message}</p>
                          )}
                        </div>
                      </div>
                      <div className="flex justify-end space-x-2">
                        <Button type="button" variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                          Cancel
                        </Button>
                        <Button type="submit" disabled={createSecurityGroupMutation.isPending}>
                          {createSecurityGroupMutation.isPending ? 'Creating...' : 'Create Security Group'}
                        </Button>
                      </div>
                    </form>
                  )}
                </DialogContent>
              </Dialog>
            )}
          </div>

          {/* VPCs Tab */}
          <TabsContent value="vpcs" className="space-y-4">
            {/* Search and Filter for VPCs */}
            {selectedProvider && watchedCredentialId && vpcs.length > 0 && (
              <Card>
                <CardContent className="pt-6">
                  <div className="flex flex-col md:flex-row gap-4">
                    <div className="flex-1">
                      <SearchBar
                        value={searchQueryVPCs}
                        onChange={setSearchQueryVPCs}
                        onClear={clearSearchVPCs}
                        placeholder="Search VPCs by name, ID, or state..."
                      />
                    </div>
                    <Button
                      variant="outline"
                      onClick={() => setShowFilters(!showFilters)}
                      className="flex items-center"
                    >
                      <Filter className="mr-2 h-4 w-4" />
                      Filters
                      {Object.keys(filters).length > 0 && (
                        <Badge variant="secondary" className="ml-2">
                          {Object.keys(filters).length}
                        </Badge>
                      )}
                    </Button>
                  </div>
                  {showFilters && (
                    <div className="mt-4">
                      <FilterPanel
                        configs={vpcFilterConfigs}
                        values={filters}
                        onChange={setFilters}
                        onClear={() => setFilters({})}
                      />
                    </div>
                  )}
                </CardContent>
              </Card>
            )}
            
            {!selectedProvider || !watchedCredentialId ? (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <Network className="h-12 w-12 text-gray-400 mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    {!selectedProvider ? 'Select a Provider' : 'Select a Credential'}
                  </h3>
                  <p className="text-sm text-gray-500 text-center">
                    {!selectedProvider
                      ? 'Please select a cloud provider to view VPCs'
                      : 'Please select a credential to view VPCs'}
                  </p>
                </CardContent>
              </Card>
            ) : isLoadingVPCs ? (
              <Card>
                <CardContent className="pt-6">
                  <TableSkeleton columns={5} rows={5} showCheckbox={true} />
                </CardContent>
              </Card>
            ) : filteredVPCs.length === 0 ? (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <Network className="h-12 w-12 text-gray-400 mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">No VPCs Found</h3>
                  <p className="text-sm text-gray-500 text-center mb-4">
                    No VPCs found. Create your first VPC to get started.
                  </p>
                  <Button onClick={() => setIsCreateDialogOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create VPC
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <>
            <BulkActionsToolbar
              items={paginatedVPCs}
                  selectedIds={selectedVPCIds}
                  onSelectionChange={setSelectedVPCIds}
                  onBulkDelete={handleBulkDeleteVPCs}
                  onBulkTag={handleBulkTag}
                  getItemDisplayName={(vpc) => vpc.name}
                />
                
                <Card>
                  <CardHeader>
                    <CardTitle>VPCs</CardTitle>
                    <CardDescription>
                      {filteredVPCs.length} of {vpcs.length} VPC{vpcs.length !== 1 ? 's' : ''} found
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-12">
                            <Checkbox
                              checked={selectedVPCIds.length === filteredVPCs.length && filteredVPCs.length > 0}
                              onCheckedChange={(checked) => {
                                if (checked) {
                                  setSelectedVPCIds(filteredVPCs.map(v => v.id));
                                } else {
                                  setSelectedVPCIds([]);
                                }
                              }}
                            />
                          </TableHead>
                          <TableHead>Name</TableHead>
                          <TableHead>State</TableHead>
                          <TableHead>CIDR Block</TableHead>
                          <TableHead>Default</TableHead>
                          <TableHead>Actions</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {paginatedVPCs.map((vpc) => {
                          const isSelected = selectedVPCIds.includes(vpc.id);
                          
                          return (
                            <TableRow key={vpc.id}>
                              <TableCell>
                                <Checkbox
                                  checked={isSelected}
                                  onCheckedChange={(checked) => {
                                    if (checked) {
                                      setSelectedVPCIds([...selectedVPCIds, vpc.id]);
                                    } else {
                                      setSelectedVPCIds(selectedVPCIds.filter(id => id !== vpc.id));
                                    }
                                  }}
                                />
                              </TableCell>
                              <TableCell className="font-medium">{vpc.name}</TableCell>
                          <TableCell>
                            <Badge variant={vpc.state === 'available' ? 'default' : 'secondary'}>
                              {vpc.state}
                            </Badge>
                          </TableCell>
                          <TableCell>{vpc.description || '-'}</TableCell>
                          <TableCell>
                            {vpc.is_default ? (
                              <Badge variant="outline">Default</Badge>
                            ) : (
                              <span className="text-gray-400">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center space-x-2">
                              <Button variant="ghost" size="sm">
                                <Edit className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeleteVPC(vpc.id, selectedRegion || vpc.region)}
                                disabled={deleteVPCMutation.isPending}
                              >
                                <Trash2 className="h-4 w-4 text-red-600" />
                              </Button>
                            </div>
                          </TableCell>
                          </TableRow>
                          );
                        })}
                      </TableBody>
                    </Table>
                    
                    {/* VPC Pagination */}
                    {filteredVPCs.length > 0 && (
                      <div className="border-t">
                        <Pagination
                          total={filteredVPCs.length}
                          page={vpcPage}
                          pageSize={pageSize}
                          onPageChange={setVPCPage}
                          onPageSizeChange={(newSize) => {
                            setPageSize(newSize);
                            setVPCPageSize(newSize);
                          }}
                          pageSizeOptions={[10, 20, 50, 100]}
                          showPageSizeSelector={true}
                        />
                      </div>
                    )}
                  </CardContent>
                </Card>
              </>
            )}
          </TabsContent>

          {/* Subnets Tab */}
          <TabsContent value="subnets" className="space-y-4">
            {!selectedProvider || !watchedCredentialId || !selectedVPCId ? (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <Layers className="h-12 w-12 text-gray-400 mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    {!selectedProvider ? 'Select a Provider' : !watchedCredentialId ? 'Select a Credential' : 'Select a VPC'}
                  </h3>
                  <p className="text-sm text-gray-500 text-center">
                    Please select a provider, credential, and VPC to view subnets
                  </p>
                </CardContent>
              </Card>
            ) : isLoadingSubnets ? (
              <Card>
                <CardContent className="pt-6">
                  <TableSkeleton columns={5} rows={5} showCheckbox={true} />
                </CardContent>
              </Card>
            ) : filteredSubnets.length === 0 ? (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <Layers className="h-12 w-12 text-gray-400 mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">No Subnets Found</h3>
                  <p className="text-sm text-gray-500 text-center mb-4">
                    No subnets found for the selected VPC. Create your first subnet.
                  </p>
                  <Button onClick={() => setIsCreateDialogOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Subnet
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <>
                <BulkActionsToolbar
                  items={paginatedSubnets}
                  selectedIds={selectedSubnetIds}
                  onSelectionChange={setSelectedSubnetIds}
                  onBulkDelete={handleBulkDeleteSubnets}
                  onBulkTag={handleBulkTag}
                  getItemDisplayName={(subnet) => subnet.name}
                />
                
                <Card>
                  <CardHeader>
                    <CardTitle>Subnets</CardTitle>
                    <CardDescription>
                      {filteredSubnets.length} of {subnets.length} subnet{subnets.length !== 1 ? 's' : ''} found
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="mb-4">
                      <SearchBar
                        value={searchQuerySubnets}
                        onChange={setSearchQuerySubnets}
                        onClear={clearSearchSubnets}
                        placeholder="Search subnets by name, CIDR, or state..."
                      />
                    </div>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-12">
                            <Checkbox
                              checked={selectedSubnetIds.length === filteredSubnets.length && filteredSubnets.length > 0}
                              onCheckedChange={(checked) => {
                                if (checked) {
                                  setSelectedSubnetIds(filteredSubnets.map(s => s.id));
                                } else {
                                  setSelectedSubnetIds([]);
                                }
                              }}
                            />
                          </TableHead>
                          <TableHead>Name</TableHead>
                          <TableHead>CIDR Block</TableHead>
                          <TableHead>Availability Zone</TableHead>
                          <TableHead>State</TableHead>
                          <TableHead>Public</TableHead>
                          <TableHead>Actions</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {paginatedSubnets.map((subnet) => {
                          const isSelected = selectedSubnetIds.includes(subnet.id);
                          
                          return (
                            <TableRow key={subnet.id}>
                              <TableCell>
                                <Checkbox
                                  checked={isSelected}
                                  onCheckedChange={(checked) => {
                                    if (checked) {
                                      setSelectedSubnetIds([...selectedSubnetIds, subnet.id]);
                                    } else {
                                      setSelectedSubnetIds(selectedSubnetIds.filter(id => id !== subnet.id));
                                    }
                                  }}
                                />
                              </TableCell>
                              <TableCell className="font-medium">{subnet.name}</TableCell>
                          <TableCell>{subnet.cidr_block}</TableCell>
                          <TableCell>{subnet.availability_zone}</TableCell>
                          <TableCell>
                            <Badge variant={subnet.state === 'available' ? 'default' : 'secondary'}>
                              {subnet.state}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            {subnet.is_public ? (
                              <Badge variant="outline">Public</Badge>
                            ) : (
                              <Badge variant="secondary">Private</Badge>
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center space-x-2">
                              <Button variant="ghost" size="sm">
                                <Edit className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeleteSubnet(subnet.id, subnet.region)}
                                disabled={deleteSubnetMutation.isPending}
                              >
                                <Trash2 className="h-4 w-4 text-red-600" />
                              </Button>
                            </div>
                              </TableCell>
                            </TableRow>
                            );
                          })}
                        </TableBody>
                      </Table>
                      
                      {/* Subnet Pagination */}
                      {filteredSubnets.length > 0 && (
                        <div className="border-t">
                          <Pagination
                            total={filteredSubnets.length}
                            page={subnetPage}
                            pageSize={pageSize}
                            onPageChange={setSubnetPage}
                            onPageSizeChange={(newSize) => {
                              setPageSize(newSize);
                              setSubnetPageSize(newSize);
                            }}
                            pageSizeOptions={[10, 20, 50, 100]}
                            showPageSizeSelector={true}
                          />
                        </div>
                      )}
                    </CardContent>
                  </Card>
                </>
              )}
            </TabsContent>

          {/* Security Groups Tab */}
          <TabsContent value="security-groups" className="space-y-4">
            {!selectedProvider || !watchedCredentialId || !selectedVPCId ? (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <Shield className="h-12 w-12 text-gray-400 mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    {!selectedProvider ? 'Select a Provider' : !watchedCredentialId ? 'Select a Credential' : 'Select a VPC'}
                  </h3>
                  <p className="text-sm text-gray-500 text-center">
                    Please select a provider, credential, and VPC to view security groups
                  </p>
                </CardContent>
              </Card>
            ) : isLoadingSecurityGroups ? (
              <Card>
                <CardContent className="pt-6">
                  <TableSkeleton columns={4} rows={5} showCheckbox={true} />
                </CardContent>
              </Card>
            ) : filteredSecurityGroups.length === 0 ? (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <Shield className="h-12 w-12 text-gray-400 mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">No Security Groups Found</h3>
                  <p className="text-sm text-gray-500 text-center mb-4">
                    No security groups found for the selected VPC. Create your first security group.
                  </p>
                  <Button onClick={() => setIsCreateDialogOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Security Group
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <>
                <BulkActionsToolbar
                  items={paginatedSecurityGroups}
                  selectedIds={selectedSecurityGroupIds}
                  onSelectionChange={setSelectedSecurityGroupIds}
                  onBulkDelete={handleBulkDeleteSecurityGroups}
                  onBulkTag={handleBulkTag}
                  getItemDisplayName={(sg) => sg.name}
                />
                
                <Card>
                  <CardHeader>
                    <CardTitle>Security Groups</CardTitle>
                    <CardDescription>
                      {filteredSecurityGroups.length} of {securityGroups.length} security group{securityGroups.length !== 1 ? 's' : ''} found
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="mb-4">
                      <SearchBar
                        value={searchQuerySecurityGroups}
                        onChange={setSearchQuerySecurityGroups}
                        onClear={clearSearchSecurityGroups}
                        placeholder="Search security groups by name, ID, or description..."
                      />
                    </div>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-12">
                            <Checkbox
                              checked={selectedSecurityGroupIds.length === filteredSecurityGroups.length && filteredSecurityGroups.length > 0}
                              onCheckedChange={(checked) => {
                                if (checked) {
                                  setSelectedSecurityGroupIds(filteredSecurityGroups.map(sg => sg.id));
                                } else {
                                  setSelectedSecurityGroupIds([]);
                                }
                              }}
                            />
                          </TableHead>
                          <TableHead>Name</TableHead>
                          <TableHead>Description</TableHead>
                          <TableHead>Rules</TableHead>
                          <TableHead>Actions</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {paginatedSecurityGroups.map((sg) => {
                          const isSelected = selectedSecurityGroupIds.includes(sg.id);
                          
                          return (
                            <TableRow key={sg.id}>
                              <TableCell>
                                <Checkbox
                                  checked={isSelected}
                                  onCheckedChange={(checked) => {
                                    if (checked) {
                                      setSelectedSecurityGroupIds([...selectedSecurityGroupIds, sg.id]);
                                    } else {
                                      setSelectedSecurityGroupIds(selectedSecurityGroupIds.filter(id => id !== sg.id));
                                    }
                                  }}
                                />
                              </TableCell>
                              <TableCell className="font-medium">{sg.name}</TableCell>
                          <TableCell>{sg.description}</TableCell>
                          <TableCell>
                            <Badge variant="outline">{sg.rules?.length || 0} rules</Badge>
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center space-x-2">
                              <Button variant="ghost" size="sm">
                                <Edit className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeleteSecurityGroup(sg.id, sg.region)}
                                disabled={deleteSecurityGroupMutation.isPending}
                              >
                                <Trash2 className="h-4 w-4 text-red-600" />
                              </Button>
                            </div>
                              </TableCell>
                            </TableRow>
                            );
                          })}
                        </TableBody>
                      </Table>
                      
                      {/* Security Group Pagination */}
                      {filteredSecurityGroups.length > 0 && (
                        <div className="border-t">
                          <Pagination
                            total={filteredSecurityGroups.length}
                            page={sgPage}
                            pageSize={pageSize}
                            onPageChange={setSGPage}
                            onPageSizeChange={(newSize) => {
                              setPageSize(newSize);
                              setSGPageSize(newSize);
                            }}
                            pageSizeOptions={[10, 20, 50, 100]}
                            showPageSizeSelector={true}
                          />
                        </div>
                      )}
                    </CardContent>
                  </Card>
                </>
              )}
            </TabsContent>
              </Tabs>

              {/* Bulk Tag Dialog */}
              <Dialog open={isTagDialogOpen} onOpenChange={setIsTagDialogOpen}>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Add Tags to Selected Resources</DialogTitle>
                    <DialogDescription>
                      Add the same tag to {
                        activeTab === 'vpcs' ? selectedVPCIds.length
                        : activeTab === 'subnets' ? selectedSubnetIds.length
                        : selectedSecurityGroupIds.length
                      } selected {activeTab === 'vpcs' ? 'VPC(s)' : activeTab === 'subnets' ? 'subnet(s)' : 'security group(s)'}
                    </DialogDescription>
                  </DialogHeader>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="bulk-tag-key">Tag Key *</Label>
                      <Input
                        id="bulk-tag-key"
                        value={bulkTagKey}
                        onChange={(e) => setBulkTagKey(e.target.value)}
                        placeholder="e.g., Environment"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="bulk-tag-value">Tag Value *</Label>
                      <Input
                        id="bulk-tag-value"
                        value={bulkTagValue}
                        onChange={(e) => setBulkTagValue(e.target.value)}
                        placeholder="e.g., Production"
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' && bulkTagKey && bulkTagValue) {
                            handleBulkTagSubmit();
                          }
                        }}
                      />
                    </div>
                    <div className="flex justify-end space-x-2">
                      <Button variant="outline" onClick={() => setIsTagDialogOpen(false)}>
                        Cancel
                      </Button>
                      <Button onClick={handleBulkTagSubmit} disabled={!bulkTagKey.trim() || !bulkTagValue.trim()}>
                        Add Tag
                      </Button>
                    </div>
                  </div>
                </DialogContent>
              </Dialog>
            </div>
          </Layout>
        </WorkspaceRequired>
      );
    }


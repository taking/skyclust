'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
// import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { SearchBar } from '@/components/ui/search-bar';
import { FilterPanel, FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useSearch } from '@/hooks/useSearch';
import * as z from 'zod';
import { vmService } from '@/services/vm';
import { useWorkspaceStore } from '@/store/workspace';
import { useRouter } from 'next/navigation';
import { Plus, Server, Play, Square, Trash2, ExternalLink } from 'lucide-react';
import { CreateVMForm } from '@/lib/types';
import { ScreenReaderOnly } from '@/components/accessibility/screen-reader-only';
import { LiveRegion } from '@/components/accessibility/live-region';
import { getStatusAriaLabel, getActionAriaLabel, getLiveRegionMessage } from '@/lib/accessibility';
import { useToast } from '@/hooks/useToast';

const createVMSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  provider: z.string().min(1, 'Provider is required'),
  instance_type: z.string().min(1, 'Instance type is required'),
  region: z.string().min(1, 'Region is required'),
  image_id: z.string().min(1, 'Image ID is required'),
});

export default function VMsPage() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [filters, setFilters] = useState<FilterValue>({});
  const [liveMessage, setLiveMessage] = useState('');
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { success: showSuccess, error: showError } = useToast();

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
    // watch,
  } = useForm<CreateVMForm>({
    resolver: zodResolver(createVMSchema),
  });

  // const selectedProvider = watch('provider');

  // Fetch VMs
  const { data: vms = [], isLoading } = useQuery({
    queryKey: ['vms', currentWorkspace?.id],
    queryFn: () => currentWorkspace ? vmService.getVMs(currentWorkspace.id) : Promise.resolve([]),
    enabled: !!currentWorkspace,
  });

  // Search functionality
  const {
    query: searchQuery,
    setQuery: setSearchQuery,
    results: searchResults,
    isSearching,
    clearSearch,
  } = useSearch(vms, {
    keys: ['name', 'provider', 'instance_type', 'region', 'status'],
    threshold: 0.3,
  });

  // Filter configurations
  const filterConfigs: FilterConfig[] = [
    {
      id: 'status',
      label: 'Status',
      type: 'multiselect',
      options: [
        { id: 'running', label: 'Running', value: 'running' },
        { id: 'stopped', label: 'Stopped', value: 'stopped' },
        { id: 'starting', label: 'Starting', value: 'starting' },
        { id: 'stopping', label: 'Stopping', value: 'stopping' },
      ],
    },
    {
      id: 'provider',
      label: 'Provider',
      type: 'multiselect',
      options: [
        { id: 'aws', label: 'AWS', value: 'aws' },
        { id: 'gcp', label: 'GCP', value: 'gcp' },
        { id: 'azure', label: 'Azure', value: 'azure' },
      ],
    },
    {
      id: 'region',
      label: 'Region',
      type: 'select',
      options: [
        { id: 'us-east-1', label: 'US East (N. Virginia)', value: 'us-east-1' },
        { id: 'us-west-2', label: 'US West (Oregon)', value: 'us-west-2' },
        { id: 'eu-west-1', label: 'Europe (Ireland)', value: 'eu-west-1' },
        { id: 'ap-southeast-1', label: 'Asia Pacific (Singapore)', value: 'ap-southeast-1' },
      ],
    },
  ];

  // Apply filters to search results
  const filteredVMs = searchResults.filter((vm) => {
    // Status filter
    if (filters.status && Array.isArray(filters.status) && filters.status.length > 0) {
      if (!filters.status.includes(vm.status)) return false;
    }

    // Provider filter
    if (filters.provider && Array.isArray(filters.provider) && filters.provider.length > 0) {
      if (!filters.provider.includes(vm.provider)) return false;
    }

    // Region filter
    if (filters.region && filters.region !== vm.region) {
      return false;
    }

    return true;
  });

  // Create VM mutation
  const createVMMutation = useMutation({
    mutationFn: vmService.createVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms', currentWorkspace?.id] });
      setIsCreateDialogOpen(false);
      reset();
    },
  });

  // Delete VM mutation
  const deleteVMMutation = useMutation({
    mutationFn: vmService.deleteVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms', currentWorkspace?.id] });
    },
  });

  // Start VM mutation
  const startVMMutation = useMutation({
    mutationFn: vmService.startVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms', currentWorkspace?.id] });
    },
  });

  // Stop VM mutation
  const stopVMMutation = useMutation({
    mutationFn: vmService.stopVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms', currentWorkspace?.id] });
    },
  });

  const handleCreateVM = (data: CreateVMForm) => {
    if (!currentWorkspace) return;
    createVMMutation.mutate({
      ...data,
      workspace_id: currentWorkspace.id,
    } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
  };

  const handleDeleteVM = (vmId: string) => {
    const vm = vms.find(v => v.id === vmId);
    if (confirm('Are you sure you want to delete this VM?')) {
      deleteVMMutation.mutate(vmId, {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ['vms', currentWorkspace?.id] });
          showSuccess('VM deleted successfully');
          setLiveMessage(getLiveRegionMessage('deleted', vm?.name || 'VM', true));
        },
        onError: (error) => {
          showError(`Failed to delete VM: ${error.message}`);
          setLiveMessage(getLiveRegionMessage('deleted', vm?.name || 'VM', false));
        },
      });
    }
  };

  const handleStartVM = (vmId: string) => {
    const vm = vms.find(v => v.id === vmId);
    startVMMutation.mutate(vmId, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['vms', currentWorkspace?.id] });
        showSuccess('VM started successfully');
        setLiveMessage(getLiveRegionMessage('started', vm?.name || 'VM', true));
      },
      onError: (error) => {
        showError(`Failed to start VM: ${error.message}`);
        setLiveMessage(getLiveRegionMessage('started', vm?.name || 'VM', false));
      },
    });
  };

  const handleStopVM = (vmId: string) => {
    const vm = vms.find(v => v.id === vmId);
    stopVMMutation.mutate(vmId, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['vms', currentWorkspace?.id] });
        showSuccess('VM stopped successfully');
        setLiveMessage(getLiveRegionMessage('stopped', vm?.name || 'VM', true));
      },
      onError: (error) => {
        showError(`Failed to stop VM: ${error.message}`);
        setLiveMessage(getLiveRegionMessage('stopped', vm?.name || 'VM', false));
      },
    });
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
        return 'default';
      case 'stopped':
        return 'secondary';
      case 'pending':
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
            Please select a workspace to manage VMs.
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
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading VMs...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Virtual Machines</h1>
            <p className="text-gray-600">
              Manage VMs in {currentWorkspace.name} workspace
            </p>
          </div>
          <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                Create VM
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>Create New VM</DialogTitle>
                <DialogDescription>
                  Create a new virtual machine in your workspace.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit(handleCreateVM)} className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="name">VM Name</Label>
                    <Input
                      id="name"
                      placeholder="Enter VM name"
                      {...register('name')}
                    />
                    {errors.name && (
                      <p className="text-sm text-red-600">{errors.name.message}</p>
                    )}
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
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="instance_type">Instance Type</Label>
                    <Input
                      id="instance_type"
                      placeholder="e.g., t3.micro"
                      {...register('instance_type')}
                    />
                    {errors.instance_type && (
                      <p className="text-sm text-red-600">{errors.instance_type.message}</p>
                    )}
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="region">Region</Label>
                    <Input
                      id="region"
                      placeholder="e.g., us-east-1"
                      {...register('region')}
                    />
                    {errors.region && (
                      <p className="text-sm text-red-600">{errors.region.message}</p>
                    )}
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="image_id">Image ID</Label>
                  <Input
                    id="image_id"
                    placeholder="e.g., ami-12345678"
                    {...register('image_id')}
                  />
                  {errors.image_id && (
                    <p className="text-sm text-red-600">{errors.image_id.message}</p>
                  )}
                </div>
                <div className="flex justify-end space-x-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setIsCreateDialogOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createVMMutation.isPending}>
                    {createVMMutation.isPending ? 'Creating...' : 'Create VM'}
                  </Button>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {/* Search and Filter Controls */}
        <div className="mb-6 space-y-4">
          <div className="flex flex-col gap-4">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="flex-1">
                <SearchBar
                  placeholder="Search VMs..."
                  value={searchQuery}
                  onChange={setSearchQuery}
                  onClear={clearSearch}
                  showFilter
                  onFilterClick={() => {}}
                  filterCount={Object.keys(filters).length}
                />
              </div>
              <div className="w-full sm:w-auto">
                <FilterPanel
                  filters={filterConfigs}
                  values={filters}
                  onChange={setFilters}
                  onClear={() => setFilters({})}
                  onApply={() => {}}
                  title="Filter VMs"
                  description="Filter VMs by status, provider, and region"
                />
              </div>
            </div>
          </div>
          
          {/* Search Results Info */}
          {isSearching && (
            <div className="text-sm text-gray-600">
              Found {filteredVMs.length} VM{filteredVMs.length !== 1 ? 's' : ''} 
              {searchQuery && ` matching "${searchQuery}"`}
            </div>
          )}
        </div>

        {filteredVMs.length === 0 ? (
          <div className="text-center py-12">
            <div className="mx-auto h-12 w-12 text-gray-400">
              <Server className="h-12 w-12" />
            </div>
            <h3 className="mt-2 text-sm font-medium text-gray-900">
              {isSearching ? 'No VMs found' : 'No VMs'}
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              {isSearching 
                ? 'Try adjusting your search or filter criteria.'
                : 'Get started by creating your first virtual machine.'
              }
            </p>
            <div className="mt-6">
              <Button onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Create VM
              </Button>
            </div>
          </div>
        ) : (
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <div className="overflow-x-auto">
              <Table role="table" aria-label="Virtual machines list">
                <TableHeader>
                  <TableRow>
                    <TableHead className="min-w-[120px]">Name</TableHead>
                    <TableHead className="hidden sm:table-cell">Provider</TableHead>
                    <TableHead className="hidden md:table-cell">Type</TableHead>
                    <TableHead className="hidden lg:table-cell">Region</TableHead>
                    <TableHead className="min-w-[80px]">Status</TableHead>
                    <TableHead className="hidden xl:table-cell">IP Address</TableHead>
                    <TableHead className="hidden lg:table-cell">Created</TableHead>
                    <TableHead className="text-right min-w-[100px]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredVMs.map((vm) => (
                    <TableRow key={vm.id} role="row">
                      <TableCell className="font-medium" role="cell">
                        <div>
                          <div className="font-medium">{vm.name}</div>
                          <div className="text-sm text-gray-500 sm:hidden">
                            {vm.provider} • {vm.instance_type}
                            <ScreenReaderOnly>
                              Provider: {vm.provider}, Instance type: {vm.instance_type}
                            </ScreenReaderOnly>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="hidden sm:table-cell" role="cell">
                        <Badge variant="outline" aria-label={`Provider: ${vm.provider}`}>
                          {vm.provider}
                        </Badge>
                      </TableCell>
                      <TableCell className="hidden md:table-cell" role="cell">
                        {vm.instance_type}
                      </TableCell>
                      <TableCell className="hidden lg:table-cell" role="cell">
                        {vm.region}
                      </TableCell>
                      <TableCell role="cell">
                        <Badge 
                          variant={getStatusBadgeVariant(vm.status)}
                          aria-label={`Status: ${getStatusAriaLabel(vm.status)}`}
                        >
                          {vm.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="hidden xl:table-cell" role="cell">
                        {vm.public_ip ? (
                          <div className="flex items-center space-x-1">
                            <span>{vm.public_ip}</span>
                            <ExternalLink className="h-3 w-3 text-gray-400" aria-hidden="true" />
                            <ScreenReaderOnly>External link</ScreenReaderOnly>
                          </div>
                        ) : (
                          <span className="text-gray-400" aria-label="No IP address">-</span>
                        )}
                      </TableCell>
                      <TableCell className="hidden lg:table-cell" role="cell">
                        {new Date(vm.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-right" role="cell">
                        <div className="flex items-center justify-end space-x-1" role="group" aria-label={`Actions for ${vm.name}`}>
                          {vm.status === 'running' ? (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleStopVM(vm.id)}
                              disabled={stopVMMutation.isPending}
                              className="h-8 w-8 p-0"
                              aria-label={getActionAriaLabel('stop', vm.name)}
                            >
                              <Square className="h-4 w-4" aria-hidden="true" />
                            </Button>
                          ) : (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleStartVM(vm.id)}
                              disabled={startVMMutation.isPending}
                              className="h-8 w-8 p-0"
                              aria-label={getActionAriaLabel('start', vm.name)}
                            >
                              <Play className="h-4 w-4" aria-hidden="true" />
                            </Button>
                          )}
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleDeleteVM(vm.id)}
                            disabled={deleteVMMutation.isPending}
                            className="h-8 w-8 p-0"
                            aria-label={getActionAriaLabel('delete', vm.name)}
                          >
                            <Trash2 className="h-4 w-4" aria-hidden="true" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        )}
      </div>
      <LiveRegion message={liveMessage} />
    </div>
  );
}

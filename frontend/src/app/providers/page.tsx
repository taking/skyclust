'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { providerService } from '@/services/provider';
import { useWorkspaceStore } from '@/store/workspace';
import { useRouter } from 'next/navigation';
import { Cloud, Server, DollarSign, ExternalLink, Search } from 'lucide-react';

export default function ProvidersPage() {
  const [selectedProvider, setSelectedProvider] = useState<string>('');
  const [selectedRegion, setSelectedRegion] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();

  // Fetch providers
  const { data: providers = [], isLoading: providersLoading } = useQuery({
    queryKey: ['providers'],
    queryFn: providerService.getProviders,
  });

  // Fetch instances for selected provider
  const { data: instances = [], isLoading: instancesLoading } = useQuery({
    queryKey: ['instances', selectedProvider, selectedRegion],
    queryFn: () => selectedProvider ? providerService.getInstances(selectedProvider, selectedRegion || undefined) : Promise.resolve([]),
    enabled: !!selectedProvider,
  });

  // Fetch regions for selected provider
  const { data: regions = [] } = useQuery({
    queryKey: ['regions', selectedProvider],
    queryFn: () => selectedProvider ? providerService.getRegions(selectedProvider) : Promise.resolve([]),
    enabled: !!selectedProvider,
  });

  // Fetch cost estimates for selected provider
  const { data: costEstimates = [] } = useQuery({
    queryKey: ['cost-estimates', selectedProvider],
    queryFn: () => selectedProvider ? providerService.getCostEstimates(selectedProvider) : Promise.resolve([]),
    enabled: !!selectedProvider,
  });

  const filteredInstances = instances.filter(instance =>
    instance.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    instance.id.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getProviderIcon = (provider: string) => {
    switch (provider.toLowerCase()) {
      case 'aws':
        return '☁️';
      case 'gcp':
        return '🌐';
      case 'azure':
        return '🔷';
      default:
        return '☁️';
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
            Please select a workspace to view cloud providers.
          </p>
          <Button onClick={() => router.push('/workspaces')}>
            Select Workspace
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-4 md:py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-6 md:mb-8">
          <h1 className="text-2xl md:text-3xl font-bold text-gray-900">Cloud Providers</h1>
          <p className="text-sm md:text-base text-gray-600">
            Manage cloud resources and view instances across providers
          </p>
        </div>

        <Tabs defaultValue="providers" className="space-y-6">
          <TabsList>
            <TabsTrigger value="providers">Providers</TabsTrigger>
            <TabsTrigger value="instances">Instances</TabsTrigger>
            <TabsTrigger value="costs">Cost Estimates</TabsTrigger>
          </TabsList>

            <TabsContent value="providers" className="space-y-4 md:space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6">
              {providersLoading ? (
                <div className="col-span-full flex items-center justify-center py-12">
                  <div className="text-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
                    <p className="mt-2 text-gray-600">Loading providers...</p>
                  </div>
                </div>
              ) : (
                providers.map((provider) => (
                  <Card key={provider.name} className="hover:shadow-lg transition-shadow cursor-pointer">
                    <CardHeader>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <span className="text-2xl">{getProviderIcon(provider.name)}</span>
                          <div>
                            <CardTitle className="text-lg">{provider.name}</CardTitle>
                            <CardDescription>Version {provider.version}</CardDescription>
                          </div>
                        </div>
                        <Badge variant={getProviderBadgeVariant(provider.name)}>
                          {provider.name}
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-3">
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-500">Status</span>
                          <Badge variant="default">Available</Badge>
                        </div>
                        <Button
                          className="w-full"
                          onClick={() => setSelectedProvider(provider.name)}
                        >
                          <Cloud className="mr-2 h-4 w-4" />
                          View Instances
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))
              )}
            </div>
          </TabsContent>

          <TabsContent value="instances" className="space-y-4 md:space-y-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div>
                <Label htmlFor="provider-select">Provider</Label>
                <Select value={selectedProvider} onValueChange={setSelectedProvider}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a provider" />
                  </SelectTrigger>
                  <SelectContent>
                    {providers.map((provider) => (
                      <SelectItem key={provider.name} value={provider.name}>
                        {provider.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label htmlFor="region-select">Region</Label>
                <Select value={selectedRegion} onValueChange={setSelectedRegion}>
                  <SelectTrigger>
                    <SelectValue placeholder="All regions" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">All regions</SelectItem>
                    {regions.map((region) => (
                      <SelectItem key={region.name} value={region.name}>
                        {region.display_name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label htmlFor="search">Search</Label>
                <div className="relative">
                  <Search className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
                  <Input
                    id="search"
                    placeholder="Search instances..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10"
                  />
                </div>
              </div>
            </div>

            {selectedProvider ? (
              <div className="bg-white shadow rounded-lg">
                <div className="px-6 py-4 border-b">
                  <h3 className="text-lg font-medium">
                    {selectedProvider} Instances
                    {selectedRegion && ` in ${selectedRegion}`}
                  </h3>
                </div>
                {instancesLoading ? (
                  <div className="flex items-center justify-center py-12">
                    <div className="text-center">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
                      <p className="mt-2 text-gray-600">Loading instances...</p>
                    </div>
                  </div>
                ) : filteredInstances.length === 0 ? (
                  <div className="text-center py-12">
                    <Server className="mx-auto h-12 w-12 text-gray-400" />
                    <h3 className="mt-2 text-sm font-medium text-gray-900">No instances found</h3>
                    <p className="mt-1 text-sm text-gray-500">
                      {searchQuery ? 'No instances match your search.' : 'No instances found for this provider.'}
                    </p>
                  </div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead className="hidden sm:table-cell">ID</TableHead>
                        <TableHead className="hidden md:table-cell">Type</TableHead>
                        <TableHead className="hidden lg:table-cell">Region</TableHead>
                        <TableHead className="hidden sm:table-cell">Status</TableHead>
                        <TableHead className="hidden md:table-cell">IP Address</TableHead>
                        <TableHead className="hidden lg:table-cell">Created</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredInstances.map((instance) => (
                        <TableRow key={instance.id}>
                          <TableCell className="font-medium">
                            <div className="space-y-1">
                              <div>{instance.name}</div>
                              <div className="sm:hidden">
                                <Badge variant={getStatusBadgeVariant(instance.status)} className="text-xs">
                                  {instance.status}
                                </Badge>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell className="hidden sm:table-cell font-mono text-sm">{instance.id}</TableCell>
                          <TableCell className="hidden md:table-cell">{instance.type}</TableCell>
                          <TableCell className="hidden lg:table-cell">{instance.region}</TableCell>
                          <TableCell className="hidden sm:table-cell">
                            <Badge variant={getStatusBadgeVariant(instance.status)}>
                              {instance.status}
                            </Badge>
                          </TableCell>
                          <TableCell className="hidden md:table-cell">
                            {instance.public_ip ? (
                              <div className="flex items-center space-x-1">
                                <span>{instance.public_ip}</span>
                                <ExternalLink className="h-3 w-3 text-gray-400" />
                              </div>
                            ) : (
                              <span className="text-gray-400">-</span>
                            )}
                          </TableCell>
                          <TableCell className="hidden lg:table-cell">
                            {new Date(instance.created_at).toLocaleDateString()}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </div>
            ) : (
              <div className="text-center py-12">
                <Cloud className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">Select a provider</h3>
                <p className="mt-1 text-sm text-gray-500">
                  Choose a cloud provider to view its instances.
                </p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="costs" className="space-y-6">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="flex-1">
                <Label htmlFor="cost-provider-select">Provider</Label>
                <Select value={selectedProvider} onValueChange={setSelectedProvider}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a provider" />
                  </SelectTrigger>
                  <SelectContent>
                    {providers.map((provider) => (
                      <SelectItem key={provider.name} value={provider.name}>
                        {provider.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            {selectedProvider ? (
              <div className="bg-white shadow rounded-lg">
                <div className="px-6 py-4 border-b">
                  <h3 className="text-lg font-medium">
                    {selectedProvider} Cost Estimates
                  </h3>
                </div>
                {costEstimates.length === 0 ? (
                  <div className="text-center py-12">
                    <DollarSign className="mx-auto h-12 w-12 text-gray-400" />
                    <h3 className="mt-2 text-sm font-medium text-gray-900">No cost estimates</h3>
                    <p className="mt-1 text-sm text-gray-500">
                      No cost estimates available for this provider.
                    </p>
                  </div>
                ) : (
                  <div className="p-6">
                    <div className="text-sm text-gray-500">
                      Cost estimation features will be implemented here.
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="text-center py-12">
                <DollarSign className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">Select a provider</h3>
                <p className="mt-1 text-sm text-gray-500">
                  Choose a cloud provider to view cost estimates.
                </p>
              </div>
            )}
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

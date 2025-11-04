/**
 * Kubernetes Nodes Page
 * Kubernetes 노드 관리 페이지
 */

'use client';

import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { kubernetesService } from '@/features/kubernetes';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentials } from '@/hooks/use-credentials';
import { Building2, Search } from 'lucide-react';
import { CloudProvider, Node } from '@/lib/types';
import { useRequireAuth } from '@/hooks/use-auth';
import { DataProcessor } from '@/lib/data-processor';
import { SearchBar } from '@/components/ui/search-bar';
import { useSSEMonitoring } from '@/hooks/use-sse-monitoring';
import { Pagination } from '@/components/ui/pagination';
import { usePagination } from '@/hooks/use-pagination';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { Layout } from '@/components/layout/layout';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import { UI } from '@/lib/constants';

export default function NodesPage() {
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const { isLoading: authLoading } = useRequireAuth();

  // Get credential context from global store
  const { selectedCredentialId, selectedRegion } = useCredentialContext();

  const [selectedClusterName, setSelectedClusterName] = useState<string>('');
  const [pageSize, setPageSize] = useState(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  useSSEMonitoring();

  // Fetch credentials using unified hook
  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  // Fetch clusters for selection
  const { data: clusters = [] } = useQuery({
    queryKey: queryKeys.clusters.list(selectedProvider, selectedCredentialId, selectedRegion),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listClusters(selectedProvider, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!currentWorkspace,
  });

  // Fetch Nodes
  const { data: nodes = [], isLoading: isLoadingNodes } = useQuery({
    queryKey: queryKeys.nodes.list(selectedProvider, selectedClusterName, selectedCredentialId, selectedRegion),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId || !selectedClusterName || !selectedRegion) return [];
      return kubernetesService.listNodes(selectedProvider, selectedClusterName, selectedCredentialId, selectedRegion || '');
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!selectedClusterName && !!selectedRegion && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 10000,
  });

  const [searchQuery, setSearchQuery] = useState('');

  // Filtered nodes using DataProcessor (memoized for consistency)
  const filteredNodes = useMemo(() => {
    return DataProcessor.search(nodes, searchQuery, {
      keys: ['name', 'cluster_name', 'status', 'private_ip', 'public_ip'],
      threshold: 0.3,
    });
  }, [nodes, searchQuery]);

  const clearSearch = () => {
    setSearchQuery('');
  };

  // Pagination
  const {
    page,
    paginatedItems: paginatedNodes,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredNodes, {
    totalItems: filteredNodes.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  };

  // Header component
  const header = (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Nodes</h1>
        <p className="text-gray-600 mt-1">
          Manage Kubernetes Nodes{currentWorkspace ? ` for ${currentWorkspace.name}` : ''}
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {/* Credential selection is now handled in Header */}
      </div>
    </div>
  );

  // Empty state
  const emptyState = credentials.length === 0 ? (
    <CredentialRequiredState serviceName="Kubernetes (Nodes)" />
  ) : !selectedProvider || !selectedCredentialId ? (
    <CredentialRequiredState
      title="Select a Credential"
      description="Please select a credential to view nodes. If you don't have any credentials, register one first."
      serviceName="Kubernetes (Nodes)"
    />
  ) : !selectedClusterName || !selectedRegion ? (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Building2 className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          Select Cluster and Region
        </h3>
        <p className="text-sm text-gray-500 text-center">
          Please select a cluster and region to view nodes
        </p>
      </CardContent>
    </Card>
  ) : filteredNodes.length === 0 ? (
    <ResourceEmptyState
      resourceName="Nodes"
      icon={Building2}
      description="No nodes found for the selected cluster."
      withCard={true}
    />
  ) : null;

  // Content component
  const content = selectedProvider && selectedCredentialId && selectedClusterName && selectedRegion && filteredNodes.length > 0 ? (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Nodes</CardTitle>
          <CardDescription>
            {filteredNodes.length} of {nodes.length} node{nodes.length !== 1 ? 's' : ''} found
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="mb-4">
            <SearchBar
              value={searchQuery}
              onChange={setSearchQuery}
              onClear={clearSearch}
              placeholder="Search nodes by name, cluster, status, or IP..."
            />
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Cluster</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Instance Type</TableHead>
                <TableHead>Zone</TableHead>
                <TableHead>Private IP</TableHead>
                <TableHead>Public IP</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {paginatedNodes.map((node) => (
                <TableRow key={node.id}>
                  <TableCell className="font-medium">{node.name}</TableCell>
                  <TableCell>{node.cluster_name}</TableCell>
                  <TableCell>
                    <Badge variant={node.status === 'Ready' ? 'default' : 'secondary'}>
                      {node.status}
                    </Badge>
                  </TableCell>
                  <TableCell>{node.instance_type}</TableCell>
                  <TableCell>{node.zone || '-'}</TableCell>
                  <TableCell>{node.private_ip || '-'}</TableCell>
                  <TableCell>{node.public_ip || '-'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          
          {filteredNodes.length > 0 && (
            <div className="border-t mt-4">
              <Pagination
                total={filteredNodes.length}
                page={page}
                pageSize={pageSize}
                onPageChange={setPage}
                onPageSizeChange={handlePageSizeChange}
                pageSizeOptions={UI.PAGINATION.PAGE_SIZE_OPTIONS}
                showPageSizeSelector={true}
              />
            </div>
          )}
        </CardContent>
      </Card>
    </>
  ) : emptyState;

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
          {header}
          
          {/* Configuration */}
          {selectedProvider && selectedCredentialId && (
            <Card>
              <CardHeader>
                <CardTitle>Configuration</CardTitle>
                <CardDescription>Select cluster, region, and credential to view nodes</CardDescription>
              </CardHeader>
              <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Cluster *</Label>
                  <Select
                    value={selectedClusterName}
                    onValueChange={setSelectedClusterName}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select Cluster" />
                    </SelectTrigger>
                    <SelectContent>
                      {clusters.map((cluster) => (
                        <SelectItem key={cluster.name} value={cluster.name}>
                          {cluster.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Region *</Label>
                  <Input
                    placeholder="e.g., ap-northeast-2"
                    value={selectedRegion || ''}
                    readOnly
                    className="bg-muted"
                  />
                  <p className="text-xs text-muted-foreground">
                    Region selection is now handled in Header
                  </p>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Content */}
          {isLoadingNodes ? (
            <Card>
              <CardContent className="pt-6">
                <TableSkeleton columns={7} rows={5} />
              </CardContent>
            </Card>
          ) : (
            content
          )}
        </div>
      </Layout>
    </WorkspaceRequired>
  );
}


'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Layout } from '@/components/layout/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { providerService } from '@/services/provider';
import { useToast } from '@/hooks/useToast';
import { useRequireAuth } from '@/hooks/useAuth';
import { 
  ArrowLeft, 
  Server, 
  MapPin, 
  DollarSign, 
  Activity,
  Plus,
  Search,
  Filter
} from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, PieChart, Pie, Cell } from 'recharts';

// Mock data for charts
const costData = [
  { month: 'Jan', cost: 1200 },
  { month: 'Feb', cost: 1350 },
  { month: 'Mar', cost: 1100 },
  { month: 'Apr', cost: 1400 },
  { month: 'May', cost: 1600 },
  { month: 'Jun', cost: 1800 },
];

const usageData = [
  { service: 'EC2', usage: 45, cost: 800 },
  { service: 'S3', usage: 25, cost: 200 },
  { service: 'RDS', usage: 20, cost: 400 },
  { service: 'Lambda', usage: 10, cost: 100 },
];

const regionData = [
  { name: 'us-east-1', instances: 12, cost: 600 },
  { name: 'us-west-2', instances: 8, cost: 400 },
  { name: 'eu-west-1', instances: 6, cost: 300 },
  { name: 'ap-southeast-1', instances: 4, cost: 200 },
];

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

export default function ProviderDetailPage() {
  const params = useParams();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const { isLoading: authLoading } = useRequireAuth();
  const providerName = params.name as string;
  const [selectedRegion, setSelectedRegion] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');

  // Fetch provider details
  const { data: provider, isLoading: providerLoading } = useQuery({
    queryKey: ['provider', providerName],
    queryFn: () => providerService.getProvider(providerName),
    enabled: !!providerName,
  });

  // Fetch instances
  const { data: instances = [] } = useQuery({
    queryKey: ['instances', providerName, selectedRegion],
    queryFn: () => providerService.getInstances(providerName, selectedRegion || undefined),
    enabled: !!providerName,
  });

  // Fetch regions
  const { data: regions = [] } = useQuery({
    queryKey: ['regions', providerName],
    queryFn: () => providerService.getRegions(providerName),
    enabled: !!providerName,
  });

  // Create instance mutation (placeholder)
  const createInstanceMutation = useMutation({
    mutationFn: (_data: unknown) => {
      // This would be implemented when the API endpoint is available
      return Promise.resolve();
    },
    onSuccess: () => {
      success('Instance created successfully');
      queryClient.invalidateQueries({ queryKey: ['instances', providerName] });
    },
    onError: (error) => {
      showError(`Failed to create instance: ${error.message}`);
    },
  });

  // Delete instance mutation (placeholder)
  const deleteInstanceMutation = useMutation({
    mutationFn: (instanceId: string) => {
      // This would be implemented when the API endpoint is available
      return Promise.resolve();
    },
    onSuccess: () => {
      success('Instance deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['instances', providerName] });
    },
    onError: (error) => {
      showError(`Failed to delete instance: ${error.message}`);
    },
  });

  const handleCreateInstance = () => {
    // This would open a modal or navigate to a create instance page
    success('Create instance functionality coming soon');
  };

  const handleDeleteInstance = (instanceId: string) => {
    if (confirm('Are you sure you want to delete this instance?')) {
      deleteInstanceMutation.mutate(instanceId);
    }
  };

  const filteredInstances = instances.filter(instance =>
    instance.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    instance.id?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (authLoading || providerLoading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">Loading provider details...</p>
          </div>
        </div>
      </Layout>
    );
  }

  if (!provider) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">Provider not found</h3>
            <p className="mt-1 text-sm text-gray-500">The requested provider could not be found.</p>
            <Button onClick={() => router.push('/providers')} className="mt-4">
              Back to Providers
            </Button>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button
              variant="outline"
              size="sm"
              onClick={() => router.push('/providers')}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back
            </Button>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{provider.name}</h1>
              <p className="text-gray-600">Cloud Provider Management</p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <Badge variant="outline">v{provider.version}</Badge>
            <Button onClick={handleCreateInstance}>
              <Plus className="mr-2 h-4 w-4" />
              Create Instance
            </Button>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Instances</CardTitle>
              <Server className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{instances.length}</div>
              <p className="text-xs text-muted-foreground">
                {instances.filter(i => i.status === 'running').length} running
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Regions</CardTitle>
              <MapPin className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{regions.length}</div>
              <p className="text-xs text-muted-foreground">Available regions</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Monthly Cost</CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">$1,800</div>
              <p className="text-xs text-muted-foreground">
                <span className="text-green-600">+12%</span> from last month
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Services</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">8</div>
              <p className="text-xs text-muted-foreground">Services in use</p>
            </CardContent>
          </Card>
        </div>

        {/* Main Content Tabs */}
        <Tabs defaultValue="instances" className="space-y-4">
          <TabsList>
            <TabsTrigger value="instances">Instances</TabsTrigger>
            <TabsTrigger value="costs">Costs</TabsTrigger>
            <TabsTrigger value="regions">Regions</TabsTrigger>
            <TabsTrigger value="analytics">Analytics</TabsTrigger>
          </TabsList>

          {/* Instances Tab */}
          <TabsContent value="instances" className="space-y-4">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Instances</CardTitle>
                    <CardDescription>Manage your cloud instances</CardDescription>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div className="relative">
                      <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="Search instances..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-8 w-64"
                      />
                    </div>
                    <Select value={selectedRegion} onValueChange={setSelectedRegion}>
                      <SelectTrigger className="w-48">
                        <SelectValue placeholder="Filter by region" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="">All Regions</SelectItem>
                        {regions.map((region) => (
                          <SelectItem key={region.name} value={region.name}>
                            {region.display_name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Button variant="outline" size="sm">
                      <Filter className="mr-2 h-4 w-4" />
                      Filter
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Region</TableHead>
                      <TableHead>Public IP</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredInstances.map((instance) => (
                      <TableRow key={instance.id}>
                        <TableCell className="font-medium">{instance.name}</TableCell>
                        <TableCell>{instance.type}</TableCell>
                        <TableCell>
                          <Badge 
                            variant={instance.status === 'running' ? 'default' : 'secondary'}
                          >
                            {instance.status}
                          </Badge>
                        </TableCell>
                        <TableCell>{instance.region}</TableCell>
                        <TableCell>{instance.public_ip || 'N/A'}</TableCell>
                        <TableCell>
                          {new Date(instance.created_at).toLocaleDateString()}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center space-x-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => router.push(`/vms/${instance.id}`)}
                            >
                              View
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleDeleteInstance(instance.id)}
                              disabled={deleteInstanceMutation.isPending}
                            >
                              Delete
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                {filteredInstances.length === 0 && (
                  <div className="text-center py-8">
                    <Server className="mx-auto h-12 w-12 text-gray-400" />
                    <h3 className="mt-2 text-sm font-medium text-gray-900">No instances found</h3>
                    <p className="mt-1 text-sm text-gray-500">
                      Get started by creating a new instance.
                    </p>
                    <div className="mt-6">
                      <Button onClick={handleCreateInstance}>
                        <Plus className="mr-2 h-4 w-4" />
                        Create Instance
                      </Button>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* Costs Tab */}
          <TabsContent value="costs" className="space-y-4">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              <Card>
                <CardHeader>
                  <CardTitle>Monthly Cost Trend</CardTitle>
                  <CardDescription>Cost over the last 6 months</CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={costData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="month" />
                      <YAxis />
                      <Tooltip formatter={(value) => [`$${value}`, 'Cost']} />
                      <Line type="monotone" dataKey="cost" stroke="#8884d8" strokeWidth={2} />
                    </LineChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Cost by Service</CardTitle>
                  <CardDescription>Breakdown of costs by service</CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={usageData}
                        cx="50%"
                        cy="50%"
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="usage"
                      >
                        {usageData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip formatter={(value, name) => [`$${value}`, name]} />
                    </PieChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>

              <Card className="lg:col-span-2">
                <CardHeader>
                  <CardTitle>Cost by Region</CardTitle>
                  <CardDescription>Monthly costs across different regions</CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={regionData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="name" />
                      <YAxis />
                      <Tooltip formatter={(value) => [`$${value}`, 'Cost']} />
                      <Bar dataKey="cost" fill="#8884d8" />
                    </BarChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* Regions Tab */}
          <TabsContent value="regions" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Available Regions</CardTitle>
                <CardDescription>Regions where you can deploy resources</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {regions.map((region) => (
                    <Card key={region.name}>
                      <CardHeader className="pb-2">
                        <CardTitle className="text-sm">{region.display_name}</CardTitle>
                        <CardDescription className="text-xs">{region.name}</CardDescription>
                      </CardHeader>
                      <CardContent>
                        <div className="text-2xl font-bold">
                          {instances.filter(i => i.region === region.name).length}
                        </div>
                        <p className="text-xs text-muted-foreground">instances</p>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Analytics Tab */}
          <TabsContent value="analytics" className="space-y-4">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              <Card>
                <CardHeader>
                  <CardTitle>Resource Utilization</CardTitle>
                  <CardDescription>Current resource usage across all instances</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div>
                      <div className="flex justify-between text-sm mb-1">
                        <span>CPU Usage</span>
                        <span>65%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div className="bg-blue-600 h-2 rounded-full" style={{ width: '65%' }}></div>
                      </div>
                    </div>
                    <div>
                      <div className="flex justify-between text-sm mb-1">
                        <span>Memory Usage</span>
                        <span>42%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div className="bg-green-600 h-2 rounded-full" style={{ width: '42%' }}></div>
                      </div>
                    </div>
                    <div>
                      <div className="flex justify-between text-sm mb-1">
                        <span>Storage Usage</span>
                        <span>78%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div className="bg-yellow-600 h-2 rounded-full" style={{ width: '78%' }}></div>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Performance Metrics</CardTitle>
                  <CardDescription>Key performance indicators</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Uptime</span>
                      <span className="text-sm font-medium">99.9%</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Response Time</span>
                      <span className="text-sm font-medium">120ms</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Throughput</span>
                      <span className="text-sm font-medium">1.2M req/min</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Error Rate</span>
                      <span className="text-sm font-medium text-green-600">0.01%</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </Layout>
  );
}

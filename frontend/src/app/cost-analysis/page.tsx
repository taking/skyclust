'use client';

import { useState } from 'react';
import dynamic from 'next/dynamic';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useWorkspaceStore } from '@/store/workspace';
import { useCostBreakdown } from '@/hooks/use-cost-analysis';
import { DollarSign, TrendingUp, PieChart as PieChartIcon, BarChart3 } from 'lucide-react';

// Dynamic imports for cost analysis components
const CostSummaryCard = dynamic(
  () => import('@/components/cost-analysis/cost-summary-card').then(mod => ({ default: mod.CostSummaryCard })),
  { 
    ssr: false,
    loading: () => (
      <Card>
        <CardHeader>
          <div className="h-6 bg-gray-200 rounded animate-pulse w-32"></div>
        </CardHeader>
        <CardContent>
          <div className="h-32 bg-gray-200 rounded animate-pulse"></div>
        </CardContent>
      </Card>
    ),
  }
);

const CostTrendChart = dynamic(
  () => import('@/components/cost-analysis/cost-trend-chart').then(mod => ({ default: mod.CostTrendChart })),
  { 
    ssr: false,
    loading: () => (
      <Card>
        <CardHeader>
          <div className="h-6 bg-gray-200 rounded animate-pulse w-32"></div>
        </CardHeader>
        <CardContent>
          <div className="h-64 bg-gray-200 rounded animate-pulse"></div>
        </CardContent>
      </Card>
    ),
  }
);

const CostPredictionChart = dynamic(
  () => import('@/components/cost-analysis/cost-prediction-chart').then(mod => ({ default: mod.CostPredictionChart })),
  { 
    ssr: false,
    loading: () => (
      <Card>
        <CardHeader>
          <div className="h-6 bg-gray-200 rounded animate-pulse w-32"></div>
        </CardHeader>
        <CardContent>
          <div className="h-64 bg-gray-200 rounded animate-pulse"></div>
        </CardContent>
      </Card>
    ),
  }
);

const BudgetAlerts = dynamic(
  () => import('@/components/cost-analysis/budget-alerts').then(mod => ({ default: mod.BudgetAlerts })),
  { 
    ssr: false,
    loading: () => (
      <Card>
        <CardHeader>
          <div className="h-6 bg-gray-200 rounded animate-pulse w-32"></div>
        </CardHeader>
        <CardContent>
          <div className="h-32 bg-gray-200 rounded animate-pulse"></div>
        </CardContent>
      </Card>
    ),
  }
);

// Dynamic import for charts (recharts is heavy)
const CostBreakdownPieChart = dynamic(
  () => import('recharts').then(mod => ({
    default: ({ data, colors }: { data: Array<{ name?: string; value: number; percent?: number }>; colors: string[] }) => (
      <mod.ResponsiveContainer width="100%" height="100%">
        <mod.PieChart>
          <mod.Pie
            data={data}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={(entry: { name?: string; percent?: number }) => `${entry.name || 'Unknown'} (${entry.percent?.toFixed(1) || 0}%)`}
            outerRadius={80}
            fill="#8884d8"
            dataKey="value"
          >
            {data.map((entry, index) => (
              <mod.Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
            ))}
          </mod.Pie>
          <mod.Tooltip formatter={(value: number) => [`$${value.toFixed(2)}`, 'Cost']} />
          <mod.Legend />
        </mod.PieChart>
      </mod.ResponsiveContainer>
    ),
  })),
  { ssr: false, loading: () => <div className="h-64 bg-gray-100 animate-pulse rounded" /> }
);

const CostBreakdownBarChart = dynamic(
  () => import('recharts').then(mod => ({
    default: ({ data }: { data: Array<{ name?: string; value: number }> }) => (
      <mod.ResponsiveContainer width="100%" height="100%">
        <mod.BarChart data={data}>
          <mod.CartesianGrid strokeDasharray="3 3" />
          <mod.XAxis 
            dataKey="name" 
            tick={{ fontSize: 12 }}
            angle={-45}
            textAnchor="end"
            height={60}
          />
          <mod.YAxis 
            tick={{ fontSize: 12 }}
            tickFormatter={(value) => `$${value.toFixed(0)}`}
          />
          <mod.Tooltip formatter={(value: number) => [`$${value.toFixed(2)}`, 'Cost']} />
          <mod.Bar dataKey="value" fill="#8884d8" />
        </mod.BarChart>
      </mod.ResponsiveContainer>
    ),
  })),
  { ssr: false, loading: () => <div className="h-64 bg-gray-100 animate-pulse rounded" /> }
);

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];

export default function CostAnalysisPage() {
  const { currentWorkspace } = useWorkspaceStore();
  const [selectedPeriod, setSelectedPeriod] = useState('30d');
  const [selectedDimension, setSelectedDimension] = useState('service');
  const [budgetLimit, setBudgetLimit] = useState(1000);
  const [predictionDays, setPredictionDays] = useState(30);

  const { data: breakdown, isLoading: breakdownLoading } = useCostBreakdown(
    currentWorkspace?.id || '',
    selectedPeriod,
    selectedDimension
  );

  if (!currentWorkspace) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="p-6">
            <p className="text-center text-gray-600">
              Please select a workspace to view cost analysis
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 헤더 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Cost Analysis</h1>
          <p className="text-gray-600">
            Monitor and analyze costs for {currentWorkspace.name}
          </p>
        </div>
        <div className="flex items-center gap-4">
          <Select value={selectedPeriod} onValueChange={setSelectedPeriod}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="7d">7 days</SelectItem>
              <SelectItem value="30d">30 days</SelectItem>
              <SelectItem value="90d">90 days</SelectItem>
              <SelectItem value="1y">1 year</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* 비용 요약 카드 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <CostSummaryCard 
          workspaceId={currentWorkspace.id} 
          period={selectedPeriod}
          className="lg:col-span-2"
        />
        <BudgetAlerts 
          workspaceId={currentWorkspace.id} 
          budgetLimit={budgetLimit}
        />
      </div>

      {/* 탭 컨텐츠 */}
      <Tabs defaultValue="trend" className="space-y-6">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="trend" className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            Trend
          </TabsTrigger>
          <TabsTrigger value="prediction" className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            Prediction
          </TabsTrigger>
          <TabsTrigger value="breakdown" className="flex items-center gap-2">
            <PieChartIcon className="h-4 w-4" />
            Breakdown
          </TabsTrigger>
          <TabsTrigger value="comparison" className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            Comparison
          </TabsTrigger>
        </TabsList>

        {/* 트렌드 탭 */}
        <TabsContent value="trend">
          <CostTrendChart 
            workspaceId={currentWorkspace.id} 
            period={selectedPeriod}
          />
        </TabsContent>

        {/* 예측 탭 */}
        <TabsContent value="prediction">
          <div className="space-y-4">
            <div className="flex items-center gap-4">
              <label className="text-sm font-medium">Prediction Days:</label>
              <Select value={predictionDays.toString()} onValueChange={(value) => setPredictionDays(parseInt(value))}>
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="7">7 days</SelectItem>
                  <SelectItem value="14">14 days</SelectItem>
                  <SelectItem value="30">30 days</SelectItem>
                  <SelectItem value="60">60 days</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <CostPredictionChart 
              workspaceId={currentWorkspace.id} 
              days={predictionDays}
            />
          </div>
        </TabsContent>

        {/* 분석 탭 */}
        <TabsContent value="breakdown">
          <div className="space-y-6">
            <div className="flex items-center gap-4">
              <label className="text-sm font-medium">Breakdown by:</label>
              <Select value={selectedDimension} onValueChange={setSelectedDimension}>
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="service">Service</SelectItem>
                  <SelectItem value="provider">Provider</SelectItem>
                  <SelectItem value="region">Region</SelectItem>
                  <SelectItem value="workspace">Workspace</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {breakdownLoading ? (
              <Card>
                <CardContent className="p-6">
                  <div className="h-64 bg-gray-200 rounded animate-pulse" />
                </CardContent>
              </Card>
            ) : breakdown ? (
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* 파이 차트 */}
                <Card>
                  <CardHeader>
                    <CardTitle>Cost Distribution</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="h-64">
                      <CostBreakdownPieChart data={breakdown.breakdown} colors={COLORS} />
                    </div>
                  </CardContent>
                </Card>

                {/* 바 차트 */}
                <Card>
                  <CardHeader>
                    <CardTitle>Cost by {selectedDimension.charAt(0).toUpperCase() + selectedDimension.slice(1)}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="h-64">
                      <CostBreakdownBarChart data={breakdown.breakdown} />
                    </div>
                  </CardContent>
                </Card>
              </div>
            ) : (
              <Card>
                <CardContent className="p-6">
                  <p className="text-center text-gray-600">No breakdown data available</p>
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>

        {/* 비교 탭 */}
        <TabsContent value="comparison">
          <Card>
            <CardHeader>
              <CardTitle>Cost Comparison</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-center text-gray-600">
                Cost comparison feature will be implemented here
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 예산 설정 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DollarSign className="h-5 w-5" />
            Budget Settings
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <label className="text-sm font-medium">Monthly Budget Limit:</label>
            <Input
              type="number"
              value={budgetLimit}
              onChange={(e) => setBudgetLimit(parseFloat(e.target.value) || 0)}
              className="w-32"
              min="0"
              step="0.01"
            />
            <span className="text-sm text-gray-600">USD</span>
            <Button size="sm">Update Budget</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

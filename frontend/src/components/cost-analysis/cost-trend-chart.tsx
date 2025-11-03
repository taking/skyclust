'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useCostTrend } from '@/hooks/use-cost-analysis';
import { costAnalysisUtils } from '@/hooks/use-cost-analysis';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Area, AreaChart } from 'recharts';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';
import { cn } from '@/lib/utils';

interface CostTrendChartProps {
  workspaceId: string;
  period?: string;
  className?: string;
}

export function CostTrendChart({ workspaceId, period = '90d', className }: CostTrendChartProps) {
  const { data: trendData, isLoading, error } = useCostTrend(workspaceId, period);

  if (isLoading) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle>Cost Trend</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 bg-gray-200 rounded animate-pulse" />
        </CardContent>
      </Card>
    );
  }

  if (error || !trendData) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle>Cost Trend</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-red-600">Failed to load cost trend</p>
        </CardContent>
      </Card>
    );
  }

  // 차트 데이터 준비
  const chartData = trendData.daily_costs.map(cost => ({
    date: costAnalysisUtils.formatDate(cost.date),
    cost: cost.amount,
    service: cost.service,
    provider: cost.provider,
  }));

  // 일별 총 비용 계산
  const dailyTotals = chartData.reduce((acc, item) => {
    const existing = acc.find(d => d.date === item.date);
    if (existing) {
      existing.cost += item.cost;
    } else {
      acc.push({ ...item });
    }
    return acc;
  }, [] as typeof chartData);

  const trendIcon = costAnalysisUtils.getTrendIcon(trendData.trend);
  const trendColor = costAnalysisUtils.getTrendColor(trendData.trend);

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <span className="text-xl">{trendIcon}</span>
          Cost Trend
          <span className={cn('text-sm font-normal', trendColor)}>
            ({costAnalysisUtils.formatPercentage(trendData.growth_rate)})
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={dailyTotals}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis 
                dataKey="date" 
                tick={{ fontSize: 12 }}
                angle={-45}
                textAnchor="end"
                height={60}
              />
              <YAxis 
                tick={{ fontSize: 12 }}
                tickFormatter={(value) => `$${value.toFixed(0)}`}
              />
              <Tooltip 
                formatter={(value: number) => [costAnalysisUtils.formatCurrency(value, 'USD'), 'Cost']}
                labelFormatter={(label) => `Date: ${label}`}
              />
              <Area
                type="monotone"
                dataKey="cost"
                stroke="#8884d8"
                fill="#8884d8"
                fillOpacity={0.3}
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
        
        {/* 요약 정보 */}
        <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-600">Total Cost:</span>
            <span className="ml-2 font-semibold">
              {costAnalysisUtils.formatCurrency(trendData.total_cost, trendData.currency)}
            </span>
          </div>
          <div>
            <span className="text-gray-600">Period:</span>
            <span className="ml-2 font-semibold">{trendData.period}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { useCostSummary } from '@/hooks/useCostAnalysis';
import { costAnalysisUtils } from '@/hooks/useCostAnalysis';
import { TrendingUp, TrendingDown, Minus, DollarSign } from 'lucide-react';
import { cn } from '@/lib/utils';

interface CostSummaryCardProps {
  workspaceId: string;
  period?: string;
  className?: string;
}

export function CostSummaryCard({ workspaceId, period = '30d', className }: CostSummaryCardProps) {
  const { data: summary, isLoading, error } = useCostSummary(workspaceId, period);

  if (isLoading) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DollarSign className="h-5 w-5" />
            Cost Summary
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="h-8 bg-gray-200 rounded animate-pulse" />
            <div className="h-4 bg-gray-200 rounded animate-pulse" />
            <div className="h-4 bg-gray-200 rounded animate-pulse" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error || !summary) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DollarSign className="h-5 w-5" />
            Cost Summary
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-red-600">Failed to load cost summary</p>
        </CardContent>
      </Card>
    );
  }

  const trendIcon = costAnalysisUtils.getTrendIcon(summary.trend);
  const trendColor = costAnalysisUtils.getTrendColor(summary.trend);

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <DollarSign className="h-5 w-5" />
          Cost Summary
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 총 비용 */}
        <div className="text-center">
          <div className="text-3xl font-bold">
            {costAnalysisUtils.formatCurrency(summary.total_cost, summary.currency)}
          </div>
          <div className="text-sm text-gray-600">
            {summary.period} period
          </div>
        </div>

        {/* 트렌드 */}
        <div className="flex items-center justify-center gap-2">
          <span className="text-2xl">{trendIcon}</span>
          <div className="text-center">
            <div className={cn('font-semibold', trendColor)}>
              {summary.trend.charAt(0).toUpperCase() + summary.trend.slice(1)}
            </div>
            <div className="text-sm text-gray-600">
              {costAnalysisUtils.formatPercentage(summary.growth_rate)}
            </div>
          </div>
        </div>

        {/* 기간 정보 */}
        <div className="text-sm text-gray-600 text-center">
          {costAnalysisUtils.formatDate(summary.start_date)} - {costAnalysisUtils.formatDate(summary.end_date)}
        </div>

        {/* 서비스별 비용 */}
        <div className="space-y-2">
          <h4 className="font-semibold text-sm">By Service</h4>
          {Object.entries(summary.by_service).map(([service, cost]) => {
            const percentage = (cost / summary.total_cost) * 100;
            return (
              <div key={service} className="flex items-center justify-between">
                <span className="text-sm capitalize">{service}</span>
                <div className="flex items-center gap-2">
                  <div className="w-20">
                    <Progress value={percentage} className="h-2" />
                  </div>
                  <span className="text-sm font-medium w-20 text-right">
                    {costAnalysisUtils.formatCurrency(cost, summary.currency)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>

        {/* 프로바이더별 비용 */}
        <div className="space-y-2">
          <h4 className="font-semibold text-sm">By Provider</h4>
          {Object.entries(summary.by_provider).map(([provider, cost]) => {
            const percentage = (cost / summary.total_cost) * 100;
            return (
              <div key={provider} className="flex items-center justify-between">
                <Badge variant="outline" className="text-xs">
                  {provider.toUpperCase()}
                </Badge>
                <div className="flex items-center gap-2">
                  <div className="w-20">
                    <Progress value={percentage} className="h-2" />
                  </div>
                  <span className="text-sm font-medium w-20 text-right">
                    {costAnalysisUtils.formatCurrency(cost, summary.currency)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}

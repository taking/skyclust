'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { useBudgetAlerts } from '@/hooks/useCostAnalysis';
import { costAnalysisUtils } from '@/hooks/useCostAnalysis';
import { AlertTriangle, DollarSign, Target } from 'lucide-react';
import { cn } from '@/lib/utils';

interface BudgetAlertsProps {
  workspaceId: string;
  budgetLimit: number;
  className?: string;
}

export function BudgetAlerts({ workspaceId, budgetLimit, className }: BudgetAlertsProps) {
  const { data: alerts, isLoading, error } = useBudgetAlerts(workspaceId, budgetLimit);

  if (isLoading) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Target className="h-5 w-5" />
            Budget Alerts
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="h-8 bg-gray-200 rounded animate-pulse" />
            <div className="h-4 bg-gray-200 rounded animate-pulse" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Target className="h-5 w-5" />
            Budget Alerts
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-red-600">Failed to load budget alerts</p>
        </CardContent>
      </Card>
    );
  }

  if (!alerts || alerts.length === 0) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Target className="h-5 w-5" />
            Budget Alerts
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <Target className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">No budget alerts</p>
            <p className="text-sm text-gray-500">
              Budget usage is within normal limits
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <AlertTriangle className="h-5 w-5" />
          Budget Alerts
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {alerts.map((alert) => {
          const usageColor = costAnalysisUtils.getBudgetUsageColor(alert.percentage);
          const usageBgColor = costAnalysisUtils.getBudgetUsageBgColor(alert.percentage);
          
          return (
            <div key={alert.id} className="space-y-3">
              {/* 알림 레벨 */}
              <div className="flex items-center justify-between">
                <Badge 
                  variant={alert.alert_level === 'critical' ? 'destructive' : 'default'}
                  className="text-xs"
                >
                  {alert.alert_level.toUpperCase()}
                </Badge>
                <span className="text-sm text-gray-600">
                  {costAnalysisUtils.formatDate(alert.created_at)}
                </span>
              </div>

              {/* 예산 사용률 */}
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">Budget Usage</span>
                  <span className={cn('font-semibold', usageColor)}>
                    {costAnalysisUtils.formatPercentage(alert.percentage)}
                  </span>
                </div>
                <Progress 
                  value={Math.min(alert.percentage, 100)} 
                  className="h-2"
                />
              </div>

              {/* 비용 정보 */}
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-gray-600">Current Cost:</span>
                  <div className="font-semibold">
                    {costAnalysisUtils.formatCurrency(alert.current_cost, 'USD')}
                  </div>
                </div>
                <div>
                  <span className="text-gray-600">Budget Limit:</span>
                  <div className="font-semibold">
                    {costAnalysisUtils.formatCurrency(alert.budget_limit, 'USD')}
                  </div>
                </div>
              </div>

              {/* 메시지 */}
              <div className={cn('p-3 rounded-md text-sm', usageBgColor)}>
                <div className="flex items-start gap-2">
                  <AlertTriangle className={cn('h-4 w-4 mt-0.5', usageColor)} />
                  <span className={usageColor}>{alert.message}</span>
                </div>
              </div>

              {/* 남은 예산 */}
              {alert.percentage < 100 && (
                <div className="text-sm text-gray-600">
                  <span>Remaining Budget: </span>
                  <span className="font-semibold">
                    {costAnalysisUtils.formatCurrency(alert.budget_limit - alert.current_cost, 'USD')}
                  </span>
                </div>
              )}
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}

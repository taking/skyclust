'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useCostPredictions } from '@/hooks/use-cost-analysis';
import { costAnalysisUtils } from '@/hooks/use-cost-analysis';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';
import { TrendingUp, Calendar } from 'lucide-react';
import { cn } from '@/lib/utils';

interface CostPredictionChartProps {
  workspaceId: string;
  days?: number;
  className?: string;
}

export function CostPredictionChart({ workspaceId, days = 30, className }: CostPredictionChartProps) {
  const { data: predictions, isLoading, error } = useCostPredictions(workspaceId, days);

  if (isLoading) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle>Cost Predictions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 bg-gray-200 rounded animate-pulse" />
        </CardContent>
      </Card>
    );
  }

  if (error || !predictions || predictions.length === 0) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader>
          <CardTitle>Cost Predictions</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-gray-600">No prediction data available</p>
        </CardContent>
      </Card>
    );
  }

  // 차트 데이터 준비
  const chartData = predictions.map(prediction => ({
    date: costAnalysisUtils.formatDate(prediction.date),
    predicted: prediction.predicted,
    lowerBound: prediction.lower_bound,
    upperBound: prediction.upper_bound,
    confidence: prediction.confidence,
  }));

  // 평균 신뢰도 계산
  const avgConfidence = predictions.reduce((sum, p) => sum + p.confidence, 0) / predictions.length;
  const confidenceColor = costAnalysisUtils.getConfidenceColor(avgConfidence);

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TrendingUp className="h-5 w-5" />
          Cost Predictions
          <span className="text-sm font-normal text-gray-600">
            ({days} days ahead)
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
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
                formatter={(value: number, name: string) => {
                  const label = name === 'predicted' ? 'Predicted' : 
                               name === 'lowerBound' ? 'Lower Bound' : 'Upper Bound';
                  return [costAnalysisUtils.formatCurrency(value, 'USD'), label];
                }}
                labelFormatter={(label) => `Date: ${label}`}
              />
              
              {/* 신뢰 구간 */}
              <Line
                type="monotone"
                dataKey="upperBound"
                stroke="#e5e7eb"
                strokeWidth={1}
                strokeDasharray="5 5"
                dot={false}
                name="upperBound"
              />
              <Line
                type="monotone"
                dataKey="lowerBound"
                stroke="#e5e7eb"
                strokeWidth={1}
                strokeDasharray="5 5"
                dot={false}
                name="lowerBound"
              />
              
              {/* 예측값 */}
              <Line
                type="monotone"
                dataKey="predicted"
                stroke="#3b82f6"
                strokeWidth={3}
                dot={{ fill: '#3b82f6', strokeWidth: 2, r: 4 }}
                name="predicted"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
        
        {/* 신뢰도 정보 */}
        <div className="mt-4 flex items-center justify-between text-sm">
          <div className="flex items-center gap-2">
            <span className="text-gray-600">Average Confidence:</span>
            <span className={cn('font-semibold', confidenceColor)}>
              {costAnalysisUtils.formatPercentage(avgConfidence * 100)}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4 text-gray-600" />
            <span className="text-gray-600">
              {predictions.length} days predicted
            </span>
          </div>
        </div>
        
        {/* 범례 */}
        <div className="mt-2 flex items-center gap-4 text-xs text-gray-600">
          <div className="flex items-center gap-1">
            <div className="w-3 h-0.5 bg-blue-500" />
            <span>Predicted</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-0.5 bg-gray-400 border-dashed border-t" />
            <span>Confidence Range</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

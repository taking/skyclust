/**
 * Chart Skeleton Component
 * 차트 로딩 상태를 표시하는 Skeleton 컴포넌트
 */

'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface ChartSkeletonProps {
  /**
   * 차트 타입
   * - 'line': 라인 차트
   * - 'bar': 바 차트
   * - 'pie': 파이 차트
   * - 'area': 영역 차트
   */
  type?: 'line' | 'bar' | 'pie' | 'area';
  
  /**
   * 카드 헤더 표시 여부
   */
  showHeader?: boolean;
  
  /**
   * 차트 높이
   */
  height?: string;
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

export function ChartSkeleton({
  type = 'line',
  showHeader = true,
  height = '300px',
  className,
}: ChartSkeletonProps) {
  const renderChartContent = () => {
    switch (type) {
      case 'line':
      case 'area':
        return (
          <div className="space-y-4">
            {/* Y-axis labels */}
            <div className="flex justify-between items-end h-[calc(100%-2rem)] px-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="flex flex-col items-center space-y-2">
                  <Skeleton size="sm" className="w-8 h-3" />
                  <Skeleton
                    variant="rect"
                    className="w-full"
                    style={{ height: `${Math.random() * 60 + 20}%` }}
                  />
                </div>
              ))}
            </div>
            {/* X-axis labels */}
            <div className="flex justify-between px-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} size="sm" className="w-12 h-3" />
              ))}
            </div>
          </div>
        );
        
      case 'bar':
        return (
          <div className="space-y-4">
            {/* Bars */}
            <div className="flex justify-between items-end h-[calc(100%-2rem)] px-4 space-x-2">
              {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="flex-1 flex flex-col items-center space-y-2">
                  <Skeleton
                    variant="rect"
                    className="w-full rounded-t"
                    style={{ height: `${Math.random() * 70 + 20}%` }}
                  />
                  <Skeleton size="sm" className="w-8 h-3" />
                </div>
              ))}
            </div>
          </div>
        );
        
      case 'pie':
        return (
          <div className="flex items-center justify-center h-full">
            <div className="relative">
              <Skeleton variant="circle" size="xl" className="w-48 h-48" />
              <div className="absolute inset-0 flex items-center justify-center">
                <Skeleton size="lg" className="w-24 h-6" />
              </div>
            </div>
          </div>
        );
        
      default:
        return (
          <div className="space-y-4">
            <Skeleton className="w-full h-4" />
            <Skeleton className="w-full h-4" />
            <Skeleton className="w-full h-4" />
          </div>
        );
    }
  };

  return (
    <Card className={className}>
      {showHeader && (
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-32 mt-2" />
        </CardHeader>
      )}
      <CardContent style={{ height }}>
        {renderChartContent()}
      </CardContent>
    </Card>
  );
}


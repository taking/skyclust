/**
 * Widget Skeleton Component
 * 대시보드 위젯 로딩 상태를 표시하는 Skeleton 컴포넌트
 */

'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface WidgetSkeletonProps {
  /**
   * 위젯 타입
   * - 'stats': 통계 위젯
   * - 'chart': 차트 위젯
   * - 'list': 리스트 위젯
   */
  type?: 'stats' | 'chart' | 'list';
  
  /**
   * 헤더 표시 여부
   */
  showHeader?: boolean;
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

export function WidgetSkeleton({
  type = 'stats',
  showHeader = true,
  className,
}: WidgetSkeletonProps) {
  const renderContent = () => {
    switch (type) {
      case 'stats':
        return (
          <div className="space-y-4">
            <Skeleton className="h-8 w-32" />
            <div className="flex items-center space-x-2">
              <Skeleton size="sm" className="w-16 h-4" />
              <Skeleton variant="circle" size="sm" className="w-4 h-4" />
            </div>
          </div>
        );
        
      case 'chart':
        return (
          <div className="space-y-4">
            {/* Chart area */}
            <div className="flex justify-between items-end h-48 px-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="flex flex-col items-center space-y-2 flex-1">
                  <Skeleton
                    variant="rect"
                    className="w-full rounded-t"
                    style={{ height: `${Math.random() * 60 + 20}%` }}
                  />
                  <Skeleton size="sm" className="w-8 h-3" />
                </div>
              ))}
            </div>
          </div>
        );
        
      case 'list':
        return (
          <div className="space-y-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center space-x-3">
                <Skeleton variant="circle" size="sm" className="w-8 h-8" />
                <div className="flex-1 space-y-1">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton size="sm" className="h-3 w-1/2" />
                </div>
              </div>
            ))}
          </div>
        );
        
      default:
        return <Skeleton className="h-32 w-full" />;
    }
  };

  return (
    <Card className={cn('h-full', className)}>
      {showHeader && (
        <CardHeader>
          <Skeleton className="h-5 w-32" />
          <Skeleton size="sm" className="h-3 w-48 mt-1" />
        </CardHeader>
      )}
      <CardContent>
        {renderContent()}
      </CardContent>
    </Card>
  );
}


/**
 * Page Skeleton Component
 * 페이지 전체 로딩 상태를 표시하는 Skeleton 컴포넌트
 */

'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface PageSkeletonProps {
  /**
   * 헤더 표시 여부
   */
  showHeader?: boolean;
  
  /**
   * 카드 개수 (그리드 레이아웃)
   */
  cards?: number;
  
  /**
   * 테이블 표시 여부
   */
  showTable?: boolean;
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

export function PageSkeleton({
  showHeader = true,
  cards = 3,
  showTable = false,
  className,
}: PageSkeletonProps) {
  return (
    <div className={cn('space-y-6', className)}>
      {/* Header */}
      {showHeader && (
        <div className="space-y-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
        </div>
      )}
      
      {/* Stats Cards */}
      {cards > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: cards }).map((_, index) => (
            <Card key={index}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <Skeleton className="h-4 w-24" />
                <Skeleton variant="circle" size="sm" className="w-4 h-4" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-32 mb-2" />
                <Skeleton size="sm" className="w-40" />
              </CardContent>
            </Card>
          ))}
        </div>
      )}
      
      {/* Table */}
      {showTable && (
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-64 mt-2" />
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {/* Table header */}
              <div className="flex space-x-4">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} size="sm" className="w-20 h-4" />
                ))}
              </div>
              {/* Table rows */}
              {Array.from({ length: 5 }).map((_, rowIndex) => (
                <div key={rowIndex} className="flex space-x-4">
                  {Array.from({ length: 5 }).map((_, colIndex) => (
                    <Skeleton key={colIndex} className="h-4 flex-1" />
                  ))}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}


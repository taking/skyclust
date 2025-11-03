/**
 * Form Skeleton Component
 * 폼 로딩 상태를 표시하는 Skeleton 컴포넌트
 */

'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface FormSkeletonProps {
  /**
   * 폼 필드 개수
   */
  fields?: number;
  
  /**
   * 카드 헤더 표시 여부
   */
  showHeader?: boolean;
  
  /**
   * 버튼 표시 여부
   */
  showButtons?: boolean;
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

export function FormSkeleton({
  fields = 5,
  showHeader = true,
  showButtons = true,
  className,
}: FormSkeletonProps) {
  return (
    <Card className={className}>
      {showHeader && (
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64 mt-2" />
        </CardHeader>
      )}
      <CardContent className="space-y-6">
        {Array.from({ length: fields }).map((_, index) => (
          <div key={index} className="space-y-2">
            {/* Label */}
            <Skeleton size="sm" className="w-24 h-4" />
            {/* Input */}
            <Skeleton className="w-full h-10" />
            {/* Optional description */}
            {index % 3 === 0 && (
              <Skeleton size="sm" className="w-3/4 h-3" />
            )}
          </div>
        ))}
        
        {showButtons && (
          <div className="flex justify-end space-x-2 pt-4">
            <Skeleton className="w-20 h-10" />
            <Skeleton className="w-20 h-10" />
          </div>
        )}
      </CardContent>
    </Card>
  );
}


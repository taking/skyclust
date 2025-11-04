/**
 * List Skeleton Component
 * 리스트 아이템 로딩 상태를 표시하는 Skeleton 컴포넌트
 */

'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

export interface ListSkeletonProps {
  /**
   * 리스트 아이템 개수
   */
  items?: number;
  
  /**
   * 각 아이템의 라인 수
   */
  linesPerItem?: number;
  
  /**
   * 아바타 표시 여부
   */
  showAvatar?: boolean;
  
  /**
   * 액션 버튼 표시 여부
   */
  showActions?: boolean;
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

export function ListSkeleton({
  items = 5,
  linesPerItem = 2,
  showAvatar = false,
  showActions = false,
  className,
}: ListSkeletonProps) {
  return (
    <div className={cn('space-y-4', className)}>
      {Array.from({ length: items }).map((_, index) => (
        <div
          key={index}
          className="flex items-start space-x-4 p-4 border-b border-gray-200 last:border-b-0"
        >
          {showAvatar && (
            <Skeleton variant="circle" size="lg" className="w-10 h-10 flex-shrink-0" />
          )}
          
          <div className="flex-1 space-y-2 min-w-0">
            {Array.from({ length: linesPerItem }).map((_, lineIndex) => (
              <Skeleton
                key={lineIndex}
                variant="text"
                className={cn(
                  lineIndex === 0 ? 'w-3/4' : 'w-full',
                  lineIndex === linesPerItem - 1 ? 'w-1/2' : ''
                )}
              />
            ))}
          </div>
          
          {showActions && (
            <div className="flex space-x-2">
              <Skeleton size="sm" className="w-8 h-8" />
              <Skeleton size="sm" className="w-8 h-8" />
            </div>
          )}
        </div>
      ))}
    </div>
  );
}


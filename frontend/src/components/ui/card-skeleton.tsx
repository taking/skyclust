/**
 * Card Skeleton Component
 * 카드 로딩 상태를 표시하는 Skeleton 컴포넌트
 */

'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export interface CardSkeletonProps {
  /**
   * 헤더 표시 여부
   */
  showHeader?: boolean;
  
  /**
   * 헤더에 설명 표시 여부
   */
  showDescription?: boolean;
  
  /**
   * 콘텐츠 라인 수
   */
  lines?: number;
  
  /**
   * 카드 타입
   * - 'default': 기본 카드
   * - 'stats': 통계 카드
   * - 'content': 콘텐츠 카드
   */
  type?: 'default' | 'stats' | 'content';
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

export function CardSkeleton({ 
  showHeader = true, 
  showDescription = true,
  lines = 3,
  type = 'default',
  className,
}: CardSkeletonProps) {
  const renderContent = () => {
    switch (type) {
      case 'stats':
        return (
          <CardContent className="space-y-2">
            <Skeleton className="h-8 w-24" />
            <Skeleton size="sm" className="h-3 w-32" />
          </CardContent>
        );
        
      case 'content':
        return (
          <CardContent className="space-y-4">
            {Array.from({ length: lines }).map((_, index) => (
              <Skeleton 
                key={index} 
                variant="text"
                className={cn(
                  index === 0 ? 'w-full' : 'w-full',
                  index === lines - 1 ? 'w-3/4' : ''
                )}
              />
            ))}
          </CardContent>
        );
        
      default:
        return (
          <CardContent className="space-y-4">
            {Array.from({ length: lines }).map((_, index) => (
              <Skeleton 
                key={index} 
                variant="text"
                className={cn(
                  index === 0 ? 'w-full' : 'w-full',
                  index === lines - 1 ? 'w-2/3' : ''
                )}
              />
            ))}
          </CardContent>
        );
    }
  };

  return (
    <Card className={className}>
      {showHeader && (
        <CardHeader>
          <CardTitle>
            <Skeleton className="h-6 w-48" />
          </CardTitle>
          {showDescription && (
            <CardDescription>
              <Skeleton size="sm" className="h-4 w-32 mt-2" />
            </CardDescription>
          )}
        </CardHeader>
      )}
      {renderContent()}
    </Card>
  );
}


/**
 * Notification List States Components
 * 알림 목록 상태 컴포넌트 (로딩, 에러, 빈 상태)
 */

'use client';

import { Bell, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';

/**
 * 알림 목록 로딩 상태 컴포넌트
 */
export function NotificationListLoading() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>알림</CardTitle>
        <CardDescription>시스템 알림을 확인하세요.</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center space-x-4">
              <Skeleton className="h-4 w-4" />
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-4 w-48" />
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-4 w-16" />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

export interface NotificationListErrorProps {
  /** 재시도 핸들러 */
  onRetry: () => void;
}

/**
 * 알림 목록 에러 상태 컴포넌트
 */
export function NotificationListError({ onRetry }: NotificationListErrorProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>알림</CardTitle>
        <CardDescription>시스템 알림을 확인하세요.</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-center py-8">
          <div className="text-center">
            <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
            <p className="text-red-600 mb-4">알림을 불러올 수 없습니다.</p>
            <Button onClick={onRetry} variant="outline">
              다시 시도
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * 알림 목록 빈 상태 컴포넌트
 */
export function NotificationListEmpty() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>알림</CardTitle>
        <CardDescription>시스템 알림을 확인하세요.</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-center py-8">
          <div className="text-center">
            <Bell className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">알림이 없습니다.</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}


/**
 * Notification List Component
 * 알림 목록 컴포넌트
 * 
 * 리팩토링: 커스텀 훅과 작은 컴포넌트로 분리하여 가독성과 유지보수성 향상
 */

'use client';

import { useMemo } from 'react';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { CheckCheck, Trash2 } from 'lucide-react';
import { useNotifications } from '@/hooks/use-notifications';
import { useNotificationFilters } from '@/hooks/use-notification-filters';
import { useNotificationSelection } from '@/hooks/use-notification-selection';
import { useNotificationPagination } from '@/hooks/use-notification-pagination';
import { useNotificationActions } from '@/hooks/use-notification-actions';
import { NotificationFilters } from './notification-filters';
import { NotificationRow } from './notification-row';
import {
  NotificationListLoading,
  NotificationListError,
  NotificationListEmpty,
} from './notification-list-states';

export interface NotificationListProps {
  /** 페이지당 항목 수 */
  limit?: number;
  /** 액션 버튼 표시 여부 */
  showActions?: boolean;
  /** 필터 UI 표시 여부 */
  showFilters?: boolean;
}

/**
 * 알림 목록 컴포넌트
 */
function NotificationListComponent({ 
  limit = 20, 
  showActions = true, 
  showFilters = true 
}: NotificationListProps) {
  // 필터링 상태 관리
  const {
    unreadOnly,
    category,
    priority,
    setUnreadOnly,
    setCategory,
    setPriority,
  } = useNotificationFilters();

  // 페이지네이션 상태 관리 (초기 total은 0)
  const pagination = useNotificationPagination({
    limit,
    total: 0,
  });

  // 알림 데이터 조회 (offset은 페이지네이션 훅에서 관리)
  const { data, isLoading, error, refetch } = useNotifications(
    limit,
    pagination.offset,
    unreadOnly,
    category || undefined,
    priority || undefined
  );

  // 페이지네이션 total 업데이트 (data 로드 후)
  const updatedPagination = useNotificationPagination({
    limit,
    total: data?.total || 0,
    initialOffset: pagination.offset,
  });

  // 선택 상태 관리
  const {
    selectedNotifications,
    handleSelectNotification,
    handleSelectAll,
    clearSelection,
  } = useNotificationSelection({
    notifications: data?.notifications,
  });

  // 액션 핸들러
  const {
    handleMarkAsRead,
    handleMarkAllAsRead,
    handleDeleteNotification,
    handleDeleteSelected,
    isMarkingAllAsRead,
    isDeletingSelected,
    isMarkingAsRead,
    isDeleting,
  } = useNotificationActions({
    selectedNotifications,
    onSelectionClear: clearSelection,
  });

  // 전체 선택 여부 계산
  const allSelected = useMemo(() => {
    if (!data?.notifications || data.notifications.length === 0) {
      return false;
    }
    return selectedNotifications.length === data.notifications.length;
  }, [selectedNotifications.length, data?.notifications]);

  // 읽지 않은 알림 수 계산
  const unreadCount = useMemo(() => {
    return data?.notifications.filter(n => !n.is_read).length || 0;
  }, [data?.notifications]);

  // 로딩 상태
  if (isLoading) {
    return <NotificationListLoading />;
  }

  // 에러 상태
  if (error) {
    return <NotificationListError onRetry={() => refetch()} />;
  }

  // 빈 상태
  if (!data || data.notifications.length === 0) {
    return <NotificationListEmpty />;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>알림</CardTitle>
            <CardDescription>
              총 {data.total}개의 알림
              {unreadOnly && unreadCount > 0 && ` (읽지 않음: ${unreadCount}개)`}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {showActions && (
              <>
                <Button
                  onClick={handleMarkAllAsRead}
                  variant="outline"
                  size="sm"
                  disabled={isMarkingAllAsRead}
                >
                  <CheckCheck className="mr-2 h-4 w-4" />
                  모두 읽음
                </Button>
                {selectedNotifications.length > 0 && (
                  <Button
                    onClick={handleDeleteSelected}
                    variant="outline"
                    size="sm"
                    disabled={isDeletingSelected}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    선택 삭제
                  </Button>
                )}
              </>
            )}
          </div>
        </div>

        {showFilters && (
          <NotificationFilters
            unreadOnly={unreadOnly}
            category={category}
            priority={priority}
            onUnreadOnlyChange={setUnreadOnly}
            onCategoryChange={setCategory}
            onPriorityChange={setPriority}
          />
        )}
      </CardHeader>

      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              {showActions && (
                <TableHead className="w-12">
                  <Checkbox
                    checked={allSelected}
                    onCheckedChange={handleSelectAll}
                  />
                </TableHead>
              )}
              <TableHead>상태</TableHead>
              <TableHead>제목</TableHead>
              <TableHead>카테고리</TableHead>
              <TableHead>우선순위</TableHead>
              <TableHead>생성일</TableHead>
              {showActions && <TableHead className="text-right">작업</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.notifications.map((notification) => (
              <NotificationRow
                key={notification.id}
                notification={notification}
                isSelected={selectedNotifications.includes(notification.id)}
                onSelectChange={(checked) => handleSelectNotification(notification.id, checked)}
                onMarkAsRead={handleMarkAsRead}
                onDelete={handleDeleteNotification}
                showActions={showActions}
                isMarkingAsRead={isMarkingAsRead}
                isDeleting={isDeleting}
              />
            ))}
          </TableBody>
        </Table>

        {/* 페이지네이션 */}
        {data.total > limit && (
          <div className="flex items-center justify-between pt-4">
            <div className="text-sm text-gray-600">
              {updatedPagination.startIndex}-{updatedPagination.endIndex} / {data.total}
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={updatedPagination.goToPreviousPage}
                disabled={!updatedPagination.canGoPrevious}
              >
                이전
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={updatedPagination.goToNextPage}
                disabled={!updatedPagination.canGoNext}
              >
                다음
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export const NotificationList = NotificationListComponent;

/**
 * Notification Row Component
 * 알림 테이블 행 컴포넌트
 */

'use client';

import { format } from 'date-fns';
import { ko } from 'date-fns/locale';
import { Check, MoreHorizontal, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import {
  TableCell,
  TableRow,
} from '@/components/ui/table';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { getNotificationTypeIcon, getNotificationTypeColor, getNotificationPriorityColor } from '@/lib/notification';
import type { Notification } from '@/lib/types/notification';

export interface NotificationRowProps {
  /** 알림 데이터 */
  notification: Notification;
  /** 선택 여부 */
  isSelected: boolean;
  /** 선택 변경 핸들러 */
  onSelectChange: (checked: boolean) => void;
  /** 읽음 처리 핸들러 */
  onMarkAsRead: (notificationId: string) => void;
  /** 삭제 핸들러 */
  onDelete: (notificationId: string) => void;
  /** 액션 표시 여부 */
  showActions?: boolean;
  /** 읽음 처리 중 여부 */
  isMarkingAsRead?: boolean;
  /** 삭제 중 여부 */
  isDeleting?: boolean;
}

/**
 * 알림 테이블 행 컴포넌트
 */
export function NotificationRow({
  notification,
  isSelected,
  onSelectChange,
  onMarkAsRead,
  onDelete,
  showActions = true,
  isMarkingAsRead = false,
  isDeleting = false,
}: NotificationRowProps) {
  return (
    <TableRow key={notification.id}>
      {showActions && (
        <TableCell>
          <Checkbox
            checked={isSelected}
            onCheckedChange={onSelectChange}
          />
        </TableCell>
      )}
      <TableCell>
        <div className="flex items-center gap-2">
          {getNotificationTypeIcon(notification.type)}
          <Badge
            variant="outline"
            className={getNotificationTypeColor(notification.type)}
          >
            {notification.type}
          </Badge>
          {!notification.is_read && (
            <div className="w-2 h-2 bg-blue-500 rounded-full" />
          )}
        </div>
      </TableCell>
      <TableCell>
        <div className="font-medium">{notification.title}</div>
        <div className="text-sm text-gray-600 truncate max-w-xs">
          {notification.message}
        </div>
      </TableCell>
      <TableCell>
        {notification.category && (
          <Badge variant="secondary">
            {notification.category}
          </Badge>
        )}
      </TableCell>
      <TableCell>
        {notification.priority && (
          <Badge
            variant="outline"
            className={getNotificationPriorityColor(notification.priority)}
          >
            {notification.priority}
          </Badge>
        )}
      </TableCell>
      <TableCell>
        {format(new Date(notification.created_at), 'MM-dd HH:mm', {
          locale: ko,
        })}
      </TableCell>
      {showActions && (
        <TableCell className="text-right">
          <div className="flex items-center justify-end gap-2">
            {!notification.is_read && (
              <Button
                size="sm"
                variant="outline"
                onClick={() => onMarkAsRead(notification.id)}
                disabled={isMarkingAsRead}
              >
                <Check className="h-4 w-4" />
              </Button>
            )}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button size="sm" variant="outline">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem
                  onClick={() => onDelete(notification.id)}
                  disabled={isDeleting}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  삭제
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </TableCell>
      )}
    </TableRow>
  );
}


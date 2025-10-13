/**
 * Notification List Component
 * 알림 목록 컴포넌트
 */

'use client';

import { useState } from 'react';
import { format } from 'date-fns';
import { ko } from 'date-fns/locale';
import {
  Bell,
  Check,
  CheckCheck,
  Trash2,
  Filter,
  Search,
  MoreHorizontal,
  AlertCircle,
  Info,
  AlertTriangle,
  CheckCircle,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
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
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Skeleton } from '@/components/ui/skeleton';
import {
  useNotifications,
  useMarkAsRead,
  useMarkAllAsRead,
  useDeleteNotification,
  useDeleteNotifications,
} from '@/hooks/useNotifications';
import { Notification } from '@/services/notification';

interface NotificationListProps {
  limit?: number;
  showActions?: boolean;
  showFilters?: boolean;
}

export function NotificationList({ 
  limit = 20, 
  showActions = true, 
  showFilters = true 
}: NotificationListProps) {
  const [offset, setOffset] = useState(0);
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [category, setCategory] = useState<string>('');
  const [priority, setPriority] = useState<string>('');
  const [selectedNotifications, setSelectedNotifications] = useState<string[]>([]);

  const { data, isLoading, error, refetch } = useNotifications(
    limit,
    offset,
    unreadOnly,
    category || undefined,
    priority || undefined
  );

  const markAsReadMutation = useMarkAsRead();
  const markAllAsReadMutation = useMarkAllAsRead();
  const deleteNotificationMutation = useDeleteNotification();
  const deleteNotificationsMutation = useDeleteNotifications();

  const handleSelectNotification = (notificationId: string, checked: boolean) => {
    if (checked) {
      setSelectedNotifications(prev => [...prev, notificationId]);
    } else {
      setSelectedNotifications(prev => prev.filter(id => id !== notificationId));
    }
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked && data) {
      setSelectedNotifications(data.notifications.map(n => n.id));
    } else {
      setSelectedNotifications([]);
    }
  };

  const handleMarkAsRead = async (notificationId: string) => {
    await markAsReadMutation.mutateAsync(notificationId);
  };

  const handleMarkAllAsRead = async () => {
    await markAllAsReadMutation.mutateAsync();
  };

  const handleDeleteNotification = async (notificationId: string) => {
    await deleteNotificationMutation.mutateAsync(notificationId);
  };

  const handleDeleteSelected = async () => {
    if (selectedNotifications.length > 0) {
      await deleteNotificationsMutation.mutateAsync(selectedNotifications);
      setSelectedNotifications([]);
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'success':
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case 'warning':
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />;
      case 'error':
        return <AlertCircle className="h-4 w-4 text-red-600" />;
      case 'info':
      default:
        return <Info className="h-4 w-4 text-blue-600" />;
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'success':
        return 'bg-green-100 text-green-800';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800';
      case 'error':
        return 'bg-red-100 text-red-800';
      case 'info':
      default:
        return 'bg-blue-100 text-blue-800';
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'urgent':
        return 'bg-red-100 text-red-800';
      case 'high':
        return 'bg-orange-100 text-orange-800';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800';
      case 'low':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (isLoading) {
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

  if (error) {
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
              <Button onClick={() => refetch()} variant="outline">
                다시 시도
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!data || data.notifications.length === 0) {
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

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>알림</CardTitle>
            <CardDescription>
              총 {data.total}개의 알림
              {unreadOnly && ` (읽지 않음: ${data.notifications.filter(n => !n.is_read).length}개)`}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {showActions && (
              <>
                <Button
                  onClick={handleMarkAllAsRead}
                  variant="outline"
                  size="sm"
                  disabled={markAllAsReadMutation.isPending}
                >
                  <CheckCheck className="mr-2 h-4 w-4" />
                  모두 읽음
                </Button>
                {selectedNotifications.length > 0 && (
                  <Button
                    onClick={handleDeleteSelected}
                    variant="outline"
                    size="sm"
                    disabled={deleteNotificationsMutation.isPending}
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
          <div className="flex items-center gap-4 pt-4">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="unread-only"
                checked={unreadOnly}
                onCheckedChange={(checked) => setUnreadOnly(checked as boolean)}
              />
              <label htmlFor="unread-only" className="text-sm font-medium">
                읽지 않음만
              </label>
            </div>
            
            <Select value={category} onValueChange={setCategory}>
              <SelectTrigger className="w-32">
                <SelectValue placeholder="카테고리" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">전체</SelectItem>
                <SelectItem value="system">시스템</SelectItem>
                <SelectItem value="vm">VM</SelectItem>
                <SelectItem value="cost">비용</SelectItem>
                <SelectItem value="security">보안</SelectItem>
              </SelectContent>
            </Select>

            <Select value={priority} onValueChange={setPriority}>
              <SelectTrigger className="w-32">
                <SelectValue placeholder="우선순위" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">전체</SelectItem>
                <SelectItem value="urgent">긴급</SelectItem>
                <SelectItem value="high">높음</SelectItem>
                <SelectItem value="medium">보통</SelectItem>
                <SelectItem value="low">낮음</SelectItem>
              </SelectContent>
            </Select>
          </div>
        )}
      </CardHeader>

      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              {showActions && (
                <TableHead className="w-12">
                  <Checkbox
                    checked={selectedNotifications.length === data.notifications.length && data.notifications.length > 0}
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
              <TableRow key={notification.id}>
                {showActions && (
                  <TableCell>
                    <Checkbox
                      checked={selectedNotifications.includes(notification.id)}
                      onCheckedChange={(checked) => 
                        handleSelectNotification(notification.id, checked as boolean)
                      }
                    />
                  </TableCell>
                )}
                <TableCell>
                  <div className="flex items-center gap-2">
                    {getTypeIcon(notification.type)}
                    <Badge
                      variant="outline"
                      className={getTypeColor(notification.type)}
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
                      className={getPriorityColor(notification.priority)}
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
                          onClick={() => handleMarkAsRead(notification.id)}
                          disabled={markAsReadMutation.isPending}
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
                            onClick={() => handleDeleteNotification(notification.id)}
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
            ))}
          </TableBody>
        </Table>

        {/* 페이지네이션 */}
        {data.total > limit && (
          <div className="flex items-center justify-between pt-4">
            <div className="text-sm text-gray-600">
              {offset + 1}-{Math.min(offset + limit, data.total)} / {data.total}
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setOffset(Math.max(0, offset - limit))}
                disabled={offset === 0}
              >
                이전
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setOffset(offset + limit)}
                disabled={offset + limit >= data.total}
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

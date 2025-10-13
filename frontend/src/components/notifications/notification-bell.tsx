'use client';

import { useState } from 'react';
import { Bell, X, AlertTriangle, AlertCircle, Info, Check, CheckCheck } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet';
import { useNotifications, useMarkAsRead, useMarkAllAsRead } from '@/hooks/useNotifications';
import { cn } from '@/lib/utils';

interface NotificationBellProps {
  className?: string;
}

export function NotificationBell({ className }: NotificationBellProps) {
  const [isOpen, setIsOpen] = useState(false);
  const { data: notifications = [], isLoading } = useNotifications(10, 0);
  const markAsRead = useMarkAsRead();
  const markAllAsRead = useMarkAllAsRead();

  const unreadCount = notifications.filter(n => !n.is_read).length;
  const sortedNotifications = [...notifications].sort((a, b) => {
    // 우선순위별 정렬 (urgent > high > medium > low)
    const priorityOrder = { urgent: 4, high: 3, medium: 2, low: 1 };
    const aPriority = priorityOrder[a.priority as keyof typeof priorityOrder] || 0;
    const bPriority = priorityOrder[b.priority as keyof typeof priorityOrder] || 0;
    
    if (aPriority !== bPriority) {
      return bPriority - aPriority;
    }
    
    // 날짜별 정렬 (최신순)
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
  });

  const handleMarkAsRead = (notificationId: string) => {
    markAsRead.mutate(notificationId);
  };

  const handleMarkAllAsRead = () => {
    markAllAsRead.mutate();
  };

  return (
    <Sheet open={isOpen} onOpenChange={setIsOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className={cn('relative', className)}>
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <Badge 
              variant="destructive" 
              className="absolute -top-1 -right-1 h-5 w-5 rounded-full p-0 flex items-center justify-center text-xs"
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </Badge>
          )}
        </Button>
      </SheetTrigger>
      <SheetContent className="w-96 sm:w-[400px]">
        <SheetHeader>
          <SheetTitle className="flex items-center justify-between">
            <span>Notifications</span>
            {unreadCount > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleMarkAllAsRead}
                disabled={markAllAsRead.isPending}
              >
                Mark all read
              </Button>
            )}
          </SheetTitle>
        </SheetHeader>
        
        <div className="mt-4 space-y-4">
          {isLoading ? (
            <div className="space-y-3">
              {[...Array(3)].map((_, i) => (
                <div key={i} className="animate-pulse">
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2" />
                  <div className="h-3 bg-gray-200 rounded w-1/2" />
                </div>
              ))}
            </div>
          ) : sortedNotifications.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <Bell className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>No notifications</p>
            </div>
          ) : (
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {sortedNotifications.map((notification) => (
                <div
                  key={notification.id}
                  className={cn(
                    'p-3 rounded-lg border cursor-pointer transition-colors',
                    notification.read 
                      ? 'bg-gray-50 hover:bg-gray-100' 
                      : 'bg-white hover:bg-gray-50 border-l-4 border-l-blue-500'
                  )}
                  onClick={() => !notification.read && handleMarkAsRead(notification.id)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-lg">
                          {notificationUtils.getTypeIcon(notification.type)}
                        </span>
                        <h4 className={cn(
                          'font-medium text-sm truncate',
                          notification.read ? 'text-gray-600' : 'text-gray-900'
                        )}>
                          {notification.title}
                        </h4>
                        {!notification.read && (
                          <div className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0" />
                        )}
                      </div>
                      <p className={cn(
                        'text-sm text-gray-600 line-clamp-2',
                        notification.read && 'text-gray-500'
                      )}>
                        {notification.message}
                      </p>
                      <div className="flex items-center gap-2 mt-2">
                        <Badge
                          variant="outline"
                          className={cn(
                            'text-xs',
                            notificationUtils.getPriorityColor(notification.priority)
                          )}
                        >
                          {notification.priority}
                        </Badge>
                        <span className="text-xs text-gray-500">
                          {notificationUtils.formatRelativeTime(notification.created_at)}
                        </span>
                      </div>
                    </div>
                    {!notification.read && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleMarkAsRead(notification.id);
                        }}
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}

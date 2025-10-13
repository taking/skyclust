'use client';

import { useEffect, useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useRealtimeMonitoring } from '@/hooks/useRealtimeMonitoring';
import { Bell, AlertTriangle, Info, AlertCircle, X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Notification {
  id: string;
  type: 'info' | 'warning' | 'error';
  message: string;
  timestamp: number;
  read: boolean;
}

interface Alert {
  id: string;
  level: 'low' | 'medium' | 'high' | 'critical';
  message: string;
  timestamp: number;
  read: boolean;
}

interface RealtimeNotificationsProps {
  className?: string;
  maxNotifications?: number;
}

export function RealtimeNotifications({ 
  className, 
  maxNotifications = 50 
}: RealtimeNotificationsProps) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);

  const {
    onSystemNotification,
    onSystemAlert,
  } = useRealtimeMonitoring();

  // 알림 핸들러
  const handleNotification = useCallback((data: { type: 'info' | 'warning' | 'error'; message: string; timestamp: number }) => {
    const notification: Notification = {
      id: `notification-${Date.now()}-${Math.random()}`,
      type: data.type,
      message: data.message,
      timestamp: data.timestamp,
      read: false,
    };

    setNotifications(prev => {
      const updated = [notification, ...prev].slice(0, maxNotifications);
      return updated;
    });
  }, [maxNotifications]);

  // 알림 핸들러
  const handleAlert = useCallback((data: { level: 'low' | 'medium' | 'high' | 'critical'; message: string; timestamp: number }) => {
    const alert: Alert = {
      id: `alert-${Date.now()}-${Math.random()}`,
      level: data.level,
      message: data.message,
      timestamp: data.timestamp,
      read: false,
    };

    setAlerts(prev => {
      const updated = [alert, ...prev].slice(0, maxNotifications);
      return updated;
    });
  }, [maxNotifications]);

  // 이벤트 리스너 등록
  useEffect(() => {
    onSystemNotification(handleNotification);
    onSystemAlert(handleAlert);
  }, [onSystemNotification, onSystemAlert, handleNotification, handleAlert]);

  // 읽지 않은 알림 수 계산
  useEffect(() => {
    const unreadNotifications = notifications.filter(n => !n.read).length;
    const unreadAlerts = alerts.filter(a => !a.read).length;
    setUnreadCount(unreadNotifications + unreadAlerts);
  }, [notifications, alerts]);

  // 알림 읽음 처리
  const markAsRead = (id: string, type: 'notification' | 'alert') => {
    if (type === 'notification') {
      setNotifications(prev => 
        prev.map(n => n.id === id ? { ...n, read: true } : n)
      );
    } else {
      setAlerts(prev => 
        prev.map(a => a.id === id ? { ...a, read: true } : a)
      );
    }
  };

  // 모든 알림 읽음 처리
  const markAllAsRead = () => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
    setAlerts(prev => prev.map(a => ({ ...a, read: true })));
  };

  // 알림 삭제
  const removeNotification = (id: string, type: 'notification' | 'alert') => {
    if (type === 'notification') {
      setNotifications(prev => prev.filter(n => n.id !== id));
    } else {
      setAlerts(prev => prev.filter(a => a.id !== id));
    }
  };

  const getNotificationIcon = (type: 'info' | 'warning' | 'error') => {
    switch (type) {
      case 'info':
        return <Info className="h-4 w-4" />;
      case 'warning':
        return <AlertTriangle className="h-4 w-4" />;
      case 'error':
        return <AlertCircle className="h-4 w-4" />;
    }
  };

  const getAlertBadgeVariant = (level: 'low' | 'medium' | 'high' | 'critical') => {
    switch (level) {
      case 'low':
        return 'outline';
      case 'medium':
        return 'secondary';
      case 'high':
        return 'default';
      case 'critical':
        return 'destructive';
    }
  };

  const getNotificationBadgeVariant = (type: 'info' | 'warning' | 'error') => {
    switch (type) {
      case 'info':
        return 'outline';
      case 'warning':
        return 'secondary';
      case 'error':
        return 'destructive';
    }
  };

  const formatTimestamp = (timestamp: number) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    
    if (diff < 60000) { // 1분 미만
      return 'Just now';
    } else if (diff < 3600000) { // 1시간 미만
      return `${Math.floor(diff / 60000)}m ago`;
    } else if (diff < 86400000) { // 1일 미만
      return `${Math.floor(diff / 3600000)}h ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  const allItems = [
    ...notifications.map(n => ({ ...n, itemType: 'notification' as const })),
    ...alerts.map(a => ({ ...a, itemType: 'alert' as const }))
  ].sort((a, b) => b.timestamp - a.timestamp);

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold flex items-center space-x-2">
            <Bell className="h-5 w-5" />
            <span>Real-time Notifications</span>
            {unreadCount > 0 && (
              <Badge variant="destructive" className="ml-2">
                {unreadCount}
              </Badge>
            )}
          </CardTitle>
          {unreadCount > 0 && (
            <Button 
              variant="outline" 
              size="sm"
              onClick={markAllAsRead}
            >
              Mark all as read
            </Button>
          )}
        </div>
      </CardHeader>
      
      <CardContent>
        <ScrollArea className="h-96">
          {allItems.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Bell className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">No notifications yet</p>
            </div>
          ) : (
            <div className="space-y-2">
              {allItems.map((item) => (
                <div
                  key={item.id}
                  className={cn(
                    'flex items-start space-x-3 p-3 rounded-lg border transition-colors',
                    !item.read && 'bg-muted/50 border-primary/20'
                  )}
                >
                  <div className="flex-shrink-0 mt-0.5">
                    {item.itemType === 'notification' ? (
                      getNotificationIcon(item.type)
                    ) : (
                      <AlertTriangle className="h-4 w-4" />
                    )}
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2 mb-1">
                      {item.itemType === 'notification' ? (
                        <Badge 
                          variant={getNotificationBadgeVariant(item.type)}
                          className="text-xs"
                        >
                          {item.type}
                        </Badge>
                      ) : (
                        <Badge 
                          variant={getAlertBadgeVariant(item.level)}
                          className="text-xs"
                        >
                          {item.level}
                        </Badge>
                      )}
                      <span className="text-xs text-muted-foreground">
                        {formatTimestamp(item.timestamp)}
                      </span>
                    </div>
                    <p className="text-sm">{item.message}</p>
                  </div>
                  
                  <div className="flex items-center space-x-1">
                    {!item.read && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => markAsRead(item.id, item.itemType)}
                        className="h-6 w-6 p-0"
                      >
                        <div className="h-2 w-2 rounded-full bg-primary" />
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => removeNotification(item.id, item.itemType)}
                      className="h-6 w-6 p-0"
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}

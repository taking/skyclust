/**
 * Notification Service
 * 알림 관련 API 호출
 */

import { api } from '@/lib/api';

export interface Notification {
  id: string;
  user_id: string;
  type: 'info' | 'warning' | 'error' | 'success';
  title: string;
  message: string;
  category?: string;
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  is_read: boolean;
  data?: string;
  created_at: string;
  read_at?: string;
}

export interface NotificationPreferences {
  id: string;
  user_id: string;
  email_enabled: boolean;
  push_enabled: boolean;
  browser_enabled: boolean;
  in_app_enabled: boolean;
  system_notifications: boolean;
  vm_notifications: boolean;
  cost_notifications: boolean;
  security_notifications: boolean;
  low_priority_enabled: boolean;
  medium_priority_enabled: boolean;
  high_priority_enabled: boolean;
  urgent_priority_enabled: boolean;
  quiet_hours_start?: string;
  quiet_hours_end?: string;
  timezone: string;
  created_at: string;
  updated_at: string;
}

export interface NotificationStats {
  total_notifications: number;
  unread_notifications: number;
  read_notifications: number;
  system_count: number;
  vm_count: number;
  cost_count: number;
  security_count: number;
  low_priority_count: number;
  medium_priority_count: number;
  high_priority_count: number;
  urgent_priority_count: number;
  last_7_days_count: number;
  last_30_days_count: number;
}

export interface NotificationListResponse {
  notifications: Notification[];
  total: number;
  limit: number;
  offset: number;
}

export const notificationService = {
  // 알림 목록 조회
  async getNotifications(
    limit: number = 20,
    offset: number = 0,
    unreadOnly: boolean = false,
    category?: string,
    priority?: string
  ): Promise<NotificationListResponse> {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
      unread_only: unreadOnly.toString(),
    });
    
    if (category) params.append('category', category);
    if (priority) params.append('priority', priority);

    const response = await api.get(`/notifications?${params}`);
    return response.data.data;
  },

  // 알림 상세 조회
  async getNotification(notificationId: string): Promise<Notification> {
    const response = await api.get(`/notifications/${notificationId}`);
    return response.data.data.notification;
  },

  // 알림 읽음 처리
  async markAsRead(notificationId: string): Promise<void> {
    await api.put(`/notifications/${notificationId}/read`);
  },

  // 모든 알림 읽음 처리
  async markAllAsRead(): Promise<void> {
    await api.put('/notifications/read');
  },

  // 알림 삭제
  async deleteNotification(notificationId: string): Promise<void> {
    await api.delete(`/notifications/${notificationId}`);
  },

  // 여러 알림 삭제
  async deleteNotifications(notificationIds: string[]): Promise<void> {
    await api.delete('/notifications', {
      data: { ids: notificationIds }
    });
  },

  // 알림 설정 조회
  async getNotificationPreferences(): Promise<NotificationPreferences> {
    const response = await api.get('/notifications/preferences');
    return response.data.data.preferences;
  },

  // 알림 설정 업데이트
  async updateNotificationPreferences(preferences: Partial<NotificationPreferences>): Promise<void> {
    await api.put('/notifications/preferences', preferences);
  },

  // 알림 통계 조회
  async getNotificationStats(): Promise<NotificationStats> {
    const response = await api.get('/notifications/stats');
    return response.data.data.stats;
  },

  // 테스트 알림 전송
  async sendTestNotification(data: {
    type: string;
    title: string;
    message: string;
    category?: string;
    priority?: string;
  }): Promise<Notification> {
    const response = await api.post('/notifications/test', data);
    return response.data.data.notification;
  },
};
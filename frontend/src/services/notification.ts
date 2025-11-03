/**
 * Notification Service
 * 알림 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import type {
  Notification,
  NotificationPreferences,
  NotificationStats,
  NotificationListResponse,
} from '@/lib/types/notification';

class NotificationService extends BaseService {
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

    return this.get<NotificationListResponse>(`/notifications?${params}`);
  }

  // 알림 상세 조회
  async getNotification(notificationId: string): Promise<Notification> {
    const data = await this.get<{ notification: Notification }>(`/notifications/${notificationId}`);
    return data.notification;
  }

  // 알림 읽음 처리
  async markAsRead(notificationId: string): Promise<void> {
    return this.put<void>(`/notifications/${notificationId}/read`);
  }

  // 모든 알림 읽음 처리
  async markAllAsRead(): Promise<void> {
    return this.put<void>('/notifications/read');
  }

  // 알림 삭제
  async deleteNotification(notificationId: string): Promise<void> {
    return this.delete<void>(`/notifications/${notificationId}`);
  }

  // 여러 알림 삭제
  async deleteNotifications(notificationIds: string[]): Promise<void> {
    // DELETE with body는 BaseService를 직접 사용
    return this.request<void>('delete', '/notifications', { ids: notificationIds });
  }

  // 알림 설정 조회
  async getNotificationPreferences(): Promise<NotificationPreferences> {
    const data = await this.get<{ preferences: NotificationPreferences }>('/notifications/preferences');
    return data.preferences;
  }

  // 알림 설정 업데이트
  async updateNotificationPreferences(preferences: Partial<NotificationPreferences>): Promise<void> {
    return this.put<void>('/notifications/preferences', preferences);
  }

  // 알림 통계 조회
  async getNotificationStats(): Promise<NotificationStats> {
    const data = await this.get<{ stats: NotificationStats }>('/notifications/stats');
    return data.stats;
  }

  // 테스트 알림 전송
  async sendTestNotification(data: {
    type: string;
    title: string;
    message: string;
    category?: string;
    priority?: string;
  }): Promise<Notification> {
    const result = await this.post<{ notification: Notification }>('/notifications/test', data);
    return result.notification;
  }
}

export const notificationService = new NotificationService();
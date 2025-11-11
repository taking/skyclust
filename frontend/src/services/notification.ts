/**
 * Notification Service
 * 알림 관련 API 호출
 */

import { BaseService } from '@/lib/api';
import { API_ENDPOINTS } from '@/lib/api';
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
    return this.get<NotificationListResponse>(API_ENDPOINTS.notifications.list(limit, offset, unreadOnly, category, priority));
  }

  // 알림 상세 조회
  async getNotification(notificationId: string): Promise<Notification> {
    const data = await this.get<{ notification: Notification }>(API_ENDPOINTS.notifications.detail(notificationId));
    return data.notification;
  }

  // 알림 읽음 처리
  async markAsRead(notificationId: string): Promise<void> {
    return this.put<void>(API_ENDPOINTS.notifications.markAsRead(notificationId));
  }

  // 모든 알림 읽음 처리
  async markAllAsRead(): Promise<void> {
    return this.put<void>(API_ENDPOINTS.notifications.markAllAsRead());
  }

  // 알림 삭제
  async deleteNotification(notificationId: string): Promise<void> {
    return this.delete<void>(API_ENDPOINTS.notifications.delete(notificationId));
  }

  // 여러 알림 삭제
  async deleteNotifications(notificationIds: string[]): Promise<void> {
    // DELETE with body는 BaseService를 직접 사용
    return this.request<void>('delete', API_ENDPOINTS.notifications.deleteMultiple(), { ids: notificationIds });
  }

  // 알림 설정 조회
  async getNotificationPreferences(): Promise<NotificationPreferences> {
    const data = await this.get<{ preferences: NotificationPreferences }>(API_ENDPOINTS.notifications.preferences());
    return data.preferences;
  }

  // 알림 설정 업데이트
  async updateNotificationPreferences(preferences: Partial<NotificationPreferences>): Promise<void> {
    return this.put<void>(API_ENDPOINTS.notifications.updatePreferences(), preferences);
  }

  // 알림 통계 조회
  async getNotificationStats(): Promise<NotificationStats> {
    const data = await this.get<{ stats: NotificationStats }>(API_ENDPOINTS.notifications.stats());
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
    const result = await this.post<{ notification: Notification }>(API_ENDPOINTS.notifications.test(), data);
    return result.notification;
  }
}

export const notificationService = new NotificationService();
/**
 * Notification 관련 타입 정의
 */

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


/**
 * Notification Filters Component
 * 알림 필터 UI 컴포넌트
 */

'use client';

import { Checkbox } from '@/components/ui/checkbox';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export interface NotificationFiltersProps {
  /** 읽지 않은 알림만 표시 여부 */
  unreadOnly: boolean;
  /** 선택된 카테고리 */
  category: string;
  /** 선택된 우선순위 */
  priority: string;
  /** 읽지 않은 알림만 표시 여부 변경 핸들러 */
  onUnreadOnlyChange: (value: boolean) => void;
  /** 카테고리 변경 핸들러 */
  onCategoryChange: (value: string) => void;
  /** 우선순위 변경 핸들러 */
  onPriorityChange: (value: string) => void;
}

/**
 * 알림 필터 UI 컴포넌트
 */
export function NotificationFilters({
  unreadOnly,
  category,
  priority,
  onUnreadOnlyChange,
  onCategoryChange,
  onPriorityChange,
}: NotificationFiltersProps) {
  return (
    <div className="flex items-center gap-4 pt-4">
      <div className="flex items-center space-x-2">
        <Checkbox
          id="unread-only"
          checked={unreadOnly}
          onCheckedChange={(checked) => onUnreadOnlyChange(checked as boolean)}
        />
        <label htmlFor="unread-only" className="text-sm font-medium">
          읽지 않음만
        </label>
      </div>

      <Select value={category} onValueChange={onCategoryChange}>
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

      <Select value={priority} onValueChange={onPriorityChange}>
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
  );
}


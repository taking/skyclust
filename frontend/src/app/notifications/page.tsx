/**
 * Notifications Page
 * 알림 관리 페이지
 */

'use client';

import { useState } from 'react';
import { Bell, Settings, BarChart3, TrendingUp, Mail, Smartphone, Monitor } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { NotificationList } from '@/components/notifications/NotificationList';
import { NotificationPreferencesComponent } from '@/components/notifications/NotificationPreferences';
import { useRequireAuth } from '@/hooks/useAuth';
import { useNotificationStats } from '@/hooks/useNotifications';

export default function NotificationsPage() {
  // 인증 확인
  useRequireAuth();

  const { data: stats, isLoading: statsLoading } = useNotificationStats();

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* 헤더 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">알림 관리</h1>
          <p className="text-gray-600 mt-2">
            시스템 알림을 확인하고 설정을 관리하세요.
          </p>
        </div>
      </div>

      {/* 통계 카드 */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <Bell className="h-8 w-8 text-blue-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">총 알림</p>
                  <p className="text-2xl font-bold">{stats.total_notifications}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <TrendingUp className="h-8 w-8 text-orange-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">읽지 않음</p>
                  <p className="text-2xl font-bold text-orange-600">{stats.unread_notifications}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <Mail className="h-8 w-8 text-green-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">읽음</p>
                  <p className="text-2xl font-bold text-green-600">{stats.read_notifications}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center">
                <Monitor className="h-8 w-8 text-purple-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">최근 7일</p>
                  <p className="text-2xl font-bold text-purple-600">{stats.last_7_days_count}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 메인 컨텐츠 */}
      <Tabs defaultValue="list" className="space-y-6">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="list">알림 목록</TabsTrigger>
          <TabsTrigger value="preferences">알림 설정</TabsTrigger>
          <TabsTrigger value="stats">상세 통계</TabsTrigger>
        </TabsList>

        {/* 알림 목록 */}
        <TabsContent value="list">
          <NotificationList limit={20} showActions={true} showFilters={true} />
        </TabsContent>

        {/* 알림 설정 */}
        <TabsContent value="preferences">
          <NotificationPreferencesComponent />
        </TabsContent>

        {/* 상세 통계 */}
        <TabsContent value="stats">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* 카테고리별 통계 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BarChart3 className="h-5 w-5" />
                  카테고리별 통계
                </CardTitle>
                <CardDescription>
                  알림 유형별 수신 현황
                </CardDescription>
              </CardHeader>
              <CardContent>
                {statsLoading ? (
                  <div className="space-y-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                      <div key={i} className="flex items-center justify-between">
                        <div className="h-4 w-24 bg-gray-200 rounded animate-pulse" />
                        <div className="h-4 w-8 bg-gray-200 rounded animate-pulse" />
                      </div>
                    ))}
                  </div>
                ) : stats ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">시스템</span>
                      <span className="text-sm text-gray-600">{stats.system_count}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">VM</span>
                      <span className="text-sm text-gray-600">{stats.vm_count}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">비용</span>
                      <span className="text-sm text-gray-600">{stats.cost_count}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">보안</span>
                      <span className="text-sm text-gray-600">{stats.security_count}</span>
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <BarChart3 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                    <p className="text-gray-600">통계를 불러올 수 없습니다.</p>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* 우선순위별 통계 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <TrendingUp className="h-5 w-5" />
                  우선순위별 통계
                </CardTitle>
                <CardDescription>
                  알림 우선순위별 수신 현황
                </CardDescription>
              </CardHeader>
              <CardContent>
                {statsLoading ? (
                  <div className="space-y-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                      <div key={i} className="flex items-center justify-between">
                        <div className="h-4 w-24 bg-gray-200 rounded animate-pulse" />
                        <div className="h-4 w-8 bg-gray-200 rounded animate-pulse" />
                      </div>
                    ))}
                  </div>
                ) : stats ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-red-600">긴급</span>
                      <span className="text-sm text-gray-600">{stats.urgent_priority_count}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-orange-600">높음</span>
                      <span className="text-sm text-gray-600">{stats.high_priority_count}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-yellow-600">보통</span>
                      <span className="text-sm text-gray-600">{stats.medium_priority_count}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-green-600">낮음</span>
                      <span className="text-sm text-gray-600">{stats.low_priority_count}</span>
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <TrendingUp className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                    <p className="text-gray-600">통계를 불러올 수 없습니다.</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
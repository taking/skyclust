/**
 * Profile Overview Card Component
 * 프로필 페이지의 개요 카드 컴포넌트
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { User } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { toLocaleDateString } from '@/lib/utils/date-format';
import type { User as UserType } from '@/lib/types';

interface ProfileOverviewCardProps {
  user: UserType | null | undefined;
}

export function ProfileOverviewCard({ user }: ProfileOverviewCardProps) {
  const { t, locale } = useTranslation();

  if (!user) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('auth.profileOverview')}</CardTitle>
        <CardDescription>{t('auth.profileOverviewDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex items-center space-x-4">
            <div className="h-16 w-16 rounded-full bg-gray-200 flex items-center justify-center">
              <User className="h-8 w-8 text-gray-600" />
            </div>
            <div className="flex-1">
              <h3 className="text-lg font-semibold">{user.username || user.email}</h3>
              <p className="text-sm text-gray-600">{user.email}</p>
            </div>
            <Badge variant={user.is_active ? 'default' : 'secondary'}>
              {user.is_active ? t('auth.active') : t('auth.inactive')}
            </Badge>
          </div>
          {user.created_at && (
            <div className="pt-4 border-t">
              <p className="text-sm text-gray-600">
                {t('auth.memberSince', { date: toLocaleDateString(user.created_at, locale as 'ko' | 'en') })}
              </p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}


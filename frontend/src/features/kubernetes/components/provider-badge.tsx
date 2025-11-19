/**
 * Provider Badge Component
 * Cloud Provider 배지 컴포넌트
 * 
 * 기능:
 * - Provider별 색상 및 아이콘
 * - 일관된 스타일
 */

'use client';

import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { CloudProvider } from '@/lib/types/kubernetes';

interface ProviderBadgeProps {
  provider: CloudProvider;
  variant?: 'default' | 'outline' | 'secondary';
  className?: string;
}

const providerConfig: Record<CloudProvider, { label: string; className: string }> = {
  aws: {
    label: 'AWS',
    className: 'bg-orange-100 text-orange-800 border-orange-200',
  },
  gcp: {
    label: 'GCP',
    className: 'bg-blue-100 text-blue-800 border-blue-200',
  },
  azure: {
    label: 'Azure',
    className: 'bg-sky-100 text-sky-800 border-sky-200',
  },
};

export function ProviderBadge({ provider, variant = 'default', className }: ProviderBadgeProps) {
  const config = providerConfig[provider];
  
  if (variant === 'outline') {
    return (
      <Badge
        variant="outline"
        className={cn(config.className, className)}
      >
        {config.label}
      </Badge>
    );
  }
  
  return (
    <Badge
      variant="secondary"
      className={cn(config.className, className)}
    >
      {config.label}
    </Badge>
  );
}


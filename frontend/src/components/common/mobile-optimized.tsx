/**
 * Mobile Optimized Components
 * 모바일 최적화 컴포넌트
 */

'use client';

import { ReactNode } from 'react';
import { useResponsive } from '@/hooks/use-responsive';
import { cn } from '@/lib/utils';

/**
 * MobileCard Component
 * 모바일에서 최적화된 카드 컴포넌트
 */
export interface MobileCardProps {
  children: ReactNode;
  className?: string;
  onClick?: () => void;
  compact?: boolean; // 모바일에서 더 작은 패딩 사용
}

export function MobileCard({ 
  children, 
  className,
  onClick,
  compact = false,
}: MobileCardProps) {
  const { isMobile } = useResponsive();
  
  return (
    <div
      onClick={onClick}
      className={cn(
        'bg-card border rounded-lg shadow-sm',
        isMobile && compact ? 'p-3' : 'p-4',
        !isMobile && 'p-6',
        onClick && 'cursor-pointer hover:shadow-md transition-shadow',
        className
      )}
    >
      {children}
    </div>
  );
}

/**
 * MobileButton Component
 * 모바일에서 터치하기 쉬운 버튼 컴포넌트
 */
export interface MobileButtonProps {
  children: ReactNode;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'outline';
  className?: string;
  fullWidth?: boolean; // 모바일에서 전체 너비
  size?: 'sm' | 'md' | 'lg';
}

export function MobileButton({
  children,
  onClick,
  variant = 'primary',
  className,
  fullWidth = true,
  size = 'md',
}: MobileButtonProps) {
  const { isMobile } = useResponsive();
  
  const variantClasses = {
    primary: 'bg-primary text-primary-foreground hover:bg-primary/90',
    secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
    outline: 'border border-input bg-background hover:bg-accent',
  };

  const sizeClasses = {
    sm: isMobile ? 'h-9 px-4 text-sm' : 'h-8 px-3 text-sm',
    md: isMobile ? 'h-11 px-6 text-base' : 'h-10 px-4 text-sm',
    lg: isMobile ? 'h-12 px-8 text-lg' : 'h-11 px-6 text-base',
  };

  return (
    <button
      onClick={onClick}
      className={cn(
        'rounded-md font-medium transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
        'disabled:opacity-50 disabled:pointer-events-none',
        variantClasses[variant],
        sizeClasses[size],
        isMobile && fullWidth && 'w-full',
        className
      )}
    >
      {children}
    </button>
  );
}

/**
 * MobileTable Component
 * 모바일에서 카드 형태로 표시되는 테이블
 */
export interface MobileTableProps {
  data: Array<Record<string, ReactNode>>;
  columns: Array<{ key: string; label: string; mobileHide?: boolean }>;
  onRowClick?: (row: Record<string, ReactNode>, index: number) => void;
  className?: string;
}

export function MobileTable({
  data,
  columns,
  onRowClick,
  className,
}: MobileTableProps) {
  const { isMobile } = useResponsive();

  if (isMobile) {
    // 모바일: 카드 형태로 표시
    const visibleColumns = columns.filter(col => !col.mobileHide);
    
    return (
      <div className={cn('space-y-3', className)}>
        {data.map((row, rowIndex) => (
          <MobileCard
            key={rowIndex}
            onClick={() => onRowClick?.(row, rowIndex)}
            compact
          >
            <div className="space-y-2">
              {visibleColumns.map((column) => (
                <div key={column.key} className="flex justify-between">
                  <span className="text-sm font-medium text-muted-foreground">
                    {column.label}
                  </span>
                  <span className="text-sm text-foreground">
                    {row[column.key]}
                  </span>
                </div>
              ))}
            </div>
          </MobileCard>
        ))}
      </div>
    );
  }

  // 데스크톱: 일반 테이블
  return (
    <div className={cn('overflow-x-auto', className)}>
      <table className="w-full border-collapse">
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.key}
                className="px-4 py-2 text-left text-sm font-medium text-muted-foreground border-b"
              >
                {column.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, rowIndex) => (
            <tr
              key={rowIndex}
              onClick={() => onRowClick?.(row, rowIndex)}
              className={cn(
                'border-b hover:bg-muted/50',
                onRowClick && 'cursor-pointer'
              )}
            >
              {columns.map((column) => (
                <td key={column.key} className="px-4 py-3 text-sm">
                  {row[column.key]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/**
 * MobileDrawer Component
 * 모바일에서 하단에서 올라오는 드로어
 */
export interface MobileDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
  title?: string;
}

export function MobileDrawer({
  open,
  onOpenChange,
  children,
  title,
}: MobileDrawerProps) {
  const { isMobile } = useResponsive();

  if (!isMobile) {
    // 데스크톱에서는 일반 모달로 표시
    if (!open) return null;
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
        <div className="bg-card rounded-lg shadow-lg max-w-md w-full max-h-[80vh] overflow-y-auto">
          {title && (
            <div className="px-6 py-4 border-b">
              <h3 className="text-lg font-semibold">{title}</h3>
            </div>
          )}
          <div className="p-6">{children}</div>
        </div>
      </div>
    );
  }

  // 모바일: 하단 드로어
  return (
    <div
      className={cn(
        'fixed inset-0 z-50 transition-opacity',
        open ? 'opacity-100' : 'opacity-0 pointer-events-none'
      )}
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50"
        onClick={() => onOpenChange(false)}
      />
      
      {/* Drawer */}
      <div
        className={cn(
          'absolute bottom-0 left-0 right-0 bg-card rounded-t-xl shadow-lg',
          'transition-transform duration-300 ease-out',
          'max-h-[90vh] overflow-y-auto',
          open ? 'translate-y-0' : 'translate-y-full'
        )}
      >
        {/* Handle */}
        <div className="flex justify-center pt-3 pb-2">
          <div className="w-12 h-1 bg-muted rounded-full" />
        </div>
        
        {title && (
          <div className="px-4 py-3 border-b">
            <h3 className="text-lg font-semibold">{title}</h3>
          </div>
        )}
        
        <div className="p-4">{children}</div>
      </div>
    </div>
  );
}


/**
 * Copyable Text Component
 * 텍스트를 복사할 수 있는 컴포넌트
 */

'use client';

import { useState, useCallback } from 'react';
import { Copy, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useToast } from '@/hooks/use-toast';
import { cn } from '@/lib/utils';
import { log } from '@/lib/logging';
import { TIMEOUTS } from '@/lib/constants/values';

interface CopyableTextProps {
  text: string;
  maxLength?: number;
  showCopyButton?: boolean;
  className?: string;
  copyButtonVariant?: 'default' | 'ghost' | 'outline' | 'link';
  copyButtonSize?: 'default' | 'sm' | 'lg' | 'icon';
  onCopy?: (text: string) => void;
}

export function CopyableText({
  text,
  maxLength = 50,
  showCopyButton = true,
  className,
  copyButtonVariant = 'ghost',
  copyButtonSize = 'sm',
  onCopy,
}: CopyableTextProps) {
  const { success } = useToast();
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      success('Copied to clipboard');
      onCopy?.(text);
      
      // Reset copied state after timeout
      setTimeout(() => {
        setCopied(false);
      }, TIMEOUTS.COPY_SUCCESS_DURATION);
    } catch (error) {
      log.error('Failed to copy text', error, {
        component: 'CopyableText',
        action: 'copy',
      });
    }
  }, [text, success, onCopy]);

  const displayText = text.length > maxLength 
    ? `${text.substring(0, maxLength)}...`
    : text;

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <span 
        className="truncate flex-1"
        title={text}
      >
        {displayText}
      </span>
      {showCopyButton && (
        <Button
          variant={copyButtonVariant}
          size={copyButtonSize}
          onClick={handleCopy}
          className="shrink-0"
          aria-label="Copy to clipboard"
        >
          {copied ? (
            <Check className="h-4 w-4 text-green-600" />
          ) : (
            <Copy className="h-4 w-4" />
          )}
        </Button>
      )}
    </div>
  );
}


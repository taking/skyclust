'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';

interface SkipLinkProps {
  href: string;
  children: React.ReactNode;
  className?: string;
}

export function SkipLink({ href, children, className }: SkipLinkProps) {
  const router = useRouter();

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault();
    router.push(href);
    
    // Focus the main content after navigation
    setTimeout(() => {
      const main = document.querySelector('main');
      if (main) {
        main.focus();
        main.scrollIntoView();
      }
    }, 100);
  };

  return (
    <Button
      variant="outline"
      onClick={handleClick}
      className={`sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 ${className || ''}`}
    >
      {children}
    </Button>
  );
}

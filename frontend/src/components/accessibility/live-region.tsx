'use client';

import { useEffect, useRef } from 'react';

interface LiveRegionProps {
  message: string;
  politeness?: 'polite' | 'assertive' | 'off';
  className?: string;
}

export function LiveRegion({ message, politeness = 'polite', className }: LiveRegionProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current && message) {
      ref.current.textContent = message;
    }
  }, [message]);

  return (
    <div
      ref={ref}
      aria-live={politeness}
      aria-atomic="true"
      className={`sr-only ${className || ''}`}
    />
  );
}

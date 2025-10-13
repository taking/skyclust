'use client';

// import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Search, X, Filter } from 'lucide-react';
import { cn } from '@/lib/utils';

interface SearchBarProps {
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  onClear?: () => void;
  className?: string;
  showFilter?: boolean;
  onFilterClick?: () => void;
  filterCount?: number;
}

export function SearchBar({
  placeholder = 'Search...',
  value,
  onChange,
  onClear,
  className,
  showFilter = false,
  onFilterClick,
  filterCount = 0,
}: SearchBarProps) {
  // const [isFocused, setIsFocused] = useState(false);

  return (
    <div className={cn('relative', className)}>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={placeholder}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          // onFocus={() => setIsFocused(true)}
          // onBlur={() => setIsFocused(false)}
          className="pl-10 pr-20"
        />
        <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center space-x-1">
          {value && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onClear}
              className="h-6 w-6 p-0 hover:bg-muted"
            >
              <X className="h-3 w-3" />
            </Button>
          )}
          {showFilter && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onFilterClick}
              className="h-6 w-6 p-0 hover:bg-muted"
            >
              <Filter className="h-3 w-3" />
              {filterCount > 0 && (
                <Badge
                  variant="destructive"
                  className="absolute -top-1 -right-1 h-4 w-4 p-0 text-xs flex items-center justify-center"
                >
                  {filterCount}
                </Badge>
              )}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}

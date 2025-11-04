'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { X, RotateCcw, Filter } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface FilterOption {
  id: string;
  label: string;
  value: string;
}

export interface FilterConfig {
  id: string;
  label: string;
  type: 'select' | 'multiselect' | 'checkbox' | 'text' | 'date' | 'range';
  options?: FilterOption[];
  placeholder?: string;
  multiple?: boolean;
}

export interface FilterValue {
  [key: string]: string | string[] | boolean | { min: number; max: number };
}

interface FilterPanelProps {
  filters: FilterConfig[];
  values: FilterValue;
  onChange: (values: FilterValue) => void;
  onClear: () => void;
  onApply: () => void;
  className?: string;
  title?: string;
  description?: string;
}

export function FilterPanel({
  filters,
  values,
  onChange,
  onClear,
  onApply,
  className,
  title = 'Filters',
  description = 'Filter your data',
}: FilterPanelProps) {
  const [isOpen, setIsOpen] = useState(false);

  const handleFilterChange = (filterId: string, value: string | string[] | boolean | { min: number; max: number }) => {
    onChange({
      ...values,
      [filterId]: value,
    });
  };

  const getActiveFilterCount = () => {
    return Object.values(values).filter(value => {
      if (Array.isArray(value)) {
        return value.length > 0;
      }
      if (typeof value === 'boolean') {
        return value;
      }
      if (typeof value === 'object' && value !== null) {
        return Object.values(value).some(v => {
          if (typeof v === 'string') return v !== '';
          if (typeof v === 'number') return v !== 0;
          return true;
        });
      }
      return value !== '' && value !== null && value !== undefined;
    }).length;
  };

  const renderFilter = (filter: FilterConfig) => {
    const value = values[filter.id];

    switch (filter.type) {
      case 'select':
        return (
          <div key={filter.id} className="space-y-2">
            <Label htmlFor={filter.id}>{filter.label}</Label>
            <Select
              value={value as string || ''}
              onValueChange={(val) => handleFilterChange(filter.id, val)}
            >
              <SelectTrigger>
                <SelectValue placeholder={filter.placeholder || `Select ${filter.label}`} />
              </SelectTrigger>
              <SelectContent>
                {filter.options?.map((option) => (
                  <SelectItem key={option.id} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        );

      case 'multiselect':
        return (
          <div key={filter.id} className="space-y-2">
            <Label>{filter.label}</Label>
            <div className="space-y-2">
              {filter.options?.map((option) => (
                <div key={option.id} className="flex items-center space-x-2">
                  <Checkbox
                    id={`${filter.id}-${option.id}`}
                    checked={(value as string[] || []).includes(option.value)}
                    onCheckedChange={(checked) => {
                      const currentValues = value as string[] || [];
                      if (checked) {
                        handleFilterChange(filter.id, [...currentValues, option.value]);
                      } else {
                        handleFilterChange(filter.id, currentValues.filter(v => v !== option.value));
                      }
                    }}
                  />
                  <Label
                    htmlFor={`${filter.id}-${option.id}`}
                    className="text-sm font-normal"
                  >
                    {option.label}
                  </Label>
                </div>
              ))}
            </div>
          </div>
        );

      case 'checkbox':
        return (
          <div key={filter.id} className="flex items-center space-x-2">
            <Checkbox
              id={filter.id}
              checked={value as boolean || false}
              onCheckedChange={(checked) => handleFilterChange(filter.id, checked)}
            />
            <Label htmlFor={filter.id} className="text-sm font-normal">
              {filter.label}
            </Label>
          </div>
        );

      case 'text':
        return (
          <div key={filter.id} className="space-y-2">
            <Label htmlFor={filter.id}>{filter.label}</Label>
            <Input
              id={filter.id}
              placeholder={filter.placeholder || `Enter ${filter.label.toLowerCase()}`}
              value={value as string || ''}
              onChange={(e) => handleFilterChange(filter.id, e.target.value)}
            />
          </div>
        );

      case 'date':
        return (
          <div key={filter.id} className="space-y-2">
            <Label htmlFor={filter.id}>{filter.label}</Label>
            <Input
              id={filter.id}
              type="date"
              value={value as string || ''}
              onChange={(e) => handleFilterChange(filter.id, e.target.value)}
            />
          </div>
        );

      case 'range':
        return (
          <div key={filter.id} className="space-y-2">
            <Label>{filter.label}</Label>
            <div className="grid grid-cols-2 gap-2">
              <Input
                placeholder="Min"
                type="number"
                value={(value as { min: number; max: number })?.min || ''}
                onChange={(e) => handleFilterChange(filter.id, {
                  ...(value as { min: number; max: number }) || { min: 0, max: 0 },
                  min: parseInt(e.target.value) || 0,
                })}
              />
              <Input
                placeholder="Max"
                type="number"
                value={(value as { min: number; max: number })?.max || ''}
                onChange={(e) => handleFilterChange(filter.id, {
                  ...(value as { min: number; max: number }) || { min: 0, max: 0 },
                  max: parseInt(e.target.value) || 0,
                })}
              />
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  const activeFilterCount = getActiveFilterCount();

  return (
    <div className={cn('relative', className)}>
      <Button
        variant="outline"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full justify-between"
      >
        <div className="flex items-center space-x-2">
          <span>{title}</span>
          {activeFilterCount > 0 && (
            <Badge variant="secondary" className="ml-2">
              {activeFilterCount}
            </Badge>
          )}
        </div>
        <Filter className="h-4 w-4" />
      </Button>

      {isOpen && (
        <Card className="absolute top-full left-0 right-0 z-50 mt-2">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="text-sm">{title}</CardTitle>
                <CardDescription className="text-xs">{description}</CardDescription>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsOpen(false)}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {filters.map((filter) => (
              <div key={filter.id}>
                {renderFilter(filter)}
              </div>
            ))}
            <Separator />
            <div className="flex justify-between">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  onClear();
                  setIsOpen(false);
                }}
              >
                <RotateCcw className="mr-2 h-3 w-3" />
                Clear All
              </Button>
              <Button
                size="sm"
                onClick={() => {
                  onApply();
                  setIsOpen(false);
                }}
              >
                Apply Filters
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

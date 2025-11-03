'use client';

import * as React from 'react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { X, Filter, Tag } from 'lucide-react';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';

interface TagFilterProps {
  availableTags: Record<string, string[]>;
  selectedTags: Record<string, string[]>;
  onTagsChange: (tags: Record<string, string[]>) => void;
}

function TagFilterComponent({ availableTags, selectedTags, onTagsChange }: TagFilterProps) {
  const [isOpen, setIsOpen] = useState(false);

  const handleTagToggle = (key: string, value: string) => {
    const currentValues = selectedTags[key] || [];
    const newValues = currentValues.includes(value)
      ? currentValues.filter(v => v !== value)
      : [...currentValues, value];
    
    const updatedTags = { ...selectedTags };
    if (newValues.length > 0) {
      updatedTags[key] = newValues;
    } else {
      delete updatedTags[key];
    }
    onTagsChange(updatedTags);
  };

  const handleClearFilter = (key?: string) => {
    if (key) {
      const newTags = { ...selectedTags };
      delete newTags[key];
      onTagsChange(newTags);
    } else {
      onTagsChange({});
    }
  };

  const activeFilterCount = Object.values(selectedTags).reduce((sum, values) => sum + (values?.length || 0), 0);

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" className="flex items-center space-x-2">
          <Filter className="h-4 w-4" />
          <span>Tag Filter</span>
          {activeFilterCount > 0 && (
            <Badge variant="secondary" className="ml-2">
              {activeFilterCount}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="start">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <Label className="text-base font-semibold">Filter by Tags</Label>
            {activeFilterCount > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleClearFilter()}
                className="h-8 text-xs"
              >
                Clear All
              </Button>
            )}
          </div>

          {Object.keys(availableTags).length === 0 ? (
            <p className="text-sm text-gray-500">No tags available</p>
          ) : (
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {Object.entries(availableTags).map(([key, values]) => {
                const selectedValues = selectedTags[key] || [];
                
                return (
                  <div key={key} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-sm font-medium flex items-center">
                        <Tag className="h-3 w-3 mr-1" />
                        {key}
                      </Label>
                      {selectedValues.length > 0 && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleClearFilter(key)}
                          className="h-6 text-xs"
                        >
                          <X className="h-3 w-3" />
                        </Button>
                      )}
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {values.map((value) => {
                        const isSelected = selectedValues.includes(value);
                        return (
                          <Badge
                            key={`${key}-${value}`}
                            variant={isSelected ? 'default' : 'outline'}
                            className="cursor-pointer hover:bg-primary/10"
                            onClick={() => handleTagToggle(key, value)}
                          >
                            {value}
                            {isSelected && <X className="ml-1 h-3 w-3" />}
                          </Badge>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {activeFilterCount > 0 && (
            <div className="pt-2 border-t">
              <div className="flex flex-wrap gap-2">
                {Object.entries(selectedTags).map(([key, values]) => {
                  if (!values || values.length === 0) return null;
                  return values.map((value) => (
                    <Badge key={`${key}-${value}`} variant="secondary" className="flex items-center space-x-1">
                      <span>{key}: {value}</span>
                      <X
                        className="h-3 w-3 cursor-pointer"
                        onClick={() => handleTagToggle(key, value)}
                      />
                    </Badge>
                  ));
                })}
              </div>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}

export const TagFilter = React.memo(TagFilterComponent);


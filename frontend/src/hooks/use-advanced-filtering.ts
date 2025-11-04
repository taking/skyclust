import { useState, useEffect, useCallback } from 'react';

export interface FilterPreset {
  id: string;
  name: string;
  filters: Record<string, unknown>;
  createdAt: string;
}

export interface SortConfig {
  field: string;
  direction: 'asc' | 'desc';
}

export interface UseAdvancedFilteringOptions<T> {
  storageKey: string;
  defaultFilters?: Record<string, unknown>;
  defaultSort?: SortConfig[];
  onFiltersChange?: (filters: Record<string, unknown>) => void;
  onSortChange?: (sort: SortConfig[]) => void;
}

/**
 * Hook for managing advanced filtering with presets and multi-sort
 */
export function useAdvancedFiltering<T>(
  options: UseAdvancedFilteringOptions<T>
) {
  const {
    storageKey,
    defaultFilters = {},
    defaultSort = [],
    onFiltersChange,
    onSortChange,
  } = options;

  const [filters, setFilters] = useState<Record<string, unknown>>(defaultFilters);
  const [sortConfig, setSortConfig] = useState<SortConfig[]>(defaultSort);
  const [presets, setPresets] = useState<FilterPreset[]>([]);

  // Load presets from localStorage
  useEffect(() => {
    try {
      const savedPresets = localStorage.getItem(`${storageKey}-presets`);
      if (savedPresets) {
        setPresets(JSON.parse(savedPresets));
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to load filter presets:', error);
      }
    }
  }, [storageKey]);

  // Load saved filters and sort
  useEffect(() => {
    try {
      const savedFilters = localStorage.getItem(`${storageKey}-filters`);
      const savedSort = localStorage.getItem(`${storageKey}-sort`);
      
      if (savedFilters) {
        const parsed = JSON.parse(savedFilters);
        setFilters(parsed);
        onFiltersChange?.(parsed);
      }
      
      if (savedSort) {
        const parsed = JSON.parse(savedSort);
        setSortConfig(parsed);
        onSortChange?.(parsed);
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to load saved filters/sort:', error);
      }
    }
  }, [storageKey, onFiltersChange, onSortChange]);

  // Save filters to localStorage
  useEffect(() => {
    try {
      localStorage.setItem(`${storageKey}-filters`, JSON.stringify(filters));
      onFiltersChange?.(filters);
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to save filters:', error);
      }
    }
  }, [filters, storageKey, onFiltersChange]);

  // Save sort to localStorage
  useEffect(() => {
    try {
      localStorage.setItem(`${storageKey}-sort`, JSON.stringify(sortConfig));
      onSortChange?.(sortConfig);
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to save sort:', error);
      }
    }
  }, [sortConfig, storageKey, onSortChange]);

  const updateFilter = useCallback((key: string, value: unknown) => {
    setFilters((prev) => {
      const newFilters = { ...prev };
      if (value === null || value === undefined || value === '') {
        delete newFilters[key];
      } else {
        newFilters[key] = value;
      }
      return newFilters;
    });
  }, []);

  const setAllFilters = useCallback((newFilters: Record<string, unknown>) => {
    setFilters(newFilters);
  }, []);

  const clearFilters = useCallback(() => {
    setFilters(defaultFilters);
  }, [defaultFilters]);

  const savePreset = useCallback((name: string) => {
    const preset: FilterPreset = {
      id: `${Date.now()}`,
      name,
      filters: { ...filters },
      createdAt: new Date().toISOString(),
    };

    const newPresets = [...presets, preset];
    setPresets(newPresets);
    
    try {
      localStorage.setItem(`${storageKey}-presets`, JSON.stringify(newPresets));
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to save preset:', error);
      }
    }
  }, [filters, presets, storageKey]);

  const loadPreset = useCallback((presetId: string) => {
    const preset = presets.find(p => p.id === presetId);
    if (preset) {
      setFilters(preset.filters);
    }
  }, [presets]);

  const deletePreset = useCallback((presetId: string) => {
    const newPresets = presets.filter(p => p.id !== presetId);
    setPresets(newPresets);
    
    try {
      localStorage.setItem(`${storageKey}-presets`, JSON.stringify(newPresets));
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to delete preset:', error);
      }
    }
  }, [presets, storageKey]);

  const toggleSort = useCallback((field: string) => {
    setSortConfig((prev) => {
      const existingIndex = prev.findIndex(s => s.field === field);
      
      if (existingIndex >= 0) {
        const existing = prev[existingIndex];
        // Toggle direction
        if (existing.direction === 'asc') {
          // Remove if toggling desc -> none
          return prev.filter((_, i) => i !== existingIndex);
        } else {
          // Change to asc
          return prev.map((s, i) => 
            i === existingIndex ? { ...s, direction: 'asc' as const } : s
          );
        }
      } else {
        // Add new sort (asc first)
        return [...prev, { field, direction: 'asc' as const }];
      }
    });
  }, []);

  const clearSort = useCallback(() => {
    setSortConfig(defaultSort);
  }, [defaultSort]);

  return {
    filters,
    sortConfig,
    presets,
    updateFilter,
    setAllFilters,
    clearFilters,
    savePreset,
    loadPreset,
    deletePreset,
    toggleSort,
    clearSort,
  };
}


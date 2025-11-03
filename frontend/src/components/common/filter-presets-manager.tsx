
'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Save, Trash2, Filter, X } from 'lucide-react';
import { FilterPreset } from '@/hooks/useAdvancedFiltering';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface FilterPresetsManagerProps {
  presets: FilterPreset[];
  currentFilters: Record<string, unknown>;
  onSavePreset: (name: string) => void;
  onLoadPreset: (presetId: string) => void;
  onDeletePreset: (presetId: string) => void;
}

export function FilterPresetsManager({
  presets,
  currentFilters,
  onSavePreset,
  onLoadPreset,
  onDeletePreset,
}: FilterPresetsManagerProps) {
  const [isDialogOpen, setIsDialogOpen] = React.useState(false);
  const [presetName, setPresetName] = React.useState('');

  const handleSave = () => {
    if (presetName.trim()) {
      onSavePreset(presetName.trim());
      setPresetName('');
      setIsDialogOpen(false);
    }
  };

  const hasActiveFilters = Object.keys(currentFilters).length > 0;

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm">
            <Filter className="mr-2 h-4 w-4" />
            Presets
            {presets.length > 0 && (
              <Badge variant="secondary" className="ml-2">
                {presets.length}
              </Badge>
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-56">
          {presets.length > 0 ? (
            <>
              {presets.map((preset) => (
                <DropdownMenuItem
                  key={preset.id}
                  onClick={() => onLoadPreset(preset.id)}
                  className="flex items-center justify-between"
                >
                  <span className="truncate flex-1">{preset.name}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-5 w-5 p-0 ml-2"
                    onClick={(e) => {
                      e.stopPropagation();
                      onDeletePreset(preset.id);
                    }}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />
            </>
          ) : (
            <DropdownMenuItem disabled>
              No presets saved
            </DropdownMenuItem>
          )}
          <DropdownMenuItem
            onClick={() => setIsDialogOpen(true)}
            disabled={!hasActiveFilters}
          >
            <Save className="mr-2 h-4 w-4" />
            Save current filters
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Save Filter Preset</DialogTitle>
            <DialogDescription>
              Save your current filter settings for quick access later
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="preset-name">Preset Name</Label>
              <Input
                id="preset-name"
                value={presetName}
                onChange={(e) => setPresetName(e.target.value)}
                placeholder="e.g., Production VMs"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && presetName.trim()) {
                    handleSave();
                  }
                }}
              />
            </div>
            <div className="text-sm text-gray-500">
              Active filters: {Object.keys(currentFilters).length}
            </div>
          </div>
          <div className="flex justify-end space-x-2">
            <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={!presetName.trim()}>
              Save Preset
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}


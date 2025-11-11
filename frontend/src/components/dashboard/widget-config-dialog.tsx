'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { WidgetData, WidgetSize, WIDGET_CONFIGS, getWidgetSizeTranslationKey } from '@/lib/widgets';
import { useTranslation } from '@/hooks/use-translation';

interface WidgetConfigDialogProps {
  widget: WidgetData | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (widgetId: string, updates: { size?: WidgetSize; title?: string; config?: Record<string, unknown> }) => void;
}

export function WidgetConfigDialog({
  widget,
  open,
  onOpenChange,
  onSave,
}: WidgetConfigDialogProps) {
  const { t } = useTranslation();
  const [size, setSize] = React.useState<WidgetSize>(widget?.size || 'medium');
  const [title, setTitle] = React.useState<string>(widget?.title || '');
  const [refreshInterval, setRefreshInterval] = React.useState<string>('');

  React.useEffect(() => {
    if (widget) {
      setSize(widget.size);
      setTitle(widget.title);
      setRefreshInterval(String(widget.config?.refreshInterval || '30'));
    }
  }, [widget]);

  if (!widget) return null;

  const config = WIDGET_CONFIGS[widget.type];
  const sizeOptions: WidgetSize[] = ['small', 'medium', 'large', 'xlarge'];
  const availableSizes = sizeOptions.filter(s => {
    const index = sizeOptions.indexOf(s);
    const minIndex = sizeOptions.indexOf(config.minSize);
    const maxIndex = sizeOptions.indexOf(config.maxSize);
    return index >= minIndex && index <= maxIndex;
  });

  const handleSave = () => {
    onSave(widget.id, {
      size,
      title: title || config.title,
      config: {
        ...widget.config,
        refreshInterval: refreshInterval ? Number(refreshInterval) : undefined,
      },
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('widgetConfig.title')}</DialogTitle>
          <DialogDescription>
            {t('widgetConfig.description')}
          </DialogDescription>
        </DialogHeader>
        
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="widget-title">{t('widgetConfig.widgetTitle')}</Label>
            <Input
              id="widget-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={config.title}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="widget-size">{t('widgetConfig.widgetSize')}</Label>
            <Select value={size} onValueChange={(value) => setSize(value as WidgetSize)}>
              <SelectTrigger id="widget-size">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {availableSizes.map((s) => (
                  <SelectItem key={s} value={s}>
                    {t(getWidgetSizeTranslationKey(s))}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-gray-500">
              {t('widgetConfig.minMax', { min: config.minSize, max: config.maxSize })}
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="refresh-interval">{t('widgetConfig.refreshInterval')}</Label>
            <Input
              id="refresh-interval"
              type="number"
              min="10"
              max="3600"
              step="10"
              value={refreshInterval}
              onChange={(e) => setRefreshInterval(e.target.value)}
              placeholder="30"
            />
            <p className="text-xs text-gray-500">
              {t('widgetConfig.refreshIntervalHint')}
            </p>
          </div>
        </div>

        <div className="flex justify-end space-x-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('widgetConfig.cancel')}
          </Button>
          <Button onClick={handleSave}>
            {t('widgetConfig.saveChanges')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}


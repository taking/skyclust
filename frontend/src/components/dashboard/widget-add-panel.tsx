'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Plus, Search, Server, DollarSign, Activity, Clock, Zap, AlertTriangle, TrendingUp, MapPin } from 'lucide-react';
import { WidgetType, WIDGET_CONFIGS, useWidgetConfigs, getWidgetCategoryTranslationKey, getWidgetSizeTranslationKey } from '@/lib/widgets';
import { useTranslation } from '@/hooks/use-translation';

interface WidgetAddPanelProps {
  onAddWidget: (type: WidgetType) => void;
  existingWidgets: string[];
}

const iconMap = {
  Server,
  DollarSign,
  Activity,
  Clock,
  Zap,
  AlertTriangle,
  TrendingUp,
  MapPin,
};

export function WidgetAddPanel({ onAddWidget, existingWidgets }: WidgetAddPanelProps) {
  const { t } = useTranslation();
  const { getWidgetConfig } = useWidgetConfigs();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  const categories = [
    { id: 'all', label: t('widgets.addWidget.categories.all'), count: Object.keys(WIDGET_CONFIGS).length },
    { id: 'overview', label: t('widgets.addWidget.categories.overview'), count: Object.values(WIDGET_CONFIGS).filter(w => w.category === 'overview').length },
    { id: 'monitoring', label: t('widgets.addWidget.categories.monitoring'), count: Object.values(WIDGET_CONFIGS).filter(w => w.category === 'monitoring').length },
    { id: 'cost', label: t('widgets.addWidget.categories.cost'), count: Object.values(WIDGET_CONFIGS).filter(w => w.category === 'cost').length },
    { id: 'management', label: t('widgets.addWidget.categories.management'), count: Object.values(WIDGET_CONFIGS).filter(w => w.category === 'management').length },
  ];

  const filteredWidgets = Object.entries(WIDGET_CONFIGS)
    .map(([type, config]) => {
      const translatedConfig = getWidgetConfig(type as WidgetType);
      return [type, translatedConfig] as [string, typeof translatedConfig];
    })
    .filter(([type, config]) => {
      const matchesSearch = config.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                           config.description.toLowerCase().includes(searchQuery.toLowerCase());
      const matchesCategory = selectedCategory === 'all' || config.category === selectedCategory;
      const notAlreadyAdded = !existingWidgets.includes(type);
      
      return matchesSearch && matchesCategory && notAlreadyAdded;
    });

  const handleAddWidget = (type: WidgetType) => {
    onAddWidget(type);
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          {t('widgets.addWidget.addButton')}
        </Button>
      </DialogTrigger>
      <DialogContent className="w-[95vw] max-w-4xl max-h-[85vh] overflow-hidden sm:w-full">
        <DialogHeader>
          <DialogTitle>{t('widgets.addWidget.title')}</DialogTitle>
          <DialogDescription>
            {t('widgets.addWidget.description')}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Search and Filter */}
          <div className="flex flex-col sm:flex-row space-y-2 sm:space-y-0 sm:space-x-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={t('widgets.addWidget.searchPlaceholder')}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
          </div>

          {/* Category Tabs */}
          <Tabs value={selectedCategory} onValueChange={setSelectedCategory}>
            <TabsList className="grid w-full grid-cols-3 sm:grid-cols-5 gap-1">
              {categories.map((category) => (
                <TabsTrigger key={category.id} value={category.id}>
                  {category.label} ({category.count})
                </TabsTrigger>
              ))}
            </TabsList>

            <TabsContent value={selectedCategory} className="mt-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 max-h-[50vh] sm:max-h-96 overflow-y-auto">
                {filteredWidgets.map(([type, config]) => {
                  const IconComponent = iconMap[config.icon as keyof typeof iconMap] || Server;
                  
                  return (
                    <Card key={type} className="cursor-pointer hover:shadow-md transition-shadow">
                      <CardHeader className="pb-3">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center">
                            <IconComponent className="mr-2 h-5 w-5 text-blue-500" />
                            <CardTitle className="text-sm">{config.title}</CardTitle>
                          </div>
                          <Badge variant="outline" className="text-xs">
                            {t(getWidgetSizeTranslationKey(config.defaultSize))}
                          </Badge>
                        </div>
                        <CardDescription className="text-xs">
                          {config.description}
                        </CardDescription>
                      </CardHeader>
                      <CardContent className="pt-0">
                        <div className="space-y-2">
                          <div className="flex items-center justify-between text-xs text-gray-500">
                            <span>{t('widgets.addWidget.size')}: {t(getWidgetSizeTranslationKey(config.defaultSize))}</span>
                            <span>{t('widgets.addWidget.category')}: {t(getWidgetCategoryTranslationKey(config.category))}</span>
                          </div>
                          <Button
                            size="sm"
                            className="w-full"
                            onClick={() => handleAddWidget(type as WidgetType)}
                          >
                            <Plus className="mr-1 h-3 w-3" />
                            {t('widgets.addWidget.addButton')}
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  );
                })}
              </div>

              {filteredWidgets.length === 0 && (
                <div className="text-center py-8">
                  <div className="mx-auto h-12 w-12 text-gray-400">
                    <Search className="h-12 w-12" />
                  </div>
                  <h3 className="mt-2 text-sm font-medium text-gray-900">{t('widgets.addWidget.noWidgetsFound')}</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    {searchQuery 
                      ? t('widgets.addWidget.tryAdjusting')
                      : t('widgets.addWidget.allAdded')
                    }
                  </p>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  );
}
